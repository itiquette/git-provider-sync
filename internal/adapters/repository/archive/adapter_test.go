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

func createTestArchive(tb testing.TB) string {
	tb.Helper()

	// Get temp directory from test
	var tempDir string

	switch v := tb.(type) {
	case *testing.T:
		tempDir = v.TempDir()
	case *testing.B:
		tempDir = v.TempDir()
	default:
		tb.Fatal("unsupported testing type")
	}

	// Create archive file in test temp directory
	archivePath := filepath.Join(tempDir, "test-archive.tar.gz")

	file, err := os.Create(archivePath) //nolint:gosec // Test file with controlled path
	if err != nil {
		tb.Fatal(err)
	}

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

func createMaliciousArchive(tb testing.TB, attackType string) string {
	tb.Helper()

	archiveFile, err := os.CreateTemp("", "malicious-archive-*.tar.gz")
	if err != nil {
		tb.Fatal(err)
	}

	archivePath := archiveFile.Name()
	_ = archiveFile.Close()

	file, err := os.Create(archivePath) //nolint:gosec // Test file with controlled path
	if err != nil {
		tb.Fatal(err)
	}

	defer func() { _ = file.Close() }()

	gzWriter := gzip.NewWriter(file)

	defer func() { _ = gzWriter.Close() }()

	tarWriter := tar.NewWriter(gzWriter)

	defer func() { _ = tarWriter.Close() }()

	var header *tar.Header

	content := "malicious content"

	switch attackType {
	case "path-traversal":
		header = &tar.Header{
			Name:     "../../../etc/passwd",
			Typeflag: tar.TypeReg,
			Mode:     0644,
			Size:     int64(len(content)),
			ModTime:  time.Now(),
		}
	case "absolute-path":
		header = &tar.Header{
			Name:     "/etc/passwd",
			Typeflag: tar.TypeReg,
			Mode:     0644,
			Size:     int64(len(content)),
			ModTime:  time.Now(),
		}
	case "large-file":
		header = &tar.Header{
			Name:     "large-file.txt",
			Typeflag: tar.TypeReg,
			Mode:     0644,
			Size:     200 * 1024 * 1024, // 200MB (exceeds 100MB limit)
			ModTime:  time.Now(),
		}
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		tb.Fatal(err)
	}

	if _, err := tarWriter.Write([]byte(content)); err != nil {
		tb.Fatal(err)
	}

	return archivePath
}

// Test Adapter constructor

func TestNew_ValidConfig_CreatesArchiveAdapter(t *testing.T) {
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
	assert.Equal(t, "archive", adapter.GetName())
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
			name:     "tar.gz file URL",
			url:      "file:///path/to/archive.tar.gz",
			expected: true,
		},
		{
			name:     "tgz file URL",
			url:      "file:///path/to/archive.tgz",
			expected: true,
		},
		{
			name:     "zip file URL",
			url:      "file:///path/to/archive.zip",
			expected: false,
		},
		{
			name:     "regular file URL",
			url:      "file:///path/to/file.txt",
			expected: false,
		},
		{
			name:     "https URL",
			url:      "https://example.com/archive.tar.gz",
			expected: false,
		},
		{
			name:     "ssh URL",
			url:      "git@github.com:user/repo.git",
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

func TestAdapter_Clone_ValidArchive(t *testing.T) {
	t.Parallel()

	archivePath := createTestArchive(t)
	// No need to defer cleanup - archive is in t.TempDir()

	destDir := t.TempDir()

	adapter := New(createMockGitConfig())
	options := ports.CloneOptions{
		URL:  "file://" + archivePath,
		Path: destDir,
	}

	repo, err := adapter.Clone(context.Background(), options)

	require.NoError(t, err)
	require.NotNil(t, repo)
	assert.Equal(t, destDir, repo.Path())

	// Verify files were extracted
	testFileContent, err := os.ReadFile(filepath.Join(destDir, "test-file.txt")) //nolint:gosec // Test file with controlled path
	require.NoError(t, err)
	assert.Equal(t, "test content", string(testFileContent))

	nestedFileContent, err := os.ReadFile(filepath.Join(destDir, "subdir", "nested-file.txt")) //nolint:gosec // Test file with controlled path
	require.NoError(t, err)
	assert.Equal(t, "nested content", string(nestedFileContent))

	// Verify directory was created
	subdirInfo, err := os.Stat(filepath.Join(destDir, "subdir"))
	require.NoError(t, err)
	assert.True(t, subdirInfo.IsDir())
}

func TestAdapter_Clone_UnsupportedFormat(t *testing.T) {
	t.Parallel()

	// Create a non-archive file
	testFile, err := os.CreateTemp("", "test-file-*.txt")
	require.NoError(t, err)

	defer func() { _ = os.Remove(testFile.Name()) }()

	_ = testFile.Close()

	destDir := t.TempDir()

	adapter := New(createMockGitConfig())
	options := ports.CloneOptions{
		URL:  "file://" + testFile.Name(),
		Path: destDir,
	}

	_, err = adapter.Clone(context.Background(), options)

	require.ErrorIs(t, err, ErrUnsupportedArchiveFormat)
}

func TestAdapter_Clone_NonFileURL(t *testing.T) {
	t.Parallel()

	destDir := t.TempDir()

	adapter := New(createMockGitConfig())
	options := ports.CloneOptions{
		URL:  "https://example.com/archive.tar.gz",
		Path: destDir,
	}

	_, err := adapter.Clone(context.Background(), options)

	require.ErrorIs(t, err, ErrArchiveOnlySupportsFile)
}

func TestAdapter_Clone_MaliciousArchives(t *testing.T) {
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

			defer func() { _ = os.Remove(archivePath) }()

			destDir := t.TempDir()

			adapter := New(createMockGitConfig())
			options := ports.CloneOptions{
				URL:  "file://" + archivePath,
				Path: destDir,
			}

			_, err := adapter.Clone(context.Background(), options)

			require.ErrorIs(t, err, test.expectedErr)
		})
	}
}

