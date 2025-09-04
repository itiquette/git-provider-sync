// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package entities

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain"
)

func TestNewTmpDir_ValidPath_CreatesValueObject(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "absolute path",
			path:     "/tmp/test",
			expected: "/tmp/test",
		},
		{
			name:     "relative path gets cleaned",
			path:     "./test/../test",
			expected: "test",
		},
		{
			name:     "path with extra separators gets cleaned",
			path:     "/tmp//test//",
			expected: "/tmp/test",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tmpDir, err := NewTmpDir(testCase.path)
			require.NoError(t, err)
			assert.Equal(t, testCase.expected, tmpDir.Path())
		})
	}
}

func TestTmpDir_SubDir_CreatesSubdirectory(t *testing.T) {
	t.Parallel()

	tmpDir, err := NewTmpDir("/tmp/parent")
	require.NoError(t, err)

	subDir := tmpDir.SubDir("child")
	assert.Equal(t, "/tmp/parent/child", subDir.Path())
}

func TestTmpDir_SubDir_CleansPath(t *testing.T) {
	t.Parallel()

	tmpDir, err := NewTmpDir("/tmp/parent")
	require.NoError(t, err)

	subDir := tmpDir.SubDir("../dangerous/../child")
	assert.Equal(t, "/tmp/parent/child", subDir.Path())
}

func TestTmpDir_Exists_ChecksFilesystem(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		setupFunc func(t *testing.T) string
		expected  bool
	}{
		{
			name: "existing directory",
			setupFunc: func(t *testing.T) string {
				t.Helper()

				return t.TempDir() // Directory exists for this specific test
			},
			expected: true,
		},
		{
			name: "non-existing directory",
			setupFunc: func(_ *testing.T) string {
				return "/non/existent/path"
			},
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			path := testCase.setupFunc(t)
			tmpDir, err := NewTmpDir(path)
			require.NoError(t, err)

			assert.Equal(t, testCase.expected, tmpDir.Exists())
		})
	}
}

// Test the legacy context-based functions for backwards compatibility.
func TestCreateTmpDir_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		name          string
		dir           string
		prefix        string
		shouldSucceed bool
	}{
		{
			name:          "valid parameters",
			dir:           "",
			prefix:        "test",
			shouldSucceed: true,
		},
		{
			name:          "custom directory",
			dir:           os.TempDir(),
			prefix:        "custom",
			shouldSucceed: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			newCtx, err := CreateTmpDir(ctx, testCase.dir, testCase.prefix)

			if testCase.shouldSucceed {
				require.NoError(t, err)

				tmpDir, err := GetTmpDirPath(newCtx)
				require.NoError(t, err)
				assert.NotEmpty(t, tmpDir)
				assert.True(t, strings.HasPrefix(filepath.Base(tmpDir), testCase.prefix))

				// Clean up
				cleanupErr := DeleteTmpDir(newCtx)
				require.NoError(t, cleanupErr)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetTmpDirPath_ContextIntegration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		name          string
		setupFunc     func() context.Context
		shouldSucceed bool
		expectedErr   error
	}{
		{
			name: "with valid temp dir in context",
			setupFunc: func() context.Context {
				newCtx, err := CreateTmpDir(ctx, "", "test-get")
				require.NoError(t, err)

				return newCtx
			},
			shouldSucceed: true,
		},
		{
			name: "empty context",
			setupFunc: func() context.Context {
				return ctx
			},
			shouldSucceed: false,
			expectedErr:   domain.ErrTempDirectoryNotFound,
		},
		{
			name: "context with wrong type",
			setupFunc: func() context.Context {
				return context.WithValue(ctx, TmpDirKey{}, 123) // Wrong type
			},
			shouldSucceed: false,
			expectedErr:   domain.ErrTempDirectoryNotFound,
		},
		{
			name: "context with empty string",
			setupFunc: func() context.Context {
				return context.WithValue(ctx, TmpDirKey{}, "") // Empty string
			},
			shouldSucceed: false,
			expectedErr:   domain.ErrTempDirectoryNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testCtx := testCase.setupFunc()

			// Defer cleanup if we created a temp dir
			if testCase.shouldSucceed {
				defer func() { _ = DeleteTmpDir(testCtx) }()
			}

			tmpDir, err := GetTmpDirPath(testCtx)

			if testCase.shouldSucceed {
				require.NoError(t, err)
				assert.NotEmpty(t, tmpDir)
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.Empty(t, tmpDir)
			}
		})
	}
}

func TestCleanPath_SanitizesInput(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple path",
			input:    "simple",
			expected: "simple",
		},
		{
			name:     "path with parent directory traversal",
			input:    "../dangerous",
			expected: "dangerous",
		},
		{
			name:     "complex path traversal",
			input:    "../../dangerous/../safe",
			expected: "safe",
		},
		{
			name:     "absolute path becomes relative",
			input:    "/absolute/path",
			expected: "absolute/path",
		},
		{
			name:     "empty path becomes current directory",
			input:    "",
			expected: ".",
		},
		{
			name:     "only separators become current directory",
			input:    "///",
			expected: ".",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := CleanPath(testCase.input)
			assert.Equal(t, testCase.expected, result)
		})
	}
}
