package hooks

import (
	"strings"
)

// gitPushRule blocks git push commands to main/master branches.
type gitPushRule struct {
	gitHelper GitHelper
}

// NewGitPushRule creates a new rule that blocks pushes to main/master branches.
func NewGitPushRule(gitHelper GitHelper) Rule {
	return &gitPushRule{
		gitHelper: gitHelper,
	}
}

// Name returns the unique identifier for this rule.
func (r *gitPushRule) Name() string {
	return "git-push"
}

// Description returns a human-readable description of what this rule does.
func (r *gitPushRule) Description() string {
	return "Blocks git push commands to main/master branches"
}

// Evaluate checks if the Bash command is a git push to main/master.
func (r *gitPushRule) Evaluate(input *ToolInput) (*RuleResult, error) {
	if input.ToolName != "Bash" {
		return NewAllowedResult(), nil
	}

	command, ok := input.GetStringArg("command")
	if !ok {
		return NewAllowedResult(), nil
	}

	command = strings.TrimSpace(command)

	// Parse the command to check if it's a git push
	args := parseGitPushArgs(command)
	if len(args) < 2 || args[0] != "git" || args[1] != "push" {
		return NewAllowedResult(), nil
	}

	// Check for explicit branch name
	if isExplicitPushToProtectedBranch(command) {
		return NewBlockedResult(
			r.Name(),
			"Direct push to main/master branch is not allowed",
		), nil
	}

	// Check for implicit push (no branch specified)
	if isImplicitPush(command) {
		// Get current branch
		currentBranch, err := r.gitHelper.GetCurrentBranch()
		if err != nil {
			// Fail open - allow the command if we can't determine the branch
			return NewAllowedResult(), nil
		}

		if isProtectedBranch(currentBranch) {
			return NewBlockedResult(
				r.Name(),
				"Direct push to main/master branch is not allowed",
			), nil
		}
	}

	return NewAllowedResult(), nil
}

// isExplicitPushToProtectedBranch checks if the command explicitly pushes to main/master.
func isExplicitPushToProtectedBranch(command string) bool {
	// Parse the command to extract arguments
	args := parseGitPushArgs(command)

	// Look for branch name in the arguments
	// Common patterns:
	// git push origin main
	// git push -u origin main
	// git push --set-upstream origin main
	// git push -f origin main
	// git push --force origin main

	// Find the last argument that doesn't start with '-' and isn't a known flag value
	var lastNonFlagArg string
	skipNext := false

	for i := 2; i < len(args); i++ { // Start from index 2 to skip "git" and "push"
		arg := args[i]

		if skipNext {
			skipNext = false
			continue
		}

		// Skip flags
		if strings.HasPrefix(arg, "-") {
			// Check if this flag takes a value
			if arg == "--repo" || arg == "--exec" || arg == "--receive-pack" {
				skipNext = true
			}
			continue
		}

		lastNonFlagArg = arg
	}

	return isProtectedBranch(lastNonFlagArg)
}

// isImplicitPush checks if the command is a git push without a branch specified.
func isImplicitPush(command string) bool {
	args := parseGitPushArgs(command)

	// Check if there's a non-flag, non-remote argument
	// git push -> implicit
	// git push origin -> implicit
	// git push -u origin -> implicit
	// git push origin feature -> explicit

	foundNonFlagArg := false
	foundRemote := false
	skipNext := false

	for i := 2; i < len(args); i++ { // Start from index 2 to skip "git" and "push"
		arg := args[i]

		if skipNext {
			skipNext = false
			continue
		}

		// Skip flags
		if strings.HasPrefix(arg, "-") {
			// Check if this flag takes a value
			if arg == "--repo" || arg == "--exec" || arg == "--receive-pack" {
				skipNext = true
			}
			continue
		}

		if !foundRemote {
			// First non-flag arg is typically the remote
			foundRemote = true
			continue
		}

		// Second non-flag arg would be the branch
		foundNonFlagArg = true
		break
	}

	// If we found a second non-flag arg, it's explicit
	// Otherwise, it's implicit
	return !foundNonFlagArg
}

// parseGitPushArgs parses a git push command into arguments.
// This is a simple parser that handles basic quoting.
func parseGitPushArgs(command string) []string {
	var args []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false

	for i := 0; i < len(command); i++ {
		ch := command[i]

		switch ch {
		case '\'':
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			} else {
				current.WriteByte(ch)
			}
		case '"':
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			} else {
				current.WriteByte(ch)
			}
		case ' ', '\t', '\n', '\r':
			if !inSingleQuote && !inDoubleQuote {
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
				}
			} else {
				current.WriteByte(ch)
			}
		default:
			current.WriteByte(ch)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}
