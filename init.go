package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// requiredGlobalConfigs defines the global git config settings to apply on startup.
var requiredGlobalConfigs = map[string]string{
	"core.autocrlf":   "true",
	"core.filemode":   "false",
	"core.trustctime": "false",
	"core.checkstat":  "minimal",
}

// requiredRepoConfigs defines the repo-level git config settings to apply on startup.
var requiredRepoConfigs = map[string]string{
	"core.autocrlf":   "true",
	"core.filemode":   "false",
	"core.trustctime": "false",
	"core.checkstat":  "minimal",
}

// initGitConfig applies required git configuration settings on startup.
// It configures global settings and then scans for existing git repos to configure.
// All logging is done to stderr to avoid interfering with MCP stdio protocol on stdout.
func initGitConfig() {
	// Ensure all logs go to stderr (not stdout which is used for MCP protocol)
	log.SetOutput(os.Stderr)

	log.Println("Initializing git configuration...")

	// 1. Apply global config settings
	applyGlobalGitConfig()

	// 2. Find and configure existing git repos
	applyRepoGitConfig()

	log.Println("Git configuration initialized successfully.")
}

// applyGlobalGitConfig sets the required global git config values.
func applyGlobalGitConfig() {
	// Add safe.directory for /mcp-git-server (and all subdirectories via *)
	cmd := exec.Command("git", "config", "--global", "--add", "safe.directory", "/mcp-git-server")
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("Warning: failed to add safe.directory: %v\n%s", err, string(output))
	} else {
		log.Println("Added /mcp-git-server to safe.directory")
	}

	// Check and apply each global config setting
	for key, value := range requiredGlobalConfigs {
		if err := ensureGlobalConfig(key, value); err != nil {
			log.Printf("Warning: failed to set global config %s=%s: %v", key, value, err)
		}
	}
}

// ensureGlobalConfig checks if a global config key has the expected value and sets it if missing or different.
func ensureGlobalConfig(key, value string) error {
	// Get current value
	cmd := exec.Command("git", "config", "--global", key)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Config key doesn't exist yet, set it
		return setGlobalConfig(key, value)
	}

	currentValue := string(output)
	// Trim newline
	for len(currentValue) > 0 && (currentValue[len(currentValue)-1] == '\n' || currentValue[len(currentValue)-1] == '\r') {
		currentValue = currentValue[:len(currentValue)-1]
	}

	if currentValue == value {
		log.Printf("Global config %s is already set to %s", key, value)
		return nil
	}

	log.Printf("Updating global config %s from %s to %s", key, currentValue, value)
	return setGlobalConfig(key, value)
}

// setGlobalConfig sets a global git config value.
func setGlobalConfig(key, value string) error {
	cmd := exec.Command("git", "config", "--global", key, value)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v\n%s", err, string(output))
	}
	log.Printf("Set global config %s=%s", key, value)
	return nil
}

// applyRepoGitConfig finds all .git directories under /mcp-git-server and applies repo-level configs.
func applyRepoGitConfig() {
	rootDir := "/mcp-git-server"

	// Walk the directory tree to find .git directories (max depth 2 from root)
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}

		// Calculate depth relative to rootDir
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return nil
		}

		// Count path separators to determine depth
		// relPath "." means depth 0 (rootDir itself)
		// relPath "repo" means depth 1
		// relPath "repo/.git" means depth 2
		depth := strings.Count(relPath, string(filepath.Separator))

		// Limit depth to 2 (i.e., rootDir/repo/.git)
		if depth > 2 {
			return filepath.SkipDir
		}

		// Check if this is a .git directory
		if info.IsDir() && info.Name() == ".git" {
			repoPath := filepath.Dir(path)
			configureRepo(repoPath)
			// Skip contents of .git directory
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		log.Printf("Warning: error scanning for git repos: %v", err)
	}
}

// configureRepo applies the required git config settings to a specific repository.
func configureRepo(repoPath string) {
	log.Printf("Configuring repo: %s", repoPath)

	// Add to safe.directory
	cmd := exec.Command("git", "config", "--global", "--add", "safe.directory", repoPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("Warning: failed to add %s to safe.directory: %v\n%s", repoPath, err, string(output))
	}

	// Apply repo-level configs
	for key, value := range requiredRepoConfigs {
		if err := ensureRepoConfig(repoPath, key, value); err != nil {
			log.Printf("Warning: failed to set repo config %s=%s in %s: %v", key, value, repoPath, err)
		}
	}
}

// ensureRepoConfig checks if a repo config key has the expected value and sets it if missing or different.
func ensureRepoConfig(repoPath, key, value string) error {
	// Get current value
	cmd := exec.Command("git", "-C", repoPath, "config", key)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Config key doesn't exist yet, set it
		return setRepoConfig(repoPath, key, value)
	}

	currentValue := string(output)
	// Trim newline
	for len(currentValue) > 0 && (currentValue[len(currentValue)-1] == '\n' || currentValue[len(currentValue)-1] == '\r') {
		currentValue = currentValue[:len(currentValue)-1]
	}

	if currentValue == value {
		log.Printf("Repo config %s is already set to %s in %s", key, value, repoPath)
		return nil
	}

	log.Printf("Updating repo config %s from %s to %s in %s", key, currentValue, value, repoPath)
	return setRepoConfig(repoPath, key, value)
}

// setRepoConfig sets a repo-level git config value.
func setRepoConfig(repoPath, key, value string) error {
	cmd := exec.Command("git", "-C", repoPath, "config", key, value)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v\n%s", err, string(output))
	}
	log.Printf("Set repo config %s=%s in %s", key, value, repoPath)
	return nil
}