// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package gogit

import (
	"context"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// createInMemoryRepository creates a go-git repository using in-memory storage for unit testing.
func createInMemoryRepository(tb testing.TB, bare bool) *git.Repository {
	tb.Helper()

	var repo *git.Repository

	var err error

	storage := memory.NewStorage()

	if bare {
		repo, err = git.Init(storage, nil)
	} else {
		fs := memfs.New()
		repo, err = git.Init(storage, fs)
	}

	require.NoError(tb, err)

	return repo
}

// TestNew tests adapter creation.
func TestNew(t *testing.T) {
	t.Parallel()

	config := ports.GitConfig{
		UserName:  "test-user",
		UserEmail: "test@example.com",
	}

	adapter := New(config)
	assert.NotNil(t, adapter)
	assert.Equal(t, config, adapter.config)
}

// TestGetName tests adapter name.
func TestGetName(t *testing.T) {
	t.Parallel()

	adapter := New(ports.GitConfig{})
	assert.Equal(t, "go-git", adapter.GetName())
}

// TestSupportsURL tests URL support detection.
func TestSupportsURL(t *testing.T) {
	t.Parallel()

	adapter := New(ports.GitConfig{})

	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "https url",
			url:  "https://github.com/test/repo.git",
			want: true,
		},
		{
			name: "http url",
			url:  "http://gitlab.com/test/repo.git",
			want: true,
		},
		{
			name: "ssh url",
			url:  "git@github.com:test/repo.git",
			want: true,
		},
		{
			name: "file url",
			url:  "file:///path/to/repo.git",
			want: true,
		},
		{
			name: "local path",
			url:  "/path/to/repo.git",
			want: true,
		},
		{
			name: "ftp url",
			url:  "ftp://server.com/repo.git",
			want: false,
		},
		{
			name: "invalid url",
			url:  "invalid://url",
			want: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := adapter.SupportsURL(testCase.url)
			assert.Equal(t, testCase.want, got)
		})
	}
}

// TestCleanup tests cleanup operation.
func TestCleanup(t *testing.T) {
	t.Parallel()

	adapter := New(ports.GitConfig{})
	ctx := context.Background()

	// Test cleanup - for go-git, this is a no-op
	err := adapter.Cleanup(ctx, "/any/path")
	require.NoError(t, err)
}

// TestRepository_BasicProperties tests basic repository properties.
func TestRepository_BasicProperties(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		bare     bool
		expected bool
	}{
		{
			name:     "working repository",
			bare:     false,
			expected: false,
		},
		{
			name:     "bare repository",
			bare:     true,
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := createInMemoryRepository(t, test.bare)

			goGitRepo := &Repository{
				repo: repo,
				path: "/virtual/test/path", // Virtual path for in-memory
			}

			// Test basic properties
			assert.Equal(t, "/virtual/test/path", goGitRepo.Path())
			assert.Equal(t, "path", goGitRepo.Name()) // Base of path
			assert.Empty(t, goGitRepo.URL())          // No remote URL for in-memory
			assert.Equal(t, test.expected, goGitRepo.IsBare())
			assert.True(t, goGitRepo.IsClean())     // In-memory repo starts clean
			assert.False(t, goGitRepo.HasChanges()) // No changes initially

			// Test close
			err := goGitRepo.Close()
			require.NoError(t, err)
		})
	}
}

// TestRepository_BranchOperations tests branch operations.
func TestRepository_BranchOperations(t *testing.T) {
	t.Parallel()

	repo := createInMemoryRepository(t, false)
	ctx := context.Background()

	goGitRepo := &Repository{
		repo: repo,
		path: "/virtual/test/path",
	}

	defer func() { _ = goGitRepo.Close() }()

	// Initially no branches
	branches, err := goGitRepo.ListBranches(ctx)
	require.NoError(t, err)
	assert.Empty(t, branches)

	// Current branch should return error for empty repo (no HEAD)
	_, err = goGitRepo.CurrentBranch()
	require.Error(t, err) // Expected error for empty repository
}

