package tftp

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type Server struct {
	*sync.Mutex
	addr         *net.UDPAddr
	readHandler  func(filename string, s io.ReaderFrom) error
	writeHandler func(filename string, s io.WriterTo) error
	timeout      time.Duration
	cancel       context.Context
	cancelFunc   context.CancelFunc
}

func NewServer(readHandler func(filename string, s io.ReaderFrom) error, writeHandler func(filename string, s io.WriterTo) error) *Server {
	s := &Server{
		readHandler:  readHandler,
		writeHandler: writeHandler,
		timeout:      defaultTimeout,
	}
	s.cancel, s.cancelFunc = context.WithCancel(context.Background())
	return s
}

func (s *Server) ListenAndServe(addr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to resolve udp address %s: %v", addr, err)
	}
	s.addr = udpAddr
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	return s.Serve(conn)
}

func (s *Server) Serve(conn net.PacketConn) error {
	buffer := make([]byte, defaultDataSize)
	for {
		select {
		case <-s.cancel.Done():
			return nil
		default:
			n, addr, err := conn.ReadFrom(buffer)
			if err != nil {
				log.Println(err)
			}
			packet := buffer[:n]
			err = s.handlePacket(packet, addr.(*net.UDPAddr))
			if err != nil {
				log.Println(fmt.Errorf("error handling packet: %v", err))
			}
		}
	}
}

func (s *Server) handlePacket(rawPacket []byte, addr *net.UDPAddr) error {
	s.Lock()
	defer s.Unlock()

	opcode := binary.BigEndian.Uint16(rawPacket[0:])
	switch opcode {
	case opRRQ:
		filename, mode, err := unpackRRQ(rawPacket)
		if err != nil {
			return err
		}
		if mode != "octet" {
			return fmt.Errorf("only octet mode supported: %s", mode)
		}
		if !checkFileExists(filename) {
			return fmt.Errorf("file not found: %s", filename)
		}
		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			return err
		}
		sn := &sender{
			localAddr:  s.addr,
			remoteAddr: addr,
			conn:       conn,
			filename:   filename,
			send:       make([]byte, defaultDataSize),
			receive:    make([]byte, defaultDataSize),
			retry:      newBackoff(),
			timeout:    s.timeout,
		}
		go func() {
			if s.readHandler != nil {
				err := s.readHandler(filename, sn)
				if err != nil {
					sn.stop(err)
				}
			} else {
				sn.stop(fmt.Errorf("server does not support reads"))
			}
		}()

	case opWRQ:
		filename, mode, err := unpackWRQ(rawPacket)
		if err != nil {
			return err
		}
		if mode != "octet" {
			return fmt.Errorf("only octet mode supported: %v", mode)
		}
		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			return err
		}
		rc := &receiver{
			localAddr:  s.addr,
			remoteAddr: addr,
			conn:       conn,
			filename:   filename,
			send:       make([]byte, defaultDataSize),
			receive:    make([]byte, defaultDataSize),
			retry:      newBackoff(),
			timeout:    s.timeout,
		}
		go func() {
			if s.writeHandler != nil {
				err := s.writeHandler(filename, rc)
				if err != nil {
					rc.stop(err)
				}
			} else {
				rc.stop(fmt.Errorf("server does not support writes"))
			}
		}()
	default:
		return fmt.Errorf("unuxpected packet with opcode: %d", opcode)
	}
	return nil
}

func (s *Server) Shutdown() {
	s.cancelFunc()
}
