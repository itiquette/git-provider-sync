// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package testutil_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/testutil"
)

func TestAferoMemFS(t *testing.T) {
	t.Parallel()

	testFS := testutil.NewMemFS(t)

	// Test write and read
	testFS.WriteFileString("/test.txt", "hello world")
	content := testFS.ReadFileString("/test.txt")
	assert.Equal(t, "hello world", content)

	// Test directory creation
	testFS.CreateDir("/my/nested/dir")
	testFS.AssertDirExists("/my/nested/dir")

	// Test file in directory
	testFS.WriteFileString("/my/nested/dir/file.txt", "nested content")
	testFS.AssertFileContent("/my/nested/dir/file.txt", "nested content")

	// Test non-existence
	testFS.AssertNotExists("/nonexistent")
}

func TestAferoStructure(t *testing.T) {
	t.Parallel()

	testFS := testutil.NewMemFS(t)

	// Create complex structure
	testFS.CreateStructure(map[string]string{
		"src/main.go":       "package main",
		"src/lib/":          "", // directory
		"src/lib/helper.go": "package lib",
		"README.md":         "# Test Project",
		"config/":           "", // directory
		"config/app.yaml":   "key: value",
	})

	// Verify structure
	testFS.AssertFileContent("src/main.go", "package main")
	testFS.AssertFileContent("src/lib/helper.go", "package lib")
	testFS.AssertFileContent("README.md", "# Test Project")
	testFS.AssertFileContent("config/app.yaml", "key: value")
	testFS.AssertDirExists("src/lib")
	testFS.AssertDirExists("config")
}

func TestAferoGitRepo(t *testing.T) {
	t.Parallel()

	testFS := testutil.NewMemFS(t)

	// Create git repository structure
	repoPath := testFS.CreateGitRepo("my-repo")

	// Verify git structure
	testFS.AssertDirExists(repoPath)
	testFS.AssertDirExists(repoPath + "/.git")
	testFS.AssertFileExists(repoPath + "/.git/config")
	testFS.AssertFileExists(repoPath + "/.git/HEAD")
	testFS.AssertFileExists(repoPath + "/README.md")
}

func TestAferoCopyFile(t *testing.T) {
	t.Parallel()

	testFS := testutil.NewMemFS(t)

	// Create source file
	testFS.WriteFileString("/source.txt", "original content")

	// Copy file
	testFS.CopyFile("/source.txt", "/destination.txt")

	// Verify copy
	testFS.AssertFileContent("/destination.txt", "original content")
	testFS.AssertFileContent("/source.txt", "original content") // Original still exists
}

func TestAferoListFiles(t *testing.T) {
	t.Parallel()

	testFS := testutil.NewMemFS(t)

	// Create files and directories
	testFS.CreateStructure(map[string]string{
		"/root/file1.txt": "content1",
		"/root/file2.txt": "content2",
		"/root/dir1/":     "",
		"/root/dir2/":     "",
	})

	// List files
	files := testFS.ListFiles("/root")
	assert.Len(t, files, 2)
	assert.Contains(t, files, "file1.txt")
	assert.Contains(t, files, "file2.txt")

	// List directories
	dirs := testFS.ListDirs("/root")
	assert.Len(t, dirs, 2)
	assert.Contains(t, dirs, "dir1")
	assert.Contains(t, dirs, "dir2")
}

func TestAferoTempFile(t *testing.T) {
	t.Parallel()

	testFS := testutil.NewMemFS(t)

	// Create temp file
	tempFile := testFS.TempFile("/tmp", "test-*.txt")
	require.NotNil(t, tempFile)

	// Write to temp file
	_, err := tempFile.WriteString("temp content")
	require.NoError(t, err)

	// File should be cleaned up automatically
}

func TestAferoWriteConfig(t *testing.T) {
	t.Parallel()

	testFS := testutil.NewMemFS(t)

	// Write config using helper
	configPath := testFS.WriteConfig(`
gitprovidersync:
  test:
    source:
      provider_type: github
      owner: test-owner
`)

	// Verify config was written
	testFS.AssertFileExists(configPath)
	content := testFS.ReadFileString(configPath)
	assert.Contains(t, content, "provider_type: github")
	assert.Contains(t, content, "owner: test-owner")
}

func BenchmarkMemFS(b *testing.B) {
	testFS := testutil.NewMemFS(b)

	b.ResetTimer()

	for i := 0; i < b.N; i++ { //nolint:intrange // Need index for unique path
		path := "/test-" + string(rune(i)) + ".txt"
		testFS.WriteFileString(path, "benchmark content")
		_ = testFS.ReadFileString(path)
		testFS.Remove(path)
	}
}

func BenchmarkOsFS(b *testing.B) {
	testFS := testutil.NewOsFS(b)

	b.ResetTimer()

	for i := 0; i < b.N; i++ { //nolint:intrange // Need index for unique path
		path := "/test-" + string(rune(i)) + ".txt"
		testFS.WriteFileString(path, "benchmark content")
		_ = testFS.ReadFileString(path)
		testFS.Remove(path)
	}
}
