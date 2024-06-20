package udp

import (
	"bufio"
	"io"
	"net"

	"github.com/go-gost/core/common/bufpool"
)

const (
	bufferSize = 4 * 1024
)

func Transport(rw1, rw2 io.ReadWriter, nw1, nw2 *int64) error {

	errc := make(chan error, 1)
	go func() {
		n, err := CopyBuffer(rw1, rw2, bufferSize)
		*nw1 = n
		errc <- err
	}()

	go func() {
		n, err := CopyBuffer(rw2, rw1, bufferSize)
		*nw2 = n
		errc <- err
	}()

	if err := <-errc; err != nil && err != io.EOF {
		return err
	}

	return nil
}

func CopyBuffer(dst io.Writer, src io.Reader, bufSize int) (int64, error) {
	buf := bufpool.Get(bufSize)
	defer bufpool.Put(buf)

	n, err := io.CopyBuffer(dst, src, buf)
	return n, err
}

type bufferReaderConn struct {
	net.Conn
	br *bufio.Reader
}

func NewBufferReaderConn(conn net.Conn, br *bufio.Reader) net.Conn {
	return &bufferReaderConn{
		Conn: conn,
		br:   br,
	}
}

func (c *bufferReaderConn) Read(b []byte) (int, error) {
	return c.br.Read(b)
}
