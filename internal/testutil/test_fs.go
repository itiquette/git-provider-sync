// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package testutil

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/adapters/filesystem"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// TestFS provides all filesystem operations for tests.
// It uses Afero MemMapFs internally for speed and isolation,
// but tests don't need to know about Afero at all.
type TestFS struct {
	t  testing.TB
	fs afero.Fs
}

// NewTestFS creates a new test filesystem.
// By default uses in-memory filesystem for speed.
//
//nolint:thelper // t is the established parameter name in this widely-used test utility
func NewTestFS(t testing.TB) *TestFS {
	t.Helper()

	return &TestFS{
		t:  t,
		fs: afero.NewMemMapFs(),
	}
}

// WriteFile writes a file with the given content.
func (tfs *TestFS) WriteFile(path string, content string) string {
	tfs.t.Helper()

	// Ensure directory exists
	dir := filepath.Dir(path)
	require.NoError(tfs.t, tfs.fs.MkdirAll(dir, 0755))

	// Write file
	err := afero.WriteFile(tfs.fs, path, []byte(content), 0644)
	require.NoError(tfs.t, err, "Failed to write file: %s", path)

	return path
}

// ReadFile reads a file and returns its content.
func (tfs *TestFS) ReadFile(path string) string {
	tfs.t.Helper()

	content, err := afero.ReadFile(tfs.fs, path)
	require.NoError(tfs.t, err, "Failed to read file: %s", path)

	return string(content)
}

// CreateDir creates a directory.
func (tfs *TestFS) CreateDir(path string) string {
	tfs.t.Helper()

	require.NoError(tfs.t, tfs.fs.MkdirAll(path, 0755))

	return path
}

// Exists checks if a path exists.
func (tfs *TestFS) Exists(path string) bool {
	exists, _ := afero.Exists(tfs.fs, path)

	return exists
}

// Remove removes a file or directory.
func (tfs *TestFS) Remove(path string) {
	tfs.t.Helper()
	require.NoError(tfs.t, tfs.fs.RemoveAll(path))
}

// WriteConfig writes a YAML configuration file.
func (tfs *TestFS) WriteConfig(content string) string {
	return tfs.WriteFile("/gitprovidersync.yaml", content)
}

// CreateGitRepo creates a basic git repository structure.
func (tfs *TestFS) CreateGitRepo(name string) string {
	tfs.t.Helper()

	repoPath := filepath.Join("/repos", name)
	gitDir := filepath.Join(repoPath, ".git")

	// Create directories
	tfs.CreateDir(gitDir)
	tfs.CreateDir(filepath.Join(gitDir, "objects"))
	tfs.CreateDir(filepath.Join(gitDir, "refs/heads"))

	// Create basic git config
	config := `[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
`
	tfs.WriteFile(filepath.Join(gitDir, "config"), config)
	tfs.WriteFile(filepath.Join(gitDir, "HEAD"), "ref: refs/heads/main\n")
	tfs.WriteFile(filepath.Join(repoPath, "README.md"), "# "+name)

	return repoPath
}

// CreateStructure creates multiple files/directories from a map.
func (tfs *TestFS) CreateStructure(structure map[string]string) {
	tfs.t.Helper()

	for path, content := range structure {
		if content == "" {
			tfs.CreateDir(path)
		} else {
			tfs.WriteFile(path, content)
		}
	}
}

// CopyFile copies a file within the filesystem.
func (tfs *TestFS) CopyFile(src, dst string) {
	tfs.t.Helper()

	content := tfs.ReadFile(src)
	tfs.WriteFile(dst, content)
}

// ListFiles returns all files in a directory.
func (tfs *TestFS) ListFiles(dir string) []string {
	tfs.t.Helper()

	files, err := afero.ReadDir(tfs.fs, dir)
	require.NoError(tfs.t, err)

	var names []string

	for _, f := range files {
		if !f.IsDir() {
			names = append(names, f.Name())
		}
	}

	return names
}

// AssertFileExists verifies a file exists.
func (tfs *TestFS) AssertFileExists(path string) {
	tfs.t.Helper()
	require.True(tfs.t, tfs.Exists(path), "File should exist: %s", path)
}

// AssertFileContent verifies file content.
func (tfs *TestFS) AssertFileContent(path string, expected string) {
	tfs.t.Helper()
	actual := tfs.ReadFile(path)
	require.Equal(tfs.t, expected, actual, "File content mismatch: %s", path)
}

// AssertDirExists verifies a directory exists.
func (tfs *TestFS) AssertDirExists(path string) {
	tfs.t.Helper()

	info, err := tfs.fs.Stat(path)
	require.NoError(tfs.t, err, "Directory should exist: %s", path)
	require.True(tfs.t, info.IsDir(), "Path should be a directory: %s", path)
}

// TempDir creates a temporary directory.
func (tfs *TestFS) TempDir(prefix string) string {
	tfs.t.Helper()

	dir, err := afero.TempDir(tfs.fs, "/tmp", prefix)
	require.NoError(tfs.t, err)

	return dir
}

// GetFileSystem returns a ports.FileSystem implementation for dependency injection.
// This allows tests to pass a memory filesystem to production code.
func (tfs *TestFS) GetFileSystem() ports.FileSystem {
	return filesystem.NewAferoFileSystem(tfs.fs)
}
