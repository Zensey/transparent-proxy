build:
	go build -o bin/main main.go
	sudo setcap cap_net_bind_service,cap_net_admin+ep bin/main

run:
	go run main.go
