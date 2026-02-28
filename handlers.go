package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/mark3labs/mcp-go/mcp"
)

func handleGitStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoPath := request.GetString("repo_path", "")
	if repoPath == "" {
		return mcp.NewToolResultError("repo_path is required"), nil
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to open repository: %v", err)), nil
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get worktree: %v", err)), nil
	}

	status, err := worktree.Status()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get status: %v", err)), nil
	}

	if status.IsClean() {
		return mcp.NewToolResultText("Working tree clean"), nil
	}

	var sb strings.Builder
	for path, fileStatus := range status {
		sb.WriteString(fmt.Sprintf("%c%c %s\n",
			fileStatus.Staging,
			fileStatus.Worktree,
			path,
		))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handleGitDiffUnstaged(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoPath := request.GetString("repo_path", "")
	if repoPath == "" {
		return mcp.NewToolResultError("repo_path is required"), nil
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to open repository: %v", err)), nil
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get worktree: %v", err)), nil
	}

	status, err := worktree.Status()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get status: %v", err)), nil
	}

	var changes []string
	for path, fileStatus := range status {
		if fileStatus.Worktree != git.Unmodified && fileStatus.Worktree != ' ' {
			changes = append(changes, fmt.Sprintf("%c %s", fileStatus.Worktree, path))
		}
	}

	if len(changes) == 0 {
		return mcp.NewToolResultText("No unstaged changes"), nil
	}

	return mcp.NewToolResultText(strings.Join(changes, "\n")), nil
}

func handleGitDiffStaged(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoPath := request.GetString("repo_path", "")
	if repoPath == "" {
		return mcp.NewToolResultError("repo_path is required"), nil
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to open repository: %v", err)), nil
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get worktree: %v", err)), nil
	}

	status, err := worktree.Status()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get status: %v", err)), nil
	}

	var changes []string
	for path, fileStatus := range status {
		if fileStatus.Staging != git.Unmodified && fileStatus.Staging != ' ' {
			changes = append(changes, fmt.Sprintf("%c %s", fileStatus.Staging, path))
		}
	}

	if len(changes) == 0 {
		return mcp.NewToolResultText("No staged changes"), nil
	}

	return mcp.NewToolResultText(strings.Join(changes, "\n")), nil
}

