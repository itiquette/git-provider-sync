// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package directory

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/adapters/filesystem"
	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/testutil"
)

const (
	testSourceDir = "/source"
)

// Create a mock GitConfig for testing.
func createMockGitConfig() ports.GitConfig {
	return ports.GitConfig{
		UserName:  "Test User",
		UserEmail: "test@example.com",
	}
}

// createTestAdapter creates an adapter with an in-memory filesystem for testing.
func createTestAdapter(testFS *testutil.TestFS) *Adapter {
	fs := testFS.GetFileSystem()

	return NewWithFileSystem(createMockGitConfig(), fs)
}

// Helper functions for tests

func createTempDirWithFiles(tb testing.TB) string {
	tb.Helper()

	// Note: This function needs to create real files because Clone operations
	// access the actual filesystem
	tempDir := tb.TempDir()

	// Create test files in real filesystem
	require.NoError(tb, os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("test content 1"), 0600))
	require.NoError(tb, os.MkdirAll(filepath.Join(tempDir, "subdir"), 0750))
	require.NoError(tb, os.WriteFile(filepath.Join(tempDir, "subdir", "file2.txt"), []byte("test content 2"), 0600))

	return tempDir
}

// Test Adapter constructor

func TestNew_ValidPath_CreatesDirectoryAdapter(t *testing.T) {
	t.Parallel()

	config := createMockGitConfig()
	adapter := New(config)

	require.NotNil(t, adapter)
	assert.Equal(t, config, adapter.config)
}

// Test Adapter methods

func TestAdapter_GetName(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	assert.Equal(t, "directory", adapter.GetName())
}

func TestAdapter_SupportsURL(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())

	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "file:// URL",
			url:      "file:///path/to/directory",
			expected: true,
		},
		{
			name:     "absolute path Unix",
			url:      "/absolute/path",
			expected: true,
		},
		{
			name:     "https URL",
			url:      "https://github.com/user/repo.git",
			expected: false,
		},
		{
			name:     "ssh URL",
			url:      "git@github.com:user/repo.git",
			expected: false,
		},
		{
			name:     "relative path",
			url:      "relative/path",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := adapter.SupportsURL(test.url)
			assert.Equal(t, test.expected, result)
		})
	}
}

// Test Clone operation

// TestAdapter_Clone_FileURL tests directory copying with file:// URLs
// using in-memory filesystem for significantly faster execution.
//
// Performance comparison (100 iterations):
//
//	Memory version: ~33µs/op, 66 allocs/op
//	Disk version:   ~450µs/op, 81 allocs/op
//	Speedup:        13.5x faster, 18% fewer allocations
//
// This demonstrates the massive performance gains from in-memory testing
// for I/O-heavy operations like directory copying.
func TestAdapter_Clone_FileURL(t *testing.T) {
	t.Parallel()

	// Use memory filesystem for entire test (no disk I/O)
	memFS := testutil.NewMemFS(t)
	fileSystem := filesystem.NewAferoFileSystem(memFS.Fs)

	// Create source directory with files in memory
	sourceDir := testSourceDir
	require.NoError(t, fileSystem.MkdirAll(sourceDir, 0750))
	require.NoError(t, fileSystem.WriteFile(filepath.Join(sourceDir, "file1.txt"), []byte("test content 1"), 0600))
	require.NoError(t, fileSystem.MkdirAll(filepath.Join(sourceDir, "subdir"), 0750))
	require.NoError(t, fileSystem.WriteFile(filepath.Join(sourceDir, "subdir", "file2.txt"), []byte("test content 2"), 0600))

	// Destination path in memory
	destDir := "/dest"

	// Create adapter with memory filesystem
	adapter := NewWithFileSystem(createMockGitConfig(), fileSystem)
	options := ports.CloneOptions{
		URL:  "file://" + sourceDir,
		Path: destDir,
	}

	repo, err := adapter.Clone(context.Background(), options)

	require.NoError(t, err)
	require.NotNil(t, repo)

	// Verify repository properties
	assert.Equal(t, destDir, repo.Path())
	assert.Equal(t, "file://"+sourceDir, repo.URL())
	assert.Equal(t, filepath.Base(destDir), repo.Name())

	// Verify files were copied in memory filesystem
	file1Content, err := fileSystem.ReadFile(filepath.Join(destDir, "file1.txt"))
	require.NoError(t, err)
	assert.Equal(t, "test content 1", string(file1Content))

	file2Content, err := fileSystem.ReadFile(filepath.Join(destDir, "subdir", "file2.txt"))
	require.NoError(t, err)
	assert.Equal(t, "test content 2", string(file2Content))
}

