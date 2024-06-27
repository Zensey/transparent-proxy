build:
	go build -o bin/f-proxy github.com/zensey/transparent-proxy/cmd/f-proxy
	sudo setcap cap_net_bind_service,cap_net_admin+ep bin/f-proxy

deploy:
	chmod +x ./deploy/setup_firewall.sh up


