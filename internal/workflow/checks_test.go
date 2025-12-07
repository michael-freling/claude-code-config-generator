package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCIOutput(t *testing.T) {
	tests := []struct {
		name           string
		output         string
		wantStatus     string
		wantFailedJobs []string
	}{
		{
			name:           "empty output",
			output:         "",
			wantStatus:     "pending",
			wantFailedJobs: []string{},
		},
		{
			name: "all checks passed",
			output: `✓ build
✓ test
✓ lint`,
			wantStatus:     "success",
			wantFailedJobs: []string{},
		},
		{
			name: "some checks failed",
			output: `✓ build
✗ test
✓ lint`,
			wantStatus:     "failure",
			wantFailedJobs: []string{"test"},
		},
		{
			name: "checks pending",
			output: `✓ build
○ test
○ lint`,
			wantStatus:     "pending",
			wantFailedJobs: []string{},
		},
		{
			name: "mixed status with pending",
			output: `✓ build
✗ test
○ lint`,
			wantStatus:     "pending",
			wantFailedJobs: []string{"test"},
		},
		{
			name: "multiple failed jobs",
			output: `✓ build
✗ test-unit
✗ test-integration
✓ lint`,
			wantStatus:     "failure",
			wantFailedJobs: []string{"test-unit", "test-integration"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, gotFailedJobs := parseCIOutput(tt.output)
			assert.Equal(t, tt.wantStatus, gotStatus)
			assert.Equal(t, tt.wantFailedJobs, gotFailedJobs)
		})
	}
}

func TestNewCIChecker(t *testing.T) {
	tests := []struct {
		name           string
		workingDir     string
		checkInterval  time.Duration
		commandTimeout time.Duration
		wantInterval   time.Duration
		wantTimeout    time.Duration
	}{
		{
			name:           "with custom interval and timeout",
			workingDir:     "/tmp/test",
			checkInterval:  10 * time.Second,
			commandTimeout: 5 * time.Minute,
			wantInterval:   10 * time.Second,
			wantTimeout:    5 * time.Minute,
		},
		{
			name:           "with default interval and timeout",
			workingDir:     "/tmp/test",
			checkInterval:  0,
			commandTimeout: 0,
			wantInterval:   30 * time.Second,
			wantTimeout:    2 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := NewCIChecker(tt.workingDir, tt.checkInterval, tt.commandTimeout)
			require.NotNil(t, checker)

			concreteChecker, ok := checker.(*ciChecker)
			require.True(t, ok)
			assert.Equal(t, tt.workingDir, concreteChecker.workingDir)
			assert.Equal(t, tt.wantInterval, concreteChecker.checkInterval)
			assert.Equal(t, tt.wantTimeout, concreteChecker.commandTimeout)
		})
	}
}

func TestCIChecker_CheckCI_NotInstalled(t *testing.T) {
	checker := NewCIChecker("/nonexistent/path/that/should/not/exist", 1*time.Second, 1*time.Second)
	ctx := context.Background()

	result, err := checker.CheckCI(ctx, 123)
	require.Error(t, err)
	// Result is nil on error since CheckCI returns nil when checkCIOnce fails
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to check CI status")
}

func TestCIChecker_CheckCI_NoPR(t *testing.T) {
	// This test verifies the error handling when gh pr checks fails
	// Running in /tmp (non-git directory) will cause an error
	checker := NewCIChecker("/tmp", 1*time.Second, 1*time.Second)
	ctx := context.Background()

	result, err := checker.CheckCI(ctx, 0)
	require.Error(t, err)
	// Result is nil on error since CheckCI returns nil when checkCIOnce fails
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to check CI status")
}

func TestParseCIOutput_PendingStatus(t *testing.T) {
	// Test that pending status from gh pr checks (exit code 8) is handled correctly
	// Exit code 8 means "checks pending" - when there's output, parse it
	output := `Vercel Preview Comments	pass	0	https://vercel.com/github
Test go/aninexus-gateway / Lint go/aninexus-gateway	pending	0	https://github.com/example/actions/runs/123
Test go/aninexus-gateway / Test go/aninexus-gateway	pending	0	https://github.com/example/actions/runs/123
Vercel – nooxac-gateway	pass	0	https://vercel.com/example	Deployment has completed`

	status, failedJobs := parseCIOutput(output)
	assert.Equal(t, "pending", status)
	assert.Empty(t, failedJobs)
}

