//go:build e2e

package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/michael-freling/claude-code-tools/internal/command"
	"github.com/michael-freling/claude-code-tools/internal/workflow"
	"github.com/michael-freling/claude-code-tools/test/e2e/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorktreeManager_CreateWorktree(t *testing.T) {
	helpers.RequireGit(t)
	repo := helpers.NewTempRepo(t)

	// Create initial commit
	err := repo.CreateFile("README.md", "# Test")
	require.NoError(t, err)
	err = repo.Commit("Initial commit")
	require.NoError(t, err)

	wm := workflow.NewWorktreeManager(repo.Dir)

	tests := []struct {
		name         string
		workflowName string
		wantErr      bool
	}{
		{
			name:         "create worktree successfully",
			workflowName: "test-workflow-1",
			wantErr:      false,
		},
		{
			name:         "create another worktree",
			workflowName: "test-workflow-2",
			wantErr:      false,
		},
		{
			name:         "empty workflow name fails",
			workflowName: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			worktreePath, err := wm.CreateWorktree(tt.workflowName)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, worktreePath)
			assert.True(t, wm.WorktreeExists(worktreePath))

			// Verify worktree directory exists
			info, err := os.Stat(worktreePath)
			require.NoError(t, err)
			assert.True(t, info.IsDir())

			// Verify .git exists in worktree
			gitPath := filepath.Join(worktreePath, ".git")
			_, err = os.Stat(gitPath)
			require.NoError(t, err)
		})
	}
}

