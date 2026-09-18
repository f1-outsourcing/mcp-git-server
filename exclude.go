package main

import (
	"io"
	"os"
	"strings"
)

// stderrLogger returns io.Discard if stderr is closed (avoids panics in
// minimal container environments) and os.Stderr otherwise.
func stderrLogger() io.Writer {
	if _, err := os.Stderr.Stat(); err != nil {
		return io.Discard
	}
	return os.Stderr
}

// envOrEmpty returns the value of env var name, or "" if unset/empty.
func envOrEmpty(name string) string {
	v := os.Getenv(name)
	return strings.TrimSpace(v)
}

// joinCSV merges two comma-separated strings, skipping empty parts.
func joinCSV(a, b string) string {
	parts := []string{}
	for _, p := range strings.Split(a, ",") {
		if s := strings.TrimSpace(p); s != "" {
			parts = append(parts, s)
		}
	}
	for _, p := range strings.Split(b, ",") {
		if s := strings.TrimSpace(p); s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, ",")
}

// normalizeExclude normalizes a single exclude entry to a repo-root-relative
// path without leading/trailing slashes, and rejects unsafe values.
func normalizeExclude(in string) string {
	s := strings.TrimSpace(in)
	if s == "" {
		return ""
	}
	// Reject anything that could be misparsed as a pathspec magic or option.
	if strings.ContainsAny(s, "\n\r\t") {
		return ""
	}
	if strings.HasPrefix(s, ":") || strings.HasPrefix(s, "--") {
		return ""
	}
	// Normalize "./x" -> "x" (but NOT ".x" -> "x"; dotfiles must keep their dot).
	s = strings.TrimPrefix(s, "./")
	// Strip any leading/trailing slashes.
	s = strings.Trim(s, "/")
	if s == "" {
		return ""
	}
	return s
}

// parseExcludeList parses a comma-separated exclude list into a deduped,
// normalized []string.
func parseExcludeList(csv string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, p := range strings.Split(csv, ",") {
		s := normalizeExclude(p)
		if s == "" {
			continue
		}
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

// isExcluded reports whether repo-relative path p is inside (or equal to)
// any of the exclude prefixes.
func isExcluded(p string, excludes []string) bool {
	for _, e := range excludes {
		if p == e || strings.HasPrefix(p, e+"/") {
			return true
		}
	}
	return false
}

// excludePathspecs returns the git ":(exclude)..." pathspec arguments for
// the configured exclude list. Returns nil if there is nothing to exclude.
func excludePathspecs(excludes []string) []string {
	if len(excludes) == 0 {
		return nil
	}
	out := make([]string, 0, len(excludes))
	for _, e := range excludes {
		out = append(out, ":(exclude)"+e)
	}
	return out
}