func TestCIChecker_WaitForCI_Timeout(t *testing.T) {
	checker := NewCIChecker("/nonexistent/path/that/should/not/exist", 100*time.Millisecond, 1*time.Second)
	ctx := context.Background()

	result, err := checker.WaitForCI(ctx, 123, 200*time.Millisecond)
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestFilterE2EFailures(t *testing.T) {
	tests := []struct {
		name       string
		result     *CIResult
		e2ePattern string
		want       *CIResult
	}{
		{
			name: "no failures",
			result: &CIResult{
				Passed:     true,
				Status:     "success",
				FailedJobs: []string{},
				Output:     "all passed",
			},
			e2ePattern: "e2e|E2E",
			want: &CIResult{
				Passed:     true,
				Status:     "success",
				FailedJobs: []string{},
				Output:     "all passed",
			},
		},
		{
			name: "only e2e failures",
			result: &CIResult{
				Passed:     false,
				Status:     "failure",
				FailedJobs: []string{"test-e2e", "integration-test"},
				Output:     "e2e tests failed",
			},
			e2ePattern: "e2e|integration",
			want: &CIResult{
				Passed:     true,
				Status:     "failure",
				FailedJobs: []string{},
				Output:     "e2e tests failed",
			},
		},
		{
			name: "mixed failures with e2e",
			result: &CIResult{
				Passed:     false,
				Status:     "failure",
				FailedJobs: []string{"test-unit", "test-e2e", "lint"},
				Output:     "multiple failures",
			},
			e2ePattern: "e2e|E2E",
			want: &CIResult{
				Passed:     false,
				Status:     "failure",
				FailedJobs: []string{"test-unit", "lint"},
				Output:     "multiple failures",
			},
		},
		{
			name: "only non-e2e failures",
			result: &CIResult{
				Passed:     false,
				Status:     "failure",
				FailedJobs: []string{"test-unit", "lint"},
				Output:     "unit tests failed",
			},
			e2ePattern: "e2e|E2E",
			want: &CIResult{
				Passed:     false,
				Status:     "failure",
				FailedJobs: []string{"test-unit", "lint"},
				Output:     "unit tests failed",
			},
		},
		{
			name: "case insensitive e2e pattern",
			result: &CIResult{
				Passed:     false,
				Status:     "failure",
				FailedJobs: []string{"test-E2E", "test-Integration"},
				Output:     "e2e tests failed",
			},
			e2ePattern: "e2e|E2E|integration|Integration",
			want: &CIResult{
				Passed:     true,
				Status:     "failure",
				FailedJobs: []string{},
				Output:     "e2e tests failed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterE2EFailures(tt.result, tt.e2ePattern)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCheckCI_ContextCancellation(t *testing.T) {
	checker := NewCIChecker("/nonexistent/path", 1*time.Second, 1*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := checker.CheckCI(ctx, 0)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, context.Canceled, err)
}

func TestCheckCI_IsolatedCommandContext(t *testing.T) {
	// Test that the command uses its own isolated context (commandTimeout)
	// rather than inheriting from the parent context
	checker := NewCIChecker("/nonexistent/path", 1*time.Second, 1*time.Second)
	ctx := context.Background()

	result, err := checker.CheckCI(ctx, 0)
	require.Error(t, err)
	// Result is nil on error since CheckCI returns nil when checkCIOnce fails
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to check CI status")
}

func TestWaitForCIWithOptions_ParentContextCancellation(t *testing.T) {
	// Test that context cancellation during the initial delay is handled properly
	checker := NewCIChecker("/nonexistent/path", 100*time.Millisecond, 1*time.Second)
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel quickly during the 1-minute initial delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	result, err := checker.WaitForCIWithOptions(ctx, 0, 10*time.Second, CheckCIOptions{})
	require.Error(t, err)
	assert.Nil(t, result)
	// Context cancellation should be detected during initial delay
	assert.Equal(t, context.Canceled, err)
}