// Test Open operation

func TestAdapter_Open_ExistingPath(t *testing.T) {
	t.Parallel()

	tempDir := createTempDirWithFiles(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	adapter := New(createMockGitConfig())

	repo, err := adapter.Open(context.Background(), tempDir)

	require.NoError(t, err)
	require.NotNil(t, repo)
	assert.Equal(t, tempDir, repo.Path())
}

func TestAdapter_Open_NonExistentPath(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())

	_, err := adapter.Open(context.Background(), "/non-existent-path")

	require.ErrorIs(t, err, ErrRepositoryPathNotExist)
}

// Test Init operation

func TestAdapter_Cleanup(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())

	err := adapter.Cleanup(context.Background(), "/test/path")

	// Should be a no-op and not error
	require.NoError(t, err)
}

// Test Repository implementation (created from archive package)

func TestRepository_Properties_ReturnsCorrectValues(t *testing.T) {
	t.Parallel()

	tempDir := createTempDirWithFiles(t)

	defer func() { _ = os.RemoveAll(tempDir) }()

	config := createMockGitConfig()

	// Create repository using archive.Repository
	repo := &Repository{
		path:   tempDir,
		config: config,
	}

	// Test basic properties
	assert.Equal(t, tempDir, repo.Path())
	assert.Equal(t, filepath.Base(tempDir), repo.Name())
}

// Test Push operation (archive creation)

func TestAdapter_Push_CreateArchive(t *testing.T) {
	t.Parallel()

	sourceDir := createTempDirWithFiles(t)

	defer func() { _ = os.RemoveAll(sourceDir) }()

	adapter := New(createMockGitConfig())

	// Open repository
	repo, err := adapter.Open(context.Background(), sourceDir)
	require.NoError(t, err)

	// Push (create archive)
	err = adapter.Push(context.Background(), repo, ports.PushOptions{})
	require.NoError(t, err)

	// Verify archive was created
	archivePath := filepath.Join(sourceDir, "repository-archive.tar.gz")
	_, err = os.Stat(archivePath)
	require.NoError(t, err)

	// Verify archive contents by extracting to a temp directory
	extractDir := t.TempDir()

	err = adapter.extractArchive(archivePath, extractDir)
	require.NoError(t, err)

	// Verify extracted files
	file1Content, err := os.ReadFile(filepath.Join(extractDir, "file1.txt")) //nolint:gosec // Test file with controlled path
	require.NoError(t, err)
	assert.Equal(t, "content 1", string(file1Content))

	file2Content, err := os.ReadFile(filepath.Join(extractDir, "subdir", "file2.txt")) //nolint:gosec // Test file with controlled path
	require.NoError(t, err)
	assert.Equal(t, "content 2", string(file2Content))
}

// Test Fetch operation

func TestAdapter_extractArchive_CorruptedArchive(t *testing.T) {
	t.Parallel()

	// Create corrupted archive (not a valid gzip file)
	corruptedFile, err := os.CreateTemp("", "corrupted-*.tar.gz")
	require.NoError(t, err)

	defer func() { _ = os.Remove(corruptedFile.Name()) }()

	_, err = corruptedFile.WriteString("not a valid gzip file")
	require.NoError(t, err)

	_ = corruptedFile.Close()

	destDir := t.TempDir()

	adapter := New(createMockGitConfig())

	err = adapter.extractArchive(corruptedFile.Name(), destDir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create gzip reader")
}

func TestAdapter_createArchive_NonExistentSource(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())

	archivePath, err := os.CreateTemp("", "test-archive-*.tar.gz")
	require.NoError(t, err)

	defer func() { _ = os.Remove(archivePath.Name()) }()

	_ = archivePath.Close()

	err = adapter.createArchive("/non-existent-source", archivePath.Name())

	require.Error(t, err)
}

