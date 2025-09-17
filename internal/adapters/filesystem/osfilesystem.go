// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package filesystem

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// OSFileSystem implements FileSystem port using standard OS operations
// adapter isolates all OS-specific file operations from the domain.
type OSFileSystem struct{}

// NewOSFileSystem creates a new OS-based file system adapter.
func NewOSFileSystem() ports.FileSystem {
	return &OSFileSystem{}
}

// Exists checks if a path exists in the file system.
func (f *OSFileSystem) Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err //nolint:wrapcheck // OS errors are clear enough
}

// MkdirAll creates a directory and all necessary parents.
func (f *OSFileSystem) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm) //nolint:wrapcheck // OS errors are clear enough
}

// RemoveAll removes a path and all its contents.
func (f *OSFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path) //nolint:wrapcheck // OS errors are clear enough
}

// Stat returns file information for the given path.
func (f *OSFileSystem) Stat(path string) (fs.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err //nolint:wrapcheck // OS errors are clear enough
	}

	return info, nil
}

// TempDir creates a new temporary directory.
func (f *OSFileSystem) TempDir(dir, pattern string) (string, error) {
	tmpDir, err := os.MkdirTemp(dir, pattern)
	if err != nil {
		return "", err //nolint:wrapcheck // OS errors are clear enough
	}

	return tmpDir, nil
}

// Join joins path elements.
func (f *OSFileSystem) Join(elem ...string) string {
	return filepath.Join(elem...)
}

// Clean cleans a path.
func (f *OSFileSystem) Clean(path string) string {
	return filepath.Clean(path)
}

// ReadFile reads the contents of a file.
func (f *OSFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path) //nolint:wrapcheck,gosec // OS errors are clear enough, path comes from controlled sources
}

// WriteFile writes data to a file.
func (f *OSFileSystem) WriteFile(path string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(path, data, perm) //nolint:wrapcheck // OS errors are clear enough
}

// Walk walks the file tree rooted at root.
func (f *OSFileSystem) Walk(root string, walkFn func(path string, info fs.FileInfo, err error) error) error {
	return filepath.Walk(root, walkFn) //nolint:wrapcheck // OS errors are clear enough
}

// Chmod changes file permissions.
func (f *OSFileSystem) Chmod(path string, mode fs.FileMode) error {
	return os.Chmod(path, mode) //nolint:wrapcheck // OS errors are clear enough
}

// Open opens a file for reading.
func (f *OSFileSystem) Open(path string) (io.ReadCloser, error) {
	file, err := os.Open(path) //nolint:gosec // Path is validated by caller
	if err != nil {
		return nil, err //nolint:wrapcheck // OS errors are clear enough
	}

	return file, nil
}

// Create creates or truncates a file for writing.
func (f *OSFileSystem) Create(path string) (io.WriteCloser, error) {
	file, err := os.Create(path) //nolint:gosec // Path is validated by caller
	if err != nil {
		return nil, err //nolint:wrapcheck // OS errors are clear enough
	}

	return file, nil
}

// OpenFile opens a file with specific flags and permissions.
func (f *OSFileSystem) OpenFile(path string, flag int, perm fs.FileMode) (io.ReadWriteCloser, error) {
	file, err := os.OpenFile(path, flag, perm) //nolint:gosec // Path is validated by caller
	if err != nil {
		return nil, err //nolint:wrapcheck // OS errors are clear enough
	}

	return file, nil
}

// Remove removes a file or empty directory.
func (f *OSFileSystem) Remove(path string) error {
	return os.Remove(path) //nolint:wrapcheck // OS errors are clear enough
}

// CreateTemp creates a new temporary file in the directory dir.
func (f *OSFileSystem) CreateTemp(dir, pattern string) (string, io.WriteCloser, error) {
	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", nil, err //nolint:wrapcheck // OS errors are clear enough
	}

	return file.Name(), file, nil
}

// Abs returns the absolute path.
func (f *OSFileSystem) Abs(path string) (string, error) {
	return filepath.Abs(path) //nolint:wrapcheck // OS errors are clear enough
}

// ReadDir reads a directory and returns its entries.
func (f *OSFileSystem) ReadDir(path string) ([]fs.DirEntry, error) {
	return os.ReadDir(path) //nolint:wrapcheck // OS errors are clear enough
}

// Base returns the last element of path.
func (f *OSFileSystem) Base(path string) string {
	return filepath.Base(path)
}

// Dir returns all but the last element of path.
func (f *OSFileSystem) Dir(path string) string {
	return filepath.Dir(path)
}

// Rel returns a relative path from basepath to targpath.
func (f *OSFileSystem) Rel(basepath, targpath string) (string, error) {
	return filepath.Rel(basepath, targpath) //nolint:wrapcheck // OS errors are clear enough
}

// IsAbs reports whether the path is absolute.
func (f *OSFileSystem) IsAbs(path string) bool {
	return filepath.IsAbs(path)
}

// GetTempDir returns the default system temporary directory.
func (f *OSFileSystem) GetTempDir() string {
	return os.TempDir()
}