// TestAdapter_Clone_FileURL_Disk tests the same scenario using disk
// for comparison purposes (can be removed once memory version is proven).
func TestAdapter_Clone_FileURL_Disk(t *testing.T) {
	t.Parallel()

	// Original disk-based test
	sourceDir := createTempDirWithFiles(t)
	tempDir := t.TempDir()
	destDir := filepath.Join(tempDir, "dest")

	adapter := New(createMockGitConfig())
	options := ports.CloneOptions{
		URL:  "file://" + sourceDir,
		Path: destDir,
	}

	repo, err := adapter.Clone(context.Background(), options)

	require.NoError(t, err)
	require.NotNil(t, repo)

	assert.Equal(t, destDir, repo.Path())
	assert.Equal(t, "file://"+sourceDir, repo.URL())
	assert.Equal(t, filepath.Base(destDir), repo.Name())

	file1Content, err := os.ReadFile(filepath.Join(destDir, "file1.txt")) //nolint:gosec
	require.NoError(t, err)
	assert.Equal(t, "test content 1", string(file1Content))

	file2Content, err := os.ReadFile(filepath.Join(destDir, "subdir", "file2.txt")) //nolint:gosec
	require.NoError(t, err)
	assert.Equal(t, "test content 2", string(file2Content))
}

func TestAdapter_Clone_NonFileURL(t *testing.T) {
	t.Parallel()
	testFS := testutil.NewTestFS(t)

	destDir := testFS.TempDir("test")

	adapter := New(createMockGitConfig())
	options := ports.CloneOptions{
		URL:  "https://github.com/user/repo.git",
		Path: destDir,
	}

	repo, err := adapter.Clone(context.Background(), options)

	require.NoError(t, err)
	require.NotNil(t, repo)

	// Should create repository even without copying files
	assert.Equal(t, destDir, repo.Path())
	assert.Equal(t, "https://github.com/user/repo.git", repo.URL())
}

func TestAdapter_Clone_CreateDirectoryError(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())

	// Try to create a directory in a non-existent parent
	options := ports.CloneOptions{
		URL:  "file:///some/source",
		Path: "/non-existent-parent/child/directory",
	}

	_, err := adapter.Clone(context.Background(), options)

	// Should fail to create directory
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create directory")
}

// Test Open operation

func TestAdapter_Open_ExistingDirectory(t *testing.T) {
	t.Parallel()

	// Use memory filesystem for faster execution
	memFS := testutil.NewMemFS(t)
	fileSystem := filesystem.NewAferoFileSystem(memFS.Fs)

	// Create test directory with files in memory
	tempDir := "/test-dir"
	require.NoError(t, fileSystem.MkdirAll(tempDir, 0750))
	require.NoError(t, fileSystem.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("test content 1"), 0600))
	require.NoError(t, fileSystem.MkdirAll(filepath.Join(tempDir, "subdir"), 0750))
	require.NoError(t, fileSystem.WriteFile(filepath.Join(tempDir, "subdir", "file2.txt"), []byte("test content 2"), 0600))

	adapter := NewWithFileSystem(createMockGitConfig(), fileSystem)

	repo, err := adapter.Open(context.Background(), tempDir)

	require.NoError(t, err)
	require.NotNil(t, repo)
	assert.Equal(t, tempDir, repo.Path())
	assert.Empty(t, repo.URL())
}

func TestAdapter_Open_NonExistentDirectory(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())

	_, err := adapter.Open(context.Background(), "/non-existent-directory")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to access directory")
}

func TestAdapter_Open_FileInsteadOfDirectory(t *testing.T) {
	t.Parallel()
	testFS := testutil.NewTestFS(t)

	// Create a file in memory filesystem
	tempFile := "/test-file.txt"
	testFS.WriteFile(tempFile, "test content")

	adapter := createTestAdapter(testFS)

	_, err := adapter.Open(context.Background(), tempFile)

	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrPathNotDirectory)
}

// Test Init operation

