// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package directory

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// Mock logger for testing.
type mockLogger struct {
	logs []logEntry
}

type logEntry struct {
	level   string
	message string
	fields  map[string]any
}

func (ml *mockLogger) Trace(_ context.Context, msg string, fields map[string]any) {
	ml.logs = append(ml.logs, logEntry{level: "TRACE", message: msg, fields: fields})
}

func (ml *mockLogger) Debug(_ context.Context, msg string, fields map[string]any) {
	ml.logs = append(ml.logs, logEntry{level: "DEBUG", message: msg, fields: fields})
}

func (ml *mockLogger) Info(_ context.Context, msg string, fields map[string]any) {
	ml.logs = append(ml.logs, logEntry{level: "INFO", message: msg, fields: fields})
}

func (ml *mockLogger) Warn(_ context.Context, msg string, fields map[string]any) {
	ml.logs = append(ml.logs, logEntry{level: "WARN", message: msg, fields: fields})
}

func (ml *mockLogger) Error(_ context.Context, msg string, fields map[string]any) {
	ml.logs = append(ml.logs, logEntry{level: "ERROR", message: msg, fields: fields})
}

func (ml *mockLogger) Fatal(_ context.Context, msg string, fields map[string]any) {
	ml.logs = append(ml.logs, logEntry{level: "FATAL", message: msg, fields: fields})
}

func (ml *mockLogger) IsLevelEnabled(_ ports.LogLevel) bool {
	return true
}

func (ml *mockLogger) hasLogMessage(level, message string) bool {
	for _, entry := range ml.logs {
		if entry.level == level && strings.Contains(entry.message, message) {
			return true
		}
	}

	return false
}

// Mock repository for testing.
type mockRepository struct {
	path string
}

func (mr *mockRepository) Path() string {
	return mr.path
}

func (mr *mockRepository) URL() string {
	return ""
}

func (mr *mockRepository) Name() string {
	return filepath.Base(mr.path)
}

func (mr *mockRepository) IsBare() bool {
	return false
}

func (mr *mockRepository) IsClean() bool {
	return true
}

func (mr *mockRepository) HasChanges() bool {
	return false
}

func (mr *mockRepository) CurrentBranch() (string, error) {
	const defaultBranch = "main"

	return defaultBranch, nil
}

func (mr *mockRepository) GetCurrentBranch(_ context.Context) (string, error) {
	const defaultBranch = "main"

	return defaultBranch, nil
}

func (mr *mockRepository) ListBranches(_ context.Context) ([]ports.BranchInfo, error) {
	return []ports.BranchInfo{}, nil
}

func (mr *mockRepository) GetBranches(_ context.Context) ([]string, error) {
	return []string{"main"}, nil
}

func (mr *mockRepository) CreateBranch(_ context.Context, _, _ string) error {
	return domain.ErrBranchOpsNotSupported
}

func (mr *mockRepository) CheckoutBranch(_ context.Context, _ string) error {
	return domain.ErrBranchOpsNotSupported
}

func (mr *mockRepository) DeleteBranch(_ context.Context, _ string, _ bool) error {
	return domain.ErrBranchOpsNotSupported
}

func (mr *mockRepository) SetDefaultBranch(_ context.Context, _ string) error {
	return domain.ErrBranchOpsNotSupported
}

func (mr *mockRepository) ListRemotes(_ context.Context) ([]ports.RemoteInfo, error) {
	return []ports.RemoteInfo{}, nil
}

func (mr *mockRepository) GetRemote(_ context.Context, _ string) (ports.RemoteInfo, error) {
	return ports.RemoteInfo{}, domain.ErrRemoteOpsNotSupported
}

func (mr *mockRepository) AddRemote(_ context.Context, _, _ string) error {
	return domain.ErrRemoteOpsNotSupported
}

func (mr *mockRepository) RemoveRemote(_ context.Context, _ string) error {
	return domain.ErrRemoteOpsNotSupported
}

