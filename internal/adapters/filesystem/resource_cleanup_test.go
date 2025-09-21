// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
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

// TestResourceCleanup_TemporaryDirectories verifies that temp directories are properly cleaned up.
func TestResourceCleanup_TemporaryDirectories(t *testing.T) {
	t.Parallel()

	t.Run("temp directory is cleaned up after context cancellation", func(t *testing.T) {
		fileSystem := NewOSFileSystem()

		t.Parallel()

		baseDir := t.TempDir()
		ctx := context.Background()

		// Create temp directory
		ctx, err := CreateTmpDir(ctx, fileSystem, baseDir, "cleanup-test")
		require.NoError(t, err)

		// Get the path and verify it exists
		tmpPath, err := GetTmpDirPath(ctx)
		require.NoError(t, err)
		require.DirExists(t, tmpPath)

		// Create some files in the temp directory
		testFile := filepath.Join(tmpPath, "test.txt")
		err = os.WriteFile(testFile, []byte("test content"), 0600)
		require.NoError(t, err)
		require.FileExists(t, testFile)

		// Clean up
		err = DeleteTmpDir(ctx, fileSystem)
		require.NoError(t, err)

		// Verify everything is gone
		assert.NoDirExists(t, tmpPath, "Temp directory should be completely removed")
		assert.NoFileExists(t, testFile, "Files in temp directory should be removed")
	})

	t.Run("cleanup handles nested directories correctly", func(t *testing.T) {
		t.Parallel()

		fileSystem := NewOSFileSystem()

		baseDir := t.TempDir()
		ctx := context.Background()

		// Create temp directory
		ctx, err := CreateTmpDir(ctx, fileSystem, baseDir, "nested-cleanup")
		require.NoError(t, err)

		tmpPath, err := GetTmpDirPath(ctx)
		require.NoError(t, err)

		// Create nested structure
		nestedDir := filepath.Join(tmpPath, "level1", "level2", "level3")
		err = os.MkdirAll(nestedDir, 0750)
		require.NoError(t, err)

		// Create files at different levels
		files := []string{
			filepath.Join(tmpPath, "root.txt"),
			filepath.Join(tmpPath, "level1", "file1.txt"),
			filepath.Join(tmpPath, "level1", "level2", "file2.txt"),
			filepath.Join(nestedDir, "deep.txt"),
		}

		for _, file := range files {
			err = os.WriteFile(file, []byte("content"), 0600)
			require.NoError(t, err)
			require.FileExists(t, file)
		}

		// Clean up
		err = DeleteTmpDir(ctx, fileSystem)
		require.NoError(t, err)

		// Verify complete removal
		assert.NoDirExists(t, tmpPath)

		for _, file := range files {
			assert.NoFileExists(t, file)
		}
	})

	t.Run("cleanup handles read-only files", func(t *testing.T) {
		t.Parallel()

		fileSystem := NewOSFileSystem()

		baseDir := t.TempDir()
		ctx := context.Background()

		ctx, err := CreateTmpDir(ctx, fileSystem, baseDir, "readonly-cleanup")
		require.NoError(t, err)

		tmpPath, err := GetTmpDirPath(ctx)
		require.NoError(t, err)

		// Create a read-only file
		readOnlyFile := filepath.Join(tmpPath, "readonly.txt")
		err = os.WriteFile(readOnlyFile, []byte("readonly content"), 0400)
		require.NoError(t, err)

		// Clean up should handle read-only files
		err = DeleteTmpDir(ctx, fileSystem)
		require.NoError(t, err)

		assert.NoDirExists(t, tmpPath)
		assert.NoFileExists(t, readOnlyFile)
	})

	t.Run("cleanup is idempotent", func(t *testing.T) {
		t.Parallel()

		fileSystem := NewOSFileSystem()

		baseDir := t.TempDir()
		ctx := context.Background()

		ctx, err := CreateTmpDir(ctx, fileSystem, baseDir, "idempotent")
		require.NoError(t, err)

		tmpPath, err := GetTmpDirPath(ctx)
		require.NoError(t, err)

		// First cleanup
		err = DeleteTmpDir(ctx, fileSystem)
		require.NoError(t, err)
		assert.NoDirExists(t, tmpPath)

		// Second cleanup should not error
		err = DeleteTmpDir(ctx, fileSystem)
		require.NoError(t, err, "Cleanup should be idempotent")
	})

	t.Run("multiple temp directories don't interfere", func(t *testing.T) {
		t.Parallel()

		fileSystem := NewOSFileSystem()

		baseDir := t.TempDir()

		// Create multiple temp directories
		ctx1, err := CreateTmpDir(context.Background(), fileSystem, baseDir, "temp1")
		require.NoError(t, err)

		ctx2, err := CreateTmpDir(context.Background(), fileSystem, baseDir, "temp2")
		require.NoError(t, err)

		ctx3, err := CreateTmpDir(context.Background(), fileSystem, baseDir, "temp3")
		require.NoError(t, err)

		// Get paths
		path1, err := GetTmpDirPath(ctx1)
		require.NoError(t, err)
		path2, err := GetTmpDirPath(ctx2)
		require.NoError(t, err)
		path3, err := GetTmpDirPath(ctx3)
		require.NoError(t, err)

		// Verify all exist and are different
		require.DirExists(t, path1)
		require.DirExists(t, path2)
		require.DirExists(t, path3)
		assert.NotEqual(t, path1, path2)
		assert.NotEqual(t, path2, path3)
		assert.NotEqual(t, path1, path3)

		// Create a file in each
		file1 := filepath.Join(path1, "file1.txt")
		file2 := filepath.Join(path2, "file2.txt")
		file3 := filepath.Join(path3, "file3.txt")

		require.NoError(t, os.WriteFile(file1, []byte("1"), 0600))
		require.NoError(t, os.WriteFile(file2, []byte("2"), 0600))
		require.NoError(t, os.WriteFile(file3, []byte("3"), 0600))

		// Delete only the second one
		err = DeleteTmpDir(ctx2, fileSystem)
		require.NoError(t, err)

		// Verify only the second is gone
		assert.DirExists(t, path1)
		assert.NoDirExists(t, path2)
		assert.DirExists(t, path3)
		assert.FileExists(t, file1)
		assert.NoFileExists(t, file2)
		assert.FileExists(t, file3)

		// Clean up the rest
		require.NoError(t, DeleteTmpDir(ctx1, fileSystem))
		require.NoError(t, DeleteTmpDir(ctx3, fileSystem))

		assert.NoDirExists(t, path1)
		assert.NoDirExists(t, path3)
	})
}

