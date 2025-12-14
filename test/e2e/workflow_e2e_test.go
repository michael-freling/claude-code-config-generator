//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/michael-freling/claude-code-tools/internal/workflow"
	"github.com/michael-freling/claude-code-tools/test/e2e/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	sandboxRepoURL   = "https://github.com/michael-freling/claude-code-sandbox"
	sandboxRepoOwner = "michael-freling"
	sandboxRepoName  = "claude-code-sandbox"
)

// TestWorkflow_SimpleFeature_E2E tests a simple feature workflow with real Claude.
// This test uses the REAL Claude CLI to verify end-to-end workflow functionality.
// It is skipped when Claude is not available (e.g., in CI).
func TestWorkflow_SimpleFeature_E2E(t *testing.T) {
	helpers.RequireClaude(t)
	helpers.RequireGit(t)

	// Create a real temp git repo
	repo := helpers.NewTempRepo(t)
	require.NoError(t, repo.CreateFile("main.go", "package main\n\nfunc main() {\n}\n"))
	require.NoError(t, repo.Commit("Initial commit"))

	workflowName := "test-simple-feature"
	// Keep description SIMPLE to minimize Claude execution time and cost
	description := "Add a hello function that returns the string 'hello'"

	// Create config with REAL Claude CLI (no mock)
	config := workflow.DefaultConfig(repo.Dir)
	// Use generous timeouts for real Claude (can be slow)
	config.Timeouts.Planning = 5 * time.Minute
	config.Timeouts.Implementation = 5 * time.Minute
	config.Timeouts.Refactoring = 5 * time.Minute
	config.SplitPR = false
	config.LogLevel = workflow.LogLevelVerbose

	// Mock CI checker since temp repos don't have real CI
	mockCI := &mockCIChecker{
		result: &workflow.CIResult{
			Passed: true,
			Status: "success",
		},
	}

	// Create orchestrator with REAL Claude executor
	orchestrator, err := workflow.NewTestOrchestrator(config, func(workingDir string, checkInterval time.Duration, commandTimeout time.Duration) workflow.CIChecker {
		return mockCI
	})
	require.NoError(t, err)

	// Auto-confirm to avoid interactive blocking
	confirmCalled := false
	orchestrator.SetConfirmFunc(func(plan *workflow.Plan) (bool, string, error) {
		confirmCalled = true
		// Verify we got a real plan from Claude
		assert.NotEmpty(t, plan.Summary, "plan summary should not be empty")
		assert.NotEmpty(t, plan.ContextType, "plan context type should not be empty")
		t.Logf("Plan received: %s", plan.Summary)
		return true, "", nil
	})

	// Run the workflow with REAL Claude
	ctx := context.Background()
	err = orchestrator.Start(ctx, workflowName, description, workflow.WorkflowTypeFeature)
	require.NoError(t, err, "workflow should complete successfully with real Claude")

	// Verify workflow completed
	assert.True(t, confirmCalled, "confirm function should have been called")

	state, err := orchestrator.Status(workflowName)
	require.NoError(t, err)
	assert.Equal(t, workflow.PhaseCompleted, state.CurrentPhase, "workflow should reach completed phase")

	// Verify planning phase completed with real Claude
	planningPhase := state.Phases[workflow.PhasePlanning]
	assert.Equal(t, workflow.StatusCompleted, planningPhase.Status, "planning phase should complete")
	assert.Greater(t, planningPhase.Attempts, 0, "planning phase should have at least one attempt")

	// Verify implementation phase completed
	implPhase := state.Phases[workflow.PhaseImplementation]
	assert.Equal(t, workflow.StatusCompleted, implPhase.Status, "implementation phase should complete")

	// Verify refactoring phase completed
	refactorPhase := state.Phases[workflow.PhaseRefactoring]
	assert.Equal(t, workflow.StatusCompleted, refactorPhase.Status, "refactoring phase should complete")

	// Verify worktree was created
	assert.NotEmpty(t, state.WorktreePath, "worktree path should be set")

	// Verify plan was saved
	stateManager := workflow.NewStateManager(repo.Dir)
	plan, err := stateManager.LoadPlan(workflowName)
	require.NoError(t, err)
	assert.NotEmpty(t, plan.Summary, "saved plan should have summary")
	t.Logf("Final plan: %+v", plan)
}

