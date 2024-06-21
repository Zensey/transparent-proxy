package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"runtime/debug"

	"github.com/LiamHaworth/go-tproxy"
	"github.com/zensey/transparent-proxy/stats"
	"github.com/zensey/transparent-proxy/tcp"
	"github.com/zensey/transparent-proxy/udp"
)

type handler struct {
	T *stats.TrafficCounter
}

func (h *handler) handlerByIP(w http.ResponseWriter, req *http.Request) {
	json.NewEncoder(w).Encode(h.T.GetStatsByIP())
}
func (h *handler) handlerBySN(w http.ResponseWriter, req *http.Request) {
	json.NewEncoder(w).Encode(h.T.GetStatsBySN())
}

func main() {
	debug.SetTraceback("crash")
	T := stats.NewtrafficCounter()
	h := &handler{T: T}

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

	http.HandleFunc("/stat/ip", h.handlerByIP)
	http.HandleFunc("/stat/sn", h.handlerBySN)
	http.ListenAndServe(":8080", nil)

	select {}
}
