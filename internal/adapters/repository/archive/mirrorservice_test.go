// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package archive

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/entities"
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

// Helper functions for creating test repositories.
func createMockRepository(name, httpsURL, description string) entities.Repository {
	builder := entities.NewRepositoryBuilder()

	var err error

	builder, err = builder.WithName(name)
	if err != nil {
		panic(err) // In tests, panic on builder errors
	}

	builder, err = builder.WithHTTPSURL(httpsURL)
	if err != nil {
		panic(err)
	}

	// Some methods don't return errors
	builder = builder.WithDescription(description)

	builder, err = builder.WithDefaultBranch("main")
	if err != nil {
		panic(err)
	}

	builder = builder.WithPrivate(false)
	builder = builder.WithArchived(false)
	builder = builder.WithFork(false)

	repo, err := builder.Build()
	if err != nil {
		panic(err)
	}

	return repo
}

func TestNewMirrorService(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	tempDir := "/tmp/test"
	archiveDir := "/archive/test"

	service := NewMirrorService(logger, tempDir, archiveDir)

	require.NotNil(t, service)
	assert.Equal(t, logger, service.logger)
	assert.Equal(t, tempDir, service.tempDir)
	assert.Equal(t, archiveDir, service.archiveDir)
	assert.Nil(t, service.progressWriter)
}

func TestMirrorService_SetProgressWriter(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	service := NewMirrorService(logger, "/tmp", "/archive")

	var buffer bytes.Buffer
	service.SetProgressWriter(&buffer)

	assert.Equal(t, &buffer, service.progressWriter)
}

func TestMirrorService_Mirror_DryRun(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	tempDir := t.TempDir()
	archiveDir := t.TempDir()

	service := NewMirrorService(logger, tempDir, archiveDir)

	sourceRepo := createMockRepository("test-repo", "https://github.com/test/repo", "Test repository")
	targetRepo := createMockRepository("target-repo", "https://gitlab.com/test/target", "Target repository")

	request := MirrorRequest{
		SourceRepository: sourceRepo,
		TargetRepository: targetRepo,
		ArchiveFormat:    "tar.gz",
		CompressionLevel: 6,
		DryRun:           true,
	}

	result, err := service.Mirror(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Success)
	assert.Equal(t, 150, result.FilesProcessed)
	assert.Equal(t, 10, result.FilesSkipped)
	assert.Equal(t, int64(1024*1024*5), result.ArchiveSize)
	assert.NotEmpty(t, result.ArchivePath)
	assert.Contains(t, result.ArchivePath, archiveDir)

	// Verify logging
	assert.True(t, logger.hasLogMessage("INFO", "Starting archive repository mirror"))
	assert.True(t, logger.hasLogMessage("INFO", "Performing archive dry run"))
	assert.True(t, logger.hasLogMessage("INFO", "Archive dry run completed"))
}

func TestMirrorService_Mirror_ActualRun(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	tempDir := t.TempDir()
	archiveDir := t.TempDir()

	service := NewMirrorService(logger, tempDir, archiveDir)

	sourceRepo := createMockRepository("test-repo", "https://github.com/test/repo", "Test repository")
	targetRepo := createMockRepository("target-repo", "https://gitlab.com/test/target", "Target repository")

	request := MirrorRequest{
		SourceRepository: sourceRepo,
		TargetRepository: targetRepo,
		ArchiveFormat:    "tar.gz",
		CompressionLevel: 6,
		DryRun:           false,
		PreservePaths:    true,
	}

	result, err := service.Mirror(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Success)
	assert.NotEmpty(t, result.ArchivePath)
	assert.Positive(t, result.ArchiveSize)
	assert.Positive(t, result.FilesProcessed)
	assert.NotNil(t, result.PerformanceInfo)

	// Verify archive was created
	_, err = os.Stat(result.ArchivePath)
	require.NoError(t, err)

	// Verify logging
	assert.True(t, logger.hasLogMessage("INFO", "Starting archive repository mirror"))
	assert.True(t, logger.hasLogMessage("DEBUG", "Created working directory"))
	assert.True(t, logger.hasLogMessage("DEBUG", "Downloading source repository"))
	assert.True(t, logger.hasLogMessage("DEBUG", "Creating archive"))
	assert.True(t, logger.hasLogMessage("INFO", "Archive mirror completed successfully"))
}

