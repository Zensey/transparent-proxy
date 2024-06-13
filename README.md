# transparent-proxy
transparent proxy for TCP

# setup firewall
```
ip rule add fwmark 1 lookup 100
ip route add local 0.0.0.0/0 dev lo table 100

iptables -t mangle -N DIVERT
iptables -t mangle -A DIVERT -j MARK --set-mark 1
iptables -t mangle -A DIVERT -j ACCEPT
iptables -t mangle -A PREROUTING -p tcp -m socket -j DIVERT

iptables -t mangle -N PROXY
iptables -t mangle -A PROXY -p tcp -d 127.0.0.0/8 -j RETURN
iptables -t mangle -A PROXY -p tcp -d 192.168.0.0/16 -j RETURN
iptables -t mangle -A PROXY -p tcp -m mark --mark 100 -j RETURN
iptables -t mangle -A PROXY -p tcp -j TPROXY --tproxy-mark 0x1/0x1 --on-ip 127.0.0.1 --on-port 8585
iptables -t mangle -A PREROUTING -p tcp -j PROXY

iptables -t mangle -N PROXY_LOCAL
iptables -t mangle -A PROXY_LOCAL -p tcp -d 127.0.0.0/8 -j RETURN
iptables -t mangle -A PROXY_LOCAL -p tcp -d 255.255.255.255/32 -j RETURN
iptables -t mangle -A PROXY_LOCAL -p tcp -d 192.168.0.0/16 -j RETURN
iptables -t mangle -A PROXY_LOCAL -p tcp -m mark --mark 100 -j RETURN
iptables -t mangle -A PROXY_LOCAL -p tcp -j MARK --set-mark 1
iptables -t mangle -A OUTPUT -p tcp -m multiport --dports 80,8443 -j PROXY_LOCAL
```