package workflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPRCreationResult_JSONParsing(t *testing.T) {
	tests := []struct {
		name        string
		jsonStr     string
		wantErr     bool
		wantPRNum   int
		wantStatus  string
		wantMessage string
		wantMeta    *PRMetadata
	}{
		{
			name: "parse JSON without metadata field (backward compatibility)",
			jsonStr: `{
				"prNumber": 123,
				"status": "created",
				"message": "PR created successfully"
			}`,
			wantErr:     false,
			wantPRNum:   123,
			wantStatus:  "created",
			wantMessage: "PR created successfully",
			wantMeta:    nil,
		},
		{
			name: "parse JSON with empty metadata field",
			jsonStr: `{
				"prNumber": 456,
				"status": "exists",
				"message": "PR already exists",
				"metadata": {}
			}`,
			wantErr:     false,
			wantPRNum:   456,
			wantStatus:  "exists",
			wantMessage: "PR already exists",
			wantMeta:    &PRMetadata{},
		},
		{
			name: "parse JSON with full metadata (issues, labels, projects)",
			jsonStr: `{
				"prNumber": 789,
				"status": "created",
				"message": "PR created with metadata",
				"metadata": {
					"issues": ["#123", "fixes #456"],
					"labels": ["bug", "enhancement"],
					"projects": ["Q1 Planning", "Roadmap"]
				}
			}`,
			wantErr:     false,
			wantPRNum:   789,
			wantStatus:  "created",
			wantMessage: "PR created with metadata",
			wantMeta: &PRMetadata{
				Issues:   []string{"#123", "fixes #456"},
				Labels:   []string{"bug", "enhancement"},
				Projects: []string{"Q1 Planning", "Roadmap"},
			},
		},
		{
			name: "parse JSON with partial metadata (only issues)",
			jsonStr: `{
				"prNumber": 111,
				"status": "created",
				"message": "PR with issues only",
				"metadata": {
					"issues": ["closes #999"]
				}
			}`,
			wantErr:     false,
			wantPRNum:   111,
			wantStatus:  "created",
			wantMessage: "PR with issues only",
			wantMeta: &PRMetadata{
				Issues: []string{"closes #999"},
			},
		},
		{
			name: "parse JSON with partial metadata (only labels)",
			jsonStr: `{
				"prNumber": 222,
				"status": "created",
				"message": "PR with labels only",
				"metadata": {
					"labels": ["documentation"]
				}
			}`,
			wantErr:     false,
			wantPRNum:   222,
			wantStatus:  "created",
			wantMessage: "PR with labels only",
			wantMeta: &PRMetadata{
				Labels: []string{"documentation"},
			},
		},
		{
			name: "parse JSON with partial metadata (only projects)",
			jsonStr: `{
				"prNumber": 333,
				"status": "created",
				"message": "PR with projects only",
				"metadata": {
					"projects": ["Backlog"]
				}
			}`,
			wantErr:     false,
			wantPRNum:   333,
			wantStatus:  "created",
			wantMessage: "PR with projects only",
			wantMeta: &PRMetadata{
				Projects: []string{"Backlog"},
			},
		},
		{
			name: "parse JSON with skipped status and no metadata",
			jsonStr: `{
				"prNumber": 0,
				"status": "skipped",
				"message": "No commits to create PR"
			}`,
			wantErr:     false,
			wantPRNum:   0,
			wantStatus:  "skipped",
			wantMessage: "No commits to create PR",
			wantMeta:    nil,
		},
		{
			name: "parse JSON with failed status",
			jsonStr: `{
				"prNumber": 0,
				"status": "failed",
				"message": "Failed to create PR: permission denied"
			}`,
			wantErr:     false,
			wantPRNum:   0,
			wantStatus:  "failed",
			wantMessage: "Failed to create PR: permission denied",
			wantMeta:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got PRCreationResult
			err := json.Unmarshal([]byte(tt.jsonStr), &got)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantPRNum, got.PRNumber)
			assert.Equal(t, tt.wantStatus, got.Status)
			assert.Equal(t, tt.wantMessage, got.Message)

			if tt.wantMeta == nil {
				assert.Nil(t, got.Metadata)
			} else {
				require.NotNil(t, got.Metadata)
				assert.Equal(t, tt.wantMeta.Issues, got.Metadata.Issues)
				assert.Equal(t, tt.wantMeta.Labels, got.Metadata.Labels)
				assert.Equal(t, tt.wantMeta.Projects, got.Metadata.Projects)
			}
		})
	}
}

