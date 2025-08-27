// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTmpDirPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupContext  func() context.Context
		expectedError error
		expectedPath  string
	}{
		{
			name: "valid_temp_dir_in_context",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), TempDirKey{}, "/tmp/test")
			},
			expectedPath: "/tmp/test",
		},
		{
			name:          "empty_context",
			setupContext:  context.Background,
			expectedError: ErrTempDirNotFound,
		},
		{
			name: "empty_string_in_context",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), TempDirKey{}, "")
			},
			expectedError: ErrTempDirNotFound,
		},
		{
			name: "wrong_type_in_context",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), TempDirKey{}, 123)
			},
			expectedError: ErrTempDirNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := test.setupContext()
			path, err := GetTmpDirPath(ctx)

			if test.expectedError != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.expectedError)
				assert.Empty(t, path)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedPath, path)
			}
		})
	}
}

func TestCreateTmpDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		dir       string
		prefix    string
		expectErr bool
	}{
		{
			name:   "valid_dir_and_prefix",
			dir:    t.TempDir(),
			prefix: "test",
		},
		{
			name:   "empty_prefix",
			dir:    t.TempDir(),
			prefix: "",
		},
		{
			name:      "invalid_directory",
			dir:       "/non/existent/path",
			prefix:    "test",
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			newCtx, err := CreateTmpDir(ctx, test.dir, test.prefix)

			if test.expectErr {
				require.Error(t, err)
				assert.Nil(t, newCtx)
				assert.Contains(t, err.Error(), "failed to create temporary directory")
			} else {
				require.NoError(t, err)
				require.NotNil(t, newCtx)

				// Verify the temp dir was stored in context
				tmpPath, err := GetTmpDirPath(newCtx)
				require.NoError(t, err)
				assert.NotEmpty(t, tmpPath)

				// Verify the directory exists
				assert.DirExists(t, tmpPath)

				// Verify it's a subdirectory of the specified dir
				assert.Contains(t, tmpPath, test.dir)

				// Verify prefix is in the name if provided
				if test.prefix != "" {
					assert.Contains(t, filepath.Base(tmpPath), test.prefix)
				}

				// Cleanup
				require.NoError(t, os.RemoveAll(tmpPath))
			}
		})
	}
}

func TestDeleteTmpDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupContext  func() (context.Context, string)
		expectedError error
	}{
		{
			name: "valid_temp_directory",
			setupContext: func() (context.Context, string) {
				tmpDir := t.TempDir()
				ctx, err := CreateTmpDir(context.Background(), tmpDir, "test")
				require.NoError(t, err)

				path, err := GetTmpDirPath(ctx)
				require.NoError(t, err)

				return ctx, path
			},
		},
		{
			name: "context_without_temp_dir",
			setupContext: func() (context.Context, string) {
				return context.Background(), ""
			},
			expectedError: ErrTempDirNotFound,
		},
		{
			name: "relative_path_in_context",
			setupContext: func() (context.Context, string) {
				ctx := context.WithValue(context.Background(), TempDirKey{}, "relative/path")

				return ctx, ""
			},
			expectedError: ErrInvalidTempDirPath,
		},
		{
			name: "path_outside_temp_directory",
			setupContext: func() (context.Context, string) {
				ctx := context.WithValue(context.Background(), TempDirKey{}, "/etc/passwd")

				return ctx, ""
			},
			expectedError: ErrInvalidTempDirPath,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx, expectedPath := test.setupContext()

			// Store whether directory existed before deletion attempt
			var dirExisted bool

			if expectedPath != "" {
				_, err := os.Stat(expectedPath)
				dirExisted = err == nil
			}

			err := DeleteTmpDir(ctx)

			if test.expectedError != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.expectedError)
			} else {
				require.NoError(t, err)

				// Verify directory was actually deleted if it existed
				if dirExisted {
					assert.NoDirExists(t, expectedPath)
				}
			}
		})
	}
}

func TestIsSubdirectoryOfTemp(t *testing.T) {
	t.Parallel()

	tempDir := os.TempDir()
	validTempPath := filepath.Join(tempDir, "test", "subdir")

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "valid_subdirectory_of_temp",
			path:     validTempPath,
			expected: true,
		},
		{
			name:     "temp_directory_itself",
			path:     tempDir,
			expected: true,
		},
		{
			name:     "root_directory",
			path:     "/",
			expected: false,
		},
		{
			name:     "etc_directory",
			path:     "/etc",
			expected: false,
		},
		{
			name:     "home_directory",
			path:     "/home/user",
			expected: false,
		},
		{
			name:     "relative_path",
			path:     "relative/path",
			expected: false,
		},
		{
			name:     "path_traversal_attempt",
			path:     tempDir + "/../etc",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := isSubdirectoryOfTemp(test.path)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestTempDirKey(t *testing.T) {
	t.Parallel()

	// Test that the key type works correctly for context storage
	key1 := TempDirKey{}
	key2 := TempDirKey{}

	// Keys should be equal
	assert.Equal(t, key1, key2)

	// Test context storage and retrieval
	ctx := context.Background()
	testValue := "/tmp/test"

	ctx = context.WithValue(ctx, key1, testValue)
	retrieved := ctx.Value(key2)

	require.NotNil(t, retrieved)
	retrievedStr, ok := retrieved.(string)
	require.True(t, ok, "expected string type")
	assert.Equal(t, testValue, retrievedStr)
}

func TestErrorConstants(t *testing.T) {
	t.Parallel()

	// Verify error constants are properly defined
	errors := []error{
		ErrTempDirNotFound,
		ErrInvalidTempDirPath,
	}

	for _, err := range errors {
		require.Error(t, err)
		assert.NotEmpty(t, err.Error())
	}

	// Verify errors are distinct
	assert.NotEqual(t, ErrTempDirNotFound.Error(), ErrInvalidTempDirPath.Error())
}

// TestTempDirWorkflow tests the complete workflow of creating, using, and cleaning up temp directories.
func TestTempDirWorkflow(t *testing.T) {
	t.Parallel()

	baseDir := t.TempDir()
	ctx := context.Background()

	// Step 1: Create temp directory
	ctx, err := CreateTmpDir(ctx, baseDir, "workflow")
	require.NoError(t, err)

	// Step 2: Get the path and verify it exists
	tmpPath, err := GetTmpDirPath(ctx)
	require.NoError(t, err)
	assert.DirExists(t, tmpPath)

	// Step 3: Use the directory (create some files)
	testFile := filepath.Join(tmpPath, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0600))
	assert.FileExists(t, testFile)

	// Step 4: Cleanup
	require.NoError(t, DeleteTmpDir(ctx))
	assert.NoDirExists(t, tmpPath)
}