func (mr *mockRepository) UpdateRemote(_ context.Context, _, _ string) error {
	return domain.ErrRemoteOpsNotSupported
}

func (mr *mockRepository) GetRemoteURL(_ context.Context, _ string) (string, error) {
	return "", domain.ErrRemoteOpsNotSupported
}

func (mr *mockRepository) SetRemoteURL(_ context.Context, _, _ string) error {
	return domain.ErrRemoteOpsNotSupported
}

func (mr *mockRepository) Fetch(_ context.Context, _ ports.FetchOptions) error {
	return nil
}

func (mr *mockRepository) Pull(_ context.Context, _ ports.PullOptions) error {
	return nil
}

func (mr *mockRepository) Push(_ context.Context, _ ports.PushOptions) error {
	return domain.ErrPushOpsNotSupported
}

func (mr *mockRepository) GetCurrentCommit(_ context.Context) (ports.CommitInfo, error) {
	return ports.CommitInfo{}, nil
}

func (mr *mockRepository) GetCommit(_ context.Context, _ string) (ports.CommitInfo, error) {
	return ports.CommitInfo{}, nil
}

func (mr *mockRepository) ListCommits(_ context.Context, _ ports.ListCommitsOptions) ([]ports.CommitInfo, error) {
	return []ports.CommitInfo{}, nil
}

func (mr *mockRepository) GetCommits(_ context.Context, _ ports.ListCommitsOptions) ([]ports.CommitInfo, error) {
	return []ports.CommitInfo{}, nil
}

func (mr *mockRepository) ListTags(_ context.Context) ([]ports.TagInfo, error) {
	return []ports.TagInfo{}, nil
}

func (mr *mockRepository) CreateTag(_ context.Context, _, _, _ string) error {
	return domain.ErrTagOpsNotSupported
}

func (mr *mockRepository) DeleteTag(_ context.Context, _ string) error {
	return domain.ErrTagOpsNotSupported
}

func (mr *mockRepository) Status(_ context.Context) (ports.StatusResult, error) {
	return ports.StatusResult{IsClean: true}, nil
}

func (mr *mockRepository) GetStatus(_ context.Context) (ports.StatusResult, error) {
	return ports.StatusResult{IsClean: true}, nil
}

func (mr *mockRepository) Diff(_ context.Context, _ ports.DiffOptions) (string, error) {
	return "", domain.ErrDiffOpsNotSupported
}

func (mr *mockRepository) Add(_ context.Context, _ []string) error {
	return domain.ErrUnsupportedOperation
}

func (mr *mockRepository) Commit(_ context.Context, _ string) error {
	return domain.ErrUnsupportedOperation
}

func (mr *mockRepository) GetConfig() any {
	return ports.GitConfig{}
}

func (mr *mockRepository) GetWorkingDirectory() string {
	return mr.path
}

func (mr *mockRepository) GetGitDirectory() string {
	return filepath.Join(mr.path, ".git")
}

func (mr *mockRepository) HasUncommittedChanges(_ context.Context) (bool, error) {
	return false, nil
}

func (mr *mockRepository) GetFileContent(_ context.Context, _ string) ([]byte, error) {
	return []byte{}, nil
}

func (mr *mockRepository) WriteFile(_ context.Context, _ string, _ []byte) error {
	return nil
}

func (mr *mockRepository) ListFiles(_ context.Context, _ string) ([]string, error) {
	return []string{}, nil
}

func (mr *mockRepository) Close() error {
	return nil
}

// Helper functions for testing