func TestPRCreationResultSchema_Validation(t *testing.T) {
	tests := []struct {
		name    string
		jsonStr string
		wantErr bool
	}{
		{
			name: "valid JSON without metadata",
			jsonStr: `{
				"prNumber": 123,
				"status": "created",
				"message": "PR created"
			}`,
			wantErr: false,
		},
		{
			name: "valid JSON with complete metadata",
			jsonStr: `{
				"prNumber": 456,
				"status": "exists",
				"message": "PR exists",
				"metadata": {
					"issues": ["#123"],
					"labels": ["bug"],
					"projects": ["Q1"]
				}
			}`,
			wantErr: false,
		},
		{
			name: "valid JSON with partial metadata",
			jsonStr: `{
				"prNumber": 789,
				"status": "created",
				"message": "PR created",
				"metadata": {
					"issues": ["fixes #999"]
				}
			}`,
			wantErr: false,
		},
		{
			name: "valid JSON with empty metadata object",
			jsonStr: `{
				"prNumber": 0,
				"status": "skipped",
				"message": "Skipped",
				"metadata": {}
			}`,
			wantErr: false,
		},
		{
			name: "valid JSON with empty arrays in metadata",
			jsonStr: `{
				"prNumber": 111,
				"status": "created",
				"message": "Created",
				"metadata": {
					"issues": [],
					"labels": [],
					"projects": []
				}
			}`,
			wantErr: false,
		},
		{
			name: "invalid JSON - wrong type for metadata.issues",
			jsonStr: `{
				"prNumber": 123,
				"status": "created",
				"message": "Test",
				"metadata": {
					"issues": "not-an-array"
				}
			}`,
			wantErr: true,
		},
		{
			name: "invalid JSON - wrong type for metadata.labels",
			jsonStr: `{
				"prNumber": 123,
				"status": "created",
				"message": "Test",
				"metadata": {
					"labels": 123
				}
			}`,
			wantErr: true,
		},
		{
			name: "invalid JSON - wrong type for metadata.projects",
			jsonStr: `{
				"prNumber": 123,
				"status": "created",
				"message": "Test",
				"metadata": {
					"projects": true
				}
			}`,
			wantErr: true,
		},
		{
			name: "invalid JSON - wrong type for prNumber",
			jsonStr: `{
				"prNumber": "not-a-number",
				"status": "created",
				"message": "Test"
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result PRCreationResult
			err := json.Unmarshal([]byte(tt.jsonStr), &result)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestLogPRMetadata(t *testing.T) {
	tests := []struct {
		name         string
		metadata     *PRMetadata
		wantContains []string
		wantEmpty    bool
	}{
		{
			name:      "handles nil metadata gracefully",
			metadata:  nil,
			wantEmpty: true,
		},
		{
			name:      "handles empty metadata (all arrays empty)",
			metadata:  &PRMetadata{},
			wantEmpty: true,
		},
		{
			name: "logs issues when present",
			metadata: &PRMetadata{
				Issues: []string{"#123", "fixes #456"},
			},
			wantContains: []string{"Applied issue references", "#123", "fixes #456"},
		},
		{
			name: "logs labels when present",
			metadata: &PRMetadata{
				Labels: []string{"bug", "enhancement"},
			},
			wantContains: []string{"Applied labels", "bug", "enhancement"},
		},
		{
			name: "logs projects when present",
			metadata: &PRMetadata{
				Projects: []string{"Q1 Planning", "Roadmap"},
			},
			wantContains: []string{"Applied to projects", "Q1 Planning", "Roadmap"},
		},
		{
			name: "logs all metadata types when all present",
			metadata: &PRMetadata{
				Issues:   []string{"closes #789"},
				Labels:   []string{"documentation"},
				Projects: []string{"Backlog"},
			},
			wantContains: []string{
				"Applied issue references", "closes #789",
				"Applied labels", "documentation",
				"Applied to projects", "Backlog",
			},
		},
		{
			name: "logs multiple items in each category",
			metadata: &PRMetadata{
				Issues:   []string{"#1", "#2", "#3"},
				Labels:   []string{"bug", "critical", "security"},
				Projects: []string{"Sprint 1", "Sprint 2"},
			},
			wantContains: []string{
				"Applied issue references", "#1", "#2", "#3",
				"Applied labels", "bug", "critical", "security",
				"Applied to projects", "Sprint 1", "Sprint 2",
			},
		},
		{
			name: "handles metadata with empty arrays",
			metadata: &PRMetadata{
				Issues:   []string{},
				Labels:   []string{},
				Projects: []string{},
			},
			wantEmpty: true,
		},
		{
			name: "logs only non-empty fields",
			metadata: &PRMetadata{
				Issues:   []string{"#999"},
				Labels:   []string{},
				Projects: []string{},
			},
			wantContains: []string{"Applied issue references", "#999"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Create orchestrator with minimal config for testing
			config := DefaultConfig(t.TempDir())
			config.LogLevel = LogLevelNormal
			o, err := NewOrchestratorWithConfig(config)
			require.NoError(t, err)

			// Execute the function
			o.logPRMetadata(tt.metadata)

			// Restore stdout
			w.Close()
			os.Stdout = old

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if tt.wantEmpty {
				assert.Empty(t, output, "expected no output for nil or empty metadata")
				return
			}

			// Check all expected strings are present
			for _, want := range tt.wantContains {
				assert.Contains(t, output, want, "output should contain %q", want)
			}
		})
	}
}

func TestLogPRMetadata_OutputFormat(t *testing.T) {
	tests := []struct {
		name        string
		metadata    *PRMetadata
		checkFormat func(t *testing.T, output string)
	}{
		{
			name: "issues are comma-separated on single line",
			metadata: &PRMetadata{
				Issues: []string{"#123", "fixes #456", "closes #789"},
			},
			checkFormat: func(t *testing.T, output string) {
				lines := strings.Split(strings.TrimSpace(output), "\n")
				require.Len(t, lines, 1, "issues should be on single line")
				assert.Contains(t, lines[0], "#123, fixes #456, closes #789")
			},
		},
		{
			name: "labels are comma-separated on single line",
			metadata: &PRMetadata{
				Labels: []string{"bug", "enhancement", "documentation"},
			},
			checkFormat: func(t *testing.T, output string) {
				lines := strings.Split(strings.TrimSpace(output), "\n")
				require.Len(t, lines, 1, "labels should be on single line")
				assert.Contains(t, lines[0], "bug, enhancement, documentation")
			},
		},
		{
			name: "projects are comma-separated on single line",
			metadata: &PRMetadata{
				Projects: []string{"Q1 Planning", "Q2 Planning", "Roadmap"},
			},
			checkFormat: func(t *testing.T, output string) {
				lines := strings.Split(strings.TrimSpace(output), "\n")
				require.Len(t, lines, 1, "projects should be on single line")
				assert.Contains(t, lines[0], "Q1 Planning, Q2 Planning, Roadmap")
			},
		},
		{
			name: "all metadata types are on separate lines",
			metadata: &PRMetadata{
				Issues:   []string{"#123"},
				Labels:   []string{"bug"},
				Projects: []string{"Sprint 1"},
			},
			checkFormat: func(t *testing.T, output string) {
				lines := strings.Split(strings.TrimSpace(output), "\n")
				require.Len(t, lines, 3, "each metadata type should be on its own line")

				assert.Contains(t, lines[0], "issue references")
				assert.Contains(t, lines[1], "labels")
				assert.Contains(t, lines[2], "projects")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Create orchestrator
			config := DefaultConfig(t.TempDir())
			config.LogLevel = LogLevelNormal
			o, err := NewOrchestratorWithConfig(config)
			require.NoError(t, err)

			// Execute the function
			o.logPRMetadata(tt.metadata)

			// Restore stdout
			w.Close()
			os.Stdout = old

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Run custom format check
			tt.checkFormat(t, output)
		})
	}
}

func TestPRMetadata_OmitEmpty(t *testing.T) {
	tests := []struct {
		name         string
		metadata     PRMetadata
		wantContains string
		wantOmit     string
	}{
		{
			name: "omits empty issues array from JSON",
			metadata: PRMetadata{
				Labels:   []string{"bug"},
				Projects: []string{"Sprint 1"},
			},
			wantContains: `"labels"`,
			wantOmit:     `"issues"`,
		},
		{
			name: "omits empty labels array from JSON",
			metadata: PRMetadata{
				Issues:   []string{"#123"},
				Projects: []string{"Sprint 1"},
			},
			wantContains: `"issues"`,
			wantOmit:     `"labels"`,
		},
		{
			name: "omits empty projects array from JSON",
			metadata: PRMetadata{
				Issues: []string{"#123"},
				Labels: []string{"bug"},
			},
			wantContains: `"issues"`,
			wantOmit:     `"projects"`,
		},
		{
			name:     "omits all fields when all arrays empty",
			metadata: PRMetadata{},
			wantOmit: `"issues"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.metadata)
			require.NoError(t, err)

			jsonStr := string(jsonBytes)

			if tt.wantContains != "" {
				assert.Contains(t, jsonStr, tt.wantContains)
			}

			if tt.wantOmit != "" {
				assert.NotContains(t, jsonStr, tt.wantOmit,
					fmt.Sprintf("JSON should omit empty fields, got: %s", jsonStr))
			}
		})
	}
}

func TestPRCreationResult_OmitEmpty(t *testing.T) {
	tests := []struct {
		name         string
		result       PRCreationResult
		wantContains []string
		wantOmit     string
	}{
		{
			name: "omits metadata field when nil",
			result: PRCreationResult{
				PRNumber: 123,
				Status:   "created",
				Message:  "PR created",
				Metadata: nil,
			},
			wantContains: []string{`"prNumber"`, `"status"`, `"message"`},
			wantOmit:     `"metadata"`,
		},
		{
			name: "includes metadata field when present",
			result: PRCreationResult{
				PRNumber: 123,
				Status:   "created",
				Message:  "PR created",
				Metadata: &PRMetadata{
					Issues: []string{"#123"},
				},
			},
			wantContains: []string{`"prNumber"`, `"status"`, `"message"`, `"metadata"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.result)
			require.NoError(t, err)

			jsonStr := string(jsonBytes)

			for _, want := range tt.wantContains {
				assert.Contains(t, jsonStr, want)
			}

			if tt.wantOmit != "" {
				assert.NotContains(t, jsonStr, tt.wantOmit,
					fmt.Sprintf("JSON should omit empty metadata field, got: %s", jsonStr))
			}
		})
	}
}
