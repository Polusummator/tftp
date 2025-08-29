package tftp

import (
	"fmt"
	"io"
	"net"
	"time"
)

type Client struct {
	addr    *net.UDPAddr
	timeout time.Duration
}

func (c *Client) Send(filename string) (io.ReaderFrom, error) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{})
	if err != nil {
		return nil, err
	}
	sn := &sender{
		localAddr: c.addr,
		conn:      conn,
		send:      make([]byte, defaultDataSize),
		receive:   make([]byte, defaultDataSize),
		retry:     newBackoff(),
		timeout:   c.timeout,
	}
	n := packWRQ(sn.send, filename, "octet")
	err = sn.sendRetry(n)
	if err != nil {
		return nil, err
	}
	return sn, nil
}

func (c *Client) Receive(filename string) (io.WriterTo, error) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{})
	if err != nil {
		return nil, err
	}
	rc := &receiver{
		localAddr: c.addr,
		conn:      conn,
		send:      make([]byte, defaultDataSize),
		receive:   make([]byte, defaultDataSize),
		retry:     newBackoff(),
		timeout:   c.timeout,
		block:     1,
	}
	n := packRRQ(rc.send, filename, "octet")
	l, err := rc.receiveRetry(n)
	if err != nil {
		return nil, err
	}
	rc.l = l
	return rc, nil
}

func NewClient(addr string) (*Client, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve udp address %s: %v", addr, err)
	}
	c := &Client{
		addr: udpAddr,
	}
	return c, nil
}
