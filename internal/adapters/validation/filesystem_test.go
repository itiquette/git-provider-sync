// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package validation

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/validation"
)

const nonexistentFilePath = "/nonexistent/path/file.txt"

func TestStatPath(t *testing.T) { //nolint:tparallel // Some subtests cannot run in parallel due to os.Chdir() usage
	t.Parallel()

	tests := []struct {
		name        string
		setupPath   func() string
		expectError bool
		expectDir   bool
	}{
		{
			name: "valid file path",
			setupPath: func() string {
				tempDir := t.TempDir()
				testFile := filepath.Join(tempDir, "test.txt")
				require.NoError(t, os.WriteFile(testFile, []byte("test"), 0600))

				return testFile
			},
			expectError: false,
			expectDir:   false,
		},
		{
			name: "valid directory path",
			setupPath: func() string {
				return t.TempDir()
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
				tempDir := t.TempDir()
				testFile := filepath.Join(tempDir, "test.txt")
				require.NoError(t, os.WriteFile(testFile, []byte("test"), 0600))

				// Change to temp dir and return relative path
				oldDir, err := os.Getwd()
				require.NoError(t, err)
				require.NoError(t, os.Chdir(tempDir))
				t.Cleanup(func() {
					_ = os.Chdir(oldDir)
				})

				return "./test.txt"
			},
			expectError: false,
			expectDir:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Note: Cannot use t.Parallel() with os.Chdir()
			if test.name != "relative path" {
				t.Parallel()
			}

			path := test.setupPath()
			info, err := statPath(path)

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

	tests := []struct {
		name          string
		setupPath     func() string
		expectedValue bool
	}{
		{
			name: "readable file",
			setupPath: func() string {
				tempDir := t.TempDir()
				testFile := filepath.Join(tempDir, "readable.txt")
				require.NoError(t, os.WriteFile(testFile, []byte("test"), 0600))

				return testFile
			},
			expectedValue: true,
		},
		{
			name: "readable directory",
			setupPath: func() string {
				return t.TempDir()
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
			result := isReadable(path)

			assert.Equal(t, test.expectedValue, result)
		})
	}
}

func TestIsWritable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupPath     func() string
		expectedValue bool
	}{
		{
			name: "writable file",
			setupPath: func() string {
				tempDir := t.TempDir()
				testFile := filepath.Join(tempDir, "writable.txt")
				require.NoError(t, os.WriteFile(testFile, []byte("test"), 0600))

				return testFile
			},
			expectedValue: true,
		},
		{
			name: "writable directory",
			setupPath: func() string {
				return t.TempDir()
			},
			expectedValue: true,
		},
		{
			name: "read-only file",
			setupPath: func() string {
				tempDir := t.TempDir()
				testFile := filepath.Join(tempDir, "readonly.txt")
				require.NoError(t, os.WriteFile(testFile, []byte("test"), 0400))

				return testFile
			},
			expectedValue: false,
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
			result := isWritable(path)

			assert.Equal(t, test.expectedValue, result)
		})
	}
}

func TestIsWritable_DirectoryCreatesAndRemovesTempFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Get initial file count
	entries, err := os.ReadDir(tempDir)
	require.NoError(t, err)

	initialCount := len(entries)

	// Test writability
	result := isWritable(tempDir)
	assert.True(t, result)

	// Verify temp file was cleaned up
	entries, err = os.ReadDir(tempDir)
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

			tempDir := t.TempDir()
			testFile := filepath.Join(tempDir, "test.txt")

			require.NoError(t, os.WriteFile(testFile, []byte("test"), test.permissions))

			result := isWritable(testFile)
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
	regularFile := filepath.Join(tempDir, "regular.txt")
	require.NoError(t, os.WriteFile(regularFile, []byte("content"), 0600))

	readOnlyFile := filepath.Join(tempDir, "readonly.txt")
	require.NoError(t, os.WriteFile(readOnlyFile, []byte("content"), 0400))

	adapter := NewFileSystemAdapter()

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

	info, err := statPath(invalidPath)

	require.Error(t, err)
	assert.Nil(t, info)
}

// Benchmark tests.
func BenchmarkStatPath(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "bench.txt")
	require.NoError(b, os.WriteFile(testFile, []byte("test"), 0600))

	b.ResetTimer()

	for range b.N {
		_, err := statPath(testFile)
		require.NoError(b, err)
	}
}

func BenchmarkIsReadable(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "bench.txt")
	require.NoError(b, os.WriteFile(testFile, []byte("test"), 0600))

	b.ResetTimer()

	for range b.N {
		result := isReadable(testFile)
		require.True(b, result)
	}
}

func BenchmarkIsWritable(b *testing.B) {
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "bench.txt")
	require.NoError(b, os.WriteFile(testFile, []byte("test"), 0600))

	b.ResetTimer()

	for range b.N {
		result := isWritable(testFile)
		require.True(b, result)
	}
}
