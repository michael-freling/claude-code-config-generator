package hooks

import (
	"fmt"
	"os/exec"
	"strings"
)

// GhHelper provides methods to interact with GitHub CLI commands.
type GhHelper interface {
	// GetPRBaseBranch returns the base branch name for a pull request.
	GetPRBaseBranch(prNumber string) (string, error)
}

// realGhHelper implements GhHelper using actual gh commands.
type realGhHelper struct{}

// NewGhHelper creates a new GhHelper instance.
func NewGhHelper() GhHelper {
	return &realGhHelper{}
}

// GetPRBaseBranch returns the base branch name for the specified PR number.
func (g *realGhHelper) GetPRBaseBranch(prNumber string) (string, error) {
	cmd := exec.Command("gh", "pr", "view", prNumber, "--json", "baseRefName", "--jq", ".baseRefName")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get PR base branch: %w", err)
	}

	baseBranch := strings.TrimSpace(string(output))
	return baseBranch, nil
}
