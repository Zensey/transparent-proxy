package tcp

import (
	"io"
	"log"
	"net"
	"sync"
	"syscall"

	"github.com/zensey/transparent-proxy/metrics"
	"github.com/zensey/transparent-proxy/stats"
)

func HandleAccept(listener net.Listener, t *stats.TrafficCounter) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept Error:", err)
			continue
		}

		go handle(conn, t)
	}
}

func handle(conn net.Conn, t *stats.TrafficCounter) {
	log.Printf("[tcp] Handle: %s -> %s", conn.RemoteAddr().String(), conn.LocalAddr().String())
	defer conn.Close()

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

	dst := conn.LocalAddr().String()
	ip, _, _ := net.SplitHostPort(dst)

	var (
		serverName string
		rx, tx     int64
	)
	streamConn := func(dst io.WriteCloser, src io.Reader, dir int, written *int64) {
		// defer src.Close()
		defer dst.Close()

		sniffer := sniffer{}
		if dir == 1 {
			src = sniffer.capture(src, &serverName)
		}
		n, _ := io.Copy(dst, src)
		*written = n

		sniffer.close()
		streamWait.Done()
	}

	go streamConn(remoteConn, conn, 1, &tx)
	go streamConn(conn, remoteConn, 0, &rx)
	streamWait.Wait()

	log.Printf("[tcp] Finish: %s -> %s", conn.RemoteAddr().String(), conn.LocalAddr().String())
	t.CollectStats(ip, serverName, rx, tx)

	metrics.ReportMetrics(ip, serverName, t)
}
