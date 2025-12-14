//go:build e2e

package e2e

import (
	"context"
	"testing"

	"github.com/michael-freling/claude-code-tools/internal/command"
	"github.com/michael-freling/claude-code-tools/internal/workflow"
	"github.com/michael-freling/claude-code-tools/test/e2e/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPRSplitBranchOperations(t *testing.T) {
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
		name string
		test func(t *testing.T)
	}{
		{
			name: "create parent and child branches",
			test: func(t *testing.T) {
				// Create parent branch
				parentBranch := "parent-pr-branch"
				err := gitRunner.CreateBranch(context.Background(), repo.Dir, parentBranch, baseBranch)
				require.NoError(t, err)

				// Create empty commit on parent
				err = gitRunner.CommitEmpty(context.Background(), repo.Dir, "Parent PR placeholder")
				require.NoError(t, err)

				// Create child branch from parent
				childBranch := "child-pr-branch-1"
				err = gitRunner.CreateBranch(context.Background(), repo.Dir, childBranch, parentBranch)
				require.NoError(t, err)

				// Verify child branch exists
				currentBranch, err := gitRunner.GetCurrentBranch(context.Background(), repo.Dir)
				require.NoError(t, err)
				assert.Equal(t, childBranch, currentBranch)
			},
		},
		{
			name: "cherry-pick commits for child PR",
			test: func(t *testing.T) {
				// Create source branch with commits
				sourceBranch := "source-branch"
				err := gitRunner.CreateBranch(context.Background(), repo.Dir, sourceBranch, baseBranch)
				require.NoError(t, err)

				// Add commits
				err = repo.CreateFile("file1.txt", "content1")
				require.NoError(t, err)
				err = repo.Commit("Add file1")
				require.NoError(t, err)

				err = repo.CreateFile("file2.txt", "content2")
				require.NoError(t, err)
				err = repo.Commit("Add file2")
				require.NoError(t, err)

				// Get commits
				commits, err := gitRunner.GetCommits(context.Background(), repo.Dir, baseBranch)
				require.NoError(t, err)
				require.Len(t, commits, 2)

				// Create parent branch
				parentBranch := "parent-cherry"
				err = gitRunner.CreateBranch(context.Background(), repo.Dir, parentBranch, baseBranch)
				require.NoError(t, err)

				err = gitRunner.CommitEmpty(context.Background(), repo.Dir, "Parent PR")
				require.NoError(t, err)

				// Create child branch and cherry-pick first commit
				childBranch := "child-cherry-1"
				err = gitRunner.CreateBranch(context.Background(), repo.Dir, childBranch, parentBranch)
				require.NoError(t, err)

				err = gitRunner.CherryPick(context.Background(), repo.Dir, commits[0].Hash)
				require.NoError(t, err)

				// Verify commit was cherry-picked
				childCommits, err := gitRunner.GetCommits(context.Background(), repo.Dir, parentBranch)
				require.NoError(t, err)
				assert.Len(t, childCommits, 1)
				assert.Equal(t, "Add file1", childCommits[0].Subject)
			},
		},
		{
			name: "checkout files for child PR",
			test: func(t *testing.T) {
				// Create source branch with files
				sourceBranch := "source-files"
				err := gitRunner.CreateBranch(context.Background(), repo.Dir, sourceBranch, baseBranch)
				require.NoError(t, err)

				err = repo.CreateFile("feature1.go", "package feature1")
				require.NoError(t, err)
				err = repo.CreateFile("feature2.go", "package feature2")
				require.NoError(t, err)
				err = repo.Commit("Add features")
				require.NoError(t, err)

				// Create parent branch
				parentBranch := "parent-files"
				err = gitRunner.CreateBranch(context.Background(), repo.Dir, parentBranch, baseBranch)
				require.NoError(t, err)

				err = gitRunner.CommitEmpty(context.Background(), repo.Dir, "Parent PR")
				require.NoError(t, err)

				// Create child branch
				childBranch := "child-files-1"
				err = gitRunner.CreateBranch(context.Background(), repo.Dir, childBranch, parentBranch)
				require.NoError(t, err)

				// Checkout specific file from source
				err = gitRunner.CheckoutFiles(context.Background(), repo.Dir, sourceBranch, []string{"feature1.go"})
				require.NoError(t, err)

				// Commit the file
				err = gitRunner.CommitAll(context.Background(), repo.Dir, "Add feature1")
				require.NoError(t, err)

				// Verify only feature1.go was added
				childCommits, err := gitRunner.GetCommits(context.Background(), repo.Dir, parentBranch)
				require.NoError(t, err)
				assert.Len(t, childCommits, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func TestPRSplitPlan_Structure(t *testing.T) {
	helpers.RequireGit(t)

	tests := []struct {
		name    string
		plan    *workflow.PRSplitPlan
		wantErr bool
	}{
		{
			name: "valid commit-based split plan",
			plan: &workflow.PRSplitPlan{
				Strategy:    workflow.SplitByCommits,
				ParentTitle: "Parent PR",
				ParentDesc:  "Parent description",
				ChildPRs: []workflow.ChildPRPlan{
					{
						Title:       "Child PR 1",
						Description: "First child",
						Commits:     []string{"commit1", "commit2"},
					},
					{
						Title:       "Child PR 2",
						Description: "Second child",
						Commits:     []string{"commit3"},
					},
				},
				Summary: "Split into 2 child PRs",
			},
			wantErr: false,
		},
		{
			name: "valid file-based split plan",
			plan: &workflow.PRSplitPlan{
				Strategy:    workflow.SplitByFiles,
				ParentTitle: "Parent PR",
				ParentDesc:  "Parent description",
				ChildPRs: []workflow.ChildPRPlan{
					{
						Title:       "Child PR 1",
						Description: "First child",
						Files:       []string{"file1.go", "file2.go"},
					},
					{
						Title:       "Child PR 2",
						Description: "Second child",
						Files:       []string{"file3.go"},
					},
				},
				Summary: "Split into 2 child PRs by files",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.plan)
			assert.NotEmpty(t, tt.plan.ParentTitle)
			assert.NotEmpty(t, tt.plan.ChildPRs)

			if tt.plan.Strategy == workflow.SplitByCommits {
				for _, child := range tt.plan.ChildPRs {
					assert.NotEmpty(t, child.Commits, "commit-based plan should have commits")
				}
			}

			if tt.plan.Strategy == workflow.SplitByFiles {
				for _, child := range tt.plan.ChildPRs {
					assert.NotEmpty(t, child.Files, "file-based plan should have files")
				}
			}
		})
	}
}

func TestStateManager_PRSplitPersistence(t *testing.T) {
	helpers.RequireGit(t)

	tmpDir := t.TempDir()
	stateManager := workflow.NewStateManager(tmpDir)

	workflowName := "pr-split-test"
	_, err := stateManager.InitState(workflowName, "Test PR split", workflow.WorkflowTypeFeature)
	require.NoError(t, err)

	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "save and load PR split result",
			test: func(t *testing.T) {
				result := &workflow.PRSplitResult{
					ParentPR: workflow.PRInfo{
						Number:      1,
						URL:         "https://github.com/org/repo/pull/1",
						Title:       "Parent PR",
						Description: "Parent description",
					},
					ChildPRs: []workflow.PRInfo{
						{
							Number:      2,
							URL:         "https://github.com/org/repo/pull/2",
							Title:       "Child PR 1",
							Description: "First child",
						},
						{
							Number:      3,
							URL:         "https://github.com/org/repo/pull/3",
							Title:       "Child PR 2",
							Description: "Second child",
						},
					},
					Summary:     "Split into 2 child PRs",
					BranchNames: []string{"parent-branch", "child-branch-1", "child-branch-2"},
				}

				err := stateManager.SavePhaseOutput(workflowName, workflow.PhasePRSplit, result)
				require.NoError(t, err)

				var loaded workflow.PRSplitResult
				err = stateManager.LoadPhaseOutput(workflowName, workflow.PhasePRSplit, &loaded)
				require.NoError(t, err)

				assert.Equal(t, result.ParentPR.Number, loaded.ParentPR.Number)
				assert.Equal(t, result.ParentPR.URL, loaded.ParentPR.URL)
				assert.Len(t, loaded.ChildPRs, 2)
				assert.Equal(t, result.ChildPRs[0].Number, loaded.ChildPRs[0].Number)
				assert.Equal(t, result.Summary, loaded.Summary)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func TestGitRunner_PRSplitBranchCleanup(t *testing.T) {
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
		name string
		test func(t *testing.T)
	}{
		{
			name: "delete child branches after split",
			test: func(t *testing.T) {
				// Create parent and child branches
				parentBranch := "cleanup-parent"
				err := gitRunner.CreateBranch(context.Background(), repo.Dir, parentBranch, baseBranch)
				require.NoError(t, err)

				childBranch1 := "cleanup-child-1"
				err = gitRunner.CreateBranch(context.Background(), repo.Dir, childBranch1, parentBranch)
				require.NoError(t, err)

				childBranch2 := "cleanup-child-2"
				err = gitRunner.CreateBranch(context.Background(), repo.Dir, childBranch2, parentBranch)
				require.NoError(t, err)

				// Go back to base branch
				err = gitRunner.CheckoutBranch(context.Background(), repo.Dir, baseBranch)
				require.NoError(t, err)

				// Delete child branches
				err = gitRunner.DeleteBranch(context.Background(), repo.Dir, childBranch1, false)
				require.NoError(t, err)

				err = gitRunner.DeleteBranch(context.Background(), repo.Dir, childBranch2, false)
				require.NoError(t, err)

				// Delete parent branch
				err = gitRunner.DeleteBranch(context.Background(), repo.Dir, parentBranch, false)
				require.NoError(t, err)

				// Verify branches are deleted
				output, err := repo.RunGit("branch")
				require.NoError(t, err)
				assert.NotContains(t, output, childBranch1)
				assert.NotContains(t, output, childBranch2)
				assert.NotContains(t, output, parentBranch)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}