// TestWorkflow_PlanningOnly_E2E tests only the planning phase to save time and cost.
// This allows testing Claude integration without running the full workflow.
func TestWorkflow_PlanningOnly_E2E(t *testing.T) {
	t.Skip("Planning-only test - implement if needed for faster testing")
}

type mockCIChecker struct {
	result    *workflow.CIResult
	err       error
	onCall    func(int)
	callCount int
	checkFunc func() (*workflow.CIResult, error)
}

func (m *mockCIChecker) CheckCI(ctx context.Context, prNumber int) (*workflow.CIResult, error) {
	m.callCount++
	if m.onCall != nil {
		m.onCall(m.callCount)
	}
	if m.checkFunc != nil {
		return m.checkFunc()
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func (m *mockCIChecker) WaitForCI(ctx context.Context, prNumber int, timeout time.Duration) (*workflow.CIResult, error) {
	return m.CheckCI(ctx, prNumber)
}

func (m *mockCIChecker) WaitForCIWithOptions(ctx context.Context, prNumber int, timeout time.Duration, opts workflow.CheckCIOptions) (*workflow.CIResult, error) {
	return m.CheckCI(ctx, prNumber)
}

func (m *mockCIChecker) WaitForCIWithProgress(ctx context.Context, prNumber int, timeout time.Duration, opts workflow.CheckCIOptions, onProgress workflow.CIProgressCallback) (*workflow.CIResult, error) {
	return m.CheckCI(ctx, prNumber)
}

// TestWorkflow_FeatureWorkflow_E2E tests a complete feature workflow with real CI checks
// using the sandbox repository. This test creates real commits, PRs, and waits for real CI.
func TestWorkflow_FeatureWorkflow_E2E(t *testing.T) {
	helpers.RequireClaude(t)
	helpers.RequireGit(t)
	helpers.RequireGHAuth(t)

	// Clone the sandbox repo to a temp directory
	repo := helpers.CloneRepo(t, sandboxRepoURL)

	// Create a unique branch name to avoid conflicts
	branchName := fmt.Sprintf("e2e-test-%d", time.Now().Unix())

	// Track PR number for cleanup
	var prNumber int

	// Setup cleanup to delete PR and branch
	t.Cleanup(func() {
		if prNumber > 0 {
			// Close the PR
			closeCmd := fmt.Sprintf("gh pr close %d --repo %s/%s --delete-branch", prNumber, sandboxRepoOwner, sandboxRepoName)
			t.Logf("Cleaning up PR: %s", closeCmd)
			output, err := repo.RunGit("sh", "-c", closeCmd)
			if err != nil {
				t.Logf("Warning: failed to close PR %d: %v: %s", prNumber, err, output)
			}
		}

		// Delete the branch if it still exists
		deleteCmd := fmt.Sprintf("git push origin --delete %s", branchName)
		t.Logf("Cleaning up branch: %s", deleteCmd)
		output, err := repo.RunGit("sh", "-c", deleteCmd)
		if err != nil {
			t.Logf("Warning: failed to delete branch %s: %v: %s", branchName, err, output)
		}
	})

	workflowName := "test-feature-sandbox"
	// Keep description SIMPLE to minimize Claude execution time and cost
	description := "Add a Subtract function to the calculator that takes two integers and returns their difference"

	// Create config with REAL Claude CLI and REAL CI checker
	config := workflow.DefaultConfig(repo.Dir)
	config.Timeouts.Planning = 5 * time.Minute
	config.Timeouts.Implementation = 5 * time.Minute
	config.Timeouts.Refactoring = 5 * time.Minute
	config.CICheckTimeout = 10 * time.Minute // CI can take several minutes
	config.SplitPR = false
	config.LogLevel = workflow.LogLevelVerbose

	// Create orchestrator with REAL CI checker (no mocks!)
	orchestrator, err := workflow.NewTestOrchestrator(config, nil)
	require.NoError(t, err)

	// Auto-confirm to avoid interactive blocking
	confirmCalled := false
	orchestrator.SetConfirmFunc(func(plan *workflow.Plan) (bool, string, error) {
		confirmCalled = true
		assert.NotEmpty(t, plan.Summary, "plan summary should not be empty")
		assert.NotEmpty(t, plan.ContextType, "plan context type should not be empty")
		t.Logf("Plan received: %s", plan.Summary)
		return true, "", nil
	})

	// Run the workflow with REAL Claude and REAL CI
	ctx := context.Background()
	err = orchestrator.Start(ctx, workflowName, description, workflow.WorkflowTypeFeature)

	// Get workflow state
	state, statusErr := orchestrator.Status(workflowName)
	require.NoError(t, statusErr)

	// Get PR number from worktree using gh CLI
	if state.WorktreePath != "" {
		prListOutput, ghErr := repo.RunGit("sh", "-c", fmt.Sprintf("cd %s && gh pr list --head $(git rev-parse --abbrev-ref HEAD) --json number --jq '.[0].number'", state.WorktreePath))
		if ghErr == nil && prListOutput != "" {
			fmt.Sscanf(prListOutput, "%d", &prNumber)
			if prNumber > 0 {
				t.Logf("Found PR #%d", prNumber)
			}
		}
	}

	// Verify workflow completed or failed (CI may pass or fail)
	if err != nil {
		t.Logf("Workflow error: %v", err)
		// Check if workflow reached a terminal state
		if state.CurrentPhase != workflow.PhaseCompleted && state.CurrentPhase != workflow.PhaseFailed {
			require.NoError(t, err, "workflow should reach completion or failure state")
		}
		t.Logf("Workflow ended in phase: %s (this is acceptable for E2E test validation)", state.CurrentPhase)
	}

	// Verify workflow execution
	assert.True(t, confirmCalled, "confirm function should have been called")

	state, err = orchestrator.Status(workflowName)
	require.NoError(t, err)

	// Verify planning phase completed
	planningPhase := state.Phases[workflow.PhasePlanning]
	assert.Equal(t, workflow.StatusCompleted, planningPhase.Status, "planning phase should complete")
	assert.Greater(t, planningPhase.Attempts, 0, "planning phase should have at least one attempt")

	// Verify implementation phase completed
	implPhase := state.Phases[workflow.PhaseImplementation]
	assert.Equal(t, workflow.StatusCompleted, implPhase.Status, "implementation phase should complete")

	// Verify refactoring phase completed
	refactorPhase := state.Phases[workflow.PhaseRefactoring]
	assert.Equal(t, workflow.StatusCompleted, refactorPhase.Status, "refactoring phase should complete")

	// Verify PR was created
	assert.Greater(t, prNumber, 0, "PR should be created")

	// Verify worktree was created
	assert.NotEmpty(t, state.WorktreePath, "worktree path should be set")

	// Verify plan was saved
	stateManager := workflow.NewStateManager(repo.Dir)
	plan, err := stateManager.LoadPlan(workflowName)
	require.NoError(t, err)
	assert.NotEmpty(t, plan.Summary, "saved plan should have summary")
	t.Logf("Final plan: %+v", plan)

	// Log final state
	t.Logf("Workflow final phase: %s", state.CurrentPhase)
	if prNumber > 0 {
		t.Logf("PR URL: https://github.com/%s/%s/pull/%d", sandboxRepoOwner, sandboxRepoName, prNumber)
	}
}
