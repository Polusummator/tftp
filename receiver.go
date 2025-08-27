package tftp

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

type receiver struct {
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
	l            int
}

func (r *receiver) WriteTo(w io.Writer) (n int64, err error) {
	binary.BigEndian.PutUint16(r.send[0:2], opACK)
	for {
		if r.l > 0 {
			l, err := w.Write(r.receive[4:r.l])
			n += int64(l)
			if err != nil {
				r.stop(err)
				return n, err
			}
			if r.l < len(r.receive) {
				return n, nil
			}
		}
		binary.BigEndian.PutUint16(r.send[2:4], r.block)
		r.block++
		lr, err := r.receiveRetry(4)
		if err != nil {
			r.stop(err)
			return n, err
		}
		r.l = lr
	}
}

func (r *receiver) receiveData(n int) (int, error) {
	err := r.conn.SetReadDeadline(time.Now().Add(r.timeout))
	if err != nil {
		return 0, err
	}
	_, err = r.conn.WriteToUDP(r.send[:r.l], r.remoteAddr)
	if err != nil {
		return 0, err
	}
	r.packetsSent++
	for {
		nr, _, err := r.conn.ReadFromUDP(r.receive)
		if err != nil {
			return 0, err
		}
		opcode := binary.BigEndian.Uint16(r.receive[:nr])
		switch opcode {
		case opDATA:
			block, err := unpackACK(r.receive[:nr])
			if err != nil {
				return 0, err
			}
			if block == r.block {
				r.packetsAcked++
				return nr, nil
			}
		case opACK:
			_, err := unpackACK(r.receive[:nr])
			if r.block != 1 {
				continue
			}
			if err != nil {
				r.stop(err)
				return 0, err
			}
			r.block = 0
			return 0, nil
		case opERROR:
			_, msg, err := unpackERROR(r.receive[:nr])
			if err != nil {
				return 0, err
			}
			return 0, fmt.Errorf(msg)
		}
	}
}

func (r *receiver) receiveRetry(n int) (int, error) {
	r.retry.reset()
	for {
		nr, err := r.receiveData(n)
		if _, ok := err.(net.Error); ok && r.retry.attempt <= r.retry.maxAttempts {
			r.retry.backoff()
			continue
		}
		return nr, err
	}
}

func (r *receiver) stop(err error) error {
	if r.conn == nil {
		return nil
	}
	defer func() {
		r.conn.Close()
		r.conn = nil
	}()
	r.send = packERROR(1, err.Error())
	_, err = r.conn.WriteToUDP(r.send, r.remoteAddr)
	return err
}