func createTestSourceRepo(t *testing.T) (ports.GitRepository, string) {
	t.Helper()

	tempDir := createTempDirWithFiles(t)

	// Add more test files for mirror testing
	testFile3 := filepath.Join(tempDir, "hidden", ".hiddenfile")
	testFile4 := filepath.Join(tempDir, "docs", "README.md")
	testFile5 := filepath.Join(tempDir, "src", "main.go")

	require.NoError(t, os.MkdirAll(filepath.Dir(testFile3), 0750))
	require.NoError(t, os.MkdirAll(filepath.Dir(testFile4), 0750))
	require.NoError(t, os.MkdirAll(filepath.Dir(testFile5), 0750))

	require.NoError(t, os.WriteFile(testFile3, []byte("hidden content"), 0600))
	require.NoError(t, os.WriteFile(testFile4, []byte("# Documentation"), 0600))
	require.NoError(t, os.WriteFile(testFile5, []byte("package main"), 0600))

	return &mockRepository{path: tempDir}, tempDir
}

// Test MirrorService constructor

func TestNewMirrorService(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}

	service := NewMirrorService(adapter, logger)

	require.NotNil(t, service)
	assert.Equal(t, adapter, service.adapter)
	assert.Equal(t, logger, service.logger)
}

// Test CreateMirror operation

func TestMirrorService_CreateMirror_Success(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	sourceRepo, sourceDir := createTestSourceRepo(t)

	defer func() { _ = os.RemoveAll(sourceDir) }()

	targetDir := t.TempDir()
	targetPath := filepath.Join(targetDir, "mirror")

	request := MirrorRequest{
		SourceRepository: sourceRepo,
		TargetPath:       targetPath,
		Options: MirrorOptions{
			Overwrite: false,
			Metadata: MirrorMetadata{
				CreatedAt:   time.Now(),
				CreatedBy:   "test-user",
				Source:      sourceDir,
				Version:     "1.0.0",
				Description: "Test mirror",
			},
		},
	}

	result, err := service.CreateMirror(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Success)
	assert.Equal(t, targetPath, result.TargetPath)
	assert.Positive(t, result.FilesCount)
	assert.Positive(t, result.TotalSize)
	assert.Greater(t, result.Duration, time.Duration(0))

	// Verify target directory was created
	_, err = os.Stat(targetPath)
	require.NoError(t, err)

	// Verify metadata file was created
	metadataPath := filepath.Join(targetPath, ".mirror-metadata.txt")
	_, err = os.Stat(metadataPath)
	require.NoError(t, err)

	// Verify some files were copied
	file1Path := filepath.Join(targetPath, "file1.txt")
	content, err := os.ReadFile(file1Path) //nolint:gosec // G304: Test file with controlled path
	require.NoError(t, err)
	assert.Equal(t, "test content 1", string(content))

	// Verify logging
	assert.True(t, logger.hasLogMessage("INFO", "Creating directory mirror"))
	assert.True(t, logger.hasLogMessage("INFO", "Directory mirror created successfully"))
}

func TestMirrorService_CreateMirror_TargetExists_Overwrite(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	sourceRepo, sourceDir := createTestSourceRepo(t)

	defer func() { _ = os.RemoveAll(sourceDir) }()

	targetDir := t.TempDir()
	targetPath := filepath.Join(targetDir, "mirror")

	// Create target directory first
	require.NoError(t, os.MkdirAll(targetPath, 0750))
	require.NoError(t, os.WriteFile(filepath.Join(targetPath, "existing.txt"), []byte("existing"), 0600))

	request := MirrorRequest{
		SourceRepository: sourceRepo,
		TargetPath:       targetPath,
		Options: MirrorOptions{
			Overwrite: true,
			Metadata: MirrorMetadata{
				CreatedAt: time.Now(),
				CreatedBy: "test-user",
				Source:    sourceDir,
			},
		},
	}

	result, err := service.CreateMirror(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Success)

	// Verify old file was removed
	existingPath := filepath.Join(targetPath, "existing.txt")
	_, err = os.Stat(existingPath)
	require.ErrorIs(t, err, fs.ErrNotExist)

	// Verify new files were created
	file1Path := filepath.Join(targetPath, "file1.txt")
	_, err = os.Stat(file1Path)
	require.NoError(t, err)

	// Verify warning was logged
	assert.True(t, logger.hasLogMessage("WARN", "Target path exists, removing"))
}