func TestMirrorService_Mirror_UnsupportedFormat(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	tempDir := t.TempDir()
	archiveDir := t.TempDir()

	service := NewMirrorService(logger, tempDir, archiveDir)

	sourceRepo := createMockRepository("test-repo", "https://github.com/test/repo", "Test repository")
	targetRepo := createMockRepository("target-repo", "https://gitlab.com/test/target", "Target repository")

	request := MirrorRequest{
		SourceRepository: sourceRepo,
		TargetRepository: targetRepo,
		ArchiveFormat:    "zip", // Unsupported format
		DryRun:           false,
	}

	result, err := service.Mirror(context.Background(), request)

	require.Error(t, err)
	require.NotNil(t, result)

	assert.False(t, result.Success)
	assert.Contains(t, err.Error(), "failed to create archive")
	assert.Contains(t, result.Errors[0], "archive creation failed")
}

func TestMirrorService_Mirror_TarFormat(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	tempDir := t.TempDir()
	archiveDir := t.TempDir()

	service := NewMirrorService(logger, tempDir, archiveDir)

	sourceRepo := createMockRepository("test-repo", "https://github.com/test/repo", "Test repository")
	targetRepo := createMockRepository("target-repo", "https://gitlab.com/test/target", "Target repository")

	request := MirrorRequest{
		SourceRepository: sourceRepo,
		TargetRepository: targetRepo,
		ArchiveFormat:    "tar", // Uncompressed tar
		DryRun:           false,
		PreservePaths:    false, // Test flattened paths
	}

	result, err := service.Mirror(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Success)
	assert.NotEmpty(t, result.ArchivePath)
	assert.True(t, strings.HasSuffix(result.ArchivePath, ".tar"))
}

func TestMirrorService_Mirror_WithPatterns(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	tempDir := t.TempDir()
	archiveDir := t.TempDir()

	service := NewMirrorService(logger, tempDir, archiveDir)

	sourceRepo := createMockRepository("test-repo", "https://github.com/test/repo", "Test repository")
	targetRepo := createMockRepository("target-repo", "https://gitlab.com/test/target", "Target repository")

	request := MirrorRequest{
		SourceRepository: sourceRepo,
		TargetRepository: targetRepo,
		ArchiveFormat:    "tar.gz",
		DryRun:           false,
		IncludePatterns:  []string{"*.go", "*.md"},
		ExcludePatterns:  []string{"*.tmp", "*.log"},
	}

	result, err := service.Mirror(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Success)
	assert.Positive(t, result.FilesProcessed)
}

func TestMirrorService_Mirror_WithProgressWriter(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	tempDir := t.TempDir()
	archiveDir := t.TempDir()

	service := NewMirrorService(logger, tempDir, archiveDir)

	var progressBuffer bytes.Buffer
	service.SetProgressWriter(&progressBuffer)

	sourceRepo := createMockRepository("test-repo", "https://github.com/test/repo", "Test repository")
	targetRepo := createMockRepository("target-repo", "https://gitlab.com/test/target", "Target repository")

	request := MirrorRequest{
		SourceRepository: sourceRepo,
		TargetRepository: targetRepo,
		ArchiveFormat:    "tar.gz",
		DryRun:           false,
	}

	result, err := service.Mirror(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Success)
	// Progress reporting happens every 100 files, so might not trigger with small test repo
}

func TestMirrorService_Mirror_WithCustomArchiveName(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	tempDir := t.TempDir()
	archiveDir := t.TempDir()

	service := NewMirrorService(logger, tempDir, archiveDir)

	sourceRepo := createMockRepository("test-repo", "https://github.com/test/repo", "Test repository")
	targetRepo := createMockRepository("target-repo", "https://gitlab.com/test/target", "Target repository")

	request := MirrorRequest{
		SourceRepository:   sourceRepo,
		TargetRepository:   targetRepo,
		ArchiveFormat:      "tar.gz",
		DryRun:             false,
		ArchiveNamePattern: "custom-{name}-backup.tar.gz",
	}

	result, err := service.Mirror(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Success)
	assert.Contains(t, result.ArchivePath, "custom-test-repo-backup.tar.gz")
}

