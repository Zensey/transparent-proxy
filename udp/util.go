package udp

import (
	"net"
	"sync"
	"time"
)

type redirConn struct {
	net.Conn
	buf  []byte
	ttl  time.Duration
	once sync.Once
}

func (c *redirConn) Read(b []byte) (n int, err error) {
	if c.ttl > 0 {
		c.SetReadDeadline(time.Now().Add(c.ttl))
		defer c.SetReadDeadline(time.Time{})
	}

	c.once.Do(func() {
		n = copy(b, c.buf)
		c.buf = make([]byte, defaultReadBufferSize)
	})

	if n == 0 {
		n, err = c.Conn.Read(b)
	}

	return
}

func (c *redirConn) Write(b []byte) (n int, err error) {
	if c.ttl > 0 {
		c.SetWriteDeadline(time.Now().Add(c.ttl))
		defer c.SetWriteDeadline(time.Time{})
	}
	return c.Conn.Write(b)
}

///////////////////////////////////////////////////////////////////
type redirConnDeadline struct {
	net.Conn
	ttl time.Duration
}

func (c *redirConnDeadline) Read(b []byte) (n int, err error) {
	if c.ttl > 0 {
		c.SetReadDeadline(time.Now().Add(c.ttl))
		defer c.SetReadDeadline(time.Time{})
	}
	n, err = c.Conn.Read(b)
	return
}

func (c *redirConnDeadline) Write(b []byte) (n int, err error) {
	if c.ttl > 0 {
		c.SetWriteDeadline(time.Now().Add(c.ttl))
		defer c.SetWriteDeadline(time.Time{})
	}
	return c.Conn.Write(b)
}
