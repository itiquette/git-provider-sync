// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

// Package testutil provides test utilities and helpers for the git-provider-sync project.
package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// AferoTestFS provides an Afero-based filesystem for testing.
// By default, it uses MemMapFs for ultra-fast in-memory operations.
// This is the recommended approach for most tests.
type AferoTestFS struct {
	t    testing.TB
	Fs   afero.Fs
	root string
}

// NewMemFS creates a new in-memory filesystem for testing.
// This is the fastest option and recommended for most unit tests.
// The filesystem exists only in memory and doesn't touch the disk.
func NewMemFS(tb testing.TB) *AferoTestFS {
	tb.Helper()

	return &AferoTestFS{
		t:    tb,
		Fs:   afero.NewMemMapFs(),
		root: "/",
	}
}

// NewMemFSWithRoot creates a new in-memory filesystem with a specific root path.
// All operations will be relative to this root.
func NewMemFSWithRoot(tb testing.TB, root string) *AferoTestFS {
	tb.Helper()

	memFs := afero.NewMemMapFs()
	// Create the root directory
	_ = memFs.MkdirAll(root, 0755)

	return &AferoTestFS{
		t:    tb,
		Fs:   afero.NewBasePathFs(memFs, root),
		root: root,
	}
}

// NewOsFS creates a filesystem backed by a real temp directory.
// Use this only when you need actual OS filesystem behavior.
func NewOsFS(tb testing.TB) *AferoTestFS {
	tb.Helper()

	var tempDir string

	switch v := tb.(type) {
	case *testing.T:
		tempDir = v.TempDir()
	case *testing.B:
		tempDir = v.TempDir()
	default:
		tb.Fatal("unsupported testing type")
	}

	return &AferoTestFS{
		t:    tb,
		Fs:   afero.NewBasePathFs(afero.NewOsFs(), tempDir),
		root: tempDir,
	}
}

// NewCopyOnWriteFS creates a layered filesystem for testing.
// Reads go through the base layer, writes go to the overlay.
// Useful for testing modifications to existing file structures.
func NewCopyOnWriteFS(tb testing.TB, baseFs afero.Fs) *AferoTestFS {
	tb.Helper()

	overlay := afero.NewMemMapFs()
	copyOnWrite := afero.NewCopyOnWriteFs(baseFs, overlay)

	return &AferoTestFS{
		t:    tb,
		Fs:   copyOnWrite,
		root: "/",
	}
}

// Root returns the root path of the filesystem.
func (fs *AferoTestFS) Root() string {
	return fs.root
}

// Path returns the full path for a relative path in the filesystem.
func (fs *AferoTestFS) Path(relative string) string {
	if fs.root == "/" {
		return filepath.Join("/", relative)
	}

	return filepath.Join(fs.root, relative)
}

// WriteFile writes a file with the given content.
func (fs *AferoTestFS) WriteFile(path string, content []byte, perm os.FileMode) {
	fs.t.Helper()

	// Ensure directory exists
	dir := filepath.Dir(path)
	require.NoError(fs.t, fs.Fs.MkdirAll(dir, 0755))

	err := afero.WriteFile(fs.Fs, path, content, perm)
	require.NoError(fs.t, err, "Failed to write file: %s", path)
}

// WriteFileString writes a string as a file.
func (fs *AferoTestFS) WriteFileString(path string, content string) {
	fs.WriteFile(path, []byte(content), 0644)
}

// ReadFile reads a file and returns its content.
func (fs *AferoTestFS) ReadFile(path string) []byte {
	fs.t.Helper()

	content, err := afero.ReadFile(fs.Fs, path)
	require.NoError(fs.t, err, "Failed to read file: %s", path)

	return content
}

// ReadFileString reads a file as a string.
func (fs *AferoTestFS) ReadFileString(path string) string {
	return string(fs.ReadFile(path))
}

// CreateDir creates a directory.
func (fs *AferoTestFS) CreateDir(path string) {
	fs.t.Helper()
	require.NoError(fs.t, fs.Fs.MkdirAll(path, 0755))
}

// CreateFile creates a file with the given content.
func (fs *AferoTestFS) CreateFile(name, content string) string {
	fs.t.Helper()

	path := fs.Path(name)
	fs.WriteFileString(path, content)

	return path
}

// CreateStructure creates a directory structure from a map.
// Keys are paths, values are file contents (empty string for directories).
func (fs *AferoTestFS) CreateStructure(structure map[string]string) {
	fs.t.Helper()

	for path, content := range structure {
		// Don't use fs.Path here as paths may already be absolute
		if content == "" {
			// It's a directory
			fs.CreateDir(path)
		} else {
			// It's a file
			fs.WriteFileString(path, content)
		}
	}
}

// Exists checks if a path exists.
func (fs *AferoTestFS) Exists(path string) bool {
	exists, _ := afero.Exists(fs.Fs, path)

	return exists
}

// IsDir checks if a path is a directory.
func (fs *AferoTestFS) IsDir(path string) bool {
	isDir, _ := afero.IsDir(fs.Fs, path)

	return isDir
}

// Remove removes a file or empty directory.
func (fs *AferoTestFS) Remove(path string) {
	fs.t.Helper()
	require.NoError(fs.t, fs.Fs.Remove(path))
}

// RemoveAll removes a path and all its contents.
func (fs *AferoTestFS) RemoveAll(path string) {
	fs.t.Helper()
	require.NoError(fs.t, fs.Fs.RemoveAll(path))
}

// AssertFileExists verifies a file exists.
func (fs *AferoTestFS) AssertFileExists(path string) {
	fs.t.Helper()
	require.True(fs.t, fs.Exists(path), "File should exist: %s", path)
}