func TestMirrorService_generateArchiveName(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	service := NewMirrorService(logger, "/tmp", "/archive")

	sourceRepo := createMockRepository("my-repo", "https://github.com/test/my-repo", "Test repository")

	tests := []struct {
		name           string
		request        MirrorRequest
		expectedSuffix string
		expectedPrefix string
	}{
		{
			name: "default tar.gz naming",
			request: MirrorRequest{
				SourceRepository: sourceRepo,
				ArchiveFormat:    "tar.gz",
			},
			expectedPrefix: "my-repo-",
			expectedSuffix: ".tar.gz",
		},
		{
			name: "default tar naming",
			request: MirrorRequest{
				SourceRepository: sourceRepo,
				ArchiveFormat:    "tar",
			},
			expectedPrefix: "my-repo-",
			expectedSuffix: ".tar",
		},
		{
			name: "custom pattern with name",
			request: MirrorRequest{
				SourceRepository:   sourceRepo,
				ArchiveFormat:      "tar.gz",
				ArchiveNamePattern: "backup-{name}.tar.gz",
			},
			expectedPrefix: "backup-my-repo",
			expectedSuffix: ".tar.gz",
		},
		{
			name: "custom pattern with repo",
			request: MirrorRequest{
				SourceRepository:   sourceRepo,
				ArchiveFormat:      "tar.gz",
				ArchiveNamePattern: "{repo}-archive.tar.gz",
			},
			expectedPrefix: "my-repo-archive",
			expectedSuffix: ".tar.gz",
		},
		{
			name: "unknown format",
			request: MirrorRequest{
				SourceRepository: sourceRepo,
				ArchiveFormat:    "unknown",
			},
			expectedPrefix: "my-repo-",
			expectedSuffix: ".archive",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := service.generateArchiveName(testCase.request)

			assert.True(t, strings.HasPrefix(result, testCase.expectedPrefix))
			assert.True(t, strings.HasSuffix(result, testCase.expectedSuffix))
		})
	}
}

func TestMirrorService_shouldSkipFile(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	service := NewMirrorService(logger, "/tmp", "/archive")

	tests := []struct {
		name            string
		filePath        string
		includePatterns []string
		excludePatterns []string
		shouldSkip      bool
	}{
		{
			name:            "no patterns - include all",
			filePath:        "file.txt",
			includePatterns: []string{},
			excludePatterns: []string{},
			shouldSkip:      false,
		},
		{
			name:            "exclude pattern match",
			filePath:        "file.log",
			includePatterns: []string{},
			excludePatterns: []string{"*.log"},
			shouldSkip:      true,
		},
		{
			name:            "include pattern match",
			filePath:        "file.go",
			includePatterns: []string{"*.go"},
			excludePatterns: []string{},
			shouldSkip:      false,
		},
		{
			name:            "include pattern no match",
			filePath:        "file.txt",
			includePatterns: []string{"*.go"},
			excludePatterns: []string{},
			shouldSkip:      true,
		},
		{
			name:            "exclude overrides include",
			filePath:        "file.go",
			includePatterns: []string{"*.go"},
			excludePatterns: []string{"*.go"},
			shouldSkip:      true,
		},
		{
			name:            "wildcard include",
			filePath:        "any-file.txt",
			includePatterns: []string{"*"},
			excludePatterns: []string{},
			shouldSkip:      false,
		},
		{
			name:            "contains pattern",
			filePath:        "src/main.go",
			includePatterns: []string{"main"},
			excludePatterns: []string{},
			shouldSkip:      false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := service.shouldSkipFile(testCase.filePath, testCase.includePatterns, testCase.excludePatterns)
			assert.Equal(t, testCase.shouldSkip, result)
		})
	}
}

func TestMirrorService_matchesPattern(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	service := NewMirrorService(logger, "/tmp", "/archive")

	tests := []struct {
		name     string
		filePath string
		pattern  string
		matches  bool
	}{
		{
			name:     "wildcard matches all",
			filePath: "any-file.txt",
			pattern:  "*",
			matches:  true,
		},
		{
			name:     "prefix and suffix wildcard",
			filePath: "test-file.go",
			pattern:  "test-*.go",
			matches:  true,
		},
		{
			name:     "prefix wildcard no match",
			filePath: "main.go",
			pattern:  "test-*.go",
			matches:  false,
		},
		{
			name:     "suffix wildcard no match",
			filePath: "test-file.txt",
			pattern:  "test-*.go",
			matches:  false,
		},
		{
			name:     "contains match",
			filePath: "src/main.go",
			pattern:  "main",
			matches:  true,
		},
		{
			name:     "contains no match",
			filePath: "src/lib.go",
			pattern:  "main",
			matches:  false,
		},
		{
			name:     "exact match",
			filePath: "README.md",
			pattern:  "README.md",
			matches:  true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := service.matchesPattern(testCase.filePath, testCase.pattern)
			assert.Equal(t, testCase.matches, result)
		})
	}
}

