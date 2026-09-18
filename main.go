// Command mcp-git-stdio is an MCP server providing Git repository tools via stdio.
//
// Configuration
// -------------
// Exclusions are configured via the MCP_GIT_EXCLUDE_DIRS environment variable
// (comma-separated list of repo-root-relative paths), mirroring the
// MCP_ALLOWED_DIRS convention used by /mcp-filesystem-server. A --exclude-dir
// CLI flag is also accepted and merged with the env var, but note that the
// stdio MCP gateway may not forward CLI arguments to the child process, so
// the env var is the recommended mechanism.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/server"
)

var version = "dev"

// excludeDirs holds normalized paths (relative to repo root) that are
// excluded from status/diff output. Populated at startup from
// MCP_GIT_EXCLUDE_DIRS (primary) and/or --exclude-dir (secondary).
var excludeDirs = []string{}

func main() {
	var (
		showVersion bool
		excludeDir  string
	)
	flag.BoolVar(&showVersion, "v", false, "Print version")
	flag.BoolVar(&showVersion, "version", false, "Print version")
	flag.StringVar(&excludeDir, "exclude-dir", "",
		"Comma-separated list of repo-root-relative paths to exclude from status/diff "+
			"(e.g. .continue,.git,node_modules). Prefer MCP_GIT_EXCLUDE_DIRS env var; "+
			"the stdio gateway may not forward CLI args to the child.")

	flag.Parse()

	if showVersion {
		fmt.Printf("mcp-git-server %s\n", version)
		return
	}

	// Primary source: env var (matches MCP_ALLOWED_DIRS convention used by the
	// filesystem MCP server, and works even when the stdio transport does not
	// forward CLI arguments to the child process).
	excludeDir = joinCSV(envOrEmpty("MCP_GIT_EXCLUDE_DIRS"), excludeDir)

	excludeDirs = parseExcludeList(excludeDir)

	log.SetOutput(stderrLogger())
	log.Printf("mcp-git-stdio starting (excludes: %v)", excludeDirs)

	// Initialize git configuration on startup
	initGitConfig()

	mcpServer := NewServer()

	if err := server.ServeStdio(mcpServer); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
