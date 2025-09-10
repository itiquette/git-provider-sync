// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package testutil_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/testutil"
)

func TestFileSystem_CreateFile(t *testing.T) {
	t.Parallel()

	fileSystem := testutil.NewFileSystem(t)

	// Create a directory
	path := fileSystem.CreateDir("subdir")

	// Verify it exists
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestFileSystem_CreateStructure(t *testing.T) {
	t.Parallel()

	fileSystem := testutil.NewFileSystem(t)

	// Create a complex structure
	fileSystem.CreateStructure(map[string]string{
		"README.md":          "# Test Project",
		"src/":               "",
		"src/main.go":        "package main",
		"src/util/helper.go": "package util",
		"test/":              "",
		"test/main_test.go":  "package main_test",
	})

	// Verify files exist
	assert.True(t, fileSystem.FileExists("README.md"))
	assert.True(t, fileSystem.FileExists("src/main.go"))
	assert.True(t, fileSystem.FileExists("src/util/helper.go"))
	assert.True(t, fileSystem.FileExists("test/main_test.go"))

	// Verify content
	assert.Equal(t, "# Test Project", string(fileSystem.ReadFile("README.md")))
	assert.Equal(t, "package main", string(fileSystem.ReadFile("src/main.go")))
}

func TestFileSystem_CreateTempFile(t *testing.T) {
	t.Parallel()

	fileSystem := testutil.NewFileSystem(t)

	// Create a temp file
	file := fileSystem.CreateTempFile("test-*.txt")
	require.NotNil(t, file)

	// Write some content
	_, err := file.WriteString("temp content")
	require.NoError(t, err)

	// File should be in the filesystem root
	assert.Contains(t, file.Name(), fileSystem.Root())
}

func TestFileSystem_Isolation(t *testing.T) {
	t.Parallel()

	// Create two isolated filesystems
	fileSystem1 := testutil.NewFileSystem(t)
	fileSystem2 := testutil.NewFileSystem(t)

	// They should have different roots
	assert.NotEqual(t, fileSystem1.Root(), fileSystem2.Root())

	// Create files in each
	fileSystem1.CreateFile("test.txt", "fs1")
	fileSystem2.CreateFile("test.txt", "fs2")

	// Verify isolation
	assert.Equal(t, "fs1", string(fileSystem1.ReadFile("test.txt")))
	assert.Equal(t, "fs2", string(fileSystem2.ReadFile("test.txt")))
}

func BenchmarkFileSystem_CreateFile(b *testing.B) {
	for range b.N {
		fileSystem := testutil.NewFileSystem(b)
		fileSystem.CreateFile("test.txt", "benchmark content")
	}
}

func TestFileSystem_CreateFileWithMode(t *testing.T) {
	t.Parallel()

	fileSystem := testutil.NewFileSystem(t)

	// Create a file with specific permissions
	path := fileSystem.CreateFileWithMode("readonly.txt", "content", 0400)

	// Verify permissions
	info, err := os.Stat(path)
	require.NoError(t, err)

	// Check if permissions match (considering umask)
	mode := info.Mode().Perm()
	assert.Equal(t, os.FileMode(0400), mode&0400, "file should be readable by owner")
}

func TestFileSystem_Remove(t *testing.T) {
	t.Parallel()

	fileSystem := testutil.NewFileSystem(t)

	// Create and then remove a file
	fileSystem.CreateFile("temp.txt", "temporary")
	assert.True(t, fileSystem.FileExists("temp.txt"))

	fileSystem.Remove("temp.txt")
	assert.False(t, fileSystem.FileExists("temp.txt"))
}
