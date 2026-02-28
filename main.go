// Command mcp-server-git is an MCP server providing Git repository tools.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/server"
)

var version = "dev"

func main() {
	var transport string
	flag.StringVar(&transport, "t", "stdio", "Transport type (stdio or http)")
	flag.StringVar(&transport, "transport", "stdio", "Transport type (stdio or http)")

	var port string
	flag.StringVar(&port, "p", "8080", "HTTP port (only used with http transport)")
	flag.StringVar(&port, "port", "8080", "HTTP port (only used with http transport)")

	var showVersion bool
	flag.BoolVar(&showVersion, "v", false, "Print version")
	flag.BoolVar(&showVersion, "version", false, "Print version")

	flag.Parse()

	if showVersion {
		fmt.Printf("mcp-server-git %s\n", version)
		return
	}

	mcpServer := NewServer()

	if transport == "http" {
		httpServer := server.NewStreamableHTTPServer(mcpServer)
		log.Printf("Git MCP server listening on :%s/mcp", port)
		if err := httpServer.Start(":" + port); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	} else {
		if err := server.ServeStdio(mcpServer); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}
