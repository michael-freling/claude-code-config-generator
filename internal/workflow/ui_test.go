package workflow

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestColorFunctions(t *testing.T) {
	tests := []struct {
		name      string
		colorFunc func(string) string
		input     string
		wantStart string
		wantEnd   string
	}{
		{
			name:      "Green wraps with green ANSI codes",
			colorFunc: Green,
			input:     "success",
			wantStart: "\033[32m",
			wantEnd:   "\033[0m",
		},
		{
			name:      "Red wraps with red ANSI codes",
			colorFunc: Red,
			input:     "error",
			wantStart: "\033[31m",
			wantEnd:   "\033[0m",
		},
		{
			name:      "Yellow wraps with yellow ANSI codes",
			colorFunc: Yellow,
			input:     "warning",
			wantStart: "\033[33m",
			wantEnd:   "\033[0m",
		},
		{
			name:      "Cyan wraps with cyan ANSI codes",
			colorFunc: Cyan,
			input:     "info",
			wantStart: "\033[36m",
			wantEnd:   "\033[0m",
		},
		{
			name:      "Bold wraps with bold ANSI codes",
			colorFunc: Bold,
			input:     "emphasis",
			wantStart: "\033[1m",
			wantEnd:   "\033[0m",
		},
		{
			name:      "Green handles empty string",
			colorFunc: Green,
			input:     "",
			wantStart: "\033[32m",
			wantEnd:   "\033[0m",
		},
		{
			name:      "Red handles special characters",
			colorFunc: Red,
			input:     "error: failed!\n",
			wantStart: "\033[31m",
			wantEnd:   "\033[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.colorFunc(tt.input)
			assert.True(t, strings.HasPrefix(got, tt.wantStart))
			assert.True(t, strings.HasSuffix(got, tt.wantEnd))
			assert.Contains(t, got, tt.input)
		})
	}
}

func TestNewSpinner(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "creates spinner with message",
			message: "Loading...",
		},
		{
			name:    "creates spinner with empty message",
			message: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSpinner(tt.message)
			require.NotNil(t, got)
			assert.Equal(t, tt.message, got.message)
			assert.False(t, got.running)
			assert.NotNil(t, got.done)
		})
	}
}

func TestSpinner_Lifecycle(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Start and Stop cycle works correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spinner := NewSpinner("Testing")

			spinner.Start()
			time.Sleep(10 * time.Millisecond)
			assert.True(t, spinner.running)

			spinner.Stop()
			time.Sleep(10 * time.Millisecond)
			assert.False(t, spinner.running)
		})
	}
}

func TestSpinner_DoubleStart(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "double-start is idempotent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spinner := NewSpinner("Testing")

			spinner.Start()
			time.Sleep(10 * time.Millisecond)
			assert.True(t, spinner.running)

			spinner.Start()
			time.Sleep(10 * time.Millisecond)
			assert.True(t, spinner.running)

			spinner.Stop()
			time.Sleep(10 * time.Millisecond)
			assert.False(t, spinner.running)
		})
	}
}

func TestSpinner_DoubleStop(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "double-stop is idempotent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spinner := NewSpinner("Testing")

			spinner.Start()
			time.Sleep(10 * time.Millisecond)

			spinner.Stop()
			time.Sleep(10 * time.Millisecond)
			assert.False(t, spinner.running)

			spinner.Stop()
			time.Sleep(10 * time.Millisecond)
			assert.False(t, spinner.running)
		})
	}
}

func TestSpinner_Success(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "Success stops spinner and prints message",
			message: "Operation completed successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spinner := NewSpinner("Testing")

			spinner.Start()
			time.Sleep(10 * time.Millisecond)

			spinner.Success(tt.message)
			time.Sleep(10 * time.Millisecond)
			assert.False(t, spinner.running)
		})
	}
}

func TestSpinner_Fail(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "Fail stops spinner and prints error message",
			message: "Operation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spinner := NewSpinner("Testing")

			spinner.Start()
			time.Sleep(10 * time.Millisecond)

			spinner.Fail(tt.message)
			time.Sleep(10 * time.Millisecond)
			assert.False(t, spinner.running)
		})
	}
}

func TestNewStreamingSpinner(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "creates streaming spinner with message",
			message: "Processing...",
		},
		{
			name:    "creates streaming spinner with empty message",
			message: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewStreamingSpinner(tt.message)
			require.NotNil(t, got)
			assert.Equal(t, tt.message, got.message)
			assert.False(t, got.running)
			assert.NotNil(t, got.done)
		})
	}
}

