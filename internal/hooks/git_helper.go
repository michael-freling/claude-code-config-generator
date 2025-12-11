package hooks

import (
	"fmt"
	"os/exec"
	"strings"
)

// GitHelper provides methods to interact with git commands.
type GitHelper interface {
	// GetCurrentBranch returns the name of the current git branch.
	GetCurrentBranch() (string, error)
}

// realGitHelper implements GitHelper using actual git commands.
type realGitHelper struct{}

// NewGitHelper creates a new GitHelper instance.
func NewGitHelper() GitHelper {
	return &realGitHelper{}
}

// GetCurrentBranch returns the current git branch name.
func (g *realGitHelper) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	branch := strings.TrimSpace(string(output))
	return branch, nil
}