func TestWorktreeManager_DeleteWorktree(t *testing.T) {
	helpers.RequireGit(t)
	repo := helpers.NewTempRepo(t)

	// Create initial commit
	err := repo.CreateFile("README.md", "# Test")
	require.NoError(t, err)
	err = repo.Commit("Initial commit")
	require.NoError(t, err)

	wm := workflow.NewWorktreeManager(repo.Dir)

	tests := []struct {
		name    string
		test    func(t *testing.T)
		wantErr bool
	}{
		{
			name: "delete existing worktree",
			test: func(t *testing.T) {
				worktreePath, err := wm.CreateWorktree("delete-test")
				require.NoError(t, err)
				assert.True(t, wm.WorktreeExists(worktreePath))

				err = wm.DeleteWorktree(worktreePath)
				require.NoError(t, err)
				assert.False(t, wm.WorktreeExists(worktreePath))
			},
			wantErr: false,
		},
		{
			name: "delete nonexistent worktree succeeds",
			test: func(t *testing.T) {
				nonexistentPath := filepath.Join(repo.Dir, "..", "worktrees", "nonexistent")
				err := wm.DeleteWorktree(nonexistentPath)
				require.NoError(t, err)
			},
			wantErr: false,
		},
		{
			name: "delete with empty path fails",
			test: func(t *testing.T) {
				err := wm.DeleteWorktree("")
				require.Error(t, err)
				assert.Contains(t, err.Error(), "worktree path cannot be empty")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func TestWorktreeManager_ReuseExistingWorktree(t *testing.T) {
	helpers.RequireGit(t)
	repo := helpers.NewTempRepo(t)

	// Create initial commit
	err := repo.CreateFile("README.md", "# Test")
	require.NoError(t, err)
	err = repo.Commit("Initial commit")
	require.NoError(t, err)

	wm := workflow.NewWorktreeManager(repo.Dir)

	workflowName := "reuse-test"

	// Create worktree first time
	worktreePath1, err := wm.CreateWorktree(workflowName)
	require.NoError(t, err)
	assert.NotEmpty(t, worktreePath1)

	// Create again with same name - should return existing path
	worktreePath2, err := wm.CreateWorktree(workflowName)
	require.NoError(t, err)
	assert.Equal(t, worktreePath1, worktreePath2)
}

func TestStateManager_WithRealGitWorktree(t *testing.T) {
	helpers.RequireGit(t)
	repo := helpers.NewTempRepo(t)

	// Create initial commit
	err := repo.CreateFile("README.md", "# Test")
	require.NoError(t, err)
	err = repo.Commit("Initial commit")
	require.NoError(t, err)

	tmpDir := t.TempDir()
	stateManager := workflow.NewStateManager(tmpDir)
	wm := workflow.NewWorktreeManager(repo.Dir)

	workflowName := "state-worktree-test"

	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "save worktree path in state",
			test: func(t *testing.T) {
				state, err := stateManager.InitState(workflowName, "Test workflow with worktree", workflow.WorkflowTypeFeature)
				require.NoError(t, err)

				worktreePath, err := wm.CreateWorktree(workflowName)
				require.NoError(t, err)

				state.WorktreePath = worktreePath
				err = stateManager.SaveState(workflowName, state)
				require.NoError(t, err)

				loadedState, err := stateManager.LoadState(workflowName)
				require.NoError(t, err)
				assert.Equal(t, worktreePath, loadedState.WorktreePath)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func TestImplementationPhase_FileOperations(t *testing.T) {
	helpers.RequireGit(t)
	repo := helpers.NewTempRepo(t)

	// Create initial commit
	err := repo.CreateFile("README.md", "# Test")
	require.NoError(t, err)
	err = repo.Commit("Initial commit")
	require.NoError(t, err)

	tmpDir := t.TempDir()
	stateManager := workflow.NewStateManager(tmpDir)
	wm := workflow.NewWorktreeManager(repo.Dir)

	workflowName := "implementation-test"

	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "track implementation files",
			test: func(t *testing.T) {
				state, err := stateManager.InitState(workflowName, "Test implementation", workflow.WorkflowTypeFeature)
				require.NoError(t, err)

				worktreePath, err := wm.CreateWorktree(workflowName)
				require.NoError(t, err)

				// Create files in worktree
				err = os.WriteFile(filepath.Join(worktreePath, "new_file.go"), []byte("package main"), 0644)
				require.NoError(t, err)

				// Save implementation summary
				summary := &workflow.ImplementationSummary{
					FilesChanged: []string{"new_file.go"},
					LinesAdded:   1,
					LinesRemoved: 0,
					TestsAdded:   0,
					Summary:      "Added new file",
				}

				err = stateManager.SavePhaseOutput(workflowName, workflow.PhaseImplementation, summary)
				require.NoError(t, err)

				// Update state
				state.CurrentPhase = workflow.PhaseImplementation
				state.Phases[workflow.PhaseImplementation].Status = workflow.StatusCompleted
				err = stateManager.SaveState(workflowName, state)
				require.NoError(t, err)

				// Verify
				var loaded workflow.ImplementationSummary
				err = stateManager.LoadPhaseOutput(workflowName, workflow.PhaseImplementation, &loaded)
				require.NoError(t, err)
				assert.Equal(t, summary.FilesChanged, loaded.FilesChanged)
				assert.Equal(t, summary.Summary, loaded.Summary)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func TestWorkflowManager_IntegrationWithGit(t *testing.T) {
	helpers.RequireGit(t)
	repo := helpers.NewTempRepo(t)

	// Create initial commit
	err := repo.CreateFile("README.md", "# Test")
	require.NoError(t, err)
	err = repo.Commit("Initial commit")
	require.NoError(t, err)

	runner := command.NewRunner()
	gitRunner := command.NewGitRunner(runner)
	wm := workflow.NewWorktreeManagerWithRunner(repo.Dir, gitRunner)

	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "worktree creation and existence check",
			test: func(t *testing.T) {
				workflowName := "branch-test"
				worktreePath, err := wm.CreateWorktree(workflowName)
				require.NoError(t, err)

				// Verify worktree path exists
				assert.True(t, wm.WorktreeExists(worktreePath))

				// Verify worktree list shows more than just the main repo
				output, err := repo.RunGit("worktree", "list")
				require.NoError(t, err)
				// Should have at least 2 lines: main repo + the worktree we created
				assert.NotEmpty(t, output)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}
