// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package archive

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/constants"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// Test helper functions

func createTestRepository(tb testing.TB) (*Repository, string) {
	tb.Helper()

	t, ok := tb.(*testing.T)
	if !ok {
		tb.Fatal("createTestRepository requires *testing.T")
	}

	tempDir := t.TempDir()

	// Create test files
	testFile1 := filepath.Join(tempDir, "file1.txt")
	testFile2 := filepath.Join(tempDir, "subdir", "file2.txt")

	if err := os.MkdirAll(filepath.Dir(testFile2), 0750); err != nil {
		tb.Fatal(err)
	}

	if err := os.WriteFile(testFile1, []byte("content 1"), 0600); err != nil {
		tb.Fatal(err)
	}

	if err := os.WriteFile(testFile2, []byte("content 2"), 0600); err != nil {
		tb.Fatal(err)
	}

	config := ports.GitConfig{
		UserName:  "Test User",
		UserEmail: "test@example.com",
	}
	repo := &Repository{
		path:   tempDir,
		config: config,
	}

	return repo, tempDir
}

// Test basic repository properties

func TestRepository_Path(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	assert.Equal(t, tempDir, repo.Path())
}

func TestRepository_URL(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	// Archive repositories don't have URLs
	assert.Empty(t, repo.URL())
}

func TestRepository_Name(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	expected := filepath.Base(tempDir)
	assert.Equal(t, expected, repo.Name())
}

func TestRepository_IsBare(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	// Archive repositories are not bare
	assert.False(t, repo.IsBare())
}