func TestMirrorService_CreateMirror_TargetExists_NoOverwrite(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	sourceRepo, sourceDir := createTestSourceRepo(t)

	defer func() { _ = os.RemoveAll(sourceDir) }()

	targetDir := t.TempDir()
	targetPath := filepath.Join(targetDir, "mirror")

	// Create target directory first
	require.NoError(t, os.MkdirAll(targetPath, 0750))

	request := MirrorRequest{
		SourceRepository: sourceRepo,
		TargetPath:       targetPath,
		Options: MirrorOptions{
			Overwrite: false,
		},
	}

	result, err := service.CreateMirror(context.Background(), request)

	require.Error(t, err)
	assert.Nil(t, result)
	require.ErrorIs(t, err, domain.ErrTargetPathAlreadyExists)
}

func TestMirrorService_CreateMirror_WithPatterns(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	sourceRepo, sourceDir := createTestSourceRepo(t)

	defer func() { _ = os.RemoveAll(sourceDir) }()

	targetDir := t.TempDir()
	targetPath := filepath.Join(targetDir, "mirror")

	request := MirrorRequest{
		SourceRepository: sourceRepo,
		TargetPath:       targetPath,
		Options: MirrorOptions{
			IncludePatterns: []string{"*.txt", "*.md"},
			ExcludePatterns: []string{"*.go"},
			IncludeHidden:   false,
			Metadata: MirrorMetadata{
				CreatedAt: time.Now(),
				CreatedBy: "test-user",
				Source:    sourceDir,
			},
		},
	}

	result, err := service.CreateMirror(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Success)

	// Verify .txt files were included
	file1Path := filepath.Join(targetPath, "file1.txt")
	_, err = os.Stat(file1Path)
	require.NoError(t, err)

	// Verify .md files were included
	readmePath := filepath.Join(targetPath, "docs", "README.md")
	_, err = os.Stat(readmePath)
	require.NoError(t, err)

	// Verify .go files were excluded
	mainGoPath := filepath.Join(targetPath, "src", "main.go")
	_, err = os.Stat(mainGoPath)
	require.ErrorIs(t, err, fs.ErrNotExist)

	// Verify hidden files were excluded
	hiddenPath := filepath.Join(targetPath, "hidden", ".hiddenfile")
	_, err = os.Stat(hiddenPath)
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestMirrorService_CreateMirror_IncludeHidden(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	sourceRepo, sourceDir := createTestSourceRepo(t)

	defer func() { _ = os.RemoveAll(sourceDir) }()

	targetDir := t.TempDir()
	targetPath := filepath.Join(targetDir, "mirror")

	request := MirrorRequest{
		SourceRepository: sourceRepo,
		TargetPath:       targetPath,
		Options: MirrorOptions{
			IncludeHidden: true,
			Metadata: MirrorMetadata{
				CreatedAt: time.Now(),
				CreatedBy: "test-user",
				Source:    sourceDir,
			},
		},
	}

	result, err := service.CreateMirror(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Success)

	// Verify hidden files were included
	hiddenPath := filepath.Join(targetPath, "hidden", ".hiddenfile")
	_, err = os.Stat(hiddenPath)
	require.NoError(t, err)
}

func TestMirrorService_CreateMirror_WithArchive(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	sourceRepo, sourceDir := createTestSourceRepo(t)

	defer func() { _ = os.RemoveAll(sourceDir) }()

	targetDir := t.TempDir()
	targetPath := filepath.Join(targetDir, "mirror")

	request := MirrorRequest{
		SourceRepository: sourceRepo,
		TargetPath:       targetPath,
		Options: MirrorOptions{
			CreateArchive: true,
			ArchiveFormat: "tar.gz",
			Metadata: MirrorMetadata{
				CreatedAt: time.Now(),
				CreatedBy: "test-user",
				Source:    sourceDir,
			},
		},
	}

	result, err := service.CreateMirror(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Success)
	assert.Len(t, result.Warnings, 1) // Archive creation warning (not implemented)
	assert.Contains(t, result.Warnings[0], "Failed to create archive")

	// Verify archive creation was requested in logs
	assert.True(t, logger.hasLogMessage("DEBUG", "Archive creation requested but not implemented"))
}

func TestMirrorService_CreateMirror_CreateTargetDirectoryError(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	sourceRepo, sourceDir := createTestSourceRepo(t)

	defer func() { _ = os.RemoveAll(sourceDir) }()

	// Try to create in non-existent parent directory
	targetPath := "/non-existent-parent/mirror"

	request := MirrorRequest{
		SourceRepository: sourceRepo,
		TargetPath:       targetPath,
		Options:          MirrorOptions{},
	}

	result, err := service.CreateMirror(context.Background(), request)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create target directory")
}

// Test UpdateMirror operation

func TestMirrorService_UpdateMirror_TargetExists(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	sourceRepo, sourceDir := createTestSourceRepo(t)

	defer func() { _ = os.RemoveAll(sourceDir) }()

	targetDir := t.TempDir()
	targetPath := filepath.Join(targetDir, "mirror")

	// Create existing target
	require.NoError(t, os.MkdirAll(targetPath, 0750))

	request := MirrorRequest{
		SourceRepository: sourceRepo,
		TargetPath:       targetPath,
		Options: MirrorOptions{
			Overwrite: true,
			Metadata: MirrorMetadata{
				CreatedAt: time.Now(),
				CreatedBy: "test-user",
				Source:    sourceDir,
			},
		},
	}

	result, err := service.UpdateMirror(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Success)
	assert.Equal(t, targetPath, result.TargetPath)

	// Verify logging
	assert.True(t, logger.hasLogMessage("INFO", "Updating directory mirror"))
}

func TestMirrorService_UpdateMirror_TargetNotExists(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	sourceRepo, sourceDir := createTestSourceRepo(t)

	defer func() { _ = os.RemoveAll(sourceDir) }()

	targetDir := t.TempDir()
	targetPath := filepath.Join(targetDir, "mirror")

	request := MirrorRequest{
		SourceRepository: sourceRepo,
		TargetPath:       targetPath,
		Options: MirrorOptions{
			Metadata: MirrorMetadata{
				CreatedAt: time.Now(),
				CreatedBy: "test-user",
				Source:    sourceDir,
			},
		},
	}

	result, err := service.UpdateMirror(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Success)
	assert.Equal(t, targetPath, result.TargetPath)

	// Verify target was created
	_, err = os.Stat(targetPath)
	require.NoError(t, err)
}

// Test VerifyMirror operation

func TestMirrorService_VerifyMirror_ValidMirror(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	// Create a valid mirror
	targetDir := t.TempDir()
	targetPath := filepath.Join(targetDir, "mirror")

	require.NoError(t, os.MkdirAll(targetPath, 0750))
	require.NoError(t, os.WriteFile(filepath.Join(targetPath, "file1.txt"), []byte("content1"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(targetPath, "file2.txt"), []byte("content2"), 0600))

	// Create metadata file
	metadataPath := filepath.Join(targetPath, ".mirror-metadata.json")
	require.NoError(t, os.WriteFile(metadataPath, []byte("{}"), 0600))

	verification, err := service.VerifyMirror(context.Background(), targetPath)

	require.NoError(t, err)
	require.NotNil(t, verification)

	assert.True(t, verification.IsValid)
	assert.Equal(t, targetPath, verification.Path)
	assert.Equal(t, 3, verification.FileCount) // 2 files + metadata
	assert.Positive(t, verification.TotalSize)
	assert.True(t, verification.HasMetadata)
	assert.Greater(t, verification.Duration, time.Duration(0))
	assert.Empty(t, verification.Errors)

	// Verify logging
	assert.True(t, logger.hasLogMessage("DEBUG", "Verifying directory mirror"))
	assert.True(t, logger.hasLogMessage("INFO", "Directory mirror verification completed"))
}

func TestMirrorService_VerifyMirror_NonExistentPath(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	targetPath := "/non-existent-path"

	verification, err := service.VerifyMirror(context.Background(), targetPath)

	require.NoError(t, err)
	require.NotNil(t, verification)

	assert.False(t, verification.IsValid)
	assert.Equal(t, targetPath, verification.Path)
	assert.Equal(t, 0, verification.FileCount)
	assert.Equal(t, int64(0), verification.TotalSize)
	assert.False(t, verification.HasMetadata)
	assert.Len(t, verification.Errors, 1)
	assert.Contains(t, verification.Errors[0], "target path does not exist")
}

func TestMirrorService_VerifyMirror_WithWarnings(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	// Create a mirror with inaccessible files
	targetDir := t.TempDir()
	targetPath := filepath.Join(targetDir, "mirror")

	require.NoError(t, os.MkdirAll(targetPath, 0750))
	require.NoError(t, os.WriteFile(filepath.Join(targetPath, "file1.txt"), []byte("content1"), 0600))

	// Create inaccessible subdirectory
	inaccessibleDir := filepath.Join(targetPath, "inaccessible")
	require.NoError(t, os.MkdirAll(inaccessibleDir, 0000)) // No permissions

	verification, err := service.VerifyMirror(context.Background(), targetPath)

	require.NoError(t, err)
	require.NotNil(t, verification)

	assert.True(t, verification.IsValid) // Still valid despite warnings
	assert.NotEmpty(t, verification.Warnings)
	assert.Contains(t, verification.Warnings[0], "Error accessing")

	// Cleanup: restore permissions so test cleanup can work
	_ = os.Chmod(inaccessibleDir, 0750) //nolint:gosec // G302: Restore permissions for test cleanup
}

// Test DeleteMirror operation

func TestMirrorService_DeleteMirror_Success(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	// Create a mirror to delete
	targetDir := t.TempDir()
	targetPath := filepath.Join(targetDir, "mirror")

	require.NoError(t, os.MkdirAll(targetPath, 0750))
	require.NoError(t, os.WriteFile(filepath.Join(targetPath, "file1.txt"), []byte("content1"), 0600))

	err := service.DeleteMirror(context.Background(), targetPath)

	require.NoError(t, err)

	// Verify mirror was deleted
	_, err = os.Stat(targetPath)
	require.ErrorIs(t, err, fs.ErrNotExist)

	// Verify logging
	assert.True(t, logger.hasLogMessage("INFO", "Deleting directory mirror"))
	assert.True(t, logger.hasLogMessage("INFO", "Directory mirror deleted successfully"))
}

func TestMirrorService_DeleteMirror_NonExistentPath(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	targetPath := "/non-existent-path"

	err := service.DeleteMirror(context.Background(), targetPath)

	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrTargetPathDoesNotExist)
}

// Test helper methods

func TestMirrorService_shouldExclude(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	tests := []struct {
		name            string
		path            string
		options         MirrorOptions
		expectedExclude bool
	}{
		{
			name:            "no patterns",
			path:            "file.txt",
			options:         MirrorOptions{},
			expectedExclude: false,
		},
		{
			name: "exclude pattern match",
			path: "file.log",
			options: MirrorOptions{
				ExcludePatterns: []string{"*.log"},
			},
			expectedExclude: true,
		},
		{
			name: "exclude pattern no match",
			path: "file.txt",
			options: MirrorOptions{
				ExcludePatterns: []string{"*.log"},
			},
			expectedExclude: false,
		},
		{
			name: "hidden file, include hidden disabled",
			path: ".hiddenfile",
			options: MirrorOptions{
				IncludeHidden: false,
			},
			expectedExclude: true,
		},
		{
			name: "hidden file, include hidden enabled",
			path: ".hiddenfile",
			options: MirrorOptions{
				IncludeHidden: true,
			},
			expectedExclude: false,
		},
		{
			name: "exclude basename match",
			path: "dir/file.log",
			options: MirrorOptions{
				ExcludePatterns: []string{"*.log"},
			},
			expectedExclude: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := service.shouldExclude(test.path, test.options)
			assert.Equal(t, test.expectedExclude, result)
		})
	}
}