func TestMirrorService_createRealisticRepositoryStructure(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	service := NewMirrorService(logger, "/tmp", "/archive")

	tempDir := t.TempDir()
	repo := createMockRepository("my-awesome-repo", "https://github.com/test/my-awesome-repo", "An awesome test repository")

	err := service.createRealisticRepositoryStructure(tempDir, repo)
	require.NoError(t, err)

	// Verify expected files were created
	expectedFiles := []string{
		"README.md",
		"main.go",
		"go.mod",
		"src/lib.go",
		"tests/main_test.go",
		"docs/ARCHITECTURE.md",
		".gitignore",
		"Makefile",
		"CHANGELOG.md",
	}

	for _, expectedFile := range expectedFiles {
		fullPath := filepath.Join(tempDir, expectedFile)
		_, err := os.Stat(fullPath)
		require.NoError(t, err, "Expected file %s should exist", expectedFile)
	}

	// Verify README contains repository information
	readmeContent, err := os.ReadFile(filepath.Join(tempDir, "README.md")) //#nosec G304 -- test file with controlled tempDir
	require.NoError(t, err)

	readmeStr := string(readmeContent)
	assert.Contains(t, readmeStr, "my-awesome-repo")
	assert.Contains(t, readmeStr, "https://github.com/test/my-awesome-repo")
	assert.Contains(t, readmeStr, "An awesome test repository")

	// Verify main.go contains repository name
	mainContent, err := os.ReadFile(filepath.Join(tempDir, "main.go")) //#nosec G304 -- test file with controlled tempDir
	require.NoError(t, err)
	assert.Contains(t, string(mainContent), "my-awesome-repo")

	// Verify go.mod contains repository name
	modContent, err := os.ReadFile(filepath.Join(tempDir, "go.mod")) //#nosec G304 -- test file with controlled tempDir
	require.NoError(t, err)
	assert.Contains(t, string(modContent), "my-awesome-repo")
}

func TestMirrorService_createWorkingDirectory(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	tempDir := t.TempDir()
	service := NewMirrorService(logger, tempDir, "/archive")

	workDir, err := service.createWorkingDirectory(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, workDir)

	// Verify directory was created
	_, err = os.Stat(workDir)
	require.NoError(t, err)

	// Verify it's within temp directory
	assert.Contains(t, workDir, tempDir)
	assert.Contains(t, workDir, "archive-mirror-")

	// Verify logging
	assert.True(t, logger.hasLogMessage("DEBUG", "Created working directory"))

	// Cleanup
	service.cleanupWorkingDirectory(context.Background(), workDir)
}

func TestMirrorService_cleanupWorkingDirectory(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	service := NewMirrorService(logger, "/tmp", "/archive")

	// Create a temp directory to cleanup
	workDir := t.TempDir()

	// Create a file in it
	testFile := filepath.Join(workDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0600)
	require.NoError(t, err)

	// Cleanup
	service.cleanupWorkingDirectory(context.Background(), workDir)

	// Verify directory was removed
	_, err = os.Stat(workDir)
	require.ErrorIs(t, err, fs.ErrNotExist)

	// Verify logging
	assert.True(t, logger.hasLogMessage("DEBUG", "Cleaned up working directory"))
}

func TestMirrorService_cleanupWorkingDirectory_Error(t *testing.T) {
	t.Parallel()

	logger := &mockLogger{}
	service := NewMirrorService(logger, "/tmp", "/archive")

	// Create a directory with a file inside and make it read-only to cause removal to fail
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "test-cleanup")
	require.NoError(t, os.MkdirAll(testDir, 0750))

	testFile := filepath.Join(testDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0600))

	// Make the directory read-only to cause RemoveAll to fail
	require.NoError(t, os.Chmod(testDir, 0444)) //nolint:gosec // G302: Test requires specific permissions for error case

	// Try to cleanup the read-only directory
	service.cleanupWorkingDirectory(context.Background(), testDir)

	// Should log warning (but not return error since it's cleanup)
	assert.True(t, logger.hasLogMessage("WARN", "Failed to cleanup working directory"))

	// Cleanup: restore permissions so test cleanup can work
	_ = os.Chmod(testDir, 0755) //nolint:gosec // G302: Restore permissions for test cleanup
}

func TestGenerateTimestamp(t *testing.T) {
	t.Parallel()

	timestamp1 := generateTimestamp()

	time.Sleep(10 * time.Millisecond) // Ensure different timestamp

	timestamp2 := generateTimestamp()

	assert.NotEmpty(t, timestamp1)
	assert.NotEmpty(t, timestamp2)
	assert.NotEqual(t, timestamp1, timestamp2)

	// Verify format (YYYYMMDD-HHMMSS-mmm)
	assert.Len(t, timestamp1, 19) // 8 + 1 + 6 + 1 + 3
	assert.Contains(t, timestamp1, "-")
}

// Archive mirror service workflow tests.
func TestMirrorServiceWorkflow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		archiveFormat    string
		compressionLevel int
		preservePaths    bool
		includePatterns  []string
		excludePatterns  []string
		dryRun           bool
		expectSuccess    bool
	}{
		{
			name:             "tar.gz archive with maximum compression",
			archiveFormat:    "tar.gz",
			compressionLevel: 9,
			preservePaths:    true,
			expectSuccess:    true,
		},
		{
			name:            "tar archive with file patterns",
			archiveFormat:   "tar",
			preservePaths:   false,
			includePatterns: []string{"*.go", "*.md", "*.yaml"},
			excludePatterns: []string{"*.tmp", "*.log", ".git/*"},
			expectSuccess:   true,
		},
		{
			name:          "dry run mode - no actual archive creation",
			archiveFormat: "tar.gz",
			dryRun:        true,
			expectSuccess: true,
		},
		{
			name:            "zip archive with path preservation",
			archiveFormat:   "zip",
			preservePaths:   true,
			includePatterns: []string{"src/*", "docs/*"},
			expectSuccess:   false, // Zip format not supported
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Arrange: Set up test environment with isolated directories
			tempDir := t.TempDir()
			archiveDir := t.TempDir()

			sourceRepo := createMockRepository("workflow-test",
				"https://github.com/test/workflow",
				"Workflow test repository")
			targetRepo := createMockRepository("workflow-target",
				"https://gitlab.com/test/workflow-target",
				"Target workflow repository")

			logger := &mockLogger{}
			service := NewMirrorService(logger, tempDir, archiveDir)

			request := MirrorRequest{
				SourceRepository: sourceRepo,
				TargetRepository: targetRepo,
				ArchiveFormat:    test.archiveFormat,
				CompressionLevel: test.compressionLevel,
				PreservePaths:    test.preservePaths,
				IncludePatterns:  test.includePatterns,
				ExcludePatterns:  test.excludePatterns,
				DryRun:           test.dryRun,
			}

			// Act: Execute mirror operation
			result, err := service.Mirror(context.Background(), request)

			// Assert: Verify operation results
			if test.expectSuccess {
				require.NoError(t, err, "Mirror operation should succeed")
				require.NotNil(t, result, "Result should not be nil")

				assert.True(t, result.Success, "Result should indicate success")
				assert.NotEmpty(t, result.ArchivePath, "Archive path should be set")

				if !test.dryRun {
					assert.Positive(t, result.ArchiveSize, "Archive size should be positive for real operations")
					assert.Positive(t, result.FilesProcessed, "Files processed should be positive")

					// Verify archive file exists for non-dry-run operations
					_, err := os.Stat(result.ArchivePath)
					require.NoError(t, err, "Archive file should exist")
				} else {
					// Dry run should not create actual files
					_, err := os.Stat(result.ArchivePath)
					require.Error(t, err, "Archive file should not exist in dry run mode")
				}

				assert.NotNil(t, result.PerformanceInfo, "Performance info should be recorded")
			} else {
				require.Error(t, err, "Mirror operation should fail for invalid configuration")
			}

			// Assert: Verify logging behavior
			assert.True(t, logger.hasLogMessage("INFO", "Starting archive repository mirror"), "Should log mirror operation start")
		})
	}
}

// Test configuration validation and error cases.
func TestMirrorServiceValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		request       MirrorRequest
		expectError   bool
		errorContains string
	}{
		{
			name: "invalid archive format",
			request: MirrorRequest{
				SourceRepository: createMockRepository("test", "https://github.com/test/repo", "Test"),
				TargetRepository: createMockRepository("target", "https://gitlab.com/test/target", "Target"),
				ArchiveFormat:    "invalid-format",
			},
			expectError:   true,
			errorContains: "unsupported archive format",
		},
		{
			name: "invalid compression level",
			request: MirrorRequest{
				SourceRepository: createMockRepository("source", "https://github.com/test/source", "Source"),
				TargetRepository: createMockRepository("target", "https://gitlab.com/test/target", "Target"),
				ArchiveFormat:    "tar.gz",
				CompressionLevel: 15, // Invalid: should be 0-9
			},
			expectError:   true,
			errorContains: "compression level",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			archiveDir := t.TempDir()
			logger := &mockLogger{}
			service := NewMirrorService(logger, tempDir, archiveDir)

			_, err := service.Mirror(context.Background(), test.request)

			if test.expectError {
				require.Error(t, err)

				if test.errorContains != "" {
					assert.Contains(t, err.Error(), test.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
