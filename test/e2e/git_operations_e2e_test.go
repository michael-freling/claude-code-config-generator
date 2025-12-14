//go:build e2e

package e2e

import (
	"context"
	"testing"

	"github.com/michael-freling/claude-code-tools/internal/command"
	"github.com/michael-freling/claude-code-tools/test/e2e/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitRunner_GetCurrentBranch(t *testing.T) {
	helpers.RequireGit(t)
	repo := helpers.NewTempRepo(t)

	// Create initial commit so we have a valid branch
	err := repo.CreateFile("README.md", "# Test")
	require.NoError(t, err)
	err = repo.Commit("Initial commit")
	require.NoError(t, err)

	runner := command.NewRunner()
	gitRunner := command.NewGitRunner(runner)

	tests := []struct {
		name       string
		setup      func(t *testing.T)
		wantBranch string
		wantErr    bool
	}{
		{
			name:       "default branch is main or master",
			setup:      func(t *testing.T) {},
			wantBranch: "", // Will check if it's either main or master
			wantErr:    false,
		},
		{
			name: "custom branch",
			setup: func(t *testing.T) {
				err := repo.CreateBranch("feature-branch")
				require.NoError(t, err)
			},
			wantBranch: "feature-branch",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)

			got, err := gitRunner.GetCurrentBranch(context.Background(), repo.Dir)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.wantBranch != "" {
				assert.Equal(t, tt.wantBranch, got)
			} else {
				// Default branch should be either main or master
				assert.Contains(t, []string{"main", "master"}, got)
			}
		})
	}
}

