// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

// Package testutil provides test utilities for isolated testing.
// Following hexagonal architecture, these utilities ensure tests
// don't affect the host system.
package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// FileSystem provides isolated file operations for tests.
// It ensures all file operations happen within a test-specific
// temporary directory that's automatically cleaned up.
type FileSystem struct {
	t    testing.TB
	root string
}

// NewFileSystem creates an isolated filesystem for testing.
// The filesystem root is automatically cleaned up when the test ends.
func NewFileSystem(tb testing.TB) *FileSystem {
	tb.Helper()

	var root string

	switch v := tb.(type) {
	case *testing.T:
		root = v.TempDir()
	case *testing.B:
		root = v.TempDir()
	default:
		tb.Fatal("unsupported testing type")
	}

	return &FileSystem{
		t:    tb,
		root: root,
	}
}

// Root returns the root directory of the isolated filesystem.
func (fs *FileSystem) Root() string {
	return fs.root
}

// Path returns the full path for a relative path in the filesystem.
func (fs *FileSystem) Path(relative string) string {
	return filepath.Join(fs.root, relative)
}

// CreateFile creates a file with content in the isolated filesystem.
// Returns the full path to the created file.
func (fs *FileSystem) CreateFile(name, content string) string {
	fs.t.Helper()

	path := fs.Path(name)
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0750); err != nil {
		fs.t.Fatalf("failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		fs.t.Fatalf("failed to write file %s: %v", path, err)
	}

	return path
}

// CreateFileWithMode creates a file with specific permissions.
func (fs *FileSystem) CreateFileWithMode(name, content string, mode os.FileMode) string {
	fs.t.Helper()

	path := fs.Path(name)
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0750); err != nil {
		fs.t.Fatalf("failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		fs.t.Fatalf("failed to write file %s: %v", path, err)
	}

	return path
}

// CreateDir creates a directory in the isolated filesystem.
// Returns the full path to the created directory.
func (fs *FileSystem) CreateDir(name string) string {
	fs.t.Helper()

	path := fs.Path(name)

	if err := os.MkdirAll(path, 0750); err != nil {
		fs.t.Fatalf("failed to create directory %s: %v", path, err)
	}

	return path
}

// CreateDirWithMode creates a directory with specific permissions.
func (fs *FileSystem) CreateDirWithMode(name string, mode os.FileMode) string {
	fs.t.Helper()

	path := fs.Path(name)

	if err := os.MkdirAll(path, mode); err != nil {
		fs.t.Fatalf("failed to create directory %s: %v", path, err)
	}

	return path
}

// CreateSymlink creates a symbolic link in the isolated filesystem.
func (fs *FileSystem) CreateSymlink(target, link string) string {
	fs.t.Helper()

	linkPath := fs.Path(link)
	targetPath := fs.Path(target)

	dir := filepath.Dir(linkPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		fs.t.Fatalf("failed to create directory %s: %v", dir, err)
	}

	if err := os.Symlink(targetPath, linkPath); err != nil {
		fs.t.Fatalf("failed to create symlink %s -> %s: %v", linkPath, targetPath, err)
	}

	return linkPath
}

// ReadFile reads a file from the isolated filesystem.
func (fs *FileSystem) ReadFile(name string) []byte {
	fs.t.Helper()

	path := fs.Path(name)

	content, err := os.ReadFile(path) //nolint:gosec // Test file with controlled path
	if err != nil {
		fs.t.Fatalf("failed to read file %s: %v", path, err)
	}

	return content
}

// FileExists checks if a file exists in the isolated filesystem.
func (fs *FileSystem) FileExists(name string) bool {
	fs.t.Helper()

	path := fs.Path(name)
	_, err := os.Stat(path)

	return err == nil
}

// Remove removes a file or directory from the isolated filesystem.
func (fs *FileSystem) Remove(name string) {
	fs.t.Helper()

	path := fs.Path(name)
	if err := os.RemoveAll(path); err != nil {
		fs.t.Fatalf("failed to remove %s: %v", path, err)
	}
}

// CreateStructure creates a directory structure from a map.
// Keys are paths, values are file contents (empty string for directories).
func (fs *FileSystem) CreateStructure(structure map[string]string) {
	fs.t.Helper()

	for path, content := range structure {
		if content == "" {
			// It's a directory
			fs.CreateDir(path)
		} else {
			// It's a file
			fs.CreateFile(path, content)
		}
	}
}

// CreateTempFile creates a temporary file in the isolated filesystem.
// The pattern follows the same rules as os.CreateTemp.
func (fs *FileSystem) CreateTempFile(pattern string) *os.File {
	fs.t.Helper()

	file, err := os.CreateTemp(fs.root, pattern)
	if err != nil {
		fs.t.Fatalf("failed to create temp file: %v", err)
	}

	// Register cleanup
	fs.t.Cleanup(func() {
		_ = file.Close()
	})

	return file
}
