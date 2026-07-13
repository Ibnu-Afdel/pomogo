.PHONY: build run test lint clean snapshot

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
	-X main.Version=$(VERSION) \
	-X main.Commit=$(COMMIT) \
	-X main.Date=$(DATE)

build:
	go build -ldflags "$(LDFLAGS)" -o pomogo ./cmd/pomogo

run: build
	./pomogo

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -f pomogo

snapshot:
	goreleaser release --snapshot --clean
