# mcp-server-git

A Model Context Protocol (MCP) server for Git repository operations. Built in Go using [go-git](https://github.com/go-git/go-git) for native Git support.

## Features

- **Pure Go**: No external Git binary required
- **Cross-platform**: Works on Linux, macOS, and Windows
- **MCP Compatible**: Works with any MCP client

## Tools

| Tool | Description |
|------|-------------|
| `git_status` | Show working tree status |
| `git_diff_unstaged` | Show changes not yet staged |
| `git_diff_staged` | Show changes staged for commit |
| `git_diff` | Show differences between branches/commits |
| `git_add` | Stage files for commit |
| `git_reset` | Unstage all staged changes |
| `git_commit` | Create a commit |
| `git_log` | Show commit history |
| `git_create_branch` | Create a new branch |
| `git_checkout` | Switch branches |
| `git_show` | Show commit contents |
| `git_branch` | List branches |

## Installation

### Download Binary

Download the latest release from [GitHub Releases](https://github.com/tosolveit/mcp-server-git/releases).

### Build from Source

```bash
go install github.com/tosolveit/mcp-server-git@latest
```

## Usage

### With Claude Desktop

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "git": {
      "command": "/path/to/mcp-server-git"
    }
  }
}
```

### Standalone

```bash
# Run with stdio transport (default)
mcp-server-git

# Run with HTTP transport
mcp-server-git --transport http --port 8080

# Print version
mcp-server-git --version
```

## Configuration

All tools require a `repo_path` parameter pointing to a Git repository.

## License

MIT
