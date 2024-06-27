#!/bin/bash

LOCAL_USE=true
SO_MARK=100
PROXY_PORT=8585


if [ -z "$1" ]; then
  echo "Usage: $0 up|down"

elif [ "$1" == "up" ]; then
	ip rule add fwmark 1 lookup $SO_MARK
	ip route add local 0.0.0.0/0 dev lo table $SO_MARK

	iptables -t mangle -N DIVERT
	iptables -t mangle -A DIVERT -j MARK --set-mark 1
	iptables -t mangle -A DIVERT -j ACCEPT
	iptables -t mangle -A PREROUTING -p tcp -m socket -j DIVERT

	iptables -t mangle -N PROXY
	iptables -t mangle -A PROXY -p tcp -d 127.0.0.0/8 -j RETURN
	iptables -t mangle -A PROXY -p tcp -d 192.168.0.0/16 -j RETURN
	iptables -t mangle -A PROXY -p tcp -m mark --mark $SO_MARK -j RETURN
	iptables -t mangle -A PROXY -p tcp -j TPROXY --tproxy-mark 0x1/0x1 --on-ip 127.0.0.1 --on-port $PROXY_PORT
	iptables -t mangle -A PREROUTING -p tcp -j PROXY

	iptables -t mangle -A PROXY -p udp -d 127.0.0.0/8 -j RETURN
	iptables -t mangle -A PROXY -p udp -d 255.255.255.255/32 -j RETURN
	iptables -t mangle -A PROXY -p udp -d 192.168.0.0/16 -j RETURN
	iptables -t mangle -A PROXY -p udp -m mark --mark $SO_MARK -j RETURN
	iptables -t mangle -A PROXY -p udp -j TPROXY --tproxy-mark 0x1/0x1 --on-ip 127.0.0.1 --on-port $PROXY_PORT
	iptables -t mangle -A PREROUTING -p udp -m multiport --dports 8443,37010 -j PROXY

    if [ "$LOCAL_USE" = true ]; then
        iptables -t mangle -N PROXY_LOCAL
        iptables -t mangle -A PROXY_LOCAL -p tcp -d 127.0.0.0/8 -j RETURN
        iptables -t mangle -A PROXY_LOCAL -p tcp -d 255.255.255.255/32 -j RETURN
        iptables -t mangle -A PROXY_LOCAL -p tcp -d 192.168.0.0/16 -j RETURN
        iptables -t mangle -A PROXY_LOCAL -p tcp -m mark --mark $SO_MARK -j RETURN
        iptables -t mangle -A PROXY_LOCAL -p tcp -j MARK --set-mark 1
        iptables -t mangle -A OUTPUT -p tcp -m multiport --dports 80,8443 -j PROXY_LOCAL

        iptables -t mangle -A PROXY_LOCAL -p udp -d 127.0.0.0/8 -j RETURN
        iptables -t mangle -A PROXY_LOCAL -p udp -d 255.255.255.255/32 -j RETURN
        iptables -t mangle -A PROXY_LOCAL -p udp -d 192.168.0.0/16 -j RETURN
        iptables -t mangle -A PROXY_LOCAL -p udp -m mark --mark $SO_MARK -j RETURN
        iptables -t mangle -A PROXY_LOCAL -p udp -j MARK --set-mark 1
        iptables -t mangle -A OUTPUT -p udp -m multiport --dports 8443,37010 -j PROXY_LOCAL
    fi
	iptables -t mangle -L -n -v
    
else
    echo down
    if [ "$LOCAL_USE" = true ]; then
        iptables -t mangle -D OUTPUT -p udp -m multiport --dports 8443,37010 -j PROXY_LOCAL
        iptables -t mangle -D OUTPUT -p tcp -m multiport --dports 80,8443 -j PROXY_LOCAL
        iptables -t mangle -F PROXY_LOCAL
        iptables -t mangle -X PROXY_LOCAL
    fi
    iptables -t mangle -D PREROUTING -p udp -m multiport --dports 8443,37010 -j PROXY
    iptables -t mangle -D PREROUTING -p tcp -j PROXY
    iptables -t mangle -F PROXY
    iptables -t mangle -X PROXY

	iptables -t mangle -D PREROUTING -p tcp -m socket -j DIVERT
	iptables -t mangle -F DIVERT
	iptables -t mangle -X DIVERT

	iptables -t mangle -L -n -v

	ip route del local 0.0.0.0/0 dev lo table $SO_MARK
	ip rule del fwmark 1 lookup $SO_MARK
fi