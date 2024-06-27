module github.com/zensey/transparent-proxy

go 1.22

toolchain go1.22.3

require (
	github.com/LiamHaworth/go-tproxy v0.0.0-20190726054950-ef7efd7f24ed
	github.com/go-gost/core v0.0.0-20240508132029-8d554ddcf77c
	github.com/go-gost/tls-dissector v0.0.1
	github.com/quic-go/quic-go v0.45.0
	golang.org/x/crypto v0.24.0
	golang.org/x/exp v0.0.0-20240613232115-7f521ea00fb8
)

require (
	github.com/VictoriaMetrics/metrics v1.34.0 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/google/pprof v0.0.0-20210407192527-94a9f03dee38 // indirect
	github.com/onsi/ginkgo/v2 v2.9.5 // indirect
	github.com/valyala/fastrand v1.1.0 // indirect
	github.com/valyala/histogram v1.2.0 // indirect
	go.uber.org/mock v0.4.0 // indirect
	golang.org/x/mod v0.18.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/tools v0.22.0 // indirect
)

replace github.com/quic-go/quic-go => github.com/zensey/quic-go v0.45.1-0.20240620110850-9cef3e4a011b
