#!/bin/bash

echo "Building MCP Gateway..."
go build -o mcp-gateway gateway.go

echo "Building MCP Git Stdio Server..."
go build -o mcp-git-stdio main-stdio.go server-stdio.go handlers.go

echo "Build complete!"