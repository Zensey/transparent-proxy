package main

import (
	"io"
	"log"
	"net"
	"sync"
	"syscall"

	"github.com/LiamHaworth/go-tproxy"
	dissector "github.com/go-gost/tls-dissector"
	dissect "github.com/zensey/transparent-proxy/dissector"
	"github.com/zensey/transparent-proxy/util"
)

func main() {

	log.Println("Binding TCP TProxy listener to 0.0.0.0:8585")
	listener, err := tproxy.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: 8585})
	if err != nil {
		log.Fatalf("Encountered error while binding listener: %s", err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept Error:", err)
			continue
		}

		go handleTcpConn(conn)
	}
}

func handleTcpConn(conn net.Conn) {
	log.Printf("%s -> %s", conn.RemoteAddr().String(), conn.LocalAddr().String())
	defer conn.Close()

	// remoteConn, err := conn.(*tproxy.Conn).DialOriginalDestination(false)
	// if err != nil {
	// 	log.Printf("Failed to connect to original destination [%s]: %s", conn.LocalAddr().String(), err)
	// 	return
	// }

	dialer := &net.Dialer{
		Control: func(network, address string, conn syscall.RawConn) error {
			var operr error
			if err := conn.Control(func(fd uintptr) {
				// set so_mark=100 to prevent loop
				operr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_MARK, 100)
			}); err != nil {
				return err
			}
			return operr
		},
	}
	remoteConn, err := dialer.Dial("tcp", conn.LocalAddr().String())
	if err != nil {
		log.Printf("Failed to connect to original destination [%s]: %s", conn.LocalAddr().String(), err)
		return
	}

	defer remoteConn.Close()

	var streamWait sync.WaitGroup
	streamWait.Add(2)

	streamConn := func(dst io.WriteCloser, src io.Reader, dir int) {
		// defer src.Close()
		defer dst.Close()

		sn := sniffer{}
		if dir == 1 {
			src = sn.capture(src)
		}
		io.Copy(dst, src)
		sn.close()

		streamWait.Done()
	}

	go streamConn(remoteConn, conn, 1)
	go streamConn(conn, remoteConn, 0)

	streamWait.Wait()
	log.Printf("Finish: %s -> %s", conn.RemoteAddr().String(), conn.LocalAddr().String())
}

type sniffer struct {
	pw *io.PipeWriter
}

func (p *sniffer) close() {
	if p.pw != nil {
		p.pw.Close()
	}
}

func (p *sniffer) capture(src io.Reader) io.Reader {
	pr, pw := io.Pipe()
	p.pw = pw
	srcCopy := io.TeeReader(src, pw)

	go func() {
		// must read the stream until the end in any case
		// due to synchronous nature of TeeReader
		defer util.ReadUntilEof(pr)

		rec := dissect.Record{}
		_, err := rec.ReadFrom(pr)
		if err != nil {
			log.Println("ReadFrom err:", err)
			return
		}
		if !rec.Valid() {
			// TLS record is not found
			return
		}

		clientHello := dissector.ClientHelloHandshake{}
		_, err = clientHello.ReadFrom(pr)
		if err != nil {
			log.Println("ReadFrom err:", err)
			return
		}

		log.Println("clientHello >")
		for _, ext := range clientHello.Extensions {
			if ext.Type() == dissector.ExtServerName {
				snExtension := ext.(*dissector.ServerNameExtension)
				log.Println("sn>", snExtension)
				break
			}
		}
	}()

	return srcCopy
}
