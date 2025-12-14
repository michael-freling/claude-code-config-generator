//go:build e2e

package e2e

import (
	"path/filepath"
	"testing"

	"github.com/michael-freling/claude-code-tools/internal/workflow"
	"github.com/michael-freling/claude-code-tools/test/e2e/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorktreeManager_CreateWorktree_Integration(t *testing.T) {
	helpers.RequireGit(t)
	repo := helpers.NewTempRepo(t)

	// Create an initial commit
	err := repo.CreateFile("README.md", "# Test")
	require.NoError(t, err)

	err = repo.Commit("Initial commit")
	require.NoError(t, err)

	// Create worktree manager
	wm := workflow.NewWorktreeManager(repo.Dir)

	// Test creating a worktree
	worktreePath, err := wm.CreateWorktree("test-workflow-1")
	require.NoError(t, err)

	// Verify worktree was created
	assert.True(t, wm.WorktreeExists(worktreePath))

	// Verify worktree is at expected location
	expectedPath := filepath.Join(repo.Dir, "..", "worktrees", "test-workflow-1")
	absExpected, err := filepath.Abs(expectedPath)
	require.NoError(t, err)
	assert.Equal(t, absExpected, worktreePath)

	// Test that creating the same worktree again returns existing path
	worktreePath2, err := wm.CreateWorktree("test-workflow-1")
	require.NoError(t, err)
	assert.Equal(t, worktreePath, worktreePath2)

	// Note: Worktree deletion is tested in workflow_implementation_e2e_test.go
	// as it requires careful cleanup order to work properly
}
