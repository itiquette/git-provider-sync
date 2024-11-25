// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package directory

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStorageHandler(t *testing.T) {
	t.Run("GetTargetPath", func(t *testing.T) {
		tests := []struct {
			name      string
			targetDir string
			repoName  string
			setup     func(string) error
			wantPath  string
			wantError bool
		}{
			{
				name:      "creates new directory path",
				targetDir: "test-target",
				repoName:  "test-repo",
				wantPath:  filepath.Join("test-target", "test-repo"),
			},
			{
				name:      "handles existing directory",
				targetDir: "existing-target",
				repoName:  "test-repo",
				setup: func(dir string) error {
					return os.MkdirAll(dir, os.ModePerm)
				},
				wantPath: filepath.Join("existing-target", "test-repo"),
			},
			{
				name:      "handles nested paths",
				targetDir: filepath.Join("deep", "nested", "path"),
				repoName:  "test-repo",
				wantPath:  filepath.Join("deep", "nested", "path", "test-repo"),
			},
		}

		for _, tabletest := range tests {
			t.Run(tabletest.name, func(t *testing.T) {
				// Setup
				tmpDir := t.TempDir()
				targetDir := filepath.Join(tmpDir, tabletest.targetDir)

				if tabletest.setup != nil {
					err := tabletest.setup(targetDir)
					require.NoError(t, err)
				}

				ctx := testContext()

				handler := NewStorageHandler()

				gotPath, err := handler.GetTargetPath(ctx, targetDir, tabletest.repoName)

				if tabletest.wantError {
					require.Error(t, err)

					return
				}

				require.NoError(t, err)
				require.Equal(t, filepath.Join(targetDir, tabletest.repoName), gotPath)

				// Verify directory was created
				exists := handler.DirectoryExists(targetDir)
				require.True(t, exists)
			})
		}
	})

	t.Run("DirectoryExists", func(t *testing.T) {
		tests := []struct {
			name       string
			setup      func(string) error
			wantExists bool
		}{
			{
				name:       "returns false for non-existent directory",
				wantExists: false,
			},
			{
				name: "returns true for existing directory",
				setup: func(dir string) error {
					return os.MkdirAll(dir, os.ModePerm)
				},
				wantExists: true,
			},
			{
				name: "returns false for file",
				setup: func(dir string) error {
					return os.WriteFile(dir, []byte("test"), 0600)
				},
				wantExists: true, // DirectoryExists doesn't distinguish files from directories
			},
		}

		for _, tabletest := range tests {
			t.Run(tabletest.name, func(t *testing.T) {
				// Setup
				tmpDir := t.TempDir()
				testPath := filepath.Join(tmpDir, "test-dir")

				if tabletest.setup != nil {
					err := tabletest.setup(testPath)
					require.NoError(t, err)
				}

				// Create handler
				handler := NewStorageHandler()

				// Test
				exists := handler.DirectoryExists(testPath)
				require.Equal(t, tabletest.wantExists, exists)
			})
		}
	})
}

func TestNewStorageHandler(t *testing.T) {
	handler := NewStorageHandler()
	require.NotNil(t, handler)
	require.IsType(t, StorageHandler{}, handler)
}
