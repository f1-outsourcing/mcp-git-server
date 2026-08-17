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

	// 1. Apply global config settings (including safe.directory *)
	applyGlobalGitConfig()

	// 2. Find and configure existing git repos across the whole filesystem root.
	applyRepoGitConfig()

	log.Println("Git configuration initialized successfully.")
}

// applyGlobalGitConfig sets the required global git config values.
//
// We use `safe.directory *` (the documented "wildcard" form) so that any
// repository on this machine is considered safe, regardless of which user
// created it.  This is the robust fix for the "dubious ownership" error
// that used to appear for repos like /php8-phalcon5.6.1.
func applyGlobalGitConfig() {
	// Add safe.directory wildcard. `--replace-all` lets us overwrite any
	// pre-existing entries so we don't accumulate duplicates on restart.
	cmd := exec.Command("git", "config", "--global", "--replace-all", "safe.directory", "*")
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("Warning: failed to set safe.directory=*: %v\n%s", err, string(output))
	} else {
		log.Println("Set safe.directory=* in global config")
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

	currentValue := trimOutput(output)
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

// ensureRepoSafe makes sure the given repository is considered "safe" by git
// (i.e., no "dubious ownership" error) before we run any git command against
// it.  We rely on the `safe.directory *` wildcard, but in case it is not in
// effect yet (e.g., HOME is unwritable), we fall back to adding the specific
// path.  Idempotent and cheap.
func ensureRepoSafe(repoPath string) {
	if repoPath == "" {
		return
	}
	// Best effort: if the wildcard is not already set, add the specific path.
	cmd := exec.Command("git", "config", "--global", "get-all", "safe.directory")
	out, _ := cmd.CombinedOutput()
	alreadyWild := false
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == "*" {
			alreadyWild = true
			break
		}
	}
	if alreadyWild {
		return
	}
	add := exec.Command("git", "config", "--global", "--add", "safe.directory", repoPath)
	if output, err := add.CombinedOutput(); err != nil {
		log.Printf("Warning: failed to add %s to safe.directory: %v\n%s", repoPath, err, string(output))
	}
}

// applyRepoGitConfig finds all .git directories reachable from the filesystem
// root and applies repo-level configs.  We skip virtual filesystems and stay
// shallow (max depth 3 from /) to keep startup fast.
func applyRepoGitConfig() {
	rootDir := "/"

	// Virtual filesystems we must never descend into.
	skip := map[string]bool{
		"/proc": true,
		"/sys":  true,
		"/dev":  true,
		"/run":  true,
		"/boot": true,
	}

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return nil
		}

		// Never descend into virtual filesystems.
		if info.IsDir() && skip[path] {
			return filepath.SkipDir
		}

		// Limit depth to 3 (e.g., /some/repo/.git).
		depth := strings.Count(relPath, string(filepath.Separator))
		if depth > 3 {
			return filepath.SkipDir
		}

		if info.IsDir() && info.Name() == ".git" {
			repoPath := filepath.Dir(path)
			configureRepo(repoPath)
			return filepath.SkipDir // do not descend into .git
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

	// Make sure this repo is safe (defense in depth; wildcard should already cover it).
	ensureRepoSafe(repoPath)

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

	currentValue := trimOutput(output)
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

// trimOutput trims trailing newlines/carriage returns from command output.
func trimOutput(b []byte) string {
	s := string(b)
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}
