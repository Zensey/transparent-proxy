package udp

import (
	"bufio"
	"io"
	"net"
	"sync"

	"github.com/go-gost/core/common/bufpool"
)

const (
	bufferSize = 4 * 1024
)

func transport(rw1, rw2 io.ReadWriter, nw1, nw2 *int64) error {

	var streamWait sync.WaitGroup
	streamWait.Add(2)

	go func() {
		defer streamWait.Done()

		n, _ := CopyBuffer(rw1, rw2, bufferSize)
		*nw1 = n
	}()

	go func() {
		defer streamWait.Done()

		n, _ := CopyBuffer(rw2, rw1, bufferSize)
		*nw2 = n
	}()

	streamWait.Wait()
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