func TestMirrorService_shouldInclude(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	tests := []struct {
		name            string
		path            string
		options         MirrorOptions
		expectedInclude bool
	}{
		{
			name:            "no patterns - include all",
			path:            "file.txt",
			options:         MirrorOptions{},
			expectedInclude: true,
		},
		{
			name: "include pattern match",
			path: "file.go",
			options: MirrorOptions{
				IncludePatterns: []string{"*.go"},
			},
			expectedInclude: true,
		},
		{
			name: "include pattern no match",
			path: "file.txt",
			options: MirrorOptions{
				IncludePatterns: []string{"*.go"},
			},
			expectedInclude: false,
		},
		{
			name: "include basename match",
			path: "dir/file.go",
			options: MirrorOptions{
				IncludePatterns: []string{"*.go"},
			},
			expectedInclude: true,
		},
		{
			name: "multiple patterns - one match",
			path: "file.md",
			options: MirrorOptions{
				IncludePatterns: []string{"*.go", "*.md"},
			},
			expectedInclude: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := service.shouldInclude(test.path, test.options)
			assert.Equal(t, test.expectedInclude, result)
		})
	}
}

func TestMirrorService_isHidden(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "hidden file",
			path:     ".hiddenfile",
			expected: true,
		},
		{
			name:     "hidden directory",
			path:     ".hidden",
			expected: true,
		},
		{
			name:     "regular file",
			path:     "file.txt",
			expected: false,
		},
		{
			name:     "file in hidden directory",
			path:     ".hidden/file.txt",
			expected: false, // Only checks basename
		},
		{
			name:     "nested hidden file",
			path:     "dir/.hiddenfile",
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := service.isHidden(test.path)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestMirrorService_copyFile(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	// Create source file
	sourceFile, err := os.CreateTemp("", "mirror-copy-source-*")
	require.NoError(t, err)

	defer func() { _ = os.Remove(sourceFile.Name()) }()

	testContent := "test mirror copy content"
	_, err = sourceFile.WriteString(testContent)
	require.NoError(t, err)

	_ = sourceFile.Close()

	// Set source file permissions
	require.NoError(t, os.Chmod(sourceFile.Name(), 0600))

	// Create target path in temp directory
	targetDir := t.TempDir()
	targetPath := filepath.Join(targetDir, "subdir", "copied-file.txt")

	err = service.copyFile(sourceFile.Name(), targetPath)

	require.NoError(t, err)

	// Verify file was copied correctly
	copiedContent, err := os.ReadFile(targetPath) //nolint:gosec // G304: Test file with controlled path
	require.NoError(t, err)
	assert.Equal(t, testContent, string(copiedContent))

	// Verify permissions were preserved
	sourceInfo, err := os.Stat(sourceFile.Name())
	require.NoError(t, err)

	targetInfo, err := os.Stat(targetPath)
	require.NoError(t, err)

	assert.Equal(t, sourceInfo.Mode(), targetInfo.Mode())

	// Verify target directory was created
	targetDirPath := filepath.Dir(targetPath)
	_, err = os.Stat(targetDirPath)
	require.NoError(t, err)
}

func TestMirrorService_copyFile_SourceError(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	targetDir := t.TempDir()
	targetPath := filepath.Join(targetDir, "copied-file.txt")

	err := service.copyFile("/non-existent-source", targetPath)

	require.Error(t, err)
}

func TestMirrorService_createMetadataFile(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	targetDir := t.TempDir()

	metadata := MirrorMetadata{
		CreatedAt:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		CreatedBy:   "test-user",
		Source:      "/source/path",
		Version:     "1.0.0",
		Description: "Test metadata",
	}

	err := service.createMetadataFile(context.Background(), targetDir, metadata)

	require.NoError(t, err)

	// Verify metadata file was created
	metadataPath := filepath.Join(targetDir, ".mirror-metadata.txt")
	content, err := os.ReadFile(metadataPath) //nolint:gosec // G304: Test file with controlled path
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "Mirror created: 2024-01-01T12:00:00Z")
	assert.Contains(t, contentStr, "Source: /source/path")
	assert.Contains(t, contentStr, "Created by: test-user")
}