// TestResourceCleanup_FileDescriptors verifies that file descriptors are properly closed.
func TestResourceCleanup_FileDescriptors(t *testing.T) {
	t.Parallel()

	t.Run("file handles are released after operations", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")

		// Write and close file
		file, err := os.Create(testFile) // #nosec G304 - testFile is a safe path constructed in test
		require.NoError(t, err)

		_, err = file.WriteString("test content")
		require.NoError(t, err)

		// Close the file
		err = file.Close()
		require.NoError(t, err)

		// Should be able to delete the file immediately after closing
		// Using t.Cleanup would be too late to test immediate deletion
		// This is testing the behavior that files can be deleted after closing
		err = os.Remove(testFile)
		require.NoError(t, err, "Should be able to delete file after closing handle")
		assert.NoFileExists(t, testFile)
	})

	fileSystem := NewOSFileSystem()

	t.Run("cleanup works even with open file handles", func(t *testing.T) {
		t.Parallel()

		baseDir := t.TempDir()
		ctx := context.Background()

		ctx, err := CreateTmpDir(ctx, fileSystem, baseDir, "open-files")
		require.NoError(t, err)

		tmpPath, err := GetTmpDirPath(ctx)
		require.NoError(t, err)

		// Create and open a file but don't close it
		openFile := filepath.Join(tmpPath, "open.txt")
		file, err := os.Create(openFile) // #nosec G304 - openFile is a safe path constructed in test
		require.NoError(t, err)

		defer func() {
			_ = file.Close() // Ensure cleanup even if test fails
		}()

		_, err = file.WriteString("content")
		require.NoError(t, err)

		// Try to clean up (behavior may vary by OS)
		// On Unix-like systems, this usually succeeds
		// On Windows, this might fail if file is still open
		deleteErr := DeleteTmpDir(ctx, fileSystem)

		// Close file before checking
		closeErr := file.Close()
		if closeErr != nil {
			t.Logf("Close file error (expected on some systems): %v", closeErr)
		}

		if deleteErr != nil {
			// If cleanup failed due to open file, that's expected on some systems
			assert.Contains(t, deleteErr.Error(), "remove")
		} else {
			// If cleanup succeeded, verify removal
			assert.NoDirExists(t, tmpPath)
		}
	})
}

