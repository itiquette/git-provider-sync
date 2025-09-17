// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// CreateTestFile creates a temporary file with the given content in the test's temp directory.
// The file is automatically cleaned up when the test ends.
func CreateTestFile(t *testing.T, filename, content string) string {
	t.Helper()

	file := filepath.Join(t.TempDir(), filename)
	require.NoError(t, os.WriteFile(file, []byte(content), 0600))

	return file
}

// CreateTestFileWithPath creates a file at a specific path within the test's temp directory.
// It creates any necessary parent directories.
func CreateTestFileWithPath(t *testing.T, relativePath, content string) string {
	t.Helper()

	fullPath := filepath.Join(t.TempDir(), relativePath)
	dir := filepath.Dir(fullPath)

	require.NoError(t, os.MkdirAll(dir, 0750))
	require.NoError(t, os.WriteFile(fullPath, []byte(content), 0600))

	return fullPath
}

// CreateTestDirectory creates a temporary directory structure for testing.
// Returns the root directory path.
func CreateTestDirectory(t *testing.T, name string) string {
	t.Helper()

	dir := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.MkdirAll(dir, 0750))

	return dir
}

// CreateTestConfig creates a temporary configuration file for testing.
func CreateTestConfig(t *testing.T, config string) string {
	t.Helper()

	return CreateTestFile(t, "gitprovidersync.yaml", config)
}

// AssertFileContents verifies that a file exists and has the expected content.
func AssertFileContents(t *testing.T, path, expectedContent string) {
	t.Helper()

	content, err := os.ReadFile(filepath.Clean(path))
	require.NoError(t, err, "Failed to read file: %s", path)
	require.Equal(t, expectedContent, string(content), "File content mismatch")
}

// AssertFileExists verifies that a file exists at the given path.
func AssertFileExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	require.NoError(t, err, "File should exist: %s", path)
	require.False(t, info.IsDir(), "Path should be a file, not a directory: %s", path)
}

// AssertDirectoryExists verifies that a directory exists at the given path.
func AssertDirectoryExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	require.NoError(t, err, "Directory should exist: %s", path)
	require.True(t, info.IsDir(), "Path should be a directory, not a file: %s", path)
}

// AssertFileNotExists verifies that no file exists at the given path.
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()

	_, err := os.Stat(path)
	require.True(t, os.IsNotExist(err), "File should not exist: %s", path)
}
