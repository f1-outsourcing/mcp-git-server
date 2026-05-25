package main

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewServer creates a new Git MCP server with all git tools registered.
func NewServer() *server.MCPServer {
	s := server.NewMCPServer(
		"mcp-server-git",
		version,
		server.WithToolCapabilities(true),
	)

	// Register all git tools
	registerTools(s)

	return s
}

func registerTools(s *server.MCPServer) {
	// git_status - Show working tree status
	s.AddTool(mcp.NewTool("git_status",
		mcp.WithDescription("Shows the working tree status"),
		mcp.WithString("repo_path",
			mcp.Required(),
			mcp.Description("Path to Git repository"),
		),
	), handleGitStatus)

	// git_diff_unstaged - Show unstaged changes
	s.AddTool(mcp.NewTool("git_diff_unstaged",
		mcp.WithDescription("Shows changes in working directory not yet staged"),
		mcp.WithString("repo_path",
			mcp.Required(),
			mcp.Description("Path to Git repository"),
		),
	), handleGitDiffUnstaged)

	// git_diff_staged - Show staged changes
	s.AddTool(mcp.NewTool("git_diff_staged",
		mcp.WithDescription("Shows changes that are staged for commit"),
		mcp.WithString("repo_path",
			mcp.Required(),
			mcp.Description("Path to Git repository"),
		),
	), handleGitDiffStaged)

	// git_diff - Show differences between branches or commits
	s.AddTool(mcp.NewTool("git_diff",
		mcp.WithDescription("Shows differences between branches or commits"),
		mcp.WithString("repo_path",
			mcp.Required(),
			mcp.Description("Path to Git repository"),
		),
		mcp.WithString("target",
			mcp.Required(),
			mcp.Description("Target branch or commit to compare with"),
		),
	), handleGitDiff)

	// git_add - Stage files
	s.AddTool(mcp.NewTool("git_add",
		mcp.WithDescription("Adds file contents to the staging area"),
		mcp.WithString("repo_path",
			mcp.Required(),
			mcp.Description("Path to Git repository"),
		),
		mcp.WithArray("files",
			mcp.Required(),
			mcp.Description("Array of file paths to stage"),
			mcp.Items(map[string]any{"type": "string"}),
		),
	), handleGitAdd)

	// git_restore - Restore a file in the working directory to HEAD
	s.AddTool(mcp.NewTool("git_restore",
		mcp.WithDescription("Restores a file in the working directory to the version in HEAD"),
		mcp.WithString("repo_path",
			mcp.Required(),
			mcp.Description("Path to Git repository"),
		),
		mcp.WithString("file",
			mcp.Required(),
			mcp.Description("Path of the file to restore"),
		),
	), handleGitRestore)

	// git_reset - Unstage all changes
	s.AddTool(mcp.NewTool("git_reset",
		mcp.WithDescription("Unstages all staged changes"),
		mcp.WithString("repo_path",
			mcp.Required(),
			mcp.Description("Path to Git repository"),
		),
	), handleGitReset)

	// git_commit - Create a commit
	s.AddTool(mcp.NewTool("git_commit",
		mcp.WithDescription("Records changes to the repository"),
		mcp.WithString("repo_path",
			mcp.Required(),
			mcp.Description("Path to Git repository"),
		),
		mcp.WithString("message",
			mcp.Required(),
			mcp.Description("Commit message"),
		),
	), handleGitCommit)

	// git_log - Show commit history
	s.AddTool(mcp.NewTool("git_log",
		mcp.WithDescription("Shows the commit logs"),
		mcp.WithString("repo_path",
			mcp.Required(),
			mcp.Description("Path to Git repository"),
		),
		mcp.WithNumber("max_count",
			mcp.Description("Maximum number of commits to show (default: 10)"),
		),
	), handleGitLog)

	// git_create_branch - Create a new branch
	s.AddTool(mcp.NewTool("git_create_branch",
		mcp.WithDescription("Creates a new branch"),
		mcp.WithString("repo_path",
			mcp.Required(),
			mcp.Description("Path to Git repository"),
		),
		mcp.WithString("branch_name",
			mcp.Required(),
			mcp.Description("Name of the new branch"),
		),
		mcp.WithString("base_branch",
			mcp.Description("Base branch to create from (defaults to current branch)"),
		),
	), handleGitCreateBranch)

	// git_checkout - Switch branches
	s.AddTool(mcp.NewTool("git_checkout",
		mcp.WithDescription("Switches branches"),
		mcp.WithString("repo_path",
			mcp.Required(),
			mcp.Description("Path to Git repository"),
		),
		mcp.WithString("branch_name",
			mcp.Required(),
			mcp.Description("Name of branch to checkout"),
		),
	), handleGitCheckout)

	// git_show - Show commit contents
	s.AddTool(mcp.NewTool("git_show",
		mcp.WithDescription("Shows the contents of a commit"),
		mcp.WithString("repo_path",
			mcp.Required(),
			mcp.Description("Path to Git repository"),
		),
		mcp.WithString("revision",
			mcp.Required(),
			mcp.Description("The revision (commit hash, branch name, tag) to show"),
		),
	), handleGitShow)

	// git_branch - List branches
	s.AddTool(mcp.NewTool("git_branch",
		mcp.WithDescription("Lists Git branches"),
		mcp.WithString("repo_path",
			mcp.Required(),
			mcp.Description("Path to Git repository"),
		),
	), handleGitBranch)
}
