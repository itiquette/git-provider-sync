// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package directory

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// Create a mock GitConfig for testing.
func createMockGitConfig() ports.GitConfig {
	return ports.GitConfig{
		UserName:  "Test User",
		UserEmail: "test@example.com",
	}
}

// Helper functions for tests

func createTempDirWithFiles(tb testing.TB) string {
	tb.Helper()

	// Use t.TempDir() for safe cleanup
	t, ok := tb.(*testing.T)
	if !ok {
		tb.Fatal("createTempDirWithFiles requires *testing.T")
	}

	tempDir := t.TempDir()

	// Create test files
	testFile1 := filepath.Join(tempDir, "file1.txt")
	testFile2 := filepath.Join(tempDir, "subdir", "file2.txt")

	if err := os.MkdirAll(filepath.Dir(testFile2), 0750); err != nil {
		tb.Fatal(err)
	}

	if err := os.WriteFile(testFile1, []byte("test content 1"), 0600); err != nil {
		tb.Fatal(err)
	}

	if err := os.WriteFile(testFile2, []byte("test content 2"), 0600); err != nil {
		tb.Fatal(err)
	}

	return tempDir
}

func createTempFile(t *testing.T) string {
	t.Helper()

	tempFile, err := os.CreateTemp("", "directory-adapter-file-*")
	require.NoError(t, err)

	filename := tempFile.Name()
	_ = tempFile.Close()

	return filename
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

func TestAdapter_Clone_FileURL(t *testing.T) {
	t.Parallel()

	// Create source directory with files
	sourceDir := createTempDirWithFiles(t)

	defer func() { _ = os.RemoveAll(sourceDir) }()

	// Create destination directory
	destDir := t.TempDir()

	adapter := New(createMockGitConfig())
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

	// Verify files were copied
	file1Content, err := os.ReadFile(filepath.Join(destDir, "file1.txt")) //nolint:gosec // Test file with controlled path
	require.NoError(t, err)
	assert.Equal(t, "test content 1", string(file1Content))

	file2Content, err := os.ReadFile(filepath.Join(destDir, "subdir", "file2.txt")) //nolint:gosec // Test file with controlled path
	require.NoError(t, err)
	assert.Equal(t, "test content 2", string(file2Content))
}

func TestAdapter_Clone_NonFileURL(t *testing.T) {
	t.Parallel()

	destDir := t.TempDir()

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

	tempDir := createTempDirWithFiles(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	adapter := New(createMockGitConfig())

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

	tempFile := createTempFile(t)

	defer func() { _ = os.Remove(tempFile) }()

	adapter := New(createMockGitConfig())

	_, err := adapter.Open(context.Background(), tempFile)

	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrPathNotDirectory)
}

// Test Init operation

func TestAdapter_Init(t *testing.T) {
	t.Parallel()

	tempDir, err := os.MkdirTemp("", "init-test-*")
	require.NoError(t, err)

	defer func() { _ = os.RemoveAll(tempDir) }()

	newRepoPath := filepath.Join(tempDir, "new-repo")

	adapter := New(createMockGitConfig())

	repo, err := adapter.Init(context.Background(), newRepoPath, ports.InitOptions{})

	require.NoError(t, err)
	require.NotNil(t, repo)
	assert.Equal(t, newRepoPath, repo.Path())

	// Verify directory was created
	info, err := os.Stat(newRepoPath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

// Test Cleanup operation

func TestAdapter_Cleanup(t *testing.T) {
	t.Parallel()

	tempDir := createTempDirWithFiles(t)
	// Don't defer removal, cleanup should handle it

	adapter := New(createMockGitConfig())

	err := adapter.Cleanup(context.Background(), tempDir)

	require.NoError(t, err)

	// Verify directory was removed
	_, err = os.Stat(tempDir)
	require.ErrorIs(t, err, fs.ErrNotExist)
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

	// Test that a directory repository correctly identifies its state
	// and can be used for directory-based operations
	tempDir := createTempDirWithFiles(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

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
	err := os.WriteFile(testFile, []byte("test"), 0600) // #nosec G306
	require.NoError(t, err, "Should be able to write to repository path")

	// Verify file exists
	_, err = os.Stat(testFile)
	require.NoError(t, err, "File should exist in repository")

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

			repo := &Repository{
				path: "/test/path",
				url:  test.url,
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

	// Create source and destination directories
	sourceDir := createTempDirWithFiles(t)

	defer func() { _ = os.RemoveAll(sourceDir) }()

	destDir := t.TempDir()

	repo := &Repository{
		path: destDir,
		url:  "file://" + sourceDir,
	}

	err := repo.Fetch(context.Background(), ports.FetchOptions{})

	require.NoError(t, err)

	// Verify files were copied
	file1Content, err := os.ReadFile(filepath.Join(destDir, "file1.txt")) //nolint:gosec // Test file with controlled path
	require.NoError(t, err)
	assert.Equal(t, "test content 1", string(file1Content))
}

func TestRepository_Pull(t *testing.T) {
	t.Parallel()

	// Create source directory
	sourceDir := createTempDirWithFiles(t)

	defer func() { _ = os.RemoveAll(sourceDir) }()

	destDir := t.TempDir()

	repo := &Repository{
		path: destDir,
		url:  "file://" + sourceDir,
	}

	err := repo.Pull(context.Background(), ports.PullOptions{Remote: "origin"})

	require.NoError(t, err)

	// Verify files were pulled
	file1Content, err := os.ReadFile(filepath.Join(destDir, "file1.txt")) //nolint:gosec // Test file with controlled path
	require.NoError(t, err)
	assert.Equal(t, "test content 1", string(file1Content))
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

			var repoPath string
			if test.expectError {
				repoPath = test.path
			} else {
				tempDir := createTempDirWithFiles(t)

				defer func() { _ = os.RemoveAll(tempDir) }()

				repoPath = tempDir
			}

			repo := &Repository{path: repoPath}

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

	sourceDir := createTempDirWithFiles(t)

	defer func() { _ = os.RemoveAll(sourceDir) }()

	destDir := t.TempDir()

	adapter := New(createMockGitConfig())

	err := adapter.copyDirectory(sourceDir, destDir)

	require.NoError(t, err)

	// Verify all files were copied
	file1Content, err := os.ReadFile(filepath.Join(destDir, "file1.txt")) //nolint:gosec // Test file with controlled path
	require.NoError(t, err)
	assert.Equal(t, "test content 1", string(file1Content))

	file2Content, err := os.ReadFile(filepath.Join(destDir, "subdir", "file2.txt")) //nolint:gosec // Test file with controlled path
	require.NoError(t, err)
	assert.Equal(t, "test content 2", string(file2Content))
}

func TestAdapter_copyFile(t *testing.T) {
	t.Parallel()

	// Create source file
	sourceFile, err := os.CreateTemp("", "copy-source-*")
	require.NoError(t, err)

	defer func() { _ = os.Remove(sourceFile.Name()) }()

	testContent := "test file content"
	_, err = sourceFile.WriteString(testContent)
	require.NoError(t, err)

	_ = sourceFile.Close()

	// Create destination path
	destDir := t.TempDir()

	destFile := filepath.Join(destDir, "copied-file.txt")

	adapter := New(createMockGitConfig())

	err = adapter.copyFile(sourceFile.Name(), destFile)

	require.NoError(t, err)

	// Verify file was copied correctly
	copiedContent, err := os.ReadFile(destFile) //nolint:gosec // Test file with controlled path
	require.NoError(t, err)
	assert.Equal(t, testContent, string(copiedContent))

	// Verify permissions were preserved
	sourceInfo, err := os.Stat(sourceFile.Name())
	require.NoError(t, err)

	destInfo, err := os.Stat(destFile)
	require.NoError(t, err)

	assert.Equal(t, sourceInfo.Mode(), destInfo.Mode())
}

// Edge cases and error conditions

func TestAdapter_copyDirectory_NonExistentSource(t *testing.T) {
	t.Parallel()

	destDir := t.TempDir()

	adapter := New(createMockGitConfig())

	err := adapter.copyDirectory("/non-existent-source", destDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to copy directory")
}

func TestAdapter_copyFile_NonExistentSource(t *testing.T) {
	t.Parallel()

	destDir := t.TempDir()

	adapter := New(createMockGitConfig())

	err := adapter.copyFile("/non-existent-file", filepath.Join(destDir, "dest"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read source file")
}

// Benchmark tests for performance regression detection

func BenchmarkAdapter_Clone(b *testing.B) {
	sourceDir := createTempDirWithFiles(b)

	defer func() { _ = os.RemoveAll(sourceDir) }()

	adapter := New(createMockGitConfig())

	b.ResetTimer()

	for range b.N {
		destDir := b.TempDir()

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
	sourceDir := createTempDirWithFiles(b)

	defer func() { _ = os.RemoveAll(sourceDir) }()

	adapter := New(createMockGitConfig())

	b.ResetTimer()

	for range b.N {
		destDir := b.TempDir()

		err := adapter.copyDirectory(sourceDir, destDir)
		if err != nil {
			b.Fatal(err)
		}
	}
}