func TestAdapter_Init(t *testing.T) {
	t.Parallel()
	testFS := testutil.NewTestFS(t)

	tempDir := testFS.TempDir("test")

	newRepoPath := filepath.Join(tempDir, "new-repo")

	adapter := createTestAdapter(testFS)

	repo, err := adapter.Init(context.Background(), newRepoPath, ports.InitOptions{})

	require.NoError(t, err)
	require.NotNil(t, repo)
	assert.Equal(t, newRepoPath, repo.Path())

	// Verify directory was created
	require.True(t, testFS.Exists(newRepoPath))
	// TestFS doesn't have Stat, but we can use AssertDirExists
	testFS.AssertDirExists(newRepoPath)
}

// Test Cleanup operation

func TestAdapter_Cleanup(t *testing.T) {
	t.Parallel()

	// Use memory filesystem for entire test
	memFS := testutil.NewMemFS(t)
	fileSystem := filesystem.NewAferoFileSystem(memFS.Fs)

	// Create test directory in memory
	tempDir := "/test-cleanup"
	require.NoError(t, fileSystem.MkdirAll(tempDir, 0750))
	require.NoError(t, fileSystem.WriteFile(filepath.Join(tempDir, "file.txt"), []byte("content"), 0600))

	adapter := NewWithFileSystem(createMockGitConfig(), fileSystem)

	err := adapter.Cleanup(context.Background(), tempDir)

	require.NoError(t, err)

	// Verify directory was removed from memory
	exists, err := fileSystem.Exists(tempDir)
	require.NoError(t, err)
	require.False(t, exists)
}

func TestAdapter_Cleanup_NonExistentPath(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())

	err := adapter.Cleanup(context.Background(), "/non-existent-path")

	// Should not error for non-existent paths
	require.NoError(t, err)
}

// Test Repository implementation

func TestRepository_BehavesAsCleanDirectoryRepository(t *testing.T) {
	t.Parallel()
	testFS := testutil.NewTestFS(t)

	// Test that a directory repository correctly identifies its state
	// And can be used for directory-based operations
	tempDir := createTempDirWithFiles(t)

	repo := &Repository{
		path: tempDir,
		url:  "file://" + tempDir,
	}

	// Behavioral test: Repository should identify as clean when no changes
	assert.True(t, repo.IsClean(), "Fresh directory repo should be clean")
	assert.False(t, repo.HasChanges(), "Fresh directory repo should have no changes")

	// Behavioral test: Repository should correctly identify its type
	assert.False(t, repo.IsBare(), "Directory adapter creates working repositories, not bare")

	// Behavioral test: Path should be usable for file operations
	testFile := filepath.Join(repo.Path(), "test.txt")
	testFS.WriteFile(testFile, "test")

	// Verify file exists
	require.True(t, testFS.Exists(testFile), "File should exist in repository")

	// Behavioral test: Name should be suitable for display/logging
	assert.NotEmpty(t, repo.Name(), "Repository should have a displayable name")
	assert.Equal(t, filepath.Base(tempDir), repo.Name(), "Name should be base of path for directory repos")
}

func TestRepository_CurrentBranch(t *testing.T) {
	t.Parallel()

	repo := &Repository{path: "/test/path"}

	branch, err := repo.CurrentBranch()

	require.NoError(t, err)
	assert.Equal(t, "main", branch)
}

func TestRepository_ListBranches(t *testing.T) {
	t.Parallel()

	repo := &Repository{path: "/test/path"}

	branches, err := repo.ListBranches(context.Background())

	require.NoError(t, err)
	require.Len(t, branches, 1)

	branch := branches[0]
	assert.Equal(t, "main", branch.Name)
	assert.Equal(t, "0000000000000000000000000000000000000000", branch.Hash)
	assert.False(t, branch.IsRemote)
	assert.True(t, branch.IsCurrent)
}

func TestRepository_BranchOperations_NotSupported(t *testing.T) {
	t.Parallel()

	repo := &Repository{path: "/test/path"}
	ctx := context.Background()

	tests := []struct {
		name string
		op   func() error
	}{
		{
			name: "CreateBranch",
			op:   func() error { return repo.CreateBranch(ctx, "new-branch", "main") },
		},
		{
			name: "CheckoutBranch",
			op:   func() error { return repo.CheckoutBranch(ctx, "main") },
		},
		{
			name: "DeleteBranch",
			op:   func() error { return repo.DeleteBranch(ctx, "branch", false) },
		},
		{
			name: "SetDefaultBranch",
			op:   func() error { return repo.SetDefaultBranch(ctx, "main") },
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.op()
			require.ErrorIs(t, err, domain.ErrBranchOpsNotSupported)
		})
	}
}

