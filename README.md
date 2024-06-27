# f-proxy
transparent forwarding proxy for TCP/UDP/QUIC

# setup firewall example
```
./deploy/setup_firewall.sh up
```

# running VictoriaMetrics
```
docker run -it --rm -v ./deploy/victoria-config.yaml:/etc/config.yaml --network="host"  victoriametrics/victoria-metrics -promscrape.config=/etc/config.yaml
```
* open http://localhost:8428/vmui/
* add queries: Rx{}, Tx{}

# api
```
curl http://localhost:8080/stat/sn | python -m json.tool

{
    "phoronix.com": {
        "rx": 14153,
        "tx": 1978
    }
}
```

```
curl http://localhost:8080/stat/ip | python -m json.tool

{
    "172.67.75.80": {
        "rx": 22483,
        "tx": 3178
    },
    "173.255.201.149": {
        "rx": 265,
        "tx": 90
    }
}
```

