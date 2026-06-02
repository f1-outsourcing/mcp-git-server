# Makefile for building MCP components

.PHONY: all build-gateway build-stdio clean

all: build-gateway build-stdio

build-gateway:
	go build -o mcp-gateway gateway.go

build-stdio:
	go build -o mcp-git-stdio main-stdio.go server-stdio.go handlers.go

clean:
	rm -f mcp-gateway mcp-git-stdio

test:
	go test ./...

.PHONY: test