// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package validation

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/adapters/filesystem"
	"itiquette/git-provider-sync/internal/domain/validation"
	"itiquette/git-provider-sync/internal/testutil"
)

const (
	nonexistentFilePath = "/nonexistent/path/file.txt"
	testFilePath        = "/test.txt"
	testDirPath         = "/testdir"
)

func TestStatPath(t *testing.T) {
	t.Parallel()

	// Create a shared in-memory filesystem for all tests
	memFS := testutil.NewMemFS(t)
	fileSystem := filesystem.NewAferoFileSystem(memFS.Fs)

	tests := []struct {
		name        string
		setupPath   func() string
		expectError bool
		expectDir   bool
	}{
		{
			name: "valid file path",
			setupPath: func() string {
				testFile := testFilePath
				memFS.WriteFileString(testFile, "test")

				return testFile
			},
			expectError: false,
			expectDir:   false,
		},
		{
			name: "valid directory path",
			setupPath: func() string {
				dirPath := testDirPath
				memFS.CreateDir(dirPath)

				return dirPath
			},
			expectError: false,
			expectDir:   true,
		},
		{
			name: "nonexistent path",
			setupPath: func() string {
				return nonexistentFilePath
			},
			expectError: true,
		},
		{
			name: "relative path",
			setupPath: func() string {
				// Create file at absolute path that relative path will resolve to
				absolutePath := testFilePath
				memFS.WriteFileString(absolutePath, "test")

				// Return relative path that will be converted to absolute
				return "./test.txt"
			},
			expectError: false,
			expectDir:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			path := test.setupPath()
			info, err := statPath(fileSystem, path)

			if test.expectError {
				require.Error(t, err)
				assert.Nil(t, info)
				assert.Contains(t, err.Error(), "failed to stat path")
			} else {
				require.NoError(t, err)
				require.NotNil(t, info)
				assert.Equal(t, test.expectDir, info.IsDir())
			}
		})
	}
}

func TestIsReadable(t *testing.T) {
	t.Parallel()

	// Create a shared in-memory filesystem for all tests
	memFS := testutil.NewMemFS(t)
	fileSystem := filesystem.NewAferoFileSystem(memFS.Fs)

	tests := []struct {
		name          string
		setupPath     func() string
		expectedValue bool
	}{
		{
			name: "readable file",
			setupPath: func() string {
				testFile := "/readable.txt"
				memFS.WriteFileString(testFile, "test")

				return testFile
			},
			expectedValue: true,
		},
		{
			name: "readable directory",
			setupPath: func() string {
				dirPath := "/readable-dir"
				memFS.CreateDir(dirPath)

				return dirPath
			},
			expectedValue: true,
		},
		{
			name: "nonexistent path",
			setupPath: func() string {
				return nonexistentFilePath
			},
			expectedValue: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			path := test.setupPath()
			result := isReadable(fileSystem, path)

			assert.Equal(t, test.expectedValue, result)
		})
	}
}

func TestIsWritable(t *testing.T) {
	t.Parallel()

	// Create a shared in-memory filesystem for all tests
	memFS := testutil.NewMemFS(t)
	fileSystem := filesystem.NewAferoFileSystem(memFS.Fs)

	tests := []struct {
		name          string
		setupPath     func() string
		expectedValue bool
	}{
		{
			name: "writable file",
			setupPath: func() string {
				testFile := "/writable.txt"
				memFS.WriteFileString(testFile, "test")

				return testFile
			},
			expectedValue: true,
		},
		{
			name: "writable directory",
			setupPath: func() string {
				dirPath := "/writable-dir"
				memFS.CreateDir(dirPath)

				return dirPath
			},
			expectedValue: true,
		},
		{
			name: "read-only file",
			setupPath: func() string {
				// Note: Memory filesystem doesn't enforce permissions
				// so this test would need real filesystem
				// For now, we skip testing read-only in memory filesystem
				testFile := "/readonly.txt"
				memFS.WriteFileString(testFile, "test")
				// In memory filesystem, this will still be writable
				return testFile
			},
			expectedValue: true, // Changed to true since memory FS doesn't enforce permissions
		},
		{
			name: "nonexistent path",
			setupPath: func() string {
				return nonexistentFilePath
			},
			expectedValue: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			path := test.setupPath()
			result := isWritable(fileSystem, path)

			assert.Equal(t, test.expectedValue, result)
		})
	}
}

func TestIsWritable_DirectoryCreatesAndRemovesTempFile(t *testing.T) {
	t.Parallel()

	memFS := testutil.NewMemFS(t)
	fileSystem := filesystem.NewAferoFileSystem(memFS.Fs)

	tempDir := "/temp-test-dir"
	memFS.CreateDir(tempDir)

	// Get initial file count
	entries, err := fileSystem.ReadDir(tempDir)
	require.NoError(t, err)

	initialCount := len(entries)

	// Test writability
	result := isWritable(fileSystem, tempDir)
	assert.True(t, result)

	// Verify temp file was cleaned up
	entries, err = fileSystem.ReadDir(tempDir)
	require.NoError(t, err)

	finalCount := len(entries)

	assert.Equal(t, initialCount, finalCount, "Temporary file should be cleaned up")
}

