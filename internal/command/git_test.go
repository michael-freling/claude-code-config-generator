package command

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewGitRunner(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRunner := NewMockRunner(ctrl)
	got := NewGitRunner(mockRunner)

	require.NotNil(t, got)
}

func TestGitRunner_GetCurrentBranch(t *testing.T) {
	tests := []struct {
		name        string
		dir         string
		setupMock   func(*MockRunner)
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name: "returns current branch successfully",
			dir:  "/test/repo",
			setupMock: func(m *MockRunner) {
				m.EXPECT().
					RunInDir(gomock.Any(), "/test/repo", "git", "rev-parse", "--abbrev-ref", "HEAD").
					Return("main", "", nil)
			},
			want:    "main",
			wantErr: false,
		},
		{
			name: "returns trimmed branch name",
			dir:  "/test/repo",
			setupMock: func(m *MockRunner) {
				m.EXPECT().
					RunInDir(gomock.Any(), "/test/repo", "git", "rev-parse", "--abbrev-ref", "HEAD").
					Return("  feature-branch  ", "", nil)
			},
			want:    "feature-branch",
			wantErr: false,
		},
		{
			name: "fails when git command fails",
			dir:  "/test/repo",
			setupMock: func(m *MockRunner) {
				m.EXPECT().
					RunInDir(gomock.Any(), "/test/repo", "git", "rev-parse", "--abbrev-ref", "HEAD").
					Return("", "fatal: not a git repository", fmt.Errorf("exit status 128"))
			},
			wantErr:     true,
			errContains: "failed to get current branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRunner := NewMockRunner(ctrl)
			tt.setupMock(mockRunner)

			gitRunner := NewGitRunner(mockRunner)
			ctx := context.Background()

			got, err := gitRunner.GetCurrentBranch(ctx, tt.dir)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGitRunner_Push(t *testing.T) {
	tests := []struct {
		name        string
		dir         string
		branch      string
		setupMock   func(*MockRunner)
		wantErr     bool
		errContains string
	}{
		{
			name:   "pushes branch successfully",
			dir:    "/test/repo",
			branch: "feature-branch",
			setupMock: func(m *MockRunner) {
				m.EXPECT().
					RunInDir(gomock.Any(), "/test/repo", "git", "push", "-u", "origin", "feature-branch").
					Return("", "", nil)
			},
			wantErr: false,
		},
		{
			name:        "fails when branch name is empty",
			dir:         "/test/repo",
			branch:      "",
			setupMock:   func(m *MockRunner) {},
			wantErr:     true,
			errContains: "branch name cannot be empty",
		},
		{
			name:   "fails when git push fails",
			dir:    "/test/repo",
			branch: "feature-branch",
			setupMock: func(m *MockRunner) {
				m.EXPECT().
					RunInDir(gomock.Any(), "/test/repo", "git", "push", "-u", "origin", "feature-branch").
					Return("", "fatal: repository not found", fmt.Errorf("exit status 128"))
			},
			wantErr:     true,
			errContains: "failed to push branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRunner := NewMockRunner(ctrl)
			tt.setupMock(mockRunner)

			gitRunner := NewGitRunner(mockRunner)
			ctx := context.Background()

			err := gitRunner.Push(ctx, tt.dir, tt.branch)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestGitRunner_WorktreeAdd(t *testing.T) {
	tests := []struct {
		name        string
		dir         string
		path        string
		branch      string
		setupMock   func(*MockRunner)
		wantErr     bool
		errContains string
	}{
		{
			name:   "creates worktree successfully",
			dir:    "/test/repo",
			path:   "/test/worktree",
			branch: "feature-branch",
			setupMock: func(m *MockRunner) {
				m.EXPECT().
					RunInDir(gomock.Any(), "/test/repo", "git", "worktree", "add", "/test/worktree", "-b", "feature-branch").
					Return("", "", nil)
			},
			wantErr: false,
		},
		{
			name:        "fails when path is empty",
			dir:         "/test/repo",
			path:        "",
			branch:      "feature-branch",
			setupMock:   func(m *MockRunner) {},
			wantErr:     true,
			errContains: "worktree path cannot be empty",
		},
		{
			name:        "fails when branch is empty",
			dir:         "/test/repo",
			path:        "/test/worktree",
			branch:      "",
			setupMock:   func(m *MockRunner) {},
			wantErr:     true,
			errContains: "branch name cannot be empty",
		},
		{
			name:   "fails when branch already exists",
			dir:    "/test/repo",
			path:   "/test/worktree",
			branch: "existing-branch",
			setupMock: func(m *MockRunner) {
				m.EXPECT().
					RunInDir(gomock.Any(), "/test/repo", "git", "worktree", "add", "/test/worktree", "-b", "existing-branch").
					Return("", "fatal: A branch named 'existing-branch' already exists", fmt.Errorf("exit status 128"))
			},
			wantErr:     true,
			errContains: "branch existing-branch already exists",
		},
		{
			name:   "fails when git worktree add fails with other error",
			dir:    "/test/repo",
			path:   "/test/worktree",
			branch: "feature-branch",
			setupMock: func(m *MockRunner) {
				m.EXPECT().
					RunInDir(gomock.Any(), "/test/repo", "git", "worktree", "add", "/test/worktree", "-b", "feature-branch").
					Return("", "fatal: some other error", fmt.Errorf("exit status 128"))
			},
			wantErr:     true,
			errContains: "failed to create worktree",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRunner := NewMockRunner(ctrl)
			tt.setupMock(mockRunner)

			gitRunner := NewGitRunner(mockRunner)
			ctx := context.Background()

			err := gitRunner.WorktreeAdd(ctx, tt.dir, tt.path, tt.branch)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestGitRunner_WorktreeRemove(t *testing.T) {
	tests := []struct {
		name        string
		dir         string
		path        string
		setupMock   func(*MockRunner)
		wantErr     bool
		errContains string
	}{
		{
			name: "removes worktree successfully",
			dir:  "/test/repo",
			path: "/test/worktree",
			setupMock: func(m *MockRunner) {
				m.EXPECT().
					RunInDir(gomock.Any(), "/test/repo", "git", "worktree", "remove", "/test/worktree").
					Return("", "", nil)
			},
			wantErr: false,
		},
		{
			name:        "fails when path is empty",
			dir:         "/test/repo",
			path:        "",
			setupMock:   func(m *MockRunner) {},
			wantErr:     true,
			errContains: "worktree path cannot be empty",
		},
		{
			name: "fails when git worktree remove fails",
			dir:  "/test/repo",
			path: "/test/worktree",
			setupMock: func(m *MockRunner) {
				m.EXPECT().
					RunInDir(gomock.Any(), "/test/repo", "git", "worktree", "remove", "/test/worktree").
					Return("", "fatal: '/test/worktree' is not a working tree", fmt.Errorf("exit status 128"))
			},
			wantErr:     true,
			errContains: "failed to remove worktree",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRunner := NewMockRunner(ctrl)
			tt.setupMock(mockRunner)

			gitRunner := NewGitRunner(mockRunner)
			ctx := context.Background()

			err := gitRunner.WorktreeRemove(ctx, tt.dir, tt.path)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
		})
	}
}