// AssertFileContent verifies file content matches expected.
func (fs *AferoTestFS) AssertFileContent(path string, expected string) {
	fs.t.Helper()
	actual := fs.ReadFileString(path)
	require.Equal(fs.t, expected, actual, "File content mismatch: %s", path)
}

// AssertDirExists verifies a directory exists.
func (fs *AferoTestFS) AssertDirExists(path string) {
	fs.t.Helper()
	require.True(fs.t, fs.IsDir(path), "Directory should exist: %s", path)
}

// AssertNotExists verifies a path does not exist.
func (fs *AferoTestFS) AssertNotExists(path string) {
	fs.t.Helper()
	require.False(fs.t, fs.Exists(path), "Path should not exist: %s", path)
}

// CopyFile copies a file within the filesystem.
func (fs *AferoTestFS) CopyFile(src, dst string) {
	fs.t.Helper()

	content := fs.ReadFile(src)

	// Get source file info for permissions
	info, err := fs.Fs.Stat(src)
	require.NoError(fs.t, err)

	fs.WriteFile(dst, content, info.Mode())
}

// TempFile creates a temporary file.
func (fs *AferoTestFS) TempFile(dir, pattern string) afero.File {
	fs.t.Helper()

	file, err := afero.TempFile(fs.Fs, dir, pattern)
	require.NoError(fs.t, err)

	// Register cleanup
	fs.t.Cleanup(func() {
		_ = file.Close()
	})

	return file
}

// TempDir creates a temporary directory.
func (fs *AferoTestFS) TempDir(dir, pattern string) string {
	fs.t.Helper()

	tempDir, err := afero.TempDir(fs.Fs, dir, pattern)
	require.NoError(fs.t, err)

	return tempDir
}

// Walk walks the file tree rooted at root.
func (fs *AferoTestFS) Walk(root string, walkFn filepath.WalkFunc) {
	fs.t.Helper()
	require.NoError(fs.t, afero.Walk(fs.Fs, root, walkFn))
}

// ListFiles returns all files in a directory (non-recursive).
func (fs *AferoTestFS) ListFiles(dir string) []string {
	fs.t.Helper()

	files, err := afero.ReadDir(fs.Fs, dir)
	require.NoError(fs.t, err)

	var names []string

	for _, f := range files {
		if !f.IsDir() {
			names = append(names, f.Name())
		}
	}

	return names
}

// ListDirs returns all directories in a directory (non-recursive).
func (fs *AferoTestFS) ListDirs(dir string) []string {
	fs.t.Helper()

	files, err := afero.ReadDir(fs.Fs, dir)
	require.NoError(fs.t, err)

	var names []string

	for _, f := range files {
		if f.IsDir() {
			names = append(names, f.Name())
		}
	}

	return names
}

// CopyFromOS copies files from the OS filesystem into the test filesystem.
// Useful for loading test fixtures from disk.
func (fs *AferoTestFS) CopyFromOS(osPath, destPath string) {
	fs.t.Helper()

	// Check if it's a directory
	info, err := os.Stat(osPath)
	require.NoError(fs.t, err)

	if !info.IsDir() {
		// Copy single file
		content, err := os.ReadFile(osPath) //nolint:gosec // Test utility reading controlled paths
		require.NoError(fs.t, err)
		fs.WriteFile(destPath, content, info.Mode())

		return
	}

	// Copy directory recursively
	err = filepath.Walk(osPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(osPath, path)
		if err != nil {
			return err //nolint:wrapcheck // Test utility error pass-through
		}

		fullDestPath := filepath.Join(destPath, relPath)

		if info.IsDir() {
			return fs.Fs.MkdirAll(fullDestPath, info.Mode())
		}

		// Copy file
		content, err := os.ReadFile(path) //nolint:gosec // Test utility reading controlled paths
		if err != nil {
			return err //nolint:wrapcheck // Test utility error pass-through
		}

		return afero.WriteFile(fs.Fs, fullDestPath, content, info.Mode())
	})
	require.NoError(fs.t, err)
}

// WriteConfig writes a test configuration file.
// This is a convenience method for the common pattern of writing YAML configs.
func (fs *AferoTestFS) WriteConfig(content string) string {
	fs.t.Helper()

	configPath := fs.Path("gitprovidersync.yaml")
	fs.WriteFileString(configPath, content)

	return configPath
}

// WriteYAML writes a YAML file with proper formatting.
func (fs *AferoTestFS) WriteYAML(path string, content string) string {
	fs.t.Helper()

	fullPath := fs.Path(path)
	fs.WriteFileString(fullPath, content)

	return fullPath
}

// CreateGitRepo creates a basic git repository structure.
// This creates the minimal structure needed for git operations.
func (fs *AferoTestFS) CreateGitRepo(name string) string {
	fs.t.Helper()

	repoPath := fs.Path(name)
	gitDir := filepath.Join(repoPath, ".git")

	// Create directories
	fs.CreateDir(gitDir)
	fs.CreateDir(filepath.Join(gitDir, "objects"))
	fs.CreateDir(filepath.Join(gitDir, "refs"))
	fs.CreateDir(filepath.Join(gitDir, "refs", "heads"))

	// Create basic git config
	config := `[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
[branch "main"]
	remote = origin
	merge = refs/heads/main
`
	fs.WriteFileString(filepath.Join(gitDir, "config"), config)

	// Create HEAD file
	fs.WriteFileString(filepath.Join(gitDir, "HEAD"), "ref: refs/heads/main\n")

	// Create a README
	fs.WriteFileString(filepath.Join(repoPath, "README.md"), "# Test Repository\n")

	return repoPath
}
