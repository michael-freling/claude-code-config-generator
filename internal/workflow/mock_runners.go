package workflow

import (
	"context"

	"github.com/michael-freling/claude-code-tools/internal/command"
	"github.com/stretchr/testify/mock"
)

// MockCommandRunner is a mock implementation of command.Runner
type MockCommandRunner struct {
	mock.Mock
}

// Ensure MockCommandRunner implements command.Runner
var _ command.Runner = (*MockCommandRunner)(nil)

func (m *MockCommandRunner) Run(ctx context.Context, name string, args ...string) (string, string, error) {
	callArgs := []interface{}{ctx, name}
	for _, arg := range args {
		callArgs = append(callArgs, arg)
	}
	mockArgs := m.Called(callArgs...)
	return mockArgs.String(0), mockArgs.String(1), mockArgs.Error(2)
}

func (m *MockCommandRunner) RunInDir(ctx context.Context, dir string, name string, args ...string) (string, string, error) {
	callArgs := []interface{}{ctx, dir, name}
	for _, arg := range args {
		callArgs = append(callArgs, arg)
	}
	mockArgs := m.Called(callArgs...)
	return mockArgs.String(0), mockArgs.String(1), mockArgs.Error(2)
}

// MockGitRunner is a mock implementation of command.GitRunner
type MockGitRunner struct {
	mock.Mock
}

// Ensure MockGitRunner implements command.GitRunner
var _ command.GitRunner = (*MockGitRunner)(nil)

func (m *MockGitRunner) GetCurrentBranch(ctx context.Context, dir string) (string, error) {
	args := m.Called(ctx, dir)
	return args.String(0), args.Error(1)
}

func (m *MockGitRunner) Push(ctx context.Context, dir string, branch string) error {
	args := m.Called(ctx, dir, branch)
	return args.Error(0)
}

func (m *MockGitRunner) WorktreeAdd(ctx context.Context, dir string, path string, branch string) error {
	args := m.Called(ctx, dir, path, branch)
	return args.Error(0)
}

func (m *MockGitRunner) WorktreeRemove(ctx context.Context, dir string, path string) error {
	args := m.Called(ctx, dir, path)
	return args.Error(0)
}

// MockGhRunner is a mock implementation of command.GhRunner
type MockGhRunner struct {
	mock.Mock
}

// Ensure MockGhRunner implements command.GhRunner
var _ command.GhRunner = (*MockGhRunner)(nil)

func (m *MockGhRunner) PRCreate(ctx context.Context, dir string, title, body, head string) (string, error) {
	args := m.Called(ctx, dir, title, body, head)
	return args.String(0), args.Error(1)
}

func (m *MockGhRunner) PRView(ctx context.Context, dir string, jsonFields string, jqQuery string) (string, error) {
	args := m.Called(ctx, dir, jsonFields, jqQuery)
	return args.String(0), args.Error(1)
}

func (m *MockGhRunner) PRChecks(ctx context.Context, dir string, prNumber int, jsonFields string) (string, error) {
	args := m.Called(ctx, dir, prNumber, jsonFields)
	return args.String(0), args.Error(1)
}

func (m *MockGhRunner) GetPRBaseBranch(ctx context.Context, dir string, prNumber string) (string, error) {
	args := m.Called(ctx, dir, prNumber)
	return args.String(0), args.Error(1)
}

func (m *MockGhRunner) RunRerun(ctx context.Context, dir string, runID int64) error {
	args := m.Called(ctx, dir, runID)
	return args.Error(0)
}

func (m *MockGhRunner) GetLatestRunID(ctx context.Context, dir string, prNumber int) (int64, error) {
	args := m.Called(ctx, dir, prNumber)
	return args.Get(0).(int64), args.Error(1)
}
