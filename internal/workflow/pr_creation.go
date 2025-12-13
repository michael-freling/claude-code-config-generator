package workflow

import (
	"fmt"
	"strings"
	"time"
)

const (
	// maxPRCreationAttempts defines the maximum number of attempts to create a PR
	// before giving up. This allows for transient failures to be retried.
	maxPRCreationAttempts = 3

	// prCreationRetryDelay is the delay between PR creation retry attempts.
	// This gives time for transient issues to resolve.
	prCreationRetryDelay = 5 * time.Second

	// prCreationTimeout is the maximum time allowed for a single PR creation attempt.
	// PR creation should be quick as it's primarily a GitHub API operation.
	prCreationTimeout = 10 * time.Minute
)

// PRMetadata contains GitHub-specific metadata extracted from user prompts
// for PR creation. This includes issue references, labels, and project assignments.
type PRMetadata struct {
	Issues   []string `json:"issues,omitempty"`
	Labels   []string `json:"labels,omitempty"`
	Projects []string `json:"projects,omitempty"`
}

// PRCreationResult represents the result of a PR creation attempt.
// It contains the PR number (if created or found), the status of the operation,
// and a message explaining the result. The status can be "created" (new PR),
// "exists" (PR already exists), "skipped" (no commits to create PR for), or
// "failed" (PR creation failed).
type PRCreationResult struct {
	PRNumber int         `json:"prNumber"`
	Status   string      `json:"status"` // "created", "exists", "skipped", "failed"
	Message  string      `json:"message"`
	Metadata *PRMetadata `json:"metadata,omitempty"`
}

// PRCreationResultSchema is the JSON schema for Claude's PR creation output
var PRCreationResultSchema = `{
    "type": "object",
    "properties": {
        "prNumber": {"type": "integer", "description": "The PR number if created or found"},
        "status": {"type": "string", "enum": ["created", "exists", "skipped", "failed"], "description": "The result status"},
        "message": {"type": "string", "description": "A message explaining the result"},
        "metadata": {
            "type": "object",
            "description": "Optional GitHub metadata extracted from user prompts",
            "properties": {
                "issues": {
                    "type": "array",
                    "items": {"type": "string"},
                    "description": "Issue references like '#123', 'fixes #456', 'closes #789'"
                },
                "labels": {
                    "type": "array",
                    "items": {"type": "string"},
                    "description": "Label names to apply to the PR like 'bug', 'enhancement', 'documentation'"
                },
                "projects": {
                    "type": "array",
                    "items": {"type": "string"},
                    "description": "Project names to add the PR to like 'Q1 Planning', 'Roadmap'"
                }
            }
        }
    },
    "required": ["status", "message"]
}`

// logPRMetadata logs applied PR metadata to the console if present.
// It formats and displays issue references, labels, and project assignments
// that were applied to the PR.
func logPRMetadata(metadata *PRMetadata) {
	if metadata == nil {
		return
	}

	if len(metadata.Issues) > 0 {
		fmt.Printf("  %s Applied issue references: %s\n", Green("✓"), strings.Join(metadata.Issues, ", "))
	}

	if len(metadata.Labels) > 0 {
		fmt.Printf("  %s Applied labels: %s\n", Green("✓"), strings.Join(metadata.Labels, ", "))
	}

	if len(metadata.Projects) > 0 {
		fmt.Printf("  %s Applied to projects: %s\n", Green("✓"), strings.Join(metadata.Projects, ", "))
	}
}
