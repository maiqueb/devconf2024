GO ?= go

build:
	$(GO) build -o plugins/bin/status-cni cmd/main.go
	$(GO) build -o plugins/bin/ipam-status-cni cmd/ipam/main.go
