// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTmpDirFunctional tests temp directory operations with real filesystem.
func TestTmpDirFunctional(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		testFunc func(t *testing.T, baseDir string)
	}{
		{
			name:     "create_and_cleanup_nested_directories",
			testFunc: testCreateAndCleanupNestedDirectories,
		},
		{
			name:     "concurrent_directory_operations",
			testFunc: testConcurrentDirectoryOperations,
		},
		{
			name:     "directory_permissions",
			testFunc: testDirectoryPermissions,
		},
		{
			name:     "symlink_handling",
			testFunc: testSymlinkHandling,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create isolated temp directory for each test
			testDir := filepath.Join(tmpDir, testCase.name)
			require.NoError(t, os.MkdirAll(testDir, 0750))

			testCase.testFunc(t, testDir)
		})
	}
}

func testCreateAndCleanupNestedDirectories(t *testing.T, baseDir string) {
	t.Helper()

	ctx := context.Background()

	// Create nested structure
	subDirs := []string{"level1", "level1/level2", "level1/level2/level3"}
	createdDirs := make([]string, 0, len(subDirs))

	for _, subDir := range subDirs {
		fullPath := filepath.Join(baseDir, subDir)
		err := os.MkdirAll(fullPath, 0750)
		require.NoError(t, err)

		// Verify directory exists
		assert.DirExists(t, fullPath)
		createdDirs = append(createdDirs, fullPath)
	}

	// Create some files in the directories
	testFile := filepath.Join(createdDirs[2], "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0600))
	assert.FileExists(t, testFile)

	// Test with tmpdir context functions
	ctx = context.WithValue(ctx, TempDirKey{}, baseDir)
	tmpPath, err := GetTmpDirPath(ctx)
	require.NoError(t, err)
	assert.Equal(t, baseDir, tmpPath)

	// Cleanup
	for i := len(createdDirs) - 1; i >= 0; i-- {
		err := os.RemoveAll(createdDirs[i])
		require.NoError(t, err)

		// Verify directory is removed
		assert.NoDirExists(t, createdDirs[i])
	}
}

func testConcurrentDirectoryOperations(t *testing.T, baseDir string) {
	t.Helper()

	// Test concurrent temp directory creation
	numGoroutines := 10
	errChan := make(chan error, numGoroutines)

	for idx := range numGoroutines {
		go func(id int) {
			ctx := context.Background()
			dirName := fmt.Sprintf("concurrent_%d", id)

			ctx, err := CreateTmpDir(ctx, baseDir, dirName)
			if err != nil {
				errChan <- err

				return
			}

			// Verify directory was created
			tmpPath, err := GetTmpDirPath(ctx)
			if err != nil {
				errChan <- err

				return
			}

			if _, err := os.Stat(tmpPath); os.IsNotExist(err) {
				errChan <- fmt.Errorf("directory %s does not exist", tmpPath)

				return
			}

			// Clean up
			if err := DeleteTmpDir(ctx); err != nil {
				errChan <- err

				return
			}

			errChan <- nil
		}(idx)
	}

	// Wait for all goroutines to complete
	for range numGoroutines {
		err := <-errChan
		assert.NoError(t, err)
	}
}

func testDirectoryPermissions(t *testing.T, baseDir string) {
	t.Helper()

	ctx := context.Background()

	// Create directory using CreateTmpDir
	ctx, err := CreateTmpDir(ctx, baseDir, "permissions-test")
	require.NoError(t, err)

	tmpPath, err := GetTmpDirPath(ctx)
	require.NoError(t, err)

	// Check default permissions
	info, err := os.Stat(tmpPath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// On Unix systems, check that we can read and write
	testFile := filepath.Join(tmpPath, "perm-test.txt")
	err = os.WriteFile(testFile, []byte("permission test"), 0600)
	require.NoError(t, err)

	content, err := os.ReadFile(testFile) //nolint:gosec // Test file with controlled path
	require.NoError(t, err)
	assert.Equal(t, "permission test", string(content))

	// Cleanup
	err = DeleteTmpDir(ctx)
	require.NoError(t, err)
}

func testSymlinkHandling(t *testing.T, baseDir string) {
	t.Helper()

	ctx := context.Background()

	// Create target directory
	targetDir := filepath.Join(baseDir, "symlink-target")
	require.NoError(t, os.MkdirAll(targetDir, 0750))

	// Create a file in target
	targetFile := filepath.Join(targetDir, "target.txt")
	require.NoError(t, os.WriteFile(targetFile, []byte("target content"), 0600))

	// Create symlink
	symlinkPath := filepath.Join(baseDir, "symlink")

	err := os.Symlink(targetDir, symlinkPath)
	require.NoError(t, err, "Failed to create symlink - OS may not support symlinks")

	// Verify symlink points to target
	linkTarget, err := os.Readlink(symlinkPath)
	require.NoError(t, err)
	assert.Equal(t, targetDir, linkTarget)

	// Test accessing file through symlink
	symlinkFile := filepath.Join(symlinkPath, "target.txt")
	content, err := os.ReadFile(symlinkFile) //nolint:gosec // Test file with controlled path
	require.NoError(t, err)
	assert.Equal(t, "target content", string(content))

	// Test context-based tmpdir with symlink
	ctx = context.WithValue(ctx, TempDirKey{}, symlinkPath)
	tmpPath, err := GetTmpDirPath(ctx)
	require.NoError(t, err)
	assert.Equal(t, symlinkPath, tmpPath)

	// Cleanup symlink
	require.NoError(t, os.Remove(symlinkPath))

	// Cleanup target
	require.NoError(t, os.RemoveAll(targetDir))
}

// TestTmpDirIntegration tests integration with other components.
func TestTmpDir_CreateAndCleanup_WorksWithFileSystem(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	ctx := context.Background()

	// Test workflow: create temp space, use it, clean up
	// Create workspace
	ctx, err := CreateTmpDir(ctx, tmpDir, "workspace")
	require.NoError(t, err)

	workspacePath, err := GetTmpDirPath(ctx)
	require.NoError(t, err)

	// Simulate git operations - create a mock git repo structure
	gitDir := filepath.Join(workspacePath, ".git")
	require.NoError(t, os.MkdirAll(gitDir, 0750))

	// Create typical git files
	gitFiles := map[string]string{
		".git/config": "[core]\n\tbare = false",
		".git/HEAD":   "ref: refs/heads/main",
		"README.md":   "# Test Repository",
		"src/main.go": "package main\n\nfunc main() {}",
		"go.mod":      "module test\n\ngo 1.21",
	}

	for filePath, content := range gitFiles {
		fullPath := filepath.Join(workspacePath, filePath)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0750))
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0600))
	}

	// Verify all files were created
	for filePath := range gitFiles {
		fullPath := filepath.Join(workspacePath, filePath)
		assert.FileExists(t, fullPath)
	}

	// Test subdirectory operations
	subPath := filepath.Join(workspacePath, "subdir")
	require.NoError(t, os.MkdirAll(subPath, 0750))

	subFile := filepath.Join(subPath, "subfile.txt")
	require.NoError(t, os.WriteFile(subFile, []byte("sub content"), 0600))
	assert.FileExists(t, subFile)

	// Cleanup everything
	require.NoError(t, DeleteTmpDir(ctx))

	// Verify cleanup
	assert.NoDirExists(t, workspacePath)
}
