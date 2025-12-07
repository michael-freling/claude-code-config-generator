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
			name:           "with default interval",
			workingDir:     "/tmp/test",
			checkInterval:  0,
			commandTimeout: 5 * time.Minute,
			wantInterval:   30 * time.Second,
			wantTimeout:    5 * time.Minute,
		},
		{
			name:           "with default timeout",
			workingDir:     "/tmp/test",
			checkInterval:  10 * time.Second,
			commandTimeout: 0,
			wantInterval:   10 * time.Second,
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
	checker := NewCIChecker("/nonexistent/path/that/should/not/exist", 1*time.Second, 10*time.Second)
	ctx := context.Background()

	result, err := checker.CheckCI(ctx, 123)
	require.Error(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Passed)
}

func TestCIChecker_CheckCI_NoPR(t *testing.T) {
	// This test verifies the error handling when gh pr checks fails
	// Running in /tmp (non-git directory) will cause an error
	checker := NewCIChecker("/tmp", 1*time.Second, 10*time.Second)
	ctx := context.Background()

	result, err := checker.CheckCI(ctx, 0)
	require.Error(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Passed)
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
	if testing.Short() {
		t.Skip("skipping test with 1 minute initial delay in short mode")
	}

	checker := NewCIChecker("/nonexistent/path/that/should/not/exist", 100*time.Millisecond, 10*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	result, err := checker.WaitForCI(ctx, 123, 2*time.Minute)
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
	checker := NewCIChecker("/tmp", 1*time.Second, 30*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := checker.CheckCI(ctx, 0)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.NotNil(t, result)
	assert.False(t, result.Passed)
}

func TestCheckCI_IsolatedCommandContext(t *testing.T) {
	checker := NewCIChecker("/tmp", 1*time.Second, 50*time.Millisecond)
	ctx := context.Background()

	result, err := checker.CheckCI(ctx, 0)

	require.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Passed)
}

// TestWaitForCIWithOptions_ParentContextCancellation is skipped because
// WaitForCIWithOptions has a hardcoded 1-minute initial delay that makes
// unit testing impractical. This should be tested in integration tests.

func TestParseCIOutput_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		output         string
		wantStatus     string
		wantFailedJobs []string
	}{
		{
			name:           "only whitespace",
			output:         "   \n\n  \t  \n   ",
			wantStatus:     "pending",
			wantFailedJobs: []string{},
		},
		{
			name: "single field line",
			output: `✓
			incomplete`,
			wantStatus:     "pending",
			wantFailedJobs: []string{},
		},
		{
			name: "multi-word job names",
			output: `✓ Build and Test Application
✗ Run Integration Tests Suite
✓ Deploy to Staging Environment`,
			wantStatus:     "failure",
			wantFailedJobs: []string{"Run Integration Tests Suite"},
		},
		{
			name: "text status keywords",
			output: `pass build
fail test
success lint`,
			wantStatus:     "failure",
			wantFailedJobs: []string{"test"},
		},
		{
			name: "mixed status symbols and text",
			output: `✓ build
fail test
success lint`,
			wantStatus:     "failure",
			wantFailedJobs: []string{"test"},
		},
		{
			name: "queued status",
			output: `✓ build
queued test
✓ lint`,
			wantStatus:     "pending",
			wantFailedJobs: []string{},
		},
		{
			name: "in_progress status",
			output: `✓ build
in_progress test
✓ lint`,
			wantStatus:     "pending",
			wantFailedJobs: []string{},
		},
		{
			name: "asterisk pending marker",
			output: `✓ build
* test
✓ lint`,
			wantStatus:     "pending",
			wantFailedJobs: []string{},
		},
		{
			name: "all failures",
			output: `✗ build
✗ test
✗ lint`,
			wantStatus:     "failure",
			wantFailedJobs: []string{"build", "test", "lint"},
		},
		{
			name: "only pending jobs",
			output: `○ build
○ test
○ lint`,
			wantStatus:     "pending",
			wantFailedJobs: []string{},
		},
		{
			name: "no completed jobs",
			output: `pending build
queued test`,
			wantStatus:     "pending",
			wantFailedJobs: []string{},
		},
		{
			name: "success after failure when pending present",
			output: `✓ build
✗ test-unit
○ test-e2e`,
			wantStatus:     "pending",
			wantFailedJobs: []string{"test-unit"},
		},
		{
			name: "lines with extra spaces get normalized",
			output: `   ✓     build    with    spaces
  ✗    test     failed   `,
			wantStatus:     "failure",
			wantFailedJobs: []string{"test failed"},
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

func TestFilterE2EFailures_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		result     *CIResult
		e2ePattern string
		want       *CIResult
	}{
		{
			name: "invalid regex pattern",
			result: &CIResult{
				Passed:     false,
				Status:     "failure",
				FailedJobs: []string{"test-e2e", "test-unit"},
				Output:     "tests failed",
			},
			e2ePattern: "[invalid(",
			want: &CIResult{
				Passed:     false,
				Status:     "failure",
				FailedJobs: []string{"test-e2e", "test-unit"},
				Output:     "tests failed",
			},
		},
		{
			name: "empty pattern matches everything",
			result: &CIResult{
				Passed:     false,
				Status:     "failure",
				FailedJobs: []string{"test-e2e"},
				Output:     "tests failed",
			},
			e2ePattern: "",
			want: &CIResult{
				Passed:     true,
				Status:     "failure",
				FailedJobs: []string{},
				Output:     "tests failed",
			},
		},
		{
			name: "pattern matches nothing",
			result: &CIResult{
				Passed:     false,
				Status:     "failure",
				FailedJobs: []string{"test-unit", "lint"},
				Output:     "tests failed",
			},
			e2ePattern: "nonexistent",
			want: &CIResult{
				Passed:     false,
				Status:     "failure",
				FailedJobs: []string{"test-unit", "lint"},
				Output:     "tests failed",
			},
		},
		{
			name: "pattern matches all failures",
			result: &CIResult{
				Passed:     false,
				Status:     "failure",
				FailedJobs: []string{"integration-api", "integration-db"},
				Output:     "integration tests failed",
			},
			e2ePattern: "integration",
			want: &CIResult{
				Passed:     true,
				Status:     "failure",
				FailedJobs: []string{},
				Output:     "integration tests failed",
			},
		},
		{
			name: "complex pattern with alternation",
			result: &CIResult{
				Passed:     false,
				Status:     "failure",
				FailedJobs: []string{"e2e-smoke", "E2E-full", "integration-test", "unit-test"},
				Output:     "multiple test failures",
			},
			e2ePattern: "(e2e|E2E|integration)",
			want: &CIResult{
				Passed:     false,
				Status:     "failure",
				FailedJobs: []string{"unit-test"},
				Output:     "multiple test failures",
			},
		},
		{
			name: "pattern at start of job name",
			result: &CIResult{
				Passed:     false,
				Status:     "failure",
				FailedJobs: []string{"e2e-browser-test", "test-unit"},
				Output:     "tests failed",
			},
			e2ePattern: "^e2e",
			want: &CIResult{
				Passed:     false,
				Status:     "failure",
				FailedJobs: []string{"test-unit"},
				Output:     "tests failed",
			},
		},
		{
			name: "pattern at end of job name",
			result: &CIResult{
				Passed:     false,
				Status:     "failure",
				FailedJobs: []string{"browser-test-e2e", "unit-test"},
				Output:     "tests failed",
			},
			e2ePattern: "e2e$",
			want: &CIResult{
				Passed:     false,
				Status:     "failure",
				FailedJobs: []string{"unit-test"},
				Output:     "tests failed",
			},
		},
		{
			name: "empty failed jobs list",
			result: &CIResult{
				Passed:     false,
				Status:     "failure",
				FailedJobs: []string{},
				Output:     "unknown failure",
			},
			e2ePattern: "e2e",
			want: &CIResult{
				Passed:     true,
				Status:     "failure",
				FailedJobs: []string{},
				Output:     "unknown failure",
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

// TestWaitForCIWithOptions_CustomE2EPattern, TestWaitForCIWithOptions_DefaultTimeout,
// and TestWaitForCI_ContextCancellation are skipped because WaitForCI methods have a
// hardcoded 1-minute initial delay that makes unit testing impractical.
// These should be tested in integration tests.