func TestGitRunner_CommitOperations(t *testing.T) {
	helpers.RequireGit(t)
	repo := helpers.NewTempRepo(t)

	runner := command.NewRunner()
	gitRunner := command.NewGitRunner(runner)

	tests := []struct {
		name    string
		test    func(t *testing.T)
		wantErr bool
	}{
		{
			name: "commit all stages and commits files",
			test: func(t *testing.T) {
				err := repo.CreateFile("test.txt", "test content")
				require.NoError(t, err)

				err = gitRunner.CommitAll(context.Background(), repo.Dir, "test commit")
				require.NoError(t, err)

				// Verify commit exists
				output, err := repo.RunGit("log", "--oneline", "-1")
				require.NoError(t, err)
				assert.Contains(t, output, "test commit")
			},
			wantErr: false,
		},
		{
			name: "commit empty creates empty commit",
			test: func(t *testing.T) {
				err := gitRunner.CommitEmpty(context.Background(), repo.Dir, "empty commit")
				require.NoError(t, err)

				output, err := repo.RunGit("log", "--oneline", "-1")
				require.NoError(t, err)
				assert.Contains(t, output, "empty commit")
			},
			wantErr: false,
		},
		{
			name: "commit empty with empty message fails",
			test: func(t *testing.T) {
				err := gitRunner.CommitEmpty(context.Background(), repo.Dir, "")
				require.Error(t, err)
				assert.Contains(t, err.Error(), "commit message cannot be empty")
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

func TestGitRunner_BranchOperations(t *testing.T) {
	helpers.RequireGit(t)
	repo := helpers.NewTempRepo(t)

	// Create initial commit
	err := repo.CreateFile("README.md", "# Test")
	require.NoError(t, err)
	err = repo.Commit("Initial commit")
	require.NoError(t, err)

	runner := command.NewRunner()
	gitRunner := command.NewGitRunner(runner)

	tests := []struct {
		name    string
		test    func(t *testing.T)
		wantErr bool
	}{
		{
			name: "create branch from base",
			test: func(t *testing.T) {
				currentBranch, err := gitRunner.GetCurrentBranch(context.Background(), repo.Dir)
				require.NoError(t, err)

				err = gitRunner.CreateBranch(context.Background(), repo.Dir, "new-branch", currentBranch)
				require.NoError(t, err)

				branch, err := gitRunner.GetCurrentBranch(context.Background(), repo.Dir)
				require.NoError(t, err)
				assert.Equal(t, "new-branch", branch)
			},
			wantErr: false,
		},
		{
			name: "checkout existing branch",
			test: func(t *testing.T) {
				currentBranch, err := gitRunner.GetCurrentBranch(context.Background(), repo.Dir)
				require.NoError(t, err)

				// Create and checkout a new branch
				err = gitRunner.CreateBranch(context.Background(), repo.Dir, "checkout-test", currentBranch)
				require.NoError(t, err)

				// Go back to original branch
				err = gitRunner.CheckoutBranch(context.Background(), repo.Dir, currentBranch)
				require.NoError(t, err)

				branch, err := gitRunner.GetCurrentBranch(context.Background(), repo.Dir)
				require.NoError(t, err)
				assert.Equal(t, currentBranch, branch)
			},
			wantErr: false,
		},
		{
			name: "delete branch",
			test: func(t *testing.T) {
				currentBranch, err := gitRunner.GetCurrentBranch(context.Background(), repo.Dir)
				require.NoError(t, err)

				// Create a branch to delete
				err = gitRunner.CreateBranch(context.Background(), repo.Dir, "delete-me", currentBranch)
				require.NoError(t, err)

				// Go back to original branch
				err = gitRunner.CheckoutBranch(context.Background(), repo.Dir, currentBranch)
				require.NoError(t, err)

				// Delete the branch
				err = gitRunner.DeleteBranch(context.Background(), repo.Dir, "delete-me", false)
				require.NoError(t, err)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func TestGitRunner_WorktreeOperations(t *testing.T) {
	helpers.RequireGit(t)
	repo := helpers.NewTempRepo(t)

	// Create initial commit
	err := repo.CreateFile("README.md", "# Test")
	require.NoError(t, err)
	err = repo.Commit("Initial commit")
	require.NoError(t, err)

	runner := command.NewRunner()
	gitRunner := command.NewGitRunner(runner)

	tests := []struct {
		name    string
		test    func(t *testing.T)
		wantErr bool
	}{
		{
			name: "add worktree",
			test: func(t *testing.T) {
				worktreePath := t.TempDir()
				err := gitRunner.WorktreeAdd(context.Background(), repo.Dir, worktreePath, "worktree-branch")
				require.NoError(t, err)

				// Verify worktree was created
				output, err := repo.RunGit("worktree", "list")
				require.NoError(t, err)
				assert.Contains(t, output, worktreePath)
			},
			wantErr: false,
		},
		{
			name: "remove worktree",
			test: func(t *testing.T) {
				worktreePath := t.TempDir()
				err := gitRunner.WorktreeAdd(context.Background(), repo.Dir, worktreePath, "remove-worktree")
				require.NoError(t, err)

				err = gitRunner.WorktreeRemove(context.Background(), repo.Dir, worktreePath)
				require.NoError(t, err)

				// Verify worktree was removed
				output, err := repo.RunGit("worktree", "list")
				require.NoError(t, err)
				assert.NotContains(t, output, worktreePath)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func TestGitRunner_GetCommits(t *testing.T) {
	helpers.RequireGit(t)
	repo := helpers.NewTempRepo(t)

	// Create initial commit
	err := repo.CreateFile("README.md", "# Test")
	require.NoError(t, err)
	err = repo.Commit("Initial commit")
	require.NoError(t, err)

	runner := command.NewRunner()
	gitRunner := command.NewGitRunner(runner)

	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		want    int
		wantErr bool
	}{
		{
			name: "no commits after base",
			setup: func(t *testing.T) string {
				branch, err := gitRunner.GetCurrentBranch(context.Background(), repo.Dir)
				require.NoError(t, err)
				return branch
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "multiple commits after base",
			setup: func(t *testing.T) string {
				baseBranch, err := gitRunner.GetCurrentBranch(context.Background(), repo.Dir)
				require.NoError(t, err)

				err = gitRunner.CreateBranch(context.Background(), repo.Dir, "feature", baseBranch)
				require.NoError(t, err)

				err = repo.CreateFile("file1.txt", "content1")
				require.NoError(t, err)
				err = repo.Commit("commit 1")
				require.NoError(t, err)

				err = repo.CreateFile("file2.txt", "content2")
				require.NoError(t, err)
				err = repo.Commit("commit 2")
				require.NoError(t, err)

				return baseBranch
			},
			want:    2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseBranch := tt.setup(t)

			got, err := gitRunner.GetCommits(context.Background(), repo.Dir, baseBranch)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, got, tt.want)
		})
	}
}

func TestGitRunner_CherryPick(t *testing.T) {
	helpers.RequireGit(t)
	repo := helpers.NewTempRepo(t)

	// Create initial commit
	err := repo.CreateFile("README.md", "# Test")
	require.NoError(t, err)
	err = repo.Commit("Initial commit")
	require.NoError(t, err)

	runner := command.NewRunner()
	gitRunner := command.NewGitRunner(runner)

	baseBranch, err := gitRunner.GetCurrentBranch(context.Background(), repo.Dir)
	require.NoError(t, err)

	// Create a commit to cherry-pick
	err = gitRunner.CreateBranch(context.Background(), repo.Dir, "source", baseBranch)
	require.NoError(t, err)

	err = repo.CreateFile("cherry.txt", "cherry pick me")
	require.NoError(t, err)
	err = repo.Commit("Cherry pick commit")
	require.NoError(t, err)

	commits, err := gitRunner.GetCommits(context.Background(), repo.Dir, baseBranch)
	require.NoError(t, err)
	require.Len(t, commits, 1)

	commitHash := commits[0].Hash

	// Go back to base and create target branch
	err = gitRunner.CheckoutBranch(context.Background(), repo.Dir, baseBranch)
	require.NoError(t, err)

	err = gitRunner.CreateBranch(context.Background(), repo.Dir, "target", baseBranch)
	require.NoError(t, err)

	// Cherry-pick the commit
	err = gitRunner.CherryPick(context.Background(), repo.Dir, commitHash)
	require.NoError(t, err)

	// Verify cherry-pick worked
	commits, err = gitRunner.GetCommits(context.Background(), repo.Dir, baseBranch)
	require.NoError(t, err)
	assert.Len(t, commits, 1)
	assert.Equal(t, "Cherry pick commit", commits[0].Subject)
}

func TestGitRunner_GetDiffStat(t *testing.T) {
	helpers.RequireGit(t)
	repo := helpers.NewTempRepo(t)

	// Create initial commit
	err := repo.CreateFile("README.md", "# Test")
	require.NoError(t, err)
	err = repo.Commit("Initial commit")
	require.NoError(t, err)

	runner := command.NewRunner()
	gitRunner := command.NewGitRunner(runner)

	baseBranch, err := gitRunner.GetCurrentBranch(context.Background(), repo.Dir)
	require.NoError(t, err)

	tests := []struct {
		name    string
		setup   func(t *testing.T)
		wantErr bool
	}{
		{
			name:    "no changes",
			setup:   func(t *testing.T) {},
			wantErr: false,
		},
		{
			name: "with changes",
			setup: func(t *testing.T) {
				err := gitRunner.CreateBranch(context.Background(), repo.Dir, "changes", baseBranch)
				require.NoError(t, err)

				err = repo.CreateFile("newfile.txt", "new content")
				require.NoError(t, err)
				err = repo.Commit("Add new file")
				require.NoError(t, err)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)

			got, err := gitRunner.GetDiffStat(context.Background(), repo.Dir, baseBranch)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, got)
		})
	}
}

func TestGitRunner_CheckoutFiles(t *testing.T) {
	helpers.RequireGit(t)
	repo := helpers.NewTempRepo(t)

	// Create initial commit
	err := repo.CreateFile("README.md", "# Test")
	require.NoError(t, err)
	err = repo.Commit("Initial commit")
	require.NoError(t, err)

	runner := command.NewRunner()
	gitRunner := command.NewGitRunner(runner)

	baseBranch, err := gitRunner.GetCurrentBranch(context.Background(), repo.Dir)
	require.NoError(t, err)

	// Create source branch with file
	err = gitRunner.CreateBranch(context.Background(), repo.Dir, "source", baseBranch)
	require.NoError(t, err)

	err = repo.CreateFile("checkout.txt", "checkout this file")
	require.NoError(t, err)
	err = repo.Commit("Add checkout file")
	require.NoError(t, err)

	// Go back to base branch
	err = gitRunner.CheckoutBranch(context.Background(), repo.Dir, baseBranch)
	require.NoError(t, err)

	// Checkout specific files from source branch
	err = gitRunner.CheckoutFiles(context.Background(), repo.Dir, "source", []string{"checkout.txt"})
	require.NoError(t, err)

	// Verify file exists
	output, err := repo.RunGit("status", "--short")
	require.NoError(t, err)
	assert.Contains(t, output, "checkout.txt")
}
