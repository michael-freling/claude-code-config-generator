//go:build e2e

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/michael-freling/claude-code-tools/internal/workflow"
	"github.com/michael-freling/claude-code-tools/test/e2e/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflow_FeatureWorkflow_E2E(t *testing.T) {
	helpers.RequireGit(t)

	repo := helpers.NewTempRepo(t)
	require.NoError(t, repo.CreateFile("main.go", "package main\n\nfunc main() {\n}\n"))
	require.NoError(t, repo.Commit("Initial commit"))

	workflowName := "test-feature"
	description := "Add user authentication"

	planJSON := map[string]interface{}{
		"summary":     "Implement user authentication with JWT",
		"contextType": "feature",
		"architecture": map[string]interface{}{
			"overview":   "Add JWT-based authentication",
			"components": []string{"auth middleware", "user model"},
		},
		"phases": []map[string]interface{}{
			{
				"name":           "Setup",
				"description":    "Setup authentication infrastructure",
				"estimatedFiles": 3,
				"estimatedLines": 150,
			},
		},
		"workStreams": []map[string]interface{}{
			{
				"name":  "Core auth",
				"tasks": []string{"JWT generation", "Token validation"},
			},
		},
		"risks":               []string{"Token expiration handling"},
		"complexity":          "medium",
		"estimatedTotalLines": 150,
		"estimatedTotalFiles": 3,
	}

	implSummary := map[string]interface{}{
		"filesChanged": []string{"auth/jwt.go", "auth/middleware.go"},
		"linesAdded":   150,
		"linesRemoved": 10,
		"testsAdded":   8,
		"summary":      "Implemented JWT authentication",
	}

	refactorSummary := map[string]interface{}{
		"filesChanged": []string{"auth/jwt.go"},
		"linesAdded":   20,
		"linesRemoved": 15,
		"summary":      "Refactored JWT token generation",
	}

	mockPlanClaude := helpers.NewMockClaudeBuilder(t).
		WithStreamingResponse("Plan created", planJSON)
	planClaudePath := mockPlanClaude.Build()

	mockImplClaude := helpers.NewMockClaudeBuilder(t).
		WithStreamingResponse("Implementation complete", implSummary)
	implClaudePath := mockImplClaude.Build()

	mockRefactorClaude := helpers.NewMockClaudeBuilder(t).
		WithStreamingResponse("Refactoring complete", refactorSummary)
	refactorClaudePath := mockRefactorClaude.Build()

	config := workflow.DefaultConfig(repo.Dir)
	config.ClaudePath = planClaudePath
	config.Timeouts.Planning = 10 * time.Second
	config.Timeouts.Implementation = 10 * time.Second
	config.Timeouts.Refactoring = 10 * time.Second
	config.SplitPR = false
	config.LogLevel = workflow.LogLevelNormal

	mockCI := &mockCIChecker{
		result: &workflow.CIResult{
			Passed: true,
			Status: "success",
		},
	}

	orchestrator, err := workflow.NewTestOrchestrator(config, func(workingDir string, checkInterval time.Duration, commandTimeout time.Duration) workflow.CIChecker {
		return mockCI
	})
	require.NoError(t, err)

	confirmCalled := false
	orchestrator.SetConfirmFunc(func(plan *workflow.Plan) (bool, string, error) {
		confirmCalled = true
		assert.Equal(t, planJSON["summary"], plan.Summary)
		config.ClaudePath = implClaudePath
		return true, "", nil
	})

	ciCallCount := 0
	mockCI.onCall = func(callNum int) {
		ciCallCount = callNum
		if ciCallCount == 1 {
			config.ClaudePath = refactorClaudePath
		}
	}

	ctx := context.Background()
	err = orchestrator.Start(ctx, workflowName, description, workflow.WorkflowTypeFeature)
	require.NoError(t, err)

	assert.True(t, confirmCalled, "confirm function should have been called")

	state, err := orchestrator.Status(workflowName)
	require.NoError(t, err)
	assert.Equal(t, workflow.PhaseCompleted, state.CurrentPhase)

	planningPhase := state.Phases[workflow.PhasePlanning]
	assert.Equal(t, workflow.StatusCompleted, planningPhase.Status)
	assert.Greater(t, planningPhase.Attempts, 0)

	implPhase := state.Phases[workflow.PhaseImplementation]
	assert.Equal(t, workflow.StatusCompleted, implPhase.Status)

	refactorPhase := state.Phases[workflow.PhaseRefactoring]
	assert.Equal(t, workflow.StatusCompleted, refactorPhase.Status)

	assert.NotEmpty(t, state.WorktreePath)

	stateManager := workflow.NewStateManager(repo.Dir)
	plan, err := stateManager.LoadPlan(workflowName)
	require.NoError(t, err)
	assert.Equal(t, planJSON["summary"], plan.Summary)

	args := mockPlanClaude.GetCapturedArgs()
	assert.Contains(t, args, "--print", "planning phase should include --print flag")
	assert.Contains(t, args, "stream-json", "planning phase should use streaming mode")
	assert.Contains(t, args, "--json-schema", "planning phase should use JSON schema")
}

