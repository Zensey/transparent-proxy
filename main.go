package main

import (
	"log"
	"net"
	"runtime/debug"

	"github.com/LiamHaworth/go-tproxy"
	"github.com/zensey/transparent-proxy/logic"
	"github.com/zensey/transparent-proxy/tcp"
	"github.com/zensey/transparent-proxy/udp"
)

func main() {
	debug.SetTraceback("crash")
	T := logic.NewtrafficCounter()

	log.Println("Binding TCP TProxy listener to 0.0.0.0:8585")
	listener, err := tproxy.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: 8585})
	if err != nil {
		log.Fatalf("Encountered error while binding listener: %s", err)
		return
	}
	defer listener.Close()
	go tcp.HandleAccept(listener, T)

	log.Println("Binding UDP TProxy listener to 0.0.0.0:8585")
	udpListener, err := tproxy.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: 8585})
	if err != nil {
		log.Fatalf("Encountered error while binding listener: %s", err)
		return
	}
	defer udpListener.Close()
	go udp.HandleAccept(udpListener, T)

	select {}
}