// TestRepository_RemoteOperations tests remote operations.
func TestRepository_RemoteOperations(t *testing.T) {
	t.Parallel()

	repo := createInMemoryRepository(t, false)
	ctx := context.Background()

	goGitRepo := &Repository{
		repo: repo,
		path: "/virtual/test/path",
	}

	defer func() { _ = goGitRepo.Close() }()

	// Initially no remotes
	remotes, err := goGitRepo.ListRemotes(ctx)
	require.NoError(t, err)
	assert.Empty(t, remotes)

	// Add a remote
	err = goGitRepo.AddRemote(ctx, "origin", "https://github.com/test/repo.git")
	require.NoError(t, err)

	// List remotes should now show the added remote
	remotes, err = goGitRepo.ListRemotes(ctx)
	require.NoError(t, err)
	require.Len(t, remotes, 1)
	assert.Equal(t, "origin", remotes[0].Name)
	assert.Equal(t, "https://github.com/test/repo.git", remotes[0].URL)

	// Add second remote
	err = goGitRepo.AddRemote(ctx, "upstream", "https://github.com/upstream/repo.git")
	require.NoError(t, err)

	// List remotes should show both
	remotes, err = goGitRepo.ListRemotes(ctx)
	require.NoError(t, err)
	assert.Len(t, remotes, 2)

	// Remove a remote
	err = goGitRepo.RemoveRemote(ctx, "upstream")
	require.NoError(t, err)

	// Should be back to one remote
	remotes, err = goGitRepo.ListRemotes(ctx)
	require.NoError(t, err)
	require.Len(t, remotes, 1)
	assert.Equal(t, "origin", remotes[0].Name)

	// Update remote URL
	err = goGitRepo.UpdateRemote(ctx, "origin", "https://github.com/new/repo.git")
	require.NoError(t, err)

	// Verify updated URL
	remotes, err = goGitRepo.ListRemotes(ctx)
	require.NoError(t, err)
	require.Len(t, remotes, 1)
	assert.Equal(t, "https://github.com/new/repo.git", remotes[0].URL)
}

// TestRepository_TagOperations tests tag operations.
func TestRepository_TagOperations(t *testing.T) {
	t.Parallel()

	repo := createInMemoryRepository(t, false)
	ctx := context.Background()

	goGitRepo := &Repository{
		repo: repo,
		path: "/virtual/test/path",
	}

	defer func() { _ = goGitRepo.Close() }()

	// Initially no tags
	tags, err := goGitRepo.ListTags(ctx)
	require.NoError(t, err)
	assert.Empty(t, tags)

	// Note: Creating tags requires commits, which requires more setup
	// For pure unit tests, we focus on the interface behavior
}

// TestRepository_Status tests status operations.
func TestRepository_Status(t *testing.T) {
	t.Parallel()

	repo := createInMemoryRepository(t, false)
	ctx := context.Background()

	goGitRepo := &Repository{
		repo: repo,
		path: "/virtual/test/path",
	}

	defer func() { _ = goGitRepo.Close() }()

	// Get status - should be clean initially
	status, err := goGitRepo.Status(ctx)
	require.NoError(t, err)
	assert.NotNil(t, status)
	assert.True(t, status.IsClean)
	assert.Empty(t, status.Modified)
	assert.Empty(t, status.Added)
	assert.Empty(t, status.Deleted)
	assert.Empty(t, status.Untracked)
}

// BenchmarkRepository_Operations benchmarks core operations with in-memory storage.
func BenchmarkRepository_Operations(b *testing.B) {
	repo := createInMemoryRepository(b, false)
	ctx := context.Background()

	goGitRepo := &Repository{
		repo: repo,
		path: "/virtual/bench/path",
	}

	defer func() { _ = goGitRepo.Close() }()

	b.ResetTimer()

	for range b.N {
		remoteName := "bench-remote"
		remoteURL := "https://github.com/test/repo.git"

		// Add remote
		_ = goGitRepo.AddRemote(ctx, remoteName, remoteURL)

		// List remotes
		_, _ = goGitRepo.ListRemotes(ctx)

		// Remove remote
		_ = goGitRepo.RemoveRemote(ctx, remoteName)
	}
}