func handleGitDiff(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoPath := request.GetString("repo_path", "")
	target := request.GetString("target", "")
	if repoPath == "" {
		return mcp.NewToolResultError("repo_path is required"), nil
	}
	if target == "" {
		return mcp.NewToolResultError("target is required"), nil
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to open repository: %v", err)), nil
	}

	// Get HEAD commit
	headRef, err := repo.Head()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get HEAD: %v", err)), nil
	}

	headCommit, err := repo.CommitObject(headRef.Hash())
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get HEAD commit: %v", err)), nil
	}

	// Get target commit
	targetHash, err := repo.ResolveRevision(plumbing.Revision(target))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to resolve target: %v", err)), nil
	}

	targetCommit, err := repo.CommitObject(*targetHash)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get target commit: %v", err)), nil
	}

	// Get trees
	headTree, err := headCommit.Tree()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get HEAD tree: %v", err)), nil
	}

	targetTree, err := targetCommit.Tree()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get target tree: %v", err)), nil
	}

	// Compare trees
	changes, err := targetTree.Diff(headTree)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to diff: %v", err)), nil
	}

	if len(changes) == 0 {
		return mcp.NewToolResultText("No differences"), nil
	}

	var sb strings.Builder
	for _, change := range changes {
		action, err := change.Action()
		if err != nil {
			continue
		}
		sb.WriteString(fmt.Sprintf("%s: %s\n", action, change.To.Name))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handleGitAdd(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoPath := request.GetString("repo_path", "")
	if repoPath == "" {
		return mcp.NewToolResultError("repo_path is required"), nil
	}

	args := request.GetArguments()
	filesRaw, ok := args["files"]
	if !ok {
		return mcp.NewToolResultError("files is required"), nil
	}

	filesSlice, ok := filesRaw.([]any)
	if !ok {
		return mcp.NewToolResultError("files must be an array of strings"), nil
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to open repository: %v", err)), nil
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get worktree: %v", err)), nil
	}

	var staged []string
	for _, f := range filesSlice {
		file, ok := f.(string)
		if !ok {
			continue
		}
		_, err := worktree.Add(file)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to add %s: %v", file, err)), nil
		}
		staged = append(staged, file)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Staged files: %s", strings.Join(staged, ", "))), nil
}

func handleGitReset(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoPath := request.GetString("repo_path", "")
	if repoPath == "" {
		return mcp.NewToolResultError("repo_path is required"), nil
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to open repository: %v", err)), nil
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get worktree: %v", err)), nil
	}

	head, err := repo.Head()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get HEAD: %v", err)), nil
	}

	err = worktree.Reset(&git.ResetOptions{
		Commit: head.Hash(),
		Mode:   git.MixedReset,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to reset: %v", err)), nil
	}

	return mcp.NewToolResultText("Reset successful"), nil
}

func handleGitCommit(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoPath := request.GetString("repo_path", "")
	message := request.GetString("message", "")
	if repoPath == "" {
		return mcp.NewToolResultError("repo_path is required"), nil
	}
	if message == "" {
		return mcp.NewToolResultError("message is required"), nil
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to open repository: %v", err)), nil
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get worktree: %v", err)), nil
	}

	hash, err := worktree.Commit(message, &git.CommitOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to commit: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Committed: %s", hash.String())), nil
}

func handleGitLog(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoPath := request.GetString("repo_path", "")
	if repoPath == "" {
		return mcp.NewToolResultError("repo_path is required"), nil
	}

	args := request.GetArguments()
	maxCount := 10
	if mc, ok := args["max_count"].(float64); ok && mc > 0 {
		maxCount = int(mc)
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to open repository: %v", err)), nil
	}

	logIter, err := repo.Log(&git.LogOptions{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get log: %v", err)), nil
	}

	var sb strings.Builder
	count := 0
	err = logIter.ForEach(func(c *object.Commit) error {
		if count >= maxCount {
			return fmt.Errorf("stop") // Stop iteration
		}
		sb.WriteString(fmt.Sprintf("commit %s\n", c.Hash.String()))
		sb.WriteString(fmt.Sprintf("Author: %s <%s>\n", c.Author.Name, c.Author.Email))
		sb.WriteString(fmt.Sprintf("Date:   %s\n", c.Author.When.Format("Mon Jan 2 15:04:05 2006 -0700")))
		sb.WriteString(fmt.Sprintf("\n    %s\n\n", strings.Split(c.Message, "\n")[0]))
		count++
		return nil
	})

	// Ignore "stop" error used to break iteration
	if err != nil && err.Error() != "stop" {
		return mcp.NewToolResultError(fmt.Sprintf("failed to iterate log: %v", err)), nil
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handleGitCreateBranch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoPath := request.GetString("repo_path", "")
	branchName := request.GetString("branch_name", "")
	baseBranch := request.GetString("base_branch", "")

	if repoPath == "" {
		return mcp.NewToolResultError("repo_path is required"), nil
	}
	if branchName == "" {
		return mcp.NewToolResultError("branch_name is required"), nil
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to open repository: %v", err)), nil
	}

	// Get the base commit
	var baseHash plumbing.Hash
	if baseBranch != "" {
		hash, err := repo.ResolveRevision(plumbing.Revision(baseBranch))
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to resolve base branch: %v", err)), nil
		}
		baseHash = *hash
	} else {
		headRef, err := repo.Head()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get HEAD: %v", err)), nil
		}
		baseHash = headRef.Hash()
	}

	// Create the new branch reference
	branchRef := plumbing.NewHashReference(
		plumbing.NewBranchReferenceName(branchName),
		baseHash,
	)

	err = repo.Storer.SetReference(branchRef)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create branch: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Created branch '%s'", branchName)), nil
}

func handleGitCheckout(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoPath := request.GetString("repo_path", "")
	branchName := request.GetString("branch_name", "")

	if repoPath == "" {
		return mcp.NewToolResultError("repo_path is required"), nil
	}
	if branchName == "" {
		return mcp.NewToolResultError("branch_name is required"), nil
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to open repository: %v", err)), nil
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get worktree: %v", err)), nil
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to checkout: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Switched to branch '%s'", branchName)), nil
}

func handleGitShow(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoPath := request.GetString("repo_path", "")
	revision := request.GetString("revision", "")

	if repoPath == "" {
		return mcp.NewToolResultError("repo_path is required"), nil
	}
	if revision == "" {
		return mcp.NewToolResultError("revision is required"), nil
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to open repository: %v", err)), nil
	}

	hash, err := repo.ResolveRevision(plumbing.Revision(revision))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to resolve revision: %v", err)), nil
	}

	commit, err := repo.CommitObject(*hash)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get commit: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("commit %s\n", commit.Hash.String()))
	sb.WriteString(fmt.Sprintf("Author: %s <%s>\n", commit.Author.Name, commit.Author.Email))
	sb.WriteString(fmt.Sprintf("Date:   %s\n", commit.Author.When.Format("Mon Jan 2 15:04:05 2006 -0700")))
	sb.WriteString(fmt.Sprintf("\n%s\n", commit.Message))

	// Show changed files
	parent, err := commit.Parent(0)
	if err == nil {
		parentTree, _ := parent.Tree()
		commitTree, _ := commit.Tree()
		if parentTree != nil && commitTree != nil {
			changes, _ := parentTree.Diff(commitTree)
			if len(changes) > 0 {
				sb.WriteString("\nChanged files:\n")
				for _, change := range changes {
					action, _ := change.Action()
					name := change.To.Name
					if name == "" {
						name = change.From.Name
					}
					sb.WriteString(fmt.Sprintf("  %s: %s\n", action, name))
				}
			}
		}
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handleGitBranch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoPath := request.GetString("repo_path", "")
	if repoPath == "" {
		return mcp.NewToolResultError("repo_path is required"), nil
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to open repository: %v", err)), nil
	}

	branchIter, err := repo.Branches()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list branches: %v", err)), nil
	}

	head, _ := repo.Head()
	var currentBranch string
	if head != nil && head.Name().IsBranch() {
		currentBranch = head.Name().Short()
	}

	var branches []string
	err = branchIter.ForEach(func(ref *plumbing.Reference) error {
		name := ref.Name().Short()
		if name == currentBranch {
			branches = append(branches, "* "+name)
		} else {
			branches = append(branches, "  "+name)
		}
		return nil
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to iterate branches: %v", err)), nil
	}

	return mcp.NewToolResultText(strings.Join(branches, "\n")), nil
}
