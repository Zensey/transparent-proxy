package udp

import (
	"bytes"
	"log"
	"net"
	"syscall"
	"time"

	"github.com/LiamHaworth/go-tproxy"
	dissector "github.com/go-gost/tls-dissector"
	"github.com/quic-go/quic-go"

	"github.com/zensey/transparent-proxy/stats"
)

const (
	defaultTTL            = 10 * time.Second
	defaultReadBufferSize = 4096
)

func HandleAccept(udpListener *net.UDPConn, t *stats.TrafficCounter) {
	for {
		buff := make([]byte, defaultReadBufferSize)
		n, srcAddr, dstAddr, err := tproxy.ReadFromUDP(udpListener, buff)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				log.Printf("Temporary error while reading data: %s", netErr)
			}

			log.Fatalf("Unrecoverable error while reading data: %s", err)
			return
		}

		go handle(buff[:n], srcAddr, dstAddr, t)
	}
}

func extractSN(data []byte) string {
	pr := bytes.NewReader(data)

	clientHello := dissector.ClientHelloHandshake{}
	_, err := clientHello.ReadFrom(pr)
	if err != nil {
		log.Println("ReadFrom err:", err)
		return ""
	}

	for _, ext := range clientHello.Extensions {
		if ext.Type() == dissector.ExtServerName {
			snExtension := ext.(*dissector.ServerNameExtension)
			return snExtension.Name
		}
	}
	return ""
}

// handle establishes a bidirectional connection
// to the original destination pretending to be
// the client. 
// If no interaction within a few seconds, then connection is closed
func handle(data []byte, srcAddr, dstAddr *net.UDPAddr, t *stats.TrafficCounter) {
	log.Printf("[udp] Handle: %s -> %s", srcAddr, dstAddr)

	tx0 := int64(len(data))
	var sn string
	crypto := quic.ExractCryptoFrame(data)
	if crypto != nil {
		sn = extractSN(crypto)
	}

	// to local
	localConn, err := tproxy.DialUDP("udp", dstAddr, srcAddr)
	if err != nil {
		log.Printf("Failed to connect to original UDP source [%s]: %s", srcAddr.String(), err)
		return
	}
	defer localConn.Close()

	// call to remote
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
	remoteConn, err := dialer.Dial("udp", dstAddr.String())
	if err != nil {
		log.Printf("Failed to connect to original UDP destination [%s]: %s", dstAddr.String(), err)
		return
	}
	defer remoteConn.Close()
	remoteConnWrap := &redirConnDeadline{
		Conn: remoteConn,
		ttl:  defaultTTL,
	}

	localConnWrap := &redirConn{
		Conn: localConn,
		buf:  data,
		ttl:  defaultTTL,
	}

	var nwL, nwR int64
	Transport(localConnWrap, remoteConnWrap, &nwL, &nwR)
	log.Printf("[udp] Finish: %s -> %s", srcAddr, dstAddr)

	t.CollectStats(dstAddr.IP.String(), sn, nwL, nwR+tx0)
}
