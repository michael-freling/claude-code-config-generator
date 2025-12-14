//go:build e2e

package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/michael-freling/claude-code-tools/internal/workflow"
	"github.com/michael-freling/claude-code-tools/test/e2e/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateManager_WorkflowCreation(t *testing.T) {
	helpers.RequireGit(t)

	tmpDir := t.TempDir()
	stateManager := workflow.NewStateManager(tmpDir)

	tests := []struct {
		name        string
		workflowName string
		description string
		wfType      workflow.WorkflowType
		wantErr     bool
	}{
		{
			name:         "create feature workflow",
			workflowName: "test-feature-1",
			description:  "Test feature workflow",
			wfType:       workflow.WorkflowTypeFeature,
			wantErr:      false,
		},
		{
			name:         "create fix workflow",
			workflowName: "test-fix-1",
			description:  "Test fix workflow",
			wfType:       workflow.WorkflowTypeFix,
			wantErr:      false,
		},
		{
			name:         "duplicate workflow fails",
			workflowName: "test-duplicate",
			description:  "Duplicate workflow",
			wfType:       workflow.WorkflowTypeFeature,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "duplicate workflow fails" {
				// Create it once first
				_, err := stateManager.InitState(tt.workflowName, tt.description, tt.wfType)
				require.NoError(t, err)
			}

			state, err := stateManager.InitState(tt.workflowName, tt.description, tt.wfType)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.workflowName, state.Name)
			assert.Equal(t, tt.description, state.Description)
			assert.Equal(t, tt.wfType, state.Type)
			assert.Equal(t, workflow.PhasePlanning, state.CurrentPhase)
			assert.NotNil(t, state.Phases[workflow.PhasePlanning])
			assert.Equal(t, workflow.StatusInProgress, state.Phases[workflow.PhasePlanning].Status)
		})
	}
}

