package tftp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

type sender struct {
	localAddr    *net.UDPAddr
	remoteAddr   *net.UDPAddr
	conn         *net.UDPConn
	filename     string
	block        uint16
	send         []byte
	receive      []byte
	packetsSent  int
	packetsAcked int
	retry        *backoff
	timeout      time.Duration
}

func (s *sender) ReadFrom(r io.Reader) (n int64, err error) {
	defer func() {
		if s.conn != nil {
			s.conn.Close()
			s.conn = nil
		}
	}()
	s.block = 1
	binary.BigEndian.PutUint16(s.send[0:2], opDATA)
	for {
		l, err := io.ReadFull(r, s.send[4:])
		n += int64(l)
		if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
			if errors.Is(err, io.EOF) {
				binary.BigEndian.PutUint16(s.send[2:4], s.block)
				err := s.sendRetry(4)
				if err != nil {
					s.stop(err)
					return n, err
				}
				return n, nil
			}
			s.stop(err)
			return n, err
		}
		binary.BigEndian.PutUint16(s.send[2:4], s.block)
		err = s.sendRetry(4 + l)
		if err != nil {
			s.stop(err)
			return n, err
		}
		if l < defaultBlockSize {
			return n, nil
		}
		s.block++
	}
}

func (s *sender) sendData(n int) error {
	err := s.conn.SetReadDeadline(time.Now().Add(s.timeout))
	if err != nil {
		return err
	}
	_, err = s.conn.WriteToUDP(s.send[:n], s.remoteAddr)
	if err != nil {
		return err
	}
	s.packetsSent += 1
	for {
		nr, _, err := s.conn.ReadFromUDP(s.receive)
		if err != nil {
			return err
		}
		opcode := binary.BigEndian.Uint16(s.receive[:nr])
		switch opcode {
		case opACK:
			block, err := unpackACK(s.receive[:nr])
			if err != nil {
				return err
			}
			if block == s.block {
				s.packetsAcked++
				return nil
			}
		case opERROR:
			_, msg, err := unpackERROR(s.receive[:nr])
			if err != nil {
				return err
			}
			return fmt.Errorf(msg)
		}
	}
}

func (s *sender) sendRetry(n int) error {
	s.retry.reset()
	for {
		err := s.sendData(n)
		if _, ok := err.(net.Error); ok && s.retry.attempt <= s.retry.maxAttempts {
			s.retry.backoff()
			continue
		}
		return err
	}
}

func (s *sender) stop(err error) error {
	if s.conn == nil {
		return nil
	}
	defer func() {
		s.conn.Close()
		s.conn = nil
	}()
	s.send = packERROR(1, err.Error())
	_, err = s.conn.WriteToUDP(s.send, s.remoteAddr)
	return err
}
