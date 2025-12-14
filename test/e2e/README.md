# E2E Testing

This directory contains end-to-end (e2e) tests that test complete workflows using real git, gh, and claude CLI commands without mocking.

## Overview

E2E tests are separated from unit tests using Go build tags (`//go:build e2e`). These tests:

- Use real git commands to create repositories, branches, and commits
- Test actual workflow phases (planning, implementation, PR split)
- Verify integration between components with real filesystem operations
- Skip gracefully when optional tools (gh, claude) are not available

## Prerequisites

### Required Tools

- **git** - Required for all e2e tests
- **Go** - For running the tests

### Optional Tools

- **gh** - GitHub CLI (tests requiring gh authentication will be skipped if not available)
- **claude** - Claude Code CLI (tests requiring claude will be skipped if not available)

### Tool Installation

```bash
# Install gh CLI (macOS)
brew install gh

# Install gh CLI (Linux)
# See https://cli.github.com/manual/installation

# Authenticate gh
gh auth login

# Install claude CLI
# See https://docs.anthropic.com/en/docs/claude-code
```

## Running E2E Tests

### Using the Script (Recommended)

```bash
# Run all e2e tests
./scripts/run-e2e-tests.sh

# Run with verbose output
E2E_VERBOSE=true ./scripts/run-e2e-tests.sh

# Run with custom timeout
E2E_TIMEOUT=2m ./scripts/run-e2e-tests.sh
```

### Using go test Directly

```bash
# Run all e2e tests
go test -tags=e2e ./test/e2e/...

# Run with verbose output
go test -tags=e2e -v ./test/e2e/...

# Run specific test file
go test -tags=e2e -v ./test/e2e/git_operations_e2e_test.go

# Run specific test
go test -tags=e2e -v -run TestGitRunner_RealCommands ./test/e2e/...
```

## Test Structure

### Directory Layout

```
test/e2e/
├── README.md                          # This file
├── helpers/                           # Test helper utilities
│   ├── repo.go                        # Temporary Git repo management
│   ├── git.go                         # Git test utilities
│   ├── gh.go                          # GitHub CLI test utilities
│   ├── claude.go                      # Claude CLI detection
│   ├── cleanup.go                     # Resource cleanup helpers
│   └── helpers_test.go                # Unit tests for helpers
├── git_operations_e2e_test.go         # Git operation tests
├── workflow_planning_e2e_test.go      # Planning phase tests
├── workflow_implementation_e2e_test.go # Implementation phase tests
├── pr_split_e2e_test.go               # PR split operation tests
└── worktree_e2e_test.go               # Worktree integration tests
```

### Helper Functions

The `helpers` package provides utilities for e2e tests:

```go
import "github.com/michael-freling/claude-code-tools/test/e2e/helpers"

func TestExample(t *testing.T) {
    // Skip if git not available
    helpers.RequireGit(t)

    // Create temporary git repository
    repo := helpers.NewTempRepo(t)
    // Cleanup is automatic via t.Cleanup()

    // Create files and commits
    err := repo.CreateFile("main.go", "package main")
    require.NoError(t, err)

    err = repo.Commit("Initial commit")
    require.NoError(t, err)

    // Create branches
    err = repo.CreateBranch("feature")
    require.NoError(t, err)

    // Run git commands
    output, err := repo.RunGit("status")
    require.NoError(t, err)
}
```

### Skip Functions

```go
// Skip if git not available
helpers.RequireGit(t)

// Skip if gh not available
helpers.RequireGH(t)

// Skip if gh not authenticated
helpers.RequireGHAuth(t)

// Skip if claude not available
helpers.RequireClaude(t)

// Check claude availability without skipping
if helpers.IsCLIAvailable() {
    // claude is available
}
```

## Writing E2E Tests

### Test Template

```go
//go:build e2e

package e2e

import (
    "testing"

    "github.com/michael-freling/claude-code-tools/test/e2e/helpers"
    "github.com/stretchr/testify/require"
)

func TestMyFeature_E2E(t *testing.T) {
    helpers.RequireGit(t)
    repo := helpers.NewTempRepo(t)

    tests := []struct {
        name    string
        setup   func(t *testing.T)
        verify  func(t *testing.T)
        wantErr bool
    }{
        {
            name: "basic operation",
            setup: func(t *testing.T) {
                err := repo.CreateFile("test.txt", "content")
                require.NoError(t, err)
            },
            verify: func(t *testing.T) {
                // verify results
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.setup != nil {
                tt.setup(t)
            }
            if tt.verify != nil {
                tt.verify(t)
            }
        })
    }
}
```

### Best Practices

1. **Always use build tags**: Start every e2e test file with `//go:build e2e`
2. **Check prerequisites**: Use `helpers.Require*` functions at the start of tests
3. **Use TempRepo**: Create isolated test repositories with `helpers.NewTempRepo(t)`
4. **Automatic cleanup**: TempRepo registers cleanup via `t.Cleanup()` automatically
5. **Table-driven tests**: Use table-driven approach for better organization
6. **Descriptive names**: Use descriptive test and subtest names
7. **Independence**: Each test should be independent and not rely on other tests

## Troubleshooting

### Tests are skipped

If tests are being skipped, check that required tools are installed and in PATH:

```bash
# Check git
git --version

# Check gh
gh --version
gh auth status

# Check claude
claude --version
```

### Permission errors

If you see permission errors when creating temporary directories:

```bash
# Check temp directory permissions
ls -la /tmp

# Set custom temp directory
export TMPDIR=/path/to/writable/dir
```

### Timeout errors

If tests timeout, increase the timeout:

```bash
# Using script
E2E_TIMEOUT=5m ./scripts/run-e2e-tests.sh

# Using go test
go test -tags=e2e -timeout=5m ./test/e2e/...
```

### Git configuration issues

If git operations fail with user configuration errors:

```bash
# Set global git config
git config --global user.email "test@test.com"
git config --global user.name "Test User"
```

Note: The TempRepo helper automatically configures git user for each test repository.

## CI Integration

E2E tests run automatically in CI via GitHub Actions. The workflow:

1. Installs Go
2. Runs `./scripts/run-e2e-tests.sh`
3. Uses 1-minute timeout
4. Skips gh/claude dependent tests if tools not authenticated

See `.github/workflows/go.yml` for the CI configuration.

## Related Documentation

- [Go Testing](https://golang.org/pkg/testing/)
- [Build Tags](https://pkg.go.dev/cmd/go#hdr-Build_constraints)
- [testify](https://github.com/stretchr/testify)