// TestResourceCleanup_ConcurrentAccess verifies cleanup with concurrent operations.
func TestResourceCleanup_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	fileSystem := NewOSFileSystem()

	t.Run("concurrent cleanup operations are safe", func(t *testing.T) {
		t.Parallel()

		baseDir := t.TempDir()
		ctx := context.Background()

		ctx, err := CreateTmpDir(ctx, fileSystem, baseDir, "concurrent")
		require.NoError(t, err)

		tmpPath, err := GetTmpDirPath(ctx)
		require.NoError(t, err)

		// Create some files
		for range 10 {
			file := filepath.Join(tmpPath, filepath.FromSlash("file%d.txt"))
			err = os.WriteFile(file, []byte("content"), 0600)
			require.NoError(t, err)
		}

		// Try concurrent cleanup (only first should succeed, others should be no-ops)
		done := make(chan bool, 5)

		for range 5 {
			go func() {
				_ = DeleteTmpDir(ctx, fileSystem) // Ignore errors - some calls might fail

				done <- true
			}()
		}

		// Wait for all goroutines
		for range 5 {
			<-done
		}

		// Directory should be gone
		assert.NoDirExists(t, tmpPath)
	})
}

// TestResourceCleanup_ErrorScenarios tests cleanup in error conditions.
func TestResourceCleanup_ErrorScenarios(t *testing.T) {
	t.Parallel()

	fileSystem := NewOSFileSystem()

	t.Run("cleanup with invalid context", func(t *testing.T) {
		t.Parallel()

		// Context without temp dir
		ctx := context.Background()
		err := DeleteTmpDir(ctx, fileSystem)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrTempDirNotFound)
	})

	t.Run("cleanup with malformed path in context", func(t *testing.T) {
		t.Parallel()

		fileSystem := NewOSFileSystem()

		// Context with invalid path
		ctx := context.WithValue(context.Background(), TempDirKey{}, "../../etc/passwd")
		err := DeleteTmpDir(ctx, fileSystem)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidTempDirPath)
	})

	t.Run("cleanup preserves base directory", func(t *testing.T) {
		t.Parallel()

		fileSystem := NewOSFileSystem()
		baseDir := t.TempDir()
		importantFile := filepath.Join(baseDir, "important.txt")
		err := os.WriteFile(importantFile, []byte("important"), 0600)
		require.NoError(t, err)

		ctx, err := CreateTmpDir(context.Background(), fileSystem, baseDir, "temp")
		require.NoError(t, err)

		// Clean up temp dir
		err = DeleteTmpDir(ctx, fileSystem)
		require.NoError(t, err)

		// Base directory and its contents should still exist
		assert.DirExists(t, baseDir)
		assert.FileExists(t, importantFile)
	})
}
