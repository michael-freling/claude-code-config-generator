package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPRManager(t *testing.T) {
	tests := []struct {
		name       string
		workingDir string
	}{
		{
			name:       "creates manager with working directory",
			workingDir: "/path/to/repo",
		},
		{
			name:       "creates manager with empty working directory",
			workingDir: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPRManager(tt.workingDir)
			require.NotNil(t, got)

			manager, ok := got.(*prManager)
			require.True(t, ok)
			assert.Equal(t, tt.workingDir, manager.workingDir)
		})
	}
}

func TestExtractPRNumberFromURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    int
		wantErr bool
	}{
		{
			name: "standard GitHub URL",
			url:  "https://github.com/owner/repo/pull/123",
			want: 123,
		},
		{
			name: "GitHub Enterprise URL",
			url:  "https://github.enterprise.com/owner/repo/pull/456",
			want: 456,
		},
		{
			name: "URL with trailing content",
			url:  "https://github.com/owner/repo/pull/789/files",
			want: 789,
		},
		{
			name: "URL with query params",
			url:  "https://github.com/owner/repo/pull/101?tab=files",
			want: 101,
		},
		{
			name: "URL with large PR number",
			url:  "https://github.com/owner/repo/pull/99999",
			want: 99999,
		},
		{
			name:    "invalid URL without PR number",
			url:     "https://github.com/owner/repo",
			wantErr: true,
		},
		{
			name:    "invalid URL with non-numeric PR",
			url:     "https://github.com/owner/repo/pull/abc",
			wantErr: true,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "URL without pull segment",
			url:     "https://github.com/owner/repo/issues/123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractPRNumberFromURL(tt.url)

			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, 0, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
