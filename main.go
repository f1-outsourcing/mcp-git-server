// Command mcp-git-stdio is an MCP server providing Git repository tools via stdio.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/server"
)

var version = "dev"

func main() {
	var showVersion bool
	flag.BoolVar(&showVersion, "v", false, "Print version")
	flag.BoolVar(&showVersion, "version", false, "Print version")

	flag.Parse()

	if showVersion {
		fmt.Printf("mcp-git-server %s\n", version)
		return
	}

	mcpServer := NewServer()

	if err := server.ServeStdio(mcpServer); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}