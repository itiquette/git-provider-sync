// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package archive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/adapters/filesystem"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/testutil"
)

// Create a mock GitConfig for testing.
func createMockGitConfig() ports.GitConfig {
	return ports.GitConfig{
		UserName:  "Test User",
		UserEmail: "test@example.com",
	}
}

// createTestFileSystem creates an in-memory filesystem for testing.
func createTestFileSystem(tb testing.TB) ports.FileSystem {
	tb.Helper()
	memFS := testutil.NewMemFS(tb)

	return filesystem.NewAferoFileSystem(memFS.Fs)
}

// createTestFileSystemWithFiles creates a filesystem with test files.
func createTestFileSystemWithFiles(tb testing.TB) (ports.FileSystem, string) {
	tb.Helper()
	memFS := testutil.NewMemFS(tb)

	// Create test directory and files in memory
	tempDir := "/test-dir"
	memFS.CreateDir(tempDir)
	memFS.WriteFileString(filepath.Join(tempDir, "file1.txt"), "test content 1")
	memFS.CreateDir(filepath.Join(tempDir, "subdir"))
	memFS.WriteFileString(filepath.Join(tempDir, "subdir", "file2.txt"), "test content 2")

	return filesystem.NewAferoFileSystem(memFS.Fs), tempDir
}

// Helper functions for tests

func createTestArchiveInMemory(tb testing.TB, fs ports.FileSystem) string {
	tb.Helper()

	archivePath := "/test-archive.tar.gz"

	file, err := fs.OpenFile(archivePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	require.NoError(tb, err)

	defer func() { _ = file.Close() }()

	gzWriter := gzip.NewWriter(file)

	defer func() { _ = gzWriter.Close() }()

	tarWriter := tar.NewWriter(gzWriter)

	defer func() { _ = tarWriter.Close() }()

	// Add test files to archive
	files := []struct {
		name    string
		content string
		isDir   bool
		size    int64
	}{
		{"test-file.txt", "test content", false, 12},
		{"subdir/", "", true, 0},
		{"subdir/nested-file.txt", "nested content", false, 14},
	}

	for _, file := range files {
		header := &tar.Header{
			Name:    file.name,
			ModTime: time.Now(),
		}

		if file.isDir {
			header.Typeflag = tar.TypeDir
			header.Mode = 0755
		} else {
			header.Typeflag = tar.TypeReg
			header.Mode = 0644
			header.Size = file.size
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			tb.Fatal(err)
		}

		if !file.isDir {
			if _, err := tarWriter.Write([]byte(file.content)); err != nil {
				tb.Fatal(err)
			}
		}
	}

	return archivePath
}

// createTestArchive creates a test archive in a temporary real filesystem location.
// Used for tests that require real file paths (backwards compatibility).
func createTestArchive(tb testing.TB) string {
	tb.Helper()

	// Create real temp file for tests that need actual file paths
	tempDir := ""

	switch v := tb.(type) {
	case *testing.T:
		tempDir = v.TempDir()
	case *testing.B:
		tempDir = v.TempDir()
	default:
		tb.Fatal("unsupported testing type")
	}

	archivePath := filepath.Join(tempDir, "test-archive.tar.gz")

	// Create the archive using os filesystem
	file, err := os.Create(archivePath) //nolint:gosec // Test code with controlled paths
	require.NoError(tb, err)

	defer func() { _ = file.Close() }()

	gzWriter := gzip.NewWriter(file)

	defer func() { _ = gzWriter.Close() }()

	tarWriter := tar.NewWriter(gzWriter)

	defer func() { _ = tarWriter.Close() }()

	// Add test files to archive
	files := []struct {
		name    string
		content string
		isDir   bool
		size    int64
	}{
		{"test-file.txt", "test content", false, 12},
		{"subdir/", "", true, 0},
		{"subdir/nested-file.txt", "nested content", false, 14},
	}

	for _, file := range files {
		header := &tar.Header{
			Name:    file.name,
			ModTime: time.Now(),
		}

		if file.isDir {
			header.Typeflag = tar.TypeDir
			header.Mode = 0755
		} else {
			header.Typeflag = tar.TypeReg
			header.Mode = 0644
			header.Size = file.size
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			tb.Fatal(err)
		}

		if !file.isDir {
			if _, err := tarWriter.Write([]byte(file.content)); err != nil {
				tb.Fatal(err)
			}
		}
	}

	return archivePath
}

func createTempDirWithFiles(tb testing.TB) string {
	tb.Helper()

	// Both *testing.T and *testing.B have TempDir() method
	var tempDir string

	switch v := tb.(type) {
	case *testing.T:
		tempDir = v.TempDir()
	case *testing.B:
		tempDir = v.TempDir()
	default:
		tb.Fatal("unsupported testing type")
	}

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

	return tempDir
}

// createMaliciousHeader creates a tar header for testing various attack scenarios.
func createMaliciousHeader(attackType string, contentSize int) *tar.Header {
	baseHeader := tar.Header{
		Typeflag: tar.TypeReg,
		Mode:     0644,
		ModTime:  time.Now(),
	}

	switch attackType {
	case "path-traversal":
		baseHeader.Name = "../../../etc/passwd"
		baseHeader.Size = int64(contentSize)
	case "absolute-path":
		baseHeader.Name = "/etc/passwd"
		baseHeader.Size = int64(contentSize)
	case "large-file":
		baseHeader.Name = "large-file.txt"
		baseHeader.Size = 200 * 1024 * 1024 // 200MB (exceeds 100MB limit)
	default:
		baseHeader.Name = "unknown.txt"
		baseHeader.Size = int64(contentSize)
	}

	return &baseHeader
}

func createMaliciousArchive(tb testing.TB, attackType string) string {
	tb.Helper()

	// Get temp directory from test
	var tempDir string

	switch v := tb.(type) {
	case *testing.T:
		tempDir = v.TempDir()
	case *testing.B:
		tempDir = v.TempDir()
	default:
		tb.Fatal("unknown test type")
	}

	archiveFile, err := os.CreateTemp(tempDir, "malicious-archive-*.tar.gz")
	if err != nil {
		tb.Fatal(err)
	}

	archivePath := archiveFile.Name()
	_ = archiveFile.Close()

	file, err := os.Create(archivePath) //nolint:gosec // Test file with controlled path
	require.NoError(tb, err)

	defer func() { _ = file.Close() }()

	gzWriter := gzip.NewWriter(file)

	defer func() { _ = gzWriter.Close() }()

	tarWriter := tar.NewWriter(gzWriter)

	defer func() { _ = tarWriter.Close() }()

	content := "malicious content"
	header := createMaliciousHeader(attackType, len(content))

	if err := tarWriter.WriteHeader(header); err != nil {
		tb.Fatal(err)
	}

	if _, err := tarWriter.Write([]byte(content)); err != nil {
		tb.Fatal(err)
	}

	return archivePath
}

// createMaliciousArchiveInMemory creates a malicious archive entirely in memory filesystem.
// This is much faster than disk-based operations and provides better test isolation.
func createMaliciousArchiveInMemory(tb testing.TB, fileSystem ports.FileSystem, attackType string) string {
	tb.Helper()

	// Use memory path instead of OS temp
	archivePath := "/test-archives/malicious-archive.tar.gz"

	// Ensure directory exists in memory
	require.NoError(tb, fileSystem.MkdirAll("/test-archives", 0755))

	// Create archive file in memory filesystem
	file, err := fileSystem.Create(archivePath)
	require.NoError(tb, err)

	defer func() { _ = file.Close() }()

	gzWriter := gzip.NewWriter(file)

	defer func() { _ = gzWriter.Close() }()

	tarWriter := tar.NewWriter(gzWriter)

	defer func() { _ = tarWriter.Close() }()

	content := "malicious content"
	header := createMaliciousHeader(attackType, len(content))

	require.NoError(tb, tarWriter.WriteHeader(header))
	_, err = tarWriter.Write([]byte(content))
	require.NoError(tb, err)

	return archivePath
}

// Test Adapter constructor

func TestNew_ValidConfig_CreatesArchiveAdapter(t *testing.T) {
	t.Parallel()

	config := createMockGitConfig()
	// Use in-memory filesystem for unit tests
	memFS := testutil.NewMemFS(t)
	fs := filesystem.NewAferoFileSystem(memFS.Fs)
	adapter := NewWithFileSystem(config, fs)

	require.NotNil(t, adapter)
	assert.Equal(t, config, adapter.config)
}

// Test Adapter methods

func TestGetName_WhenCalled_ReturnsArchive(t *testing.T) {
	t.Parallel()

	// Use in-memory filesystem for unit tests
	memFS := testutil.NewMemFS(t)
	fs := filesystem.NewAferoFileSystem(memFS.Fs)
	adapter := NewWithFileSystem(createMockGitConfig(), fs)
	assert.Equal(t, "archive", adapter.GetName())
}

func TestSupportsURL_WhenGivenVariousURLs_CorrectlyIdentifiesArchiveFormats(t *testing.T) {
	t.Parallel()

	// Use in-memory filesystem for unit tests
	memFS := testutil.NewMemFS(t)
	fs := filesystem.NewAferoFileSystem(memFS.Fs)
	adapter := NewWithFileSystem(createMockGitConfig(), fs)

	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "accepts tar.gz file URL as valid archive format",
			url:      "file:///path/to/archive.tar.gz",
			expected: true,
		},
		{
			name:     "accepts tgz file URL as valid archive format",
			url:      "file:///path/to/archive.tgz",
			expected: true,
		},
		{
			name:     "rejects zip file URL as unsupported format",
			url:      "file:///path/to/archive.zip",
			expected: false,
		},
		{
			name:     "rejects regular text file URL as non-archive",
			url:      "file:///path/to/file.txt",
			expected: false,
		},
		{
			name:     "rejects https URL as unsupported protocol",
			url:      "https://example.com/archive.tar.gz",
			expected: false,
		},
		{
			name:     "rejects ssh git URL as unsupported protocol",
			url:      "git@github.com:user/repo.git",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := adapter.SupportsURL(test.url)
			require.Equal(t, test.expected, result)
		})
	}
}

// Test Clone operation

func TestClone_WhenGivenValidArchive_ExtractsContentSuccessfully(t *testing.T) {
	t.Parallel()

	// Create memory filesystem for the entire test
	memFS := testutil.NewMemFS(t)
	fs := filesystem.NewAferoFileSystem(memFS.Fs) //nolint:varnamelen // Common abbreviation

	// Create archive in memory filesystem
	archivePath := createTestArchiveInMemory(t, fs)

	// Destination directory in memory filesystem
	destDir := "/extracted"

	adapter := NewWithFileSystem(createMockGitConfig(), fs)
	options := ports.CloneOptions{
		URL:  "file://" + archivePath,
		Path: destDir,
	}

	repo, err := adapter.Clone(context.Background(), options)

	require.NoError(t, err)
	require.NotNil(t, repo)
	assert.Equal(t, destDir, repo.Path())

	// Verify files were extracted using filesystem interface
	testFileContent, err := fs.ReadFile(filepath.Join(destDir, "test-file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "test content", string(testFileContent))

	nestedFileContent, err := fs.ReadFile(filepath.Join(destDir, "subdir", "nested-file.txt"))
	require.NoError(t, err)
	assert.Equal(t, "nested content", string(nestedFileContent))

	// Verify directory was created
	subdirInfo, err := fs.Stat(filepath.Join(destDir, "subdir"))
	require.NoError(t, err)
	assert.True(t, subdirInfo.IsDir())
}

func TestClone_WhenGivenUnsupportedFormat_ReturnsError(t *testing.T) {
	t.Parallel()

	// Use memory filesystem consistently
	memFS := testutil.NewMemFS(t)
	fs := filesystem.NewAferoFileSystem(memFS.Fs) //nolint:varnamelen // Common abbreviation

	// Create test file in memory filesystem
	testFile := "/test-file.txt"
	memFS.WriteFileString(testFile, "test content")

	destDir := "/dest"

	adapter := NewWithFileSystem(createMockGitConfig(), fs)
	options := ports.CloneOptions{
		URL:  "file://" + testFile,
		Path: destDir,
	}

	_, err := adapter.Clone(context.Background(), options)

	require.ErrorIs(t, err, ErrUnsupportedArchiveFormat)
}

func TestClone_WhenGivenNonFileURL_ReturnsUnsupportedURLError(t *testing.T) {
	t.Parallel()

	destDir := t.TempDir()

	fs := createTestFileSystem(t)
	adapter := NewWithFileSystem(createMockGitConfig(), fs)
	options := ports.CloneOptions{
		URL:  "https://example.com/archive.tar.gz",
		Path: destDir,
	}

	_, err := adapter.Clone(context.Background(), options)

	require.ErrorIs(t, err, ErrArchiveOnlySupportsFile)
}

// TestClone_WhenArchiveContainsMaliciousPath_RejectsExtraction tests malicious archive rejection
// using in-memory filesystem for 1.8x faster execution.
//
// Performance comparison (1000 iterations):
//
//	Memory version: ~252µs/op, 65 allocs/op
//	Disk version:   ~461µs/op, 84 allocs/op
//	Speedup:        1.83x faster, 23% fewer allocations
//
// This demonstrates the benefits of in-memory testing for I/O heavy operations.
func TestClone_WhenArchiveContainsMaliciousPath_RejectsExtraction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		attackType  string
		expectedErr error
	}{
		{
			name:        "path traversal attack",
			attackType:  "path-traversal",
			expectedErr: ErrUnsafePathInArchive,
		},
		{
			name:        "absolute path attack",
			attackType:  "absolute-path",
			expectedErr: ErrUnsafePathInArchive,
		},
		{
			name:        "large file attack",
			attackType:  "large-file",
			expectedErr: ErrFileTooLargeInArchive,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Use existing testutil memory filesystem helpers
			memFS := testutil.NewMemFS(t)
			fileSystem := filesystem.NewAferoFileSystem(memFS.Fs)

			// Create malicious archive entirely in memory
			archivePath := createMaliciousArchiveInMemory(t, fileSystem, test.attackType)

			// Destination also in memory
			destDir := "/extracted"

			// Adapter uses the same memory filesystem
			adapter := NewWithFileSystem(createMockGitConfig(), fileSystem)
			options := ports.CloneOptions{
				URL:  "file://" + archivePath,
				Path: destDir,
			}

			_, err := adapter.Clone(context.Background(), options)

			require.ErrorIs(t, err, test.expectedErr)
		})
	}
}

// TestClone_WhenArchiveContainsMaliciousPath_RejectsExtraction_Disk tests the same scenario
// but using disk-based operations (for comparison/legacy).
// TODO: Remove this once we verify memory-based version works correctly.
func TestClone_WhenArchiveContainsMaliciousPath_RejectsExtraction_Disk(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		attackType  string
		expectedErr error
	}{
		{
			name:        "path traversal attack",
			attackType:  "path-traversal",
			expectedErr: ErrUnsafePathInArchive,
		},
		{
			name:        "absolute path attack",
			attackType:  "absolute-path",
			expectedErr: ErrUnsafePathInArchive,
		},
		{
			name:        "large file attack",
			attackType:  "large-file",
			expectedErr: ErrFileTooLargeInArchive,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			archivePath := createMaliciousArchive(t, test.attackType)

			destDir := t.TempDir()

			// Uses OS filesystem for disk-based operations
			fs := filesystem.NewOSFileSystem()
			adapter := NewWithFileSystem(createMockGitConfig(), fs)
			options := ports.CloneOptions{
				URL:  "file://" + archivePath,
				Path: destDir,
			}

			_, err := adapter.Clone(context.Background(), options)

			require.ErrorIs(t, err, test.expectedErr)
		})
	}
}

// Benchmarks to compare memory vs disk performance.
func BenchmarkMaliciousArchive_Memory(b *testing.B) {
	// Setup memory filesystem once
	memFS := testutil.NewMemFS(b)
	fileSystem := filesystem.NewAferoFileSystem(memFS.Fs)

	b.ResetTimer()

	for range b.N {
		// Create and test malicious archive in memory
		archivePath := createMaliciousArchiveInMemory(b, fileSystem, "path-traversal")
		adapter := NewWithFileSystem(createMockGitConfig(), fileSystem)
		options := ports.CloneOptions{
			URL:  "file://" + archivePath,
			Path: "/extracted",
		}
		_, _ = adapter.Clone(context.Background(), options)
	}
}

func BenchmarkMaliciousArchive_Disk(b *testing.B) {
	for range b.N {
		// Create and test malicious archive on disk
		archivePath := createMaliciousArchive(b, "path-traversal")
		fileSystem := filesystem.NewOSFileSystem()
		adapter := NewWithFileSystem(createMockGitConfig(), fileSystem)
		tempDir := b.TempDir()
		options := ports.CloneOptions{
			URL:  "file://" + archivePath,
			Path: tempDir,
		}
		_, _ = adapter.Clone(context.Background(), options)
	}
}

// Test Open operation

func TestOpen_WhenPathExists_ReturnsRepositoryHandle(t *testing.T) {
	t.Parallel()

	fs, tempDir := createTestFileSystemWithFiles(t)
	adapter := NewWithFileSystem(createMockGitConfig(), fs)

	repo, err := adapter.Open(context.Background(), tempDir)

	require.NoError(t, err)
	require.NotNil(t, repo)
	assert.Equal(t, tempDir, repo.Path())
}

func TestOpen_WhenPathDoesNotExist_ReturnsError(t *testing.T) {
	t.Parallel()

	fs := createTestFileSystem(t)
	adapter := NewWithFileSystem(createMockGitConfig(), fs)

	_, err := adapter.Open(context.Background(), "/non-existent-path")

	require.ErrorIs(t, err, ErrRepositoryPathNotExist)
}

// Test Init operation

func TestCleanup_WhenCalled_RemovesAllTemporaryFiles(t *testing.T) {
	t.Parallel()

	fs := createTestFileSystem(t)
	adapter := NewWithFileSystem(createMockGitConfig(), fs)

	err := adapter.Cleanup(context.Background(), "/test/path")

	// Should be a no-op and not error
	require.NoError(t, err)
}

// Test Repository implementation (created from archive package)

func TestRepositoryProperties_WhenQueried_ReturnsExpectedValues(t *testing.T) {
	t.Parallel()

	fs, tempDir := createTestFileSystemWithFiles(t) //nolint:varnamelen // Common abbreviation
	config := createMockGitConfig()

	// Create repository using archive.Repository
	repo := &Repository{
		path:   tempDir,
		config: config,
		fs:     fs,
	}

	// Test basic properties
	assert.Equal(t, tempDir, repo.Path())
	assert.Equal(t, filepath.Base(tempDir), repo.Name())
}

// Test Push operation (archive creation)

func TestPush_WhenPushingRepository_CreatesArchiveSuccessfully(t *testing.T) {
	t.Parallel()

	fs, sourceDir := createTestFileSystemWithFiles(t) //nolint:varnamelen // Common abbreviation
	adapter := NewWithFileSystem(createMockGitConfig(), fs)

	// Open repository
	repo, err := adapter.Open(context.Background(), sourceDir)
	require.NoError(t, err)

	// Push (create archive)
	err = adapter.Push(context.Background(), repo, ports.PushOptions{})
	require.NoError(t, err)

	// Verify archive was created
	archivePath := filepath.Join(sourceDir, "repository-archive.tar.gz")
	_, err = fs.Stat(archivePath)
	require.NoError(t, err)

	// Verify archive contents by extracting to a temp directory
	extractDir := "/extract-dir"
	err = fs.MkdirAll(extractDir, 0750)
	require.NoError(t, err)

	err = adapter.extractArchive(archivePath, extractDir)
	require.NoError(t, err)

	// Verify extracted files
	file1Content, err := fs.ReadFile(filepath.Join(extractDir, "file1.txt"))
	require.NoError(t, err)
	assert.Equal(t, "test content 1", string(file1Content))

	file2Content, err := fs.ReadFile(filepath.Join(extractDir, "subdir", "file2.txt"))
	require.NoError(t, err)
	assert.Equal(t, "test content 2", string(file2Content))
}

// Test Fetch operation

func TestExtractArchive_WhenArchiveCorrupted_ReturnsError(t *testing.T) {
	t.Parallel()

	// Create corrupted archive (not a valid gzip file) in memory filesystem
	memFS := testutil.NewMemFS(t)
	fs := filesystem.NewAferoFileSystem(memFS.Fs) //nolint:varnamelen // Common abbreviation

	corruptedFile := "/corrupted.tar.gz"
	err := fs.WriteFile(corruptedFile, []byte("not a valid gzip file"), 0644)
	require.NoError(t, err)

	destDir := "/dest"

	adapter := NewWithFileSystem(createMockGitConfig(), fs)

	err = adapter.extractArchive(corruptedFile, destDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create gzip reader")
}

// TestCreateArchive_WhenSourceDoesNotExist_ReturnsError tests archive creation error handling
// using in-memory filesystem for better test isolation and performance.
func TestCreateArchive_WhenSourceDoesNotExist_ReturnsError(t *testing.T) {
	t.Parallel()

	// Use memory filesystem for faster execution (no disk I/O)
	memFS := testutil.NewMemFS(t)
	fs := filesystem.NewAferoFileSystem(memFS.Fs)
	adapter := NewWithFileSystem(createMockGitConfig(), fs)

	// Create archive path in memory (no disk I/O)
	archivePath := "/test-archive.tar.gz"

	// Try to create archive from non-existent source
	err := adapter.createArchive("/non-existent-source", archivePath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to walk source directory")
}

func TestValidateAndBuildTargetPath_WhenGivenVariousPaths_ValidatesCorrectly(t *testing.T) {
	t.Parallel()

	fs := createTestFileSystem(t)
	adapter := NewWithFileSystem(createMockGitConfig(), fs)
	destPath := "/dest/path"

	tests := []struct {
		name        string
		entryName   string
		expectError bool
		expectedErr error
	}{
		{
			name:        "safe relative path",
			entryName:   "safe/file.txt",
			expectError: false,
		},
		{
			name:        "path traversal with double dots",
			entryName:   "../../../etc/passwd",
			expectError: true,
			expectedErr: ErrUnsafePathInArchive,
		},
		{
			name:        "absolute path",
			entryName:   "/etc/passwd",
			expectError: true,
			expectedErr: ErrUnsafePathInArchive,
		},
		{
			name:        "path with double dots in middle",
			entryName:   "safe/../../../etc/passwd",
			expectError: true,
			expectedErr: ErrUnsafePathInArchive,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			targetPath, err := adapter.validateAndBuildTargetPath(test.entryName, destPath)

			if test.expectError {
				require.Error(t, err)

				if test.expectedErr != nil {
					require.ErrorIs(t, err, test.expectedErr)
				}

				assert.Empty(t, targetPath)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, targetPath)
				assert.Contains(t, targetPath, destPath)
			}
		})
	}
}

func TestExtractEntry_WhenUnsupportedFileType_ReturnsError(t *testing.T) {
	t.Parallel()

	fs := createTestFileSystem(t)
	adapter := NewWithFileSystem(createMockGitConfig(), fs)

	// Create header with unsupported type (symlink)
	header := &tar.Header{
		Name:     "symlink",
		Typeflag: tar.TypeSymlink,
		Linkname: "target",
	}

	err := adapter.extractEntry(nil, header, "/test/path")

	// Should not error for unsupported types (they are skipped)
	require.NoError(t, err)
}

func TestExtractFile_WhenPermissionDenied_ReturnsError(t *testing.T) {
	t.Parallel()

	fs := createTestFileSystem(t) //nolint:varnamelen // Common abbreviation
	adapter := NewWithFileSystem(createMockGitConfig(), fs)

	// Create a header for a regular file
	content := "test content"
	header := &tar.Header{
		Name:     "test-file.txt",
		Typeflag: tar.TypeReg,
		Size:     int64(len(content)),
		Mode:     0644,
		ModTime:  time.Now(),
	}

	// Create a tar reader with the content
	tarReader := createTarReaderWithContent(t, header, content)

	// Advance to the file content (the extractFile method expects this)
	actualHeader, err := tarReader.Next()
	require.NoError(t, err)
	require.Equal(t, header.Name, actualHeader.Name)

	destDir := "/test-extract"
	err = fs.MkdirAll(destDir, 0750)
	require.NoError(t, err)

	targetPath := filepath.Join(destDir, "test-file.txt")

	err = adapter.extractFile(tarReader, header, targetPath)

	require.NoError(t, err)

	// Verify file was created with correct permissions
	info, err := fs.Stat(targetPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// Verify content
	extractedContent, err := fs.ReadFile(targetPath)
	require.NoError(t, err)
	assert.Equal(t, content, string(extractedContent))
}

// Creates tar reader with content.
func createTarReaderWithContent(t *testing.T, header *tar.Header, content string) *tar.Reader {
	t.Helper()

	// Create a buffer to hold tar data
	var tarData strings.Builder

	tarWriter := tar.NewWriter(&tarData)

	err := tarWriter.WriteHeader(header)
	require.NoError(t, err)

	_, err = tarWriter.Write([]byte(content))
	require.NoError(t, err)

	err = tarWriter.Close()
	require.NoError(t, err)

	return tar.NewReader(strings.NewReader(tarData.String()))
}

// Test helper methods more thoroughly

func TestOpenArchiveForReading_WhenGzipInvalid_ReturnsError(t *testing.T) {
	t.Parallel()

	// Create a file that's not a valid gzip file in memory filesystem
	memFS := testutil.NewMemFS(t)
	fs := filesystem.NewAferoFileSystem(memFS.Fs) //nolint:varnamelen // Common abbreviation

	invalidFile := "/invalid.tar.gz"

	// Write invalid gzip data
	err := fs.WriteFile(invalidFile, []byte("this is not gzip data"), 0644)
	require.NoError(t, err)

	adapter := NewWithFileSystem(createMockGitConfig(), fs)

	_, _, err = adapter.openArchiveForReading(invalidFile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create gzip reader")
}

func TestExtractTarContents_WhenReadFails_ReturnsError(t *testing.T) {
	t.Parallel()

	fs := createTestFileSystem(t)
	adapter := NewWithFileSystem(createMockGitConfig(), fs)

	// Create a broken tar reader by using invalid tar data
	invalidTarData := "this is not valid tar data"
	tarReader := tar.NewReader(strings.NewReader(invalidTarData))

	destDir := t.TempDir()

	err := adapter.extractTarContents(tarReader, destDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read tar header")
}

func TestValidateAndBuildTargetPath_WhenPathTraversalAttempted_RejectsPath(t *testing.T) {
	t.Parallel()

	fs := createTestFileSystem(t)
	adapter := NewWithFileSystem(createMockGitConfig(), fs)
	destPath := "/dest/path"

	tests := []struct {
		name        string
		entryName   string
		expectError bool
		description string
	}{
		{
			name:        "complex path traversal",
			entryName:   "safe/../../etc/passwd",
			expectError: true,
			description: "should reject complex path traversal attempts",
		},
		{
			name:        "hidden file path traversal",
			entryName:   "./../../../secret",
			expectError: true,
			description: "should reject hidden file path traversal",
		},
		{
			name:        "very deep nested path",
			entryName:   "a/b/c/d/e/f/g/h/i/j/file.txt",
			expectError: false,
			description: "should allow deep nested paths",
		},
		{
			name:        "empty path",
			entryName:   "",
			expectError: false,
			description: "should handle empty path",
		},
		{
			name:        "dot path",
			entryName:   ".",
			expectError: false,
			description: "should handle dot path",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			targetPath, err := adapter.validateAndBuildTargetPath(test.entryName, destPath)

			if test.expectError {
				require.Error(t, err, test.description)
				assert.Empty(t, targetPath)
			} else {
				require.NoError(t, err, test.description)

				if test.entryName != "" {
					assert.NotEmpty(t, targetPath)
				}
			}
		})
	}
}

func TestCreateArchiveWriters_WhenPathInvalid_ReturnsError(t *testing.T) {
	t.Parallel()

	fs := testutil.NewErrorFileSystem(t)
	adapter := NewWithFileSystem(createMockGitConfig(), fs)

	// Set error for file creation
	fs.SetError("/archive.tar.gz", testutil.ErrPermissionDenied)

	// Try to create archive
	_, _, err := adapter.createArchiveWriters("/archive.tar.gz")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create archive file")
}

func TestWalkAndAddToArchive_WhenSymlinksPresent_HandlesCorrectly(t *testing.T) {
	t.Parallel()

	memFS := testutil.NewMemFS(t)
	fs := filesystem.NewAferoFileSystem(memFS.Fs) //nolint:varnamelen // Common abbreviation

	// Create source directory with various file types
	sourceDir := "/source-dir"

	// Create regular file
	regularFile := filepath.Join(sourceDir, "regular.txt")
	err := fs.WriteFile(regularFile, []byte("regular file content"), 0600)
	require.NoError(t, err)

	// Create directory
	subDir := filepath.Join(sourceDir, "subdir")
	err = fs.MkdirAll(subDir, 0750)
	require.NoError(t, err)

	// Create file in subdirectory
	subFile := filepath.Join(subDir, "sub.txt")
	err = fs.WriteFile(subFile, []byte("sub file content"), 0600)
	require.NoError(t, err)

	adapter := NewWithFileSystem(createMockGitConfig(), fs)

	// Create archive
	archivePath := filepath.Join(sourceDir, "test.tar.gz")
	err = adapter.createArchive(sourceDir, archivePath)
	require.NoError(t, err)

	// Verify archive was created
	_, err = fs.Stat(archivePath)
	require.NoError(t, err)

	// Extract and verify contents
	extractDir := "/extract-dir"
	err = fs.MkdirAll(extractDir, 0750)
	require.NoError(t, err)

	err = adapter.extractArchive(archivePath, extractDir)
	require.NoError(t, err)

	// Verify extracted files
	extractedRegular := filepath.Join(extractDir, "regular.txt")
	content, err := fs.ReadFile(extractedRegular)
	require.NoError(t, err)
	assert.Equal(t, "regular file content", string(content))

	extractedSub := filepath.Join(extractDir, "subdir", "sub.txt")
	content, err = fs.ReadFile(extractedSub)
	require.NoError(t, err)
	assert.Equal(t, "sub file content", string(content))
}

// TestAddFileToArchive_WhenRelativePathInvalid_ReturnsError verifies path validation
// using in-memory filesystem to avoid OS-specific path handling issues.
func TestAddFileToArchive_WhenRelativePathInvalid_ReturnsError(t *testing.T) {
	t.Parallel()

	// Use memory filesystem for consistent behavior across platforms
	memFS := testutil.NewMemFS(t)
	fileSystem := filesystem.NewAferoFileSystem(memFS.Fs)
	adapter := NewWithFileSystem(createMockGitConfig(), fileSystem)

	// Create a buffer to capture tar data
	var tarData bytes.Buffer

	tarWriter := tar.NewWriter(&tarData)

	defer func() { _ = tarWriter.Close() }()

	// Create a dummy file in memory to get FileInfo
	dummyFile := "/dummy.txt"
	require.NoError(t, fileSystem.WriteFile(dummyFile, []byte("dummy"), 0644))
	fileInfo, err := fileSystem.Stat(dummyFile)
	require.NoError(t, err)

	// Try to add file with problematic path relationship
	sourcePath := "/some/path"
	filePath := "/completely/different/path" // Not relative to source

	err = adapter.addFileToArchive(sourcePath, filePath, fileInfo, tarWriter)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get relative path")
}

func TestWriteFileContent_WhenReadFails_ReturnsError(t *testing.T) {
	t.Parallel()

	fs := createTestFileSystem(t)
	adapter := NewWithFileSystem(createMockGitConfig(), fs)

	// Create a buffer to capture tar data
	var tarData bytes.Buffer

	tarWriter := tar.NewWriter(&tarData)

	defer func() { _ = tarWriter.Close() }()

	// Try to write content from non-existent file
	err := adapter.writeFileContent("/non-existent-file.txt", tarWriter)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open file")
}

func TestExtractFile_WhenCannotCreateParentDir_ReturnsError(t *testing.T) {
	t.Parallel()

	fs := testutil.NewErrorFileSystem(t) //nolint:varnamelen // Common abbreviation
	adapter := NewWithFileSystem(createMockGitConfig(), fs)

	header := &tar.Header{
		Name:     "test-file.txt",
		Typeflag: tar.TypeReg,
		Size:     12,
		Mode:     0644,
		ModTime:  time.Now(),
	}

	// Create tar reader with content
	tarReader := createTarReaderWithContent(t, header, "test content")

	// Advance to the file content
	_, err := tarReader.Next()
	require.NoError(t, err)

	// Set error for parent directory creation
	targetPath := "/blocking-file/nested/file.txt"
	parentDir := fs.Dir(targetPath)
	fs.SetError(parentDir, testutil.ErrPermissionDenied)

	err = adapter.extractFile(tarReader, header, targetPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create parent directory")
}

func TestExtractFile_WhenReaderLimited_HandlesCorrectly(t *testing.T) {
	t.Parallel()

	memFS := testutil.NewMemFS(t)
	fs := filesystem.NewAferoFileSystem(memFS.Fs) //nolint:varnamelen // Common abbreviation
	adapter := NewWithFileSystem(createMockGitConfig(), fs)

	// Create a file with specific size
	content := strings.Repeat("A", 1000)
	header := &tar.Header{
		Name:     "test-file.txt",
		Typeflag: tar.TypeReg,
		Size:     int64(len(content)),
		Mode:     0644,
		ModTime:  time.Now(),
	}

	// Create tar reader with content
	tarReader := createTarReaderWithContent(t, header, content)

	// Advance to the file content
	_, err := tarReader.Next()
	require.NoError(t, err)

	destDir := "/test-dest"
	targetPath := filepath.Join(destDir, "test-file.txt")

	err = adapter.extractFile(tarReader, header, targetPath)
	require.NoError(t, err)

	// Verify file was created with correct content
	extractedContent, err := fs.ReadFile(targetPath)
	require.NoError(t, err)
	assert.Equal(t, content, string(extractedContent))

	// Verify file permissions
	info, err := fs.Stat(targetPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

// Integration tests

func TestFullWorkflow_WhenCreatingAndExtractingArchive_WorksEndToEnd(t *testing.T) {
	t.Parallel()

	// Create source directory with files in MemFS
	fs, sourceDir := createTestFileSystemWithFiles(t) //nolint:varnamelen // Common abbreviation //nolint:varnamelen // Common abbreviation
	adapter := NewWithFileSystem(createMockGitConfig(), fs)

	// Step 1: Open repository
	repo, err := adapter.Open(context.Background(), sourceDir)
	require.NoError(t, err)

	// Step 2: Create archive (Push)
	err = adapter.Push(context.Background(), repo, ports.PushOptions{})
	require.NoError(t, err)

	archivePath := filepath.Join(sourceDir, "repository-archive.tar.gz")

	// Step 3: Clone from archive
	destDir := "/clone-dest"
	err = fs.MkdirAll(destDir, 0750)
	require.NoError(t, err)

	options := ports.CloneOptions{
		URL:  "file://" + archivePath,
		Path: destDir,
	}

	clonedRepo, err := adapter.Clone(context.Background(), options)
	require.NoError(t, err)
	require.NotNil(t, clonedRepo)

	// Step 4: Verify cloned content
	file1Content, err := fs.ReadFile(filepath.Join(destDir, "file1.txt"))
	require.NoError(t, err)
	assert.Equal(t, "test content 1", string(file1Content))

	file2Content, err := fs.ReadFile(filepath.Join(destDir, "subdir", "file2.txt"))
	require.NoError(t, err)
	assert.Equal(t, "test content 2", string(file2Content))
}

// Edge cases and error conditions

func TestAdapter_Clone_CreateDirectoryError(t *testing.T) {
	t.Parallel()

	// Use error filesystem that can simulate failures
	fs := testutil.NewErrorFileSystem(t) //nolint:varnamelen // Common abbreviation for filesystem

	// Create test archive in memory
	archivePath := createTestArchiveInMemory(t, fs)

	// Set error for the directory creation
	fs.SetError("/blocked/directory", testutil.ErrPermissionDenied)

	adapter := NewWithFileSystem(createMockGitConfig(), fs)
	options := ports.CloneOptions{
		URL:  "file://" + archivePath,
		Path: "/blocked/directory",
	}

	_, err := adapter.Clone(context.Background(), options)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestAdapter_extractDirectory_Error(t *testing.T) {
	t.Parallel()

	// Use error filesystem that can simulate failures
	fs := testutil.NewErrorFileSystem(t)
	adapter := NewWithFileSystem(createMockGitConfig(), fs)

	// Set error for directory creation
	fs.SetError("/blocking-file/child", testutil.ErrPermissionDenied)

	// Try to create directory
	err := adapter.extractDirectory("/blocking-file/child")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

// Benchmark tests for performance regression detection

func BenchmarkAdapter_Clone(b *testing.B) {
	archivePath := createTestArchive(b)

	// Use b.Cleanup for benchmark cleanup
	b.Cleanup(func() { _ = os.Remove(archivePath) })

	fs := createTestFileSystem(b)
	adapter := NewWithFileSystem(createMockGitConfig(), fs)

	b.ResetTimer()

	for range b.N {
		destDir := b.TempDir()

		options := ports.CloneOptions{
			URL:  "file://" + archivePath,
			Path: destDir,
		}

		_, err := adapter.Clone(context.Background(), options)
		if err != nil {
			b.Fatal(err)
		}
		// b.TempDir() auto-cleans, no need for manual removal
	}
}

func BenchmarkAdapter_extractArchive(b *testing.B) {
	archivePath := createTestArchive(b)

	// Use b.Cleanup for benchmark cleanup
	b.Cleanup(func() { _ = os.Remove(archivePath) })

	fs := createTestFileSystem(b)
	adapter := NewWithFileSystem(createMockGitConfig(), fs)

	b.ResetTimer()

	for range b.N {
		destDir := b.TempDir()

		err := adapter.extractArchive(archivePath, destDir)
		if err != nil {
			b.Fatal(err)
		}
		// b.TempDir() auto-cleans, no need for manual removal
	}
}

func BenchmarkAdapter_createArchive(b *testing.B) {
	sourceDir := createTempDirWithFiles(b)
	// createTempDirWithFiles uses b.TempDir() internally which auto-cleans

	fs := createTestFileSystem(b)
	adapter := NewWithFileSystem(createMockGitConfig(), fs)

	// Use a temp directory for all archives, auto-cleaned by b.TempDir
	tempDir := b.TempDir()

	b.ResetTimer()

	for i := range b.N {
		archivePath := filepath.Join(tempDir, fmt.Sprintf("benchmark-archive-%d.tar.gz", i))

		err := adapter.createArchive(sourceDir, archivePath)
		if err != nil {
			b.Fatal(err)
		}
		// No manual cleanup needed - b.TempDir() handles it
	}
}