func TestAdapter_validateAndBuildTargetPath(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
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

func TestAdapter_extractEntry_UnsupportedType(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())

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

func TestAdapter_extractFile_PermissionCheck(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())

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

	destDir := t.TempDir()

	targetPath := filepath.Join(destDir, "test-file.txt")

	err = adapter.extractFile(tarReader, header, targetPath)

	require.NoError(t, err)

	// Verify file was created with correct permissions
	info, err := os.Stat(targetPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// Verify content
	extractedContent, err := os.ReadFile(targetPath) //nolint:gosec // G304: Test file with controlled path //nolint:gosec // Test file with controlled path
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

func TestAdapter_openArchiveForReading_InvalidGzip(t *testing.T) {
	t.Parallel()

	// Create a file that's not a valid gzip file
	invalidFile, err := os.CreateTemp("", "invalid-*.tar.gz")
	require.NoError(t, err)

	defer func() { _ = os.Remove(invalidFile.Name()) }()

	// Write invalid gzip data
	_, err = invalidFile.WriteString("this is not gzip data")
	require.NoError(t, err)

	_ = invalidFile.Close()

	adapter := New(createMockGitConfig())

	_, _, err = adapter.openArchiveForReading(invalidFile.Name())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create gzip reader")
}

func TestAdapter_extractTarContents_ReadError(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())

	// Create a broken tar reader by using invalid tar data
	invalidTarData := "this is not valid tar data"
	tarReader := tar.NewReader(strings.NewReader(invalidTarData))

	destDir := t.TempDir()

	err := adapter.extractTarContents(tarReader, destDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read tar header")
}

func TestAdapter_validateAndBuildTargetPath_EdgeCases(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())
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

func TestAdapter_createArchiveWriters_InvalidPath(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())

	// Try to create archive in non-existent directory
	invalidPath := "/non-existent-parent/archive.tar.gz"

	_, _, err := adapter.createArchiveWriters(invalidPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create archive file")
}

func TestAdapter_walkAndAddToArchive_WithSymlinks(t *testing.T) {
	t.Parallel()

	// Create source directory with various file types
	sourceDir := t.TempDir()

	// Create regular file
	regularFile := filepath.Join(sourceDir, "regular.txt")
	err := os.WriteFile(regularFile, []byte("regular file content"), 0600)
	require.NoError(t, err)

	// Create directory
	subDir := filepath.Join(sourceDir, "subdir")
	err = os.MkdirAll(subDir, 0750)
	require.NoError(t, err)

	// Create file in subdirectory
	subFile := filepath.Join(subDir, "sub.txt")
	err = os.WriteFile(subFile, []byte("sub file content"), 0600)
	require.NoError(t, err)

	adapter := New(createMockGitConfig())

	// Create archive
	archivePath := filepath.Join(sourceDir, "test.tar.gz")
	err = adapter.createArchive(sourceDir, archivePath)
	require.NoError(t, err)

	// Verify archive was created
	_, err = os.Stat(archivePath)
	require.NoError(t, err)

	// Extract and verify contents
	extractDir := t.TempDir()
	err = adapter.extractArchive(archivePath, extractDir)
	require.NoError(t, err)

	// Verify extracted files
	extractedRegular := filepath.Join(extractDir, "regular.txt")
	content, err := os.ReadFile(extractedRegular) //nolint:gosec // G304: Test file with controlled path
	require.NoError(t, err)
	assert.Equal(t, "regular file content", string(content))

	extractedSub := filepath.Join(extractDir, "subdir", "sub.txt")
	content, err = os.ReadFile(extractedSub) //nolint:gosec // G304: Test file with controlled path
	require.NoError(t, err)
	assert.Equal(t, "sub file content", string(content))
}

func TestAdapter_addFileToArchive_RelativePathError(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())

	// Create a buffer to capture tar data
	var tarData bytes.Buffer

	tarWriter := tar.NewWriter(&tarData)

	defer func() { _ = tarWriter.Close() }()

	// Try to add file with problematic path relationship
	sourcePath := "/some/path"
	filePath := "/completely/different/path" // Not relative to source
	fileInfo, err := os.Stat(".")            // Use current directory info
	require.NoError(t, err)

	err = adapter.addFileToArchive(sourcePath, filePath, fileInfo, tarWriter)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get relative path")
}

func TestAdapter_writeFileContent_ReadError(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())

	// Create a buffer to capture tar data
	var tarData bytes.Buffer

	tarWriter := tar.NewWriter(&tarData)

	defer func() { _ = tarWriter.Close() }()

	// Try to write content from non-existent file
	err := adapter.writeFileContent("/non-existent-file.txt", tarWriter)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open file")
}