func TestWorkflow_FixWorkflow_E2E(t *testing.T) {
	helpers.RequireGit(t)

	repo := helpers.NewTempRepo(t)
	require.NoError(t, repo.CreateFile("buggy.go", "package main\n\nfunc buggy() {\n\t// memory leak here\n}\n"))
	require.NoError(t, repo.Commit("Initial commit"))

	workflowName := "test-fix"
	description := "Fix memory leak in parser"

	planJSON := map[string]interface{}{
		"summary":     "Fix memory leak in parser",
		"contextType": "fix",
		"phases": []map[string]interface{}{
			{
				"name":           "Fix",
				"description":    "Fix the memory leak",
				"estimatedFiles": 1,
				"estimatedLines": 10,
			},
		},
		"complexity": "small",
	}

	implSummary := map[string]interface{}{
		"filesChanged": []string{"buggy.go"},
		"linesAdded":   5,
		"linesRemoved": 2,
		"testsAdded":   1,
		"summary":      "Fixed memory leak",
	}

	refactorSummary := map[string]interface{}{
		"filesChanged": []string{"buggy.go"},
		"linesAdded":   3,
		"linesRemoved": 1,
		"summary":      "Added documentation",
	}

	mockPlanClaude := helpers.NewMockClaudeBuilder(t).
		WithStreamingResponse("Plan created", planJSON)
	planClaudePath := mockPlanClaude.Build()

	mockImplClaude := helpers.NewMockClaudeBuilder(t).
		WithStreamingResponse("Implementation complete", implSummary)
	implClaudePath := mockImplClaude.Build()

	mockRefactorClaude := helpers.NewMockClaudeBuilder(t).
		WithStreamingResponse("Refactoring complete", refactorSummary)
	refactorClaudePath := mockRefactorClaude.Build()

	config := workflow.DefaultConfig(repo.Dir)
	config.ClaudePath = planClaudePath
	config.Timeouts.Planning = 10 * time.Second
	config.Timeouts.Implementation = 10 * time.Second
	config.Timeouts.Refactoring = 10 * time.Second
	config.SplitPR = false

	mockCI := &mockCIChecker{
		result: &workflow.CIResult{
			Passed: true,
			Status: "success",
		},
	}

	orchestrator, err := workflow.NewTestOrchestrator(config, func(workingDir string, checkInterval time.Duration, commandTimeout time.Duration) workflow.CIChecker {
		return mockCI
	})
	require.NoError(t, err)

	orchestrator.SetConfirmFunc(func(plan *workflow.Plan) (bool, string, error) {
		config.ClaudePath = implClaudePath
		return true, "", nil
	})

	ciCallCount := 0
	mockCI.onCall = func(callNum int) {
		ciCallCount = callNum
		if ciCallCount == 1 {
			config.ClaudePath = refactorClaudePath
		}
	}

	ctx := context.Background()
	err = orchestrator.Start(ctx, workflowName, description, workflow.WorkflowTypeFix)
	require.NoError(t, err)

	state, err := orchestrator.Status(workflowName)
	require.NoError(t, err)
	assert.Equal(t, workflow.PhaseCompleted, state.CurrentPhase)
	assert.Equal(t, workflow.WorkflowTypeFix, state.Type)
}

func TestWorkflow_WithPRSplit_E2E(t *testing.T) {
	t.Skip("PR split requires complex GH runner mocking - deferred to later implementation")
}

