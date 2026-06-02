// Command mcp-gateway is an MCP gateway that can run any stdio MCP server.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
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

	var serverPath string
	flag.StringVar(&serverPath, "s", "", "Path to the stdio MCP server executable")
	flag.StringVar(&serverPath, "server", "", "Path to the stdio MCP server executable")

	flag.Parse()

	if showVersion {
		fmt.Printf("mcp-gateway %s\n", version)
		return
	}

	if serverPath == "" && transport == "stdio" {
		log.Fatal("Server path is required for stdio transport")
	}

	if transport == "http" {
		// HTTP transport - create a proxy server
		mcpServer := server.NewMCPServer(
			"mcp-gateway",
			version,
			server.WithToolCapabilities(true),
		)

		// Add a tool to run the stdio server
		mcpServer.AddTool(mcp.NewTool("run_stdio_server",
			mcp.WithDescription("Runs a stdio MCP server"),
			mcp.WithString("server_path",
				mcp.Required(),
				mcp.Description("Path to the stdio MCP server executable"),
			),
		), handleRunStdioServer)

		httpServer := server.NewStreamableHTTPServer(mcpServer)
		log.Printf("MCP Gateway listening on :%s/mcp", port)
		if err := httpServer.Start(":" + port); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	} else {
		// Stdio transport - directly run the server
		if serverPath != "" {
			runStdioServer(serverPath)
		} else {
			// If no server specified, just run the gateway as stdio server
			log.Fatal("Server path required for stdio transport")
		}
	}
}

func handleRunStdioServer(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	serverPath := request.GetString("server_path", "")
	if serverPath == "" {
		return mcp.NewToolResultError("server_path is required"), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Running stdio server: %s", serverPath)), nil
}

func runStdioServer(serverPath string) {
	// Resolve the absolute path
	absPath, err := filepath.Abs(serverPath)
	if err != nil {
		log.Fatalf("Failed to resolve server path: %v", err)
	}

	// Create command
	cmd := exec.Command(absPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}