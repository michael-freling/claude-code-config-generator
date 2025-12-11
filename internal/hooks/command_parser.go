package hooks

import (
	"strings"
)

// isGhApiCommand checks if the command starts with "gh api".
func isGhApiCommand(command string) bool {
	tokens := strings.Fields(command)
	if len(tokens) < 2 {
		return false
	}
	return tokens[0] == "gh" && tokens[1] == "api"
}

// extractHTTPMethod extracts the HTTP method from a gh api command.
// Returns empty string if no method is specified (defaults to GET).
func extractHTTPMethod(command string) string {
	tokens := strings.Fields(command)

	for i := 0; i < len(tokens); i++ {
		if tokens[i] == "-X" || tokens[i] == "--method" {
			if i+1 < len(tokens) {
				return strings.ToUpper(tokens[i+1])
			}
		}
	}

	return ""
}

// isProtectedBranch checks if a branch name is main or master.
func isProtectedBranch(branch string) bool {
	branch = strings.TrimSpace(branch)
	return branch == "main" || branch == "master"
}