func TestStateManager_PersistenceOperations(t *testing.T) {
	helpers.RequireGit(t)

	tmpDir := t.TempDir()
	stateManager := workflow.NewStateManager(tmpDir)

	workflowName := "persistence-test"
	description := "Test workflow persistence"

	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "save and load state",
			test: func(t *testing.T) {
				state, err := stateManager.InitState(workflowName, description, workflow.WorkflowTypeFeature)
				require.NoError(t, err)

				// Modify state
				state.CurrentPhase = workflow.PhaseImplementation
				err = stateManager.SaveState(workflowName, state)
				require.NoError(t, err)

				// Load and verify
				loadedState, err := stateManager.LoadState(workflowName)
				require.NoError(t, err)
				assert.Equal(t, workflow.PhaseImplementation, loadedState.CurrentPhase)
			},
		},
		{
			name: "workflow directory created",
			test: func(t *testing.T) {
				err := stateManager.EnsureWorkflowDir("dir-test")
				require.NoError(t, err)

				workflowDir := stateManager.WorkflowDir("dir-test")
				info, err := os.Stat(workflowDir)
				require.NoError(t, err)
				assert.True(t, info.IsDir())

				phasesDir := filepath.Join(workflowDir, "phases")
				info, err = os.Stat(phasesDir)
				require.NoError(t, err)
				assert.True(t, info.IsDir())
			},
		},
		{
			name: "workflow exists check",
			test: func(t *testing.T) {
				assert.False(t, stateManager.WorkflowExists("nonexistent"))

				_, err := stateManager.InitState("exists-test", description, workflow.WorkflowTypeFeature)
				require.NoError(t, err)

				assert.True(t, stateManager.WorkflowExists("exists-test"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func TestStateManager_PlanOperations(t *testing.T) {
	helpers.RequireGit(t)

	tmpDir := t.TempDir()
	stateManager := workflow.NewStateManager(tmpDir)

	workflowName := "plan-test"
	_, err := stateManager.InitState(workflowName, "Test plan", workflow.WorkflowTypeFeature)
	require.NoError(t, err)

	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "save and load plan JSON",
			test: func(t *testing.T) {
				plan := &workflow.Plan{
					Summary:     "Test plan summary",
					ContextType: "feature",
					Architecture: workflow.Architecture{
						Overview:   "Test architecture",
						Components: []string{"component1", "component2"},
					},
					Phases: []workflow.PlanPhase{
						{
							Name:           "Phase 1",
							Description:    "First phase",
							EstimatedFiles: 5,
							EstimatedLines: 100,
						},
					},
					WorkStreams:         []workflow.WorkStream{},
					Risks:               []string{"risk1", "risk2"},
					Complexity:          "medium",
					EstimatedTotalLines: 100,
					EstimatedTotalFiles: 5,
				}

				err := stateManager.SavePlan(workflowName, plan)
				require.NoError(t, err)

				loadedPlan, err := stateManager.LoadPlan(workflowName)
				require.NoError(t, err)
				assert.Equal(t, plan.Summary, loadedPlan.Summary)
				assert.Equal(t, plan.ContextType, loadedPlan.ContextType)
				assert.Len(t, loadedPlan.Phases, 1)
				assert.Equal(t, "Phase 1", loadedPlan.Phases[0].Name)
			},
		},
		{
			name: "save plan markdown",
			test: func(t *testing.T) {
				markdown := "# Test Plan\n\nThis is a test plan"
				err := stateManager.SavePlanMarkdown(workflowName, markdown)
				require.NoError(t, err)

				planPath := filepath.Join(stateManager.WorkflowDir(workflowName), "plan.md")
				content, err := os.ReadFile(planPath)
				require.NoError(t, err)
				assert.Equal(t, markdown, string(content))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func TestStateManager_PhaseOutputOperations(t *testing.T) {
	helpers.RequireGit(t)

	tmpDir := t.TempDir()
	stateManager := workflow.NewStateManager(tmpDir)

	workflowName := "phase-output-test"
	_, err := stateManager.InitState(workflowName, "Test phase output", workflow.WorkflowTypeFeature)
	require.NoError(t, err)

	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "save and load phase output",
			test: func(t *testing.T) {
				output := &workflow.ImplementationSummary{
					FilesChanged: []string{"file1.go", "file2.go"},
					LinesAdded:   50,
					LinesRemoved: 10,
					TestsAdded:   5,
					Summary:      "Implementation complete",
					NextSteps:    []string{"Review", "Test"},
				}

				err := stateManager.SavePhaseOutput(workflowName, workflow.PhaseImplementation, output)
				require.NoError(t, err)

				var loaded workflow.ImplementationSummary
				err = stateManager.LoadPhaseOutput(workflowName, workflow.PhaseImplementation, &loaded)
				require.NoError(t, err)

				assert.Equal(t, output.FilesChanged, loaded.FilesChanged)
				assert.Equal(t, output.LinesAdded, loaded.LinesAdded)
				assert.Equal(t, output.TestsAdded, loaded.TestsAdded)
				assert.Equal(t, output.Summary, loaded.Summary)
			},
		},
		{
			name: "save raw output",
			test: func(t *testing.T) {
				rawOutput := "Raw Claude output for debugging"
				err := stateManager.SaveRawOutput(workflowName, workflow.PhasePlanning, rawOutput)
				require.NoError(t, err)

				rawPath := filepath.Join(stateManager.WorkflowDir(workflowName), "phases", "PLANNING_raw.txt")
				content, err := os.ReadFile(rawPath)
				require.NoError(t, err)
				assert.Equal(t, rawOutput, string(content))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func TestStateManager_ListAndDeleteWorkflows(t *testing.T) {
	helpers.RequireGit(t)

	tmpDir := t.TempDir()
	stateManager := workflow.NewStateManager(tmpDir)

	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "list workflows",
			test: func(t *testing.T) {
				_, err := stateManager.InitState("workflow1", "First workflow", workflow.WorkflowTypeFeature)
				require.NoError(t, err)

				_, err = stateManager.InitState("workflow2", "Second workflow", workflow.WorkflowTypeFix)
				require.NoError(t, err)

				workflows, err := stateManager.ListWorkflows()
				require.NoError(t, err)
				assert.GreaterOrEqual(t, len(workflows), 2)

				var found1, found2 bool
				for _, wf := range workflows {
					if wf.Name == "workflow1" {
						found1 = true
						assert.Equal(t, workflow.WorkflowTypeFeature, wf.Type)
					}
					if wf.Name == "workflow2" {
						found2 = true
						assert.Equal(t, workflow.WorkflowTypeFix, wf.Type)
					}
				}
				assert.True(t, found1, "workflow1 not found")
				assert.True(t, found2, "workflow2 not found")
			},
		},
		{
			name: "delete workflow",
			test: func(t *testing.T) {
				_, err := stateManager.InitState("delete-me", "Workflow to delete", workflow.WorkflowTypeFeature)
				require.NoError(t, err)

				assert.True(t, stateManager.WorkflowExists("delete-me"))

				err = stateManager.DeleteWorkflow("delete-me")
				require.NoError(t, err)

				assert.False(t, stateManager.WorkflowExists("delete-me"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

func TestStateManager_PromptSaving(t *testing.T) {
	helpers.RequireGit(t)

	tmpDir := t.TempDir()
	stateManager := workflow.NewStateManager(tmpDir)

	workflowName := "prompt-test"
	_, err := stateManager.InitState(workflowName, "Test prompt saving", workflow.WorkflowTypeFeature)
	require.NoError(t, err)

	tests := []struct {
		name    string
		phase   workflow.Phase
		attempt int
		prompt  string
		wantErr bool
	}{
		{
			name:    "save planning prompt",
			phase:   workflow.PhasePlanning,
			attempt: 1,
			prompt:  "This is a planning phase prompt",
			wantErr: false,
		},
		{
			name:    "save implementation prompt",
			phase:   workflow.PhaseImplementation,
			attempt: 2,
			prompt:  "This is an implementation phase prompt for attempt 2",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			savedPath, err := stateManager.SavePrompt(workflowName, tt.phase, tt.attempt, tt.prompt)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, savedPath)

			content, err := os.ReadFile(savedPath)
			require.NoError(t, err)
			assert.Equal(t, tt.prompt, string(content))
		})
	}
}