func TestIsWritable_FilePermissions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		permissions   os.FileMode
		expectedValue bool
	}{
		{
			name:          "full permissions",
			permissions:   0600,
			expectedValue: true,
		},
		{
			name:          "read-only permissions",
			permissions:   0400,
			expectedValue: false,
		},
		{
			name:          "write-only permissions",
			permissions:   0200,
			expectedValue: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			fileSystem := filesystem.NewOSFileSystem()
			tempDir := t.TempDir()
			testFile := filepath.Join(tempDir, "test.txt")

			require.NoError(t, os.WriteFile(testFile, []byte("test"), test.permissions))

			result := isWritable(fileSystem, testFile)
			assert.Equal(t, test.expectedValue, result)
		})
	}
}

func TestFileSystemValidation_Integration(t *testing.T) {
	t.Parallel()

	// Create a complex directory structure for testing
	tempDir := t.TempDir()

	// Create subdirectories
	subDir := filepath.Join(tempDir, "subdir")
	require.NoError(t, os.MkdirAll(subDir, 0750))

	readOnlyDir := filepath.Join(tempDir, "readonly")
	require.NoError(t, os.MkdirAll(readOnlyDir, 0500))
	t.Cleanup(func() {
		// Restore permissions for cleanup
		_ = os.Chmod(readOnlyDir, 0750) //nolint:gosec // Need to restore permissions for cleanup
	})

	// Create files
	fileSystem := filesystem.NewOSFileSystem()
	regularFile := filepath.Join(tempDir, "regular.txt")
	require.NoError(t, os.WriteFile(regularFile, []byte("content"), 0600))

	readOnlyFile := filepath.Join(tempDir, "readonly.txt")
	require.NoError(t, os.WriteFile(readOnlyFile, []byte("content"), 0400))

	adapter := NewFileSystemAdapter(fileSystem)

	tests := []struct {
		name           string
		path           string
		fsType         validation.FileSystemType
		writable       bool
		expectSuccess  bool
		expectReadable bool
		expectWritable bool
	}{
		{
			name:           "writable directory",
			path:           tempDir,
			fsType:         validation.FileSystemTypeDirectory,
			writable:       true,
			expectSuccess:  true,
			expectReadable: true,
			expectWritable: true,
		},
		{
			name:           "read-only directory",
			path:           readOnlyDir,
			fsType:         validation.FileSystemTypeDirectory,
			writable:       true,
			expectSuccess:  false,
			expectReadable: true,
			expectWritable: false,
		},
		{
			name:           "regular file",
			path:           regularFile,
			fsType:         validation.FileSystemTypeFile,
			writable:       false,
			expectSuccess:  true,
			expectReadable: true,
			expectWritable: true,
		},
		{
			name:           "read-only file",
			path:           readOnlyFile,
			fsType:         validation.FileSystemTypeFile,
			writable:       true,
			expectSuccess:  false,
			expectReadable: true,
			expectWritable: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			val := validation.FileSystemValidation{
				Type:     test.fsType,
				Path:     test.path,
				Writable: test.writable,
			}

			result := adapter.ValidateFileSystem(context.Background(), val)

			assert.Equal(t, test.expectSuccess, result.Success)
			assert.True(t, result.Exists)
			assert.Equal(t, test.expectReadable, result.Readable)
			assert.Equal(t, test.expectWritable, result.Writable)

			if test.expectSuccess {
				require.NoError(t, result.Error)
			}
		})
	}
}

// Edge case tests.
func TestStatPath_InvalidCharacters(t *testing.T) {
	t.Parallel()

	// Test with path containing invalid characters (depending on OS)
	invalidPath := "/invalid\x00path"
	fileSystem := filesystem.NewOSFileSystem()

	info, err := statPath(fileSystem, invalidPath)

	require.Error(t, err)
	assert.Nil(t, info)
}

// Benchmark tests.
func BenchmarkStatPath(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "bench.txt")
	require.NoError(b, os.WriteFile(testFile, []byte("test"), 0600))

	fileSystem := filesystem.NewOSFileSystem()

	b.ResetTimer()

	for range b.N {
		_, err := statPath(fileSystem, testFile)
		require.NoError(b, err)
	}
}

func BenchmarkIsReadable(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "bench.txt")
	require.NoError(b, os.WriteFile(testFile, []byte("test"), 0600))

	fileSystem := filesystem.NewOSFileSystem()

	b.ResetTimer()

	for range b.N {
		result := isReadable(fileSystem, testFile)
		require.True(b, result)
	}
}

func BenchmarkIsWritable(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "bench.txt")
	require.NoError(b, os.WriteFile(testFile, []byte("test"), 0600))

	fileSystem := filesystem.NewOSFileSystem()

	b.ResetTimer()

	for range b.N {
		result := isWritable(fileSystem, testFile)
		require.True(b, result)
	}
}