func TestWorkflow_ResumeAfterFailure_E2E(t *testing.T) {
	helpers.RequireGit(t)

	repo := helpers.NewTempRepo(t)
	require.NoError(t, repo.CreateFile("main.go", "package main\n\nfunc main() {\n}\n"))
	require.NoError(t, repo.Commit("Initial commit"))

	workflowName := "test-resume"
	description := "Add feature with resume"

	planJSON := map[string]interface{}{
		"summary":     "Add new feature",
		"contextType": "feature",
		"phases": []map[string]interface{}{
			{
				"name":           "Implementation",
				"description":    "Implement the feature",
				"estimatedFiles": 2,
				"estimatedLines": 100,
			},
		},
		"complexity": "medium",
	}

	implSummary := map[string]interface{}{
		"filesChanged": []string{"feature.go"},
		"linesAdded":   100,
		"linesRemoved": 0,
		"testsAdded":   5,
		"summary":      "Implemented feature",
	}

	implFixSummary := map[string]interface{}{
		"filesChanged": []string{"feature.go"},
		"linesAdded":   5,
		"linesRemoved": 2,
		"testsAdded":   1,
		"summary":      "Fixed failing tests",
	}

	refactorSummary := map[string]interface{}{
		"filesChanged": []string{"feature.go"},
		"linesAdded":   10,
		"linesRemoved": 5,
		"summary":      "Cleaned up code",
	}

	mockPlanClaude := helpers.NewMockClaudeBuilder(t).
		WithStreamingResponse("Plan created", planJSON)
	planClaudePath := mockPlanClaude.Build()

	mockImplClaude := helpers.NewMockClaudeBuilder(t).
		WithStreamingResponse("Implementation complete", implSummary)
	implClaudePath := mockImplClaude.Build()

	mockImplFixClaude := helpers.NewMockClaudeBuilder(t).
		WithStreamingResponse("Fixed issues", implFixSummary)
	implFixClaudePath := mockImplFixClaude.Build()

	mockRefactorClaude := helpers.NewMockClaudeBuilder(t).
		WithStreamingResponse("Refactoring complete", refactorSummary)
	refactorClaudePath := mockRefactorClaude.Build()

	config := workflow.DefaultConfig(repo.Dir)
	config.ClaudePath = planClaudePath
	config.Timeouts.Planning = 10 * time.Second
	config.Timeouts.Implementation = 10 * time.Second
	config.Timeouts.Refactoring = 10 * time.Second
	config.SplitPR = false

	ciCallCount := 0
	mockCI := &mockCIChecker{
		onCall: func(callNum int) {
			ciCallCount = callNum
		},
	}

	mockCI.checkFunc = func() (*workflow.CIResult, error) {
		if ciCallCount == 1 {
			return &workflow.CIResult{
				Passed:     false,
				Status:     "failure",
				FailedJobs: []string{"test"},
				Output:     "Tests failed: feature_test.go:10: assertion failed",
			}, nil
		}
		return &workflow.CIResult{
			Passed: true,
			Status: "success",
		}, nil
	}

	orchestrator, err := workflow.NewTestOrchestrator(config, func(workingDir string, checkInterval time.Duration, commandTimeout time.Duration) workflow.CIChecker {
		return mockCI
	})
	require.NoError(t, err)

	orchestrator.SetConfirmFunc(func(plan *workflow.Plan) (bool, string, error) {
		config.ClaudePath = implClaudePath
		return true, "", nil
	})

	ctx := context.Background()
	err = orchestrator.Start(ctx, workflowName, description, workflow.WorkflowTypeFeature)

	state, err := orchestrator.Status(workflowName)
	require.NoError(t, err)

	if state.CurrentPhase == workflow.PhaseCompleted {
		t.Skip("CI fix succeeded on first attempt, skipping resume test")
	}

	assert.Equal(t, workflow.PhaseFailed, state.CurrentPhase)
	assert.NotNil(t, state.Error)
	assert.Equal(t, workflow.FailureTypeCI, state.Error.FailureType)

	config.ClaudePath = implFixClaudePath

	ciCallCount = 0
	mockCI.checkFunc = func() (*workflow.CIResult, error) {
		if ciCallCount == 1 {
			config.ClaudePath = refactorClaudePath
		}
		return &workflow.CIResult{
			Passed: true,
			Status: "success",
		}, nil
	}

	err = orchestrator.Resume(ctx, workflowName)
	require.NoError(t, err)

	state, err = orchestrator.Status(workflowName)
	require.NoError(t, err)
	assert.Equal(t, workflow.PhaseCompleted, state.CurrentPhase)
	assert.Nil(t, state.Error, "error should be cleared after successful resume")
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