func TestRepository_ListRemotes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		url          string
		expectedLen  int
		expectedName string
	}{
		{
			name:         "repository with URL",
			url:          "file:///source/path",
			expectedLen:  1,
			expectedName: "origin",
		},
		{
			name:        "repository without URL",
			url:         "",
			expectedLen: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := &Repository{
				path: "/test/path",
				url:  test.url,
			}

			remotes, err := repo.ListRemotes(context.Background())

			require.NoError(t, err)
			assert.Len(t, remotes, test.expectedLen)

			if test.expectedLen > 0 {
				remote := remotes[0]
				assert.Equal(t, test.expectedName, remote.Name)
				assert.Equal(t, test.url, remote.URL)
				assert.Equal(t, test.url, remote.FetchURL)
				assert.Equal(t, test.url, remote.PushURL)
			}
		})
	}
}

func TestRepository_RemoteOperations_NotSupported(t *testing.T) {
	t.Parallel()

	repo := &Repository{path: "/test/path"}
	ctx := context.Background()

	tests := []struct {
		name string
		op   func() error
	}{
		{
			name: "AddRemote",
			op:   func() error { return repo.AddRemote(ctx, "origin", "https://example.com") },
		},
		{
			name: "RemoveRemote",
			op:   func() error { return repo.RemoveRemote(ctx, "origin") },
		},
		{
			name: "UpdateRemote",
			op:   func() error { return repo.UpdateRemote(ctx, "origin", "https://new.com") },
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.op()
			require.ErrorIs(t, err, domain.ErrRemoteOpsNotSupported)
		})
	}
}

func TestRepository_Fetch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		url         string
		expectError bool
		errorType   error
	}{
		{
			name:        "no source URL",
			url:         "",
			expectError: true,
			errorType:   domain.ErrNoSourceURLConfigured,
		},
		{
			name:        "unsupported URL type",
			url:         "https://github.com/user/repo.git",
			expectError: true,
			errorType:   domain.ErrFetchNotSupportedForURLType,
		},
		{
			name:        "file URL with non-existent source",
			url:         "file:///non-existent-source",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			testFS := testutil.NewTestFS(t)

			adapter := createTestAdapter(testFS)
			repo := &Repository{
				path: "/test/path",
				url:  test.url,
				fs:   adapter.fs,
			}

			err := repo.Fetch(context.Background(), ports.FetchOptions{})

			if test.expectError {
				require.Error(t, err)

				if test.errorType != nil {
					require.ErrorIs(t, err, test.errorType)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRepository_Fetch_FileURL_Success(t *testing.T) {
	t.Parallel()
	testFS := testutil.NewTestFS(t)

	// Create source and destination directories
	sourceDir := testFS.TempDir("test")
	testFS.WriteFile(filepath.Join(sourceDir, "file1.txt"), "test content 1")
	testFS.WriteFile(filepath.Join(sourceDir, "subdir", "file2.txt"), "test content 2")

	destDir := testFS.TempDir("test")

	// Create adapter with test filesystem
	adapter := createTestAdapter(testFS)
	repo := &Repository{
		path: destDir,
		url:  "file://" + sourceDir,
		fs:   adapter.fs,
	}

	err := repo.Fetch(context.Background(), ports.FetchOptions{})

	require.NoError(t, err)

	// Verify files were copied
	file1Content := testFS.ReadFile(filepath.Join(destDir, "file1.txt"))
	assert.Equal(t, "test content 1", file1Content)
}

func TestRepository_Pull(t *testing.T) {
	t.Parallel()
	testFS := testutil.NewTestFS(t)

	// Create source directory
	sourceDir := testFS.TempDir("test")
	testFS.WriteFile(filepath.Join(sourceDir, "file1.txt"), "test content 1")
	testFS.WriteFile(filepath.Join(sourceDir, "subdir", "file2.txt"), "test content 2")

	destDir := testFS.TempDir("test")

	// Create adapter with test filesystem
	adapter := createTestAdapter(testFS)
	repo := &Repository{
		path: destDir,
		url:  "file://" + sourceDir,
		fs:   adapter.fs,
	}

	err := repo.Pull(context.Background(), ports.PullOptions{Remote: "origin"})

	require.NoError(t, err)

	// Verify files were pulled
	file1Content := testFS.ReadFile(filepath.Join(destDir, "file1.txt"))
	assert.Equal(t, "test content 1", file1Content)
}

func TestRepository_Push_NotSupported(t *testing.T) {
	t.Parallel()

	repo := &Repository{path: "/test/path"}

	err := repo.Push(context.Background(), ports.PushOptions{})

	require.ErrorIs(t, err, domain.ErrPushOpsNotSupported)
}

func TestRepository_GetCommit(t *testing.T) {
	t.Parallel()

	repo := &Repository{path: "/test/path"}

	commit, err := repo.GetCommit(context.Background(), "main")

	require.NoError(t, err)
	assert.Equal(t, "0000000000000000000000000000000000000000", commit.Hash)
	assert.Equal(t, "Directory snapshot", commit.Message)
	assert.Equal(t, "Directory System", commit.Author.Name)
	assert.Equal(t, "system@directory", commit.Author.Email)
}

func TestRepository_ListCommits(t *testing.T) {
	t.Parallel()

	repo := &Repository{path: "/test/path"}

	commits, err := repo.ListCommits(context.Background(), ports.ListCommitsOptions{})

	require.NoError(t, err)
	assert.Empty(t, commits)
}

func TestRepository_TagOperations(t *testing.T) {
	t.Parallel()

	repo := &Repository{path: "/test/path"}
	ctx := context.Background()

	// Test ListTags
	tags, err := repo.ListTags(ctx)
	require.NoError(t, err)
	assert.Empty(t, tags)

	// Test CreateTag (not supported)
	err = repo.CreateTag(ctx, "v1.0.0", "main", "Release v1.0.0")
	require.ErrorIs(t, err, domain.ErrTagOpsNotSupported)

	// Test DeleteTag (not supported)
	err = repo.DeleteTag(ctx, "v1.0.0")
	require.ErrorIs(t, err, domain.ErrTagOpsNotSupported)
}

func TestRepository_Status(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "existing directory",
			expectError: false,
		},
		{
			name:        "non-existent directory",
			path:        "/non-existent-directory",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			testFS := testutil.NewTestFS(t)
			adapter := createTestAdapter(testFS)

			var repoPath string
			if test.expectError {
				repoPath = test.path
			} else {
				tempDir := testFS.TempDir("test")
				testFS.WriteFile(filepath.Join(tempDir, "file1.txt"), "test content 1")
				repoPath = tempDir
			}

			repo := &Repository{
				path: repoPath,
				fs:   adapter.fs,
			}

			status, err := repo.Status(context.Background())

			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.True(t, status.IsClean)
			}
		})
	}
}