func TestStreamingSpinner_OnProgress(t *testing.T) {
	tests := []struct {
		name  string
		event ProgressEvent
	}{
		{
			name: "tool_use event with ToolInput",
			event: ProgressEvent{
				Type:      "tool_use",
				ToolName:  "Read",
				ToolInput: "/path/to/file.go",
			},
		},
		{
			name: "tool_use event without ToolInput",
			event: ProgressEvent{
				Type:     "tool_use",
				ToolName: "Bash",
			},
		},
		{
			name: "tool_result event with IsError true",
			event: ProgressEvent{
				Type:    "tool_result",
				Text:    "File not found",
				IsError: true,
			},
		},
		{
			name: "tool_result event with IsError false",
			event: ProgressEvent{
				Type:    "tool_result",
				Text:    "Success",
				IsError: false,
			},
		},
		{
			name: "text event",
			event: ProgressEvent{
				Type: "text",
				Text: "Some output",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spinner := NewStreamingSpinner("Testing")
			spinner.Start()
			time.Sleep(10 * time.Millisecond)

			spinner.OnProgress(tt.event)

			if tt.event.Type == "tool_use" {
				if tt.event.ToolInput != "" {
					assert.Contains(t, spinner.lastTool, tt.event.ToolName)
				} else {
					assert.Equal(t, tt.event.ToolName, spinner.lastTool)
				}
			}

			spinner.Stop()
			time.Sleep(10 * time.Millisecond)
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "0 seconds",
			duration: 0 * time.Second,
			want:     "0s",
		},
		{
			name:     "30 seconds",
			duration: 30 * time.Second,
			want:     "30s",
		},
		{
			name:     "90 seconds",
			duration: 90 * time.Second,
			want:     "1m 30s",
		},
		{
			name:     "3600 seconds",
			duration: 3600 * time.Second,
			want:     "1h 0m 0s",
		},
		{
			name:     "3661 seconds",
			duration: 3661 * time.Second,
			want:     "1h 1m 1s",
		},
		{
			name:     "1 hour exactly",
			duration: 1 * time.Hour,
			want:     "1h 0m 0s",
		},
		{
			name:     "2 hours 30 minutes 45 seconds",
			duration: 2*time.Hour + 30*time.Minute + 45*time.Second,
			want:     "2h 30m 45s",
		},
		{
			name:     "1 minute exactly",
			duration: 1 * time.Minute,
			want:     "1m 0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDuration(tt.duration)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatPhase(t *testing.T) {
	tests := []struct {
		name  string
		phase Phase
		total int
		want  string
	}{
		{
			name:  "PhasePlanning",
			phase: PhasePlanning,
			total: 5,
			want:  "Phase 1/5: Planning",
		},
		{
			name:  "PhaseConfirmation",
			phase: PhaseConfirmation,
			total: 5,
			want:  "Phase 2/5: Confirmation",
		},
		{
			name:  "PhaseImplementation",
			phase: PhaseImplementation,
			total: 5,
			want:  "Phase 3/5: Implementation",
		},
		{
			name:  "PhaseRefactoring",
			phase: PhaseRefactoring,
			total: 5,
			want:  "Phase 4/5: Refactoring",
		},
		{
			name:  "PhasePRSplit",
			phase: PhasePRSplit,
			total: 5,
			want:  "Phase 5/5: PR Split",
		},
		{
			name:  "PhaseCompleted",
			phase: PhaseCompleted,
			total: 5,
			want:  "Phase 0/5: Completed",
		},
		{
			name:  "PhaseFailed",
			phase: PhaseFailed,
			total: 5,
			want:  "Phase 0/5: Failed",
		},
		{
			name:  "Unknown phase",
			phase: Phase("UNKNOWN"),
			total: 5,
			want:  "Phase 0/5: UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatPhase(tt.phase, tt.total)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTruncateForDisplay(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "short string unchanged",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "long string truncated with ellipsis",
			input:  "this is a very long string that needs truncation",
			maxLen: 20,
			want:   "this is a very lo...",
		},
		{
			name:   "string with newlines converted to spaces",
			input:  "line1\nline2\nline3",
			maxLen: 50,
			want:   "line1 line2 line3",
		},
		{
			name:   "string with multiple spaces collapsed",
			input:  "hello    world   test",
			maxLen: 50,
			want:   "hello world test",
		},
		{
			name:   "string with tabs converted to spaces",
			input:  "hello\tworld",
			maxLen: 50,
			want:   "hello world",
		},
		{
			name:   "exact length string unchanged",
			input:  "12345678901234567890",
			maxLen: 20,
			want:   "12345678901234567890",
		},
		{
			name:   "string longer by one character truncated",
			input:  "123456789012345678901",
			maxLen: 20,
			want:   "12345678901234567...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateForDisplay(tt.input, tt.maxLen)
			assert.Equal(t, tt.want, got)
			if len(tt.input) > tt.maxLen {
				assert.LessOrEqual(t, len(got), tt.maxLen)
			}
		})
	}
}

func TestFormatPlanSummary(t *testing.T) {
	tests := []struct {
		name         string
		plan         *Plan
		wantContains []string
	}{
		{
			name: "complete plan with all fields",
			plan: &Plan{
				Summary:             "Add authentication feature",
				Complexity:          "Medium",
				EstimatedTotalLines: 500,
				EstimatedTotalFiles: 10,
				Architecture: Architecture{
					Overview:   "JWT-based authentication",
					Components: []string{"Auth Handler", "Token Service"},
				},
				Phases: []PlanPhase{
					{
						Name:           "Setup",
						Description:    "Initial setup",
						EstimatedFiles: 3,
						EstimatedLines: 150,
					},
				},
				WorkStreams: []WorkStream{
					{
						Name:      "Backend",
						Tasks:     []string{"Create API", "Add middleware"},
						DependsOn: []string{"Setup"},
					},
				},
				Risks: []string{"Security vulnerability", "Performance impact"},
			},
			wantContains: []string{
				"Plan Summary",
				"Add authentication feature",
				"Complexity: ",
				"Medium",
				"~500 lines across 10 files",
				"Architecture",
				"JWT-based authentication",
				"Auth Handler",
				"Token Service",
				"Phases (1 total)",
				"Setup",
				"3 files, ~150 lines",
				"Work Streams",
				"Backend",
				"Create API",
				"Dependencies: Setup",
				"Risks",
				"Security vulnerability",
				"Performance impact",
			},
		},
		{
			name: "minimal plan without optional fields",
			plan: &Plan{
				Summary:             "Simple feature",
				Complexity:          "Low",
				EstimatedTotalLines: 50,
				EstimatedTotalFiles: 2,
				Phases: []PlanPhase{
					{
						Name:           "Implementation",
						EstimatedFiles: 2,
						EstimatedLines: 50,
					},
				},
			},
			wantContains: []string{
				"Plan Summary",
				"Simple feature",
				"Complexity: ",
				"Low",
				"~50 lines across 2 files",
				"Phases (1 total)",
				"Implementation",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatPlanSummary(tt.plan)
			for _, want := range tt.wantContains {
				assert.Contains(t, got, want)
			}
		})
	}
}

func TestFormatWorkflowStatus(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-30 * time.Second)

	tests := []struct {
		name         string
		state        *WorkflowState
		wantContains []string
	}{
		{
			name: "completed workflow with green status",
			state: &WorkflowState{
				Name:         "feature/auth",
				Type:         WorkflowTypeFeature,
				Description:  "Add authentication",
				CurrentPhase: PhaseCompleted,
				CreatedAt:    earlier,
				Phases: map[Phase]*PhaseState{
					PhasePlanning: {
						Status:   StatusCompleted,
						Attempts: 1,
					},
				},
			},
			wantContains: []string{
				"Workflow: ",
				"feature/auth",
				"Type: feature",
				"Description: Add authentication",
				"Status: ",
				"Completed",
				"Current Phase: Completed",
				"Elapsed: ",
				"Phase History:",
			},
		},
		{
			name: "failed workflow with red status",
			state: &WorkflowState{
				Name:         "fix/bug",
				Type:         WorkflowTypeFix,
				Description:  "Fix critical bug",
				CurrentPhase: PhaseFailed,
				CreatedAt:    earlier,
				Error: &WorkflowError{
					Message:     "Build failed",
					Phase:       PhaseImplementation,
					Recoverable: false,
				},
				Phases: map[Phase]*PhaseState{},
			},
			wantContains: []string{
				"Workflow: ",
				"fix/bug",
				"Type: fix",
				"Status: ",
				"Failed",
				"Error: ",
				"Build failed",
			},
		},
		{
			name: "in progress workflow with yellow status",
			state: &WorkflowState{
				Name:         "feature/new",
				Type:         WorkflowTypeFeature,
				Description:  "New feature",
				CurrentPhase: PhaseImplementation,
				CreatedAt:    earlier,
				Phases: map[Phase]*PhaseState{
					PhaseImplementation: {
						Status:   StatusInProgress,
						Attempts: 1,
					},
				},
			},
			wantContains: []string{
				"Workflow: ",
				"feature/new",
				"Status: ",
				"In Progress",
				"Current Phase: Implementation",
			},
		},
		{
			name: "workflow with recoverable error shows recovery hint",
			state: &WorkflowState{
				Name:         "feature/retry",
				Type:         WorkflowTypeFeature,
				Description:  "Retry test",
				CurrentPhase: PhaseImplementation,
				CreatedAt:    earlier,
				Error: &WorkflowError{
					Message:     "Temporary failure",
					Phase:       PhaseImplementation,
					Recoverable: true,
				},
				Phases: map[Phase]*PhaseState{},
			},
			wantContains: []string{
				"Error: ",
				"Temporary failure",
				"This error is recoverable. Use 'resume' to retry.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatWorkflowStatus(tt.state)
			for _, want := range tt.wantContains {
				assert.Contains(t, got, want)
			}
		})
	}
}
