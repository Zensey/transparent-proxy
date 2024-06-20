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
	"github.com/zensey/transparent-proxy/logic"
	// "github.com/zensey/transparent-proxy/quic"
)

const (
	defaultTTL            = 30 * time.Second
	defaultReadBufferSize = 4096
)

func HandleAccept(udpListener *net.UDPConn, T *logic.TrafficCounter) {
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

		// log.Printf("Accepting UDP connection from %s with destination of %s", srcAddr.String(), dstAddr.String())
		go handle(buff[:n], srcAddr, dstAddr, T)
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
	log.Println("clientHello >")
	for _, ext := range clientHello.Extensions {
		if ext.Type() == dissector.ExtServerName {
			snExtension := ext.(*dissector.ServerNameExtension)
			log.Println("SN >", snExtension.Name)
			return snExtension.Name
		}
	}
	return ""
}

// handle will open a connection
// to the original destination pretending
// to be the client. It will when right
// the received data to the remote host
// and wait a few seconds for any possible
// response data
func handle(data []byte, srcAddr, dstAddr *net.UDPAddr, T *logic.TrafficCounter) {

	// T.IncRx(dstAddr.IP.String(), int64(len(data)))
	tx0 := int64(len(data))

	var sn string
	crypto := quic.ExractCryptoFrame(data)
	if crypto != nil {
		sn = extractSN(crypto)
	}
	_ = sn

	log.Printf("Accepting UDP connection from %s with destination of %s", srcAddr, dstAddr)

	// to local
	localConn, err := tproxy.DialUDP("udp", dstAddr, srcAddr)
	if err != nil {
		log.Printf("Failed to connect to original UDP source [%s]: %s", srcAddr.String(), err)
		return
	}
	defer localConn.Close()

	// call to remote
	// remoteConn, err := tproxy.DialUDP("udp", srcAddr, dstAddr)
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

	localConnWrap := &redirConn{
		Conn: localConn,
		buf:  data,
		ttl:  time.Second * 2,
	}

	log.Println("Copy > src<->dst ")
	remoteConn.SetReadDeadline(time.Now().Add(2 * time.Second)) // Add deadline to ensure it doesn't block forever
	var nwL, nwR int64
	err = Transport(localConnWrap, remoteConn, &nwL, &nwR)
	if err != nil {
		log.Printf("Encountered error while Transport %s", err)
	}

	T.CollectStats(dstAddr.IP.String(), sn, nwL, nwR+tx0)
	
	T_ := T.GetTable()
	for k, v := range T_ {
		_, _ = k, v
		log.Printf("> %v %v", k, v)
	}
}