func TestRepository_Diff_NotSupported(t *testing.T) {
	t.Parallel()

	repo := &Repository{path: "/test/path"}

	diff, err := repo.Diff(context.Background(), ports.DiffOptions{})

	assert.Empty(t, diff)
	require.ErrorIs(t, err, domain.ErrDiffOpsNotSupported)
}

func TestRepository_Close(t *testing.T) {
	t.Parallel()

	repo := &Repository{path: "/test/path"}

	err := repo.Close()

	require.NoError(t, err)
}

// Test helper method copyDirectory

func TestAdapter_copyDirectory(t *testing.T) {
	t.Parallel()
	testFS := testutil.NewTestFS(t)

	// Create source directory with files in memory filesystem
	sourceDir := testSourceDir
	testFS.CreateDir(sourceDir)
	testFS.WriteFile(filepath.Join(sourceDir, "file1.txt"), "test content 1")
	testFS.CreateDir(filepath.Join(sourceDir, "subdir"))
	testFS.WriteFile(filepath.Join(sourceDir, "subdir", "file2.txt"), "test content 2")

	destDir := "/dest"

	adapter := createTestAdapter(testFS)

	err := adapter.copyDirectory(sourceDir, destDir)

	require.NoError(t, err)

	// Verify all files were copied
	file1Content := testFS.ReadFile(filepath.Join(destDir, "file1.txt"))
	assert.Equal(t, "test content 1", file1Content)

	file2Content := testFS.ReadFile(filepath.Join(destDir, "subdir", "file2.txt"))
	assert.Equal(t, "test content 2", file2Content)
}