func TestMirrorService_createArchive(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	targetDir := t.TempDir()
	options := MirrorOptions{
		ArchiveFormat: "tar.gz",
	}

	// is a placeholder implementation that returns an error
	err := service.createArchive(context.Background(), targetDir, options)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "archive creation not implemented")

	// Verify debug message was logged
	assert.True(t, logger.hasLogMessage("DEBUG", "Archive creation requested but not implemented"))
}

func TestMirrorService_pathExists(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	// Test existing path
	tempDir := t.TempDir()
	assert.True(t, service.pathExists(tempDir))

	// Test non-existent path
	assert.False(t, service.pathExists("/non-existent-path"))
}

// Edge cases and error conditions

func TestMirrorService_CreateMirror_SourceWalkError(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	// Create source with inaccessible directory
	sourceDir := t.TempDir()
	inaccessibleDir := filepath.Join(sourceDir, "inaccessible")
	require.NoError(t, os.MkdirAll(inaccessibleDir, 0000)) // No permissions

	sourceRepo := &mockRepository{path: sourceDir}
	targetDir := t.TempDir()
	targetPath := filepath.Join(targetDir, "mirror")

	request := MirrorRequest{
		SourceRepository: sourceRepo,
		TargetPath:       targetPath,
		Options: MirrorOptions{
			Metadata: MirrorMetadata{
				CreatedAt: time.Now(),
				CreatedBy: "test-user",
				Source:    sourceDir,
			},
		},
	}

	result, err := service.CreateMirror(context.Background(), request)

	require.NoError(t, err) // Should succeed but with warnings
	require.NotNil(t, result)

	assert.True(t, result.Success)
	assert.NotEmpty(t, result.Warnings)
	assert.Contains(t, result.Warnings[0], "Error accessing")

	// Cleanup: restore permissions so test cleanup can work
	_ = os.Chmod(inaccessibleDir, 0750) //nolint:gosec // G302: Restore permissions for test cleanup
}

// Benchmark tests for performance regression detection

func BenchmarkMirrorService_CreateMirror(b *testing.B) {
	adapter := New(createMockGitConfig())
	logger := &mockLogger{}
	service := NewMirrorService(adapter, logger)

	// Create test source repo using a helper approach
	tempDir := createTempDirWithFiles(b)
	sourceRepo := &mockRepository{path: tempDir}

	defer func() { _ = os.RemoveAll(tempDir) }()

	b.ResetTimer()

	for range b.N {
		targetDir, err := os.MkdirTemp("", "benchmark-mirror-*")
		if err != nil {
			b.Fatal(err)
		}

		targetPath := filepath.Join(targetDir, "mirror")

		request := MirrorRequest{
			SourceRepository: sourceRepo,
			TargetPath:       targetPath,
			Options: MirrorOptions{
				Metadata: MirrorMetadata{
					CreatedAt: time.Now(),
					CreatedBy: "benchmark",
					Source:    tempDir,
				},
			},
		}

		_, err = service.CreateMirror(context.Background(), request)
		if err != nil {
			b.Fatal(err)
		}

		_ = os.RemoveAll(targetDir)
	}
}