func TestRepository_IsClean(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(t *testing.T) *Repository
		expected bool
	}{
		{
			name: "existing directory",
			setup: func(t *testing.T) *Repository {
				t.Helper()
				repo, _ := createTestRepository(t)

				return repo
			},
			expected: true,
		},
		{
			name: "non-existent directory",
			setup: func(_ *testing.T) *Repository {
				config := ports.GitConfig{
					UserName:  "Test User",
					UserEmail: "test@example.com",
				}

				return &Repository{
					path:   "/non-existent-directory",
					config: config,
				}
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			repo := test.setup(t)
			if test.expected {
				defer func() { _ = os.RemoveAll(repo.Path()) }()
			}

			result := repo.IsClean()
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestRepository_HasChanges(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	// Archive repositories never have changes
	assert.False(t, repo.HasChanges())
}

// Test commit operations

func TestRepository_GetCurrentCommit(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	commit, err := repo.GetCurrentCommit(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "archive-extraction", commit.Hash)
	assert.Equal(t, "Archive extraction", commit.Message)
	assert.Equal(t, "Archive Adapter", commit.Author.Name)
	assert.Equal(t, "archive@git-provider-sync", commit.Author.Email)
	assert.Equal(t, "Archive Adapter", commit.Committer.Name)
	assert.Equal(t, "archive@git-provider-sync", commit.Committer.Email)
	assert.Empty(t, commit.Parents)

	// Check that timestamps are recent
	now := time.Now()
	assert.WithinDuration(t, now, commit.Author.When, time.Minute)
	assert.WithinDuration(t, now, commit.Committer.When, time.Minute)
}

func TestRepository_GetCommit(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)
	t.Cleanup(func() { _ = os.RemoveAll(tempDir) })

	tests := []struct {
		name        string
		hash        string
		expectError bool
		expectedErr error
	}{
		{
			name:        "archive-extraction hash",
			hash:        "archive-extraction",
			expectError: false,
		},
		{
			name:        "HEAD reference",
			hash:        "HEAD",
			expectError: false,
		},
		{
			name:        "non-existent hash",
			hash:        "non-existent-hash",
			expectError: true,
			expectedErr: ErrCommitNotFoundInArchive,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			commit, err := repo.GetCommit(context.Background(), test.hash)

			if test.expectError {
				require.Error(t, err)

				if test.expectedErr != nil {
					require.ErrorIs(t, err, test.expectedErr)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, "archive-extraction", commit.Hash)
				assert.Equal(t, "Archive extraction", commit.Message)
			}
		})
	}
}

func TestRepository_ListCommits(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	commits, err := repo.ListCommits(context.Background(), ports.ListCommitsOptions{})

	require.NoError(t, err)
	require.Len(t, commits, 1)

	commit := commits[0]
	assert.Equal(t, "archive-extraction", commit.Hash)
	assert.Equal(t, "Archive extraction", commit.Message)
}

func TestRepository_GetCommits(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	commits, err := repo.GetCommits(context.Background(), ports.ListCommitsOptions{})

	require.NoError(t, err)
	require.Len(t, commits, 1)

	commit := commits[0]
	assert.Equal(t, "archive-extraction", commit.Hash)
	assert.Equal(t, "Archive extraction", commit.Message)
}

// Test branch operations

func TestRepository_CurrentBranch(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	branch, err := repo.CurrentBranch()

	require.NoError(t, err)
	assert.Equal(t, constants.DefaultBranch, branch)
}

func TestRepository_GetCurrentBranch(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	branch, err := repo.GetCurrentBranch(context.Background())

	require.NoError(t, err)
	assert.Equal(t, constants.DefaultBranch, branch)
}

func TestRepository_ListBranches(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	branches, err := repo.ListBranches(context.Background())

	require.NoError(t, err)
	require.Len(t, branches, 1)

	branch := branches[0]
	assert.Equal(t, "main", branch.Name)
	assert.Equal(t, "archive-extraction", branch.Hash)
	assert.False(t, branch.IsRemote)
	assert.True(t, branch.IsCurrent)
	assert.Empty(t, branch.Upstream)
}

func TestRepository_GetBranches(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	branches, err := repo.GetBranches(context.Background())

	require.NoError(t, err)
	require.Len(t, branches, 1)
	assert.Equal(t, "main", branches[0])
}

func TestRepository_CreateBranch_NotSupported(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	err := repo.CreateBranch(context.Background(), "new-branch", "main")

	assert.ErrorIs(t, err, ErrCreateBranchNotSupported)
}

func TestRepository_CheckoutBranch_NotSupportedOperation(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)
	t.Cleanup(func() { _ = os.RemoveAll(tempDir) })

	tests := []struct {
		name        string
		branchName  string
		expectError bool
		expectedErr error
	}{
		{
			name:        "checkout main branch",
			branchName:  "main",
			expectError: false,
		},
		{
			name:        "checkout non-existent branch",
			branchName:  "feature-branch",
			expectError: true,
			expectedErr: ErrOnlyMainBranchExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := repo.CheckoutBranch(context.Background(), test.branchName)

			if test.expectError {
				require.Error(t, err)

				if test.expectedErr != nil {
					require.ErrorIs(t, err, test.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRepository_DeleteBranch_NotSupported(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	err := repo.DeleteBranch(context.Background(), "branch", false)

	assert.ErrorIs(t, err, ErrDeleteBranchNotSupported)
}

func TestRepository_SetDefaultBranch_NotSupported(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	err := repo.SetDefaultBranch(context.Background(), "main")

	assert.ErrorIs(t, err, ErrSetDefaultBranchNotSupported)
}

// Test remote operations

func TestRepository_ListRemotes(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	remotes, err := repo.ListRemotes(context.Background())

	require.NoError(t, err)
	assert.Empty(t, remotes)
}

func TestRepository_GetRemote_NotSupported(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	_, err := repo.GetRemote(context.Background(), "origin")

	assert.ErrorIs(t, err, ErrArchiveReposNoRemotes)
}

func TestRepository_AddRemote_NotSupported(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	err := repo.AddRemote(context.Background(), "origin", "https://example.com")

	assert.ErrorIs(t, err, ErrAddRemoteNotSupported)
}

func TestRepository_RemoveRemote_NotSupported(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	err := repo.RemoveRemote(context.Background(), "origin")

	assert.ErrorIs(t, err, ErrRemoveRemoteNotSupported)
}

func TestRepository_UpdateRemote_NotSupported(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	err := repo.UpdateRemote(context.Background(), "origin", "https://new.com")

	assert.ErrorIs(t, err, ErrUpdateRemoteNotSupported)
}

func TestRepository_GetRemoteURL_NotSupported(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	_, err := repo.GetRemoteURL(context.Background(), "origin")

	assert.ErrorIs(t, err, ErrArchiveReposNoRemotes)
}

func TestRepository_SetRemoteURL_NotSupported(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	err := repo.SetRemoteURL(context.Background(), "origin", "https://new.com")

	assert.ErrorIs(t, err, domain.ErrRemoteURLNotSupportedArchive)
}

// Test tag operations

func TestRepository_ListTags(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	tags, err := repo.ListTags(context.Background())

	require.NoError(t, err)
	assert.Empty(t, tags)
}

func TestRepository_CreateTag_NotSupported(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	err := repo.CreateTag(context.Background(), "v1.0.0", "main", "Release v1.0.0")

	assert.ErrorIs(t, err, ErrCreateTagNotSupported)
}

func TestRepository_DeleteTag_NotSupported(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	err := repo.DeleteTag(context.Background(), "v1.0.0")

	assert.ErrorIs(t, err, ErrDeleteTagNotSupported)
}

// Test status and diff operations

func TestRepository_Status(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	status, err := repo.Status(context.Background())

	require.NoError(t, err)

	// Should list files as untracked since archive repos don't have git tracking
	assert.Contains(t, status.Untracked, "file1.txt")
	// Note: subdir/file2.txt won't be listed directly as ReadDir only reads top level
}

func TestRepository_Status_NonExistentDirectory(t *testing.T) {
	t.Parallel()

	config := ports.GitConfig{
		UserName:  "Test User",
		UserEmail: "test@example.com",
	}
	repo := &Repository{
		path:   "/non-existent-directory",
		config: config,
	}

	_, err := repo.Status(context.Background())

	assert.Error(t, err)
}

func TestRepository_GetStatus(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	status, err := repo.GetStatus(context.Background())

	require.NoError(t, err)
	assert.Contains(t, status.Untracked, "file1.txt")
}

func TestRepository_Diff_NotSupported(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	diff, err := repo.Diff(context.Background(), ports.DiffOptions{})

	assert.Empty(t, diff)
	assert.ErrorIs(t, err, ErrDiffNotSupported)
}

// Test git operations (not supported)

func TestRepository_Fetch_NotSupported(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	err := repo.Fetch(context.Background(), ports.FetchOptions{})

	assert.ErrorIs(t, err, domain.ErrUnsupportedOperation)
}

func TestRepository_Pull_NotSupported(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	err := repo.Pull(context.Background(), ports.PullOptions{})

	assert.ErrorIs(t, err, domain.ErrPullNotSupportedArchive)
}

func TestRepository_Push_NotSupported(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	err := repo.Push(context.Background(), ports.PushOptions{})

	assert.ErrorIs(t, err, domain.ErrPushNotSupportedArchive)
}

func TestRepository_Add_NotSupported(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	err := repo.Add(context.Background(), []string{"file1.txt"})

	assert.ErrorIs(t, err, domain.ErrAddFilesNotSupportedArchive)
}

func TestRepository_Commit_NotSupported(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	err := repo.Commit(context.Background(), "Test commit")

	assert.ErrorIs(t, err, domain.ErrCommitNotSupportedArchive)
}

// Test configuration and metadata operations

func TestRepository_GetConfig(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	config := repo.GetConfig()

	assert.NotNil(t, config)
	assert.IsType(t, ports.GitConfig{}, config)
}

func TestRepository_GetWorkingDirectory(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	workDir := repo.GetWorkingDirectory()

	assert.Equal(t, tempDir, workDir)
}

func TestRepository_GetGitDirectory(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	gitDir := repo.GetGitDirectory()

	expected := filepath.Join(tempDir, ".archive-metadata")
	assert.Equal(t, expected, gitDir)
}

func TestRepository_HasUncommittedChanges(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	hasChanges, err := repo.HasUncommittedChanges(context.Background())

	require.NoError(t, err)
	assert.False(t, hasChanges)
}

// Test file operations

func TestRepository_GetFileContent(t *testing.T) { //nolint:tparallel // Race condition with shared tempDir
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	tests := []struct {
		name        string
		filePath    string
		expectError bool
		expectedErr error
		expected    string
	}{
		{
			name:        "existing file",
			filePath:    "file1.txt",
			expectError: false,
			expected:    "content 1",
		},
		{
			name:        "nested file",
			filePath:    "subdir/file2.txt",
			expectError: false,
			expected:    "content 2",
		},
		{
			name:        "non-existent file",
			filePath:    "non-existent.txt",
			expectError: true,
		},
		{
			name:        "path traversal attempt",
			filePath:    "../../../etc/passwd",
			expectError: true,
			expectedErr: domain.ErrFilePathOutsideRepository,
		},
	}

	for _, test := range tests { //nolint:paralleltest // Race condition with shared tempDir
		t.Run(test.name, func(t *testing.T) {
			// Removed t.Parallel() to avoid race condition with shared tempDir
			content, err := repo.GetFileContent(context.Background(), test.filePath)

			if test.expectError {
				require.Error(t, err)

				if test.expectedErr != nil {
					require.ErrorIs(t, err, test.expectedErr)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected, string(content))
			}
		})
	}
}

func TestRepository_WriteFile(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)
	t.Cleanup(func() { _ = os.RemoveAll(tempDir) })

	tests := []struct {
		name        string
		filePath    string
		content     string
		expectError bool
		expectedErr error
	}{
		{
			name:        "write new file",
			filePath:    "new-file.txt",
			content:     "new content",
			expectError: false,
		},
		{
			name:        "overwrite existing file",
			filePath:    "file1.txt",
			content:     "updated content",
			expectError: false,
		},
		{
			name:        "write nested file",
			filePath:    "newdir/nested.txt",
			content:     "nested content",
			expectError: false,
		},
		{
			name:        "path traversal attempt",
			filePath:    "../../../tmp/malicious.txt",
			content:     "malicious content",
			expectError: true,
			expectedErr: domain.ErrFilePathOutsideRepository,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := repo.WriteFile(context.Background(), test.filePath, []byte(test.content))

			if test.expectError {
				require.Error(t, err)

				if test.expectedErr != nil {
					require.ErrorIs(t, err, test.expectedErr)
				}
			} else {
				require.NoError(t, err)

				// Verify file was written correctly
				content, err := repo.GetFileContent(context.Background(), test.filePath)
				require.NoError(t, err)
				assert.Equal(t, test.content, string(content))

				// Verify permissions
				fullPath := filepath.Join(tempDir, test.filePath)
				info, err := os.Stat(fullPath)
				require.NoError(t, err)
				assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
			}
		})
	}
}

func TestRepository_ListFiles(t *testing.T) { //nolint:tparallel // Race condition with shared tempDir
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "list all files",
			path:     "",
			expected: []string{"file1.txt", "subdir/file2.txt"},
		},
		{
			name:     "list files in subdirectory",
			path:     "subdir",
			expected: []string{"subdir/file2.txt"},
		},
	}

	for _, test := range tests { //nolint:paralleltest // Race condition with shared tempDir
		t.Run(test.name, func(t *testing.T) {
			// Removed t.Parallel() to avoid race condition with shared tempDir
			files, err := repo.ListFiles(context.Background(), test.path)

			require.NoError(t, err)
			assert.ElementsMatch(t, test.expected, files)
		})
	}
}

func TestRepository_ListFiles_NonExistentPath(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	_, err := repo.ListFiles(context.Background(), "non-existent-dir")

	assert.Error(t, err)
}

// Test cleanup

func TestRepository_Close(t *testing.T) {
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	err := repo.Close()

	assert.NoError(t, err)
}

// Test file security validation

func TestRepository_FileOperations_SecurityValidation(t *testing.T) { //nolint:tparallel // Race condition with shared tempDir
	t.Parallel()

	repo, tempDir := createTestRepository(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	// Test Unix-style path traversal attempts
	maliciousPaths := []string{
		"../../../etc/passwd",
		"../malicious.txt",
		"../../../tmp/outside.txt",
	}

	for _, maliciousPath := range maliciousPaths { //nolint:paralleltest // Race condition with shared tempDir
		t.Run("malicious path: "+maliciousPath, func(t *testing.T) {
			// Removed t.Parallel() to avoid race condition with shared tempDir

			// Test GetFileContent
			_, err := repo.GetFileContent(context.Background(), maliciousPath)
			require.ErrorIs(t, err, domain.ErrFilePathOutsideRepository)

			// Test WriteFile
			err = repo.WriteFile(context.Background(), maliciousPath, []byte("malicious"))
			require.ErrorIs(t, err, domain.ErrFilePathOutsideRepository)
		})
	}
}

// Edge cases and error conditions

func TestRepository_FileOperations_AbsolutePathErrors(t *testing.T) {
	t.Parallel()

	// Create repository with path that will cause Abs() to fail
	config := ports.GitConfig{
		UserName:  "Test User",
		UserEmail: "test@example.com",
	}
	repo := &Repository{
		path:   string([]byte{0}), // Invalid path that should cause filepath.Abs to fail
		config: config,
	}

	// Test GetFileContent
	_, err := repo.GetFileContent(context.Background(), "test.txt")
	require.Error(t, err)

	// Test WriteFile
	err = repo.WriteFile(context.Background(), "test.txt", []byte("content"))
	require.Error(t, err)
}

// Benchmark tests for performance regression detection

func BenchmarkRepository_GetFileContent(b *testing.B) {
	repo, tempDir := createTestRepository(b)

	defer func() { _ = os.RemoveAll(tempDir) }()

	b.ResetTimer()

	for range b.N {
		_, err := repo.GetFileContent(context.Background(), "file1.txt")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRepository_WriteFile(b *testing.B) {
	repo, tempDir := createTestRepository(b)

	defer func() { _ = os.RemoveAll(tempDir) }()

	content := []byte("benchmark content")

	b.ResetTimer()

	for i := range b.N {
		filename := "benchmark-file-" + string(rune(i)) + ".txt"

		err := repo.WriteFile(context.Background(), filename, content)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRepository_ListFiles(b *testing.B) {
	repo, tempDir := createTestRepository(b)

	defer func() { _ = os.RemoveAll(tempDir) }()

	b.ResetTimer()

	for range b.N {
		_, err := repo.ListFiles(context.Background(), "")
		if err != nil {
			b.Fatal(err)
		}
	}
}