func TestAdapter_copyFile(t *testing.T) {
	t.Parallel()
	testFS := testutil.NewTestFS(t)

	// Create source file
	sourceDir := testFS.TempDir("test")
	sourceFile := filepath.Join(sourceDir, "copy-source.txt")
	testContent := "test file content"
	testFS.WriteFile(sourceFile, testContent)

	// Create destination path
	destDir := testFS.TempDir("test")
	destFile := filepath.Join(destDir, "copied-file.txt")

	adapter := createTestAdapter(testFS)

	err := adapter.copyFile(sourceFile, destFile)

	require.NoError(t, err)

	// Verify file was copied correctly
	copiedContent := testFS.ReadFile(destFile)
	assert.Equal(t, testContent, copiedContent)

	// Verify permissions were preserved
	// Note: TestFS doesn't support Stat/file mode checking
	// The test verifies content copying instead
}

// Benchmarks to compare memory vs disk performance for directory operations.
func BenchmarkDirectoryClone_Memory(b *testing.B) {
	// Setup memory filesystem once
	memFS := testutil.NewMemFS(b)
	fileSystem := filesystem.NewAferoFileSystem(memFS.Fs)

	// Create source once
	sourceDir := testSourceDir
	require.NoError(b, fileSystem.MkdirAll(sourceDir, 0750))
	require.NoError(b, fileSystem.WriteFile(filepath.Join(sourceDir, "file1.txt"), []byte("test content 1"), 0600))
	require.NoError(b, fileSystem.MkdirAll(filepath.Join(sourceDir, "subdir"), 0750))
	require.NoError(b, fileSystem.WriteFile(filepath.Join(sourceDir, "subdir", "file2.txt"), []byte("test content 2"), 0600))

	adapter := NewWithFileSystem(createMockGitConfig(), fileSystem)

	b.ResetTimer()

	for i := range b.N {
		destDir := fmt.Sprintf("/dest-%d", i)
		options := ports.CloneOptions{
			URL:  "file://" + sourceDir,
			Path: destDir,
		}
		_, _ = adapter.Clone(context.Background(), options)
		_ = fileSystem.RemoveAll(destDir)
	}
}

func BenchmarkDirectoryClone_Disk(b *testing.B) {
	// Create source once
	sourceDir := b.TempDir()
	require.NoError(b, os.WriteFile(filepath.Join(sourceDir, "file1.txt"), []byte("test content 1"), 0600))
	require.NoError(b, os.MkdirAll(filepath.Join(sourceDir, "subdir"), 0750))
	require.NoError(b, os.WriteFile(filepath.Join(sourceDir, "subdir", "file2.txt"), []byte("test content 2"), 0600))

	adapter := New(createMockGitConfig())

	b.ResetTimer()

	for i := range b.N {
		destDir := filepath.Join(b.TempDir(), fmt.Sprintf("dest-%d", i))
		options := ports.CloneOptions{
			URL:  "file://" + sourceDir,
			Path: destDir,
		}
		_, _ = adapter.Clone(context.Background(), options)
	}
}

// Edge cases and error conditions

func TestAdapter_copyDirectory_NonExistentSource(t *testing.T) {
	t.Parallel()
	testFS := testutil.NewTestFS(t)

	destDir := testFS.TempDir("test")

	adapter := New(createMockGitConfig())

	err := adapter.copyDirectory("/non-existent-source", destDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to copy directory")
}

func TestAdapter_copyFile_NonExistentSource(t *testing.T) {
	t.Parallel()
	testFS := testutil.NewTestFS(t)

	destDir := testFS.TempDir("test")

	adapter := New(createMockGitConfig())

	err := adapter.copyFile("/non-existent-file", filepath.Join(destDir, "dest"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read source file")
}

// Benchmark tests for performance regression detection

func BenchmarkAdapter_Clone(b *testing.B) {
	testFS := testutil.NewTestFS(b)
	sourceDir := createTempDirWithFiles(b)

	adapter := New(createMockGitConfig())

	b.ResetTimer()

	for range b.N {
		destDir := testFS.TempDir("test")

		options := ports.CloneOptions{
			URL:  "file://" + sourceDir,
			Path: destDir,
		}

		_, err := adapter.Clone(context.Background(), options)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAdapter_copyDirectory(b *testing.B) {
	testFS := testutil.NewTestFS(b)
	sourceDir := createTempDirWithFiles(b)

	adapter := New(createMockGitConfig())

	b.ResetTimer()

	for range b.N {
		destDir := testFS.TempDir("test")

		err := adapter.copyDirectory(sourceDir, destDir)
		if err != nil {
			b.Fatal(err)
		}
	}
}