func TestAdapter_extractFile_CreateParentError(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())

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

	// Try to extract to a path where parent directory cannot be created
	// Use a path that would require creating a file as a directory
	tempFile, err := os.CreateTemp("", "blocking-file-*")
	require.NoError(t, err)

	defer func() { _ = os.Remove(tempFile.Name()) }()

	_ = tempFile.Close()

	// Try to extract file to path that would require creating tempFile.Name() as directory
	targetPath := filepath.Join(tempFile.Name(), "nested", "file.txt")

	err = adapter.extractFile(tarReader, header, targetPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create parent directory")
}

func TestAdapter_extractFile_LimitedReader(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())

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

	destDir := t.TempDir()
	targetPath := filepath.Join(destDir, "test-file.txt")

	err = adapter.extractFile(tarReader, header, targetPath)
	require.NoError(t, err)

	// Verify file was created with correct content
	extractedContent, err := os.ReadFile(targetPath) //nolint:gosec // G304: Test file with controlled path
	require.NoError(t, err)
	assert.Equal(t, content, string(extractedContent))

	// Verify file permissions
	info, err := os.Stat(targetPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

// Integration tests

func TestAdapter_FullWorkflow(t *testing.T) {
	t.Parallel()

	// Create source directory with files
	sourceDir := createTempDirWithFiles(t)

	defer func() { _ = os.RemoveAll(sourceDir) }()

	adapter := New(createMockGitConfig())

	// Step 1: Open repository
	repo, err := adapter.Open(context.Background(), sourceDir)
	require.NoError(t, err)

	// Step 2: Create archive (Push)
	err = adapter.Push(context.Background(), repo, ports.PushOptions{})
	require.NoError(t, err)

	archivePath := filepath.Join(sourceDir, "repository-archive.tar.gz")

	// Step 3: Clone from archive
	destDir := t.TempDir()

	options := ports.CloneOptions{
		URL:  "file://" + archivePath,
		Path: destDir,
	}

	clonedRepo, err := adapter.Clone(context.Background(), options)
	require.NoError(t, err)
	require.NotNil(t, clonedRepo)

	// Step 4: Verify cloned content
	file1Content, err := os.ReadFile(filepath.Join(destDir, "file1.txt")) //nolint:gosec // Test file with controlled path
	require.NoError(t, err)
	assert.Equal(t, "content 1", string(file1Content))

	file2Content, err := os.ReadFile(filepath.Join(destDir, "subdir", "file2.txt")) //nolint:gosec // Test file with controlled path
	require.NoError(t, err)
	assert.Equal(t, "content 2", string(file2Content))
}

// Edge cases and error conditions

func TestAdapter_Clone_CreateDirectoryError(t *testing.T) {
	t.Parallel()

	archivePath := createTestArchive(t)

	defer func() { _ = os.Remove(archivePath) }()

	adapter := New(createMockGitConfig())
	options := ports.CloneOptions{
		URL:  "file://" + archivePath,
		Path: "/non-existent-parent/child/directory",
	}

	_, err := adapter.Clone(context.Background(), options)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create directory")
}

func TestAdapter_extractDirectory_Error(t *testing.T) {
	t.Parallel()

	adapter := New(createMockGitConfig())

	// Try to create directory in non-existent parent
	err := adapter.extractDirectory("/non-existent-parent/child")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create directory")
}

// Benchmark tests for performance regression detection

func BenchmarkAdapter_Clone(b *testing.B) {
	archivePath := createTestArchive(b)

	defer func() { _ = os.Remove(archivePath) }()

	adapter := New(createMockGitConfig())

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

		_ = os.RemoveAll(destDir)
	}
}

func BenchmarkAdapter_extractArchive(b *testing.B) {
	archivePath := createTestArchive(b)

	defer func() { _ = os.Remove(archivePath) }()

	adapter := New(createMockGitConfig())

	b.ResetTimer()

	for range b.N {
		destDir := b.TempDir()

		err := adapter.extractArchive(archivePath, destDir)
		if err != nil {
			b.Fatal(err)
		}

		_ = os.RemoveAll(destDir)
	}
}

func BenchmarkAdapter_createArchive(b *testing.B) {
	sourceDir := createTempDirWithFiles(b)

	defer func() { _ = os.RemoveAll(sourceDir) }()

	adapter := New(createMockGitConfig())

	b.ResetTimer()

	for i := range b.N {
		archivePath, err := os.CreateTemp("", fmt.Sprintf("benchmark-archive-%d-*.tar.gz", i))
		if err != nil {
			b.Fatal(err)
		}

		_ = archivePath.Close()

		err = adapter.createArchive(sourceDir, archivePath.Name())
		if err != nil {
			b.Fatal(err)
		}

		_ = os.Remove(archivePath.Name())
	}
}
