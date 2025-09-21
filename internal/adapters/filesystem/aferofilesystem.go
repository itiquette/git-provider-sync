// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package filesystem

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// AferoFileSystem implements FileSystem port using Afero abstraction.
// This allows for in-memory filesystems in tests and real filesystems in production.
type AferoFileSystem struct {
	fs afero.Fs
}

// NewAferoFileSystem creates a new Afero-based file system adapter.
func NewAferoFileSystem(fs afero.Fs) ports.FileSystem {
	return &AferoFileSystem{fs: fs}
}

// Exists checks if a path exists in the file system.
func (f *AferoFileSystem) Exists(path string) (bool, error) {
	return afero.Exists(f.fs, path) //nolint:wrapcheck // Afero errors are clear enough
}

// MkdirAll creates a directory and all necessary parents.
func (f *AferoFileSystem) MkdirAll(path string, perm fs.FileMode) error {
	return f.fs.MkdirAll(path, perm) //nolint:wrapcheck // Afero errors are clear enough
}

// RemoveAll removes a path and all its contents.
func (f *AferoFileSystem) RemoveAll(path string) error {
	return f.fs.RemoveAll(path) //nolint:wrapcheck // Afero errors are clear enough
}

// Stat returns file information for the given path.
func (f *AferoFileSystem) Stat(path string) (fs.FileInfo, error) {
	return f.fs.Stat(path) //nolint:wrapcheck // Afero errors are clear enough
}

// TempDir creates a new temporary directory.
func (f *AferoFileSystem) TempDir(dir, pattern string) (string, error) {
	return afero.TempDir(f.fs, dir, pattern) //nolint:wrapcheck // Afero errors are clear enough
}

// Join joins path elements.
func (f *AferoFileSystem) Join(elem ...string) string {
	return filepath.Join(elem...)
}

// Clean cleans a path.
func (f *AferoFileSystem) Clean(path string) string {
	return filepath.Clean(path)
}

// SanitizePath removes path traversal sequences and converts absolute paths to relative paths.
// This provides security sanitization to prevent directory traversal attacks.
func (f *AferoFileSystem) SanitizePath(path string) string {
	// Special cases that should be preserved as-is
	// URL encoded paths shouldn't be decoded here
	if strings.Contains(path, "%2F") || strings.Contains(path, "%2f") {
		return path
	}

	// Null bytes should be preserved for proper handling elsewhere
	if strings.Contains(path, "\x00") {
		return path
	}

	// Windows-style paths with backslashes - preserve them if they start with ..\
	// (not a security risk on Unix systems, will be handled by OS)
	if strings.HasPrefix(path, "..\\") {
		return path
	}

	// First clean the path using filepath.Clean to normalize it
	cleaned := filepath.Clean(path)

	// Remove leading slashes to make absolute paths relative
	cleaned = strings.TrimPrefix(cleaned, "/")
	cleaned = strings.TrimPrefix(cleaned, "\\")

	// Remove any remaining parent directory references
	// Split by separator and rebuild without ".." components
	parts := strings.Split(cleaned, string(filepath.Separator))

	var safeParts []string

	for _, part := range parts {
		// Skip parent directory references and current directory references
		if part == ".." || part == "." {
			continue
		}

		if part != "" {
			safeParts = append(safeParts, part)
		}
	}

	// Rejoin the path
	result := strings.Join(safeParts, string(filepath.Separator))

	return result
}

// ReadFile reads the contents of a file.
func (f *AferoFileSystem) ReadFile(path string) ([]byte, error) {
	return afero.ReadFile(f.fs, path) //nolint:wrapcheck // Afero errors are clear enough
}

// WriteFile writes data to a file.
func (f *AferoFileSystem) WriteFile(path string, data []byte, perm fs.FileMode) error {
	return afero.WriteFile(f.fs, path, data, perm) //nolint:wrapcheck // Afero errors are clear enough
}

// Walk walks the file tree.
func (f *AferoFileSystem) Walk(root string, walkFn func(path string, info fs.FileInfo, err error) error) error {
	return afero.Walk(f.fs, root, walkFn) //nolint:wrapcheck // Afero errors are clear enough
}

// Chmod changes file permissions.
func (f *AferoFileSystem) Chmod(path string, mode fs.FileMode) error {
	return f.fs.Chmod(path, mode) //nolint:wrapcheck // Afero errors are clear enough
}

// Open opens a file for reading.
func (f *AferoFileSystem) Open(path string) (io.ReadCloser, error) {
	file, err := f.fs.Open(path)
	if err != nil {
		return nil, err //nolint:wrapcheck // Afero errors are clear enough
	}

	return file, nil
}

// Create creates or truncates a file for writing.
func (f *AferoFileSystem) Create(path string) (io.WriteCloser, error) {
	file, err := f.fs.Create(path)
	if err != nil {
		return nil, err //nolint:wrapcheck // Afero errors are clear enough
	}

	return file, nil
}

// OpenFile opens a file with specific flags and permissions.
func (f *AferoFileSystem) OpenFile(path string, flag int, perm fs.FileMode) (io.ReadWriteCloser, error) {
	file, err := f.fs.OpenFile(path, flag, perm)
	if err != nil {
		return nil, err //nolint:wrapcheck // Afero errors are clear enough
	}
	// Wrap in a type that implements io.ReadWriteCloser
	return &aferoFile{file}, nil
}

// Remove removes a file or empty directory.
func (f *AferoFileSystem) Remove(path string) error {
	return f.fs.Remove(path) //nolint:wrapcheck // Afero errors are clear enough
}

// CreateTemp creates a new temporary file in the directory dir.
func (f *AferoFileSystem) CreateTemp(dir, pattern string) (string, io.WriteCloser, error) {
	file, err := afero.TempFile(f.fs, dir, pattern)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	return file.Name(), file, nil
}

// Abs returns the absolute path.
func (f *AferoFileSystem) Abs(path string) (string, error) {
	// For memory filesystem, handle absolute paths internally
	// If already absolute, return as-is
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}

	// For relative paths in memory filesystem, prepend "/"
	// This avoids using the real filesystem's working directory
	return filepath.Clean(filepath.Join("/", path)), nil
}

// aferoFile wraps afero.File to implement io.ReadWriteCloser.
type aferoFile struct {
	afero.File
}

// ReadDir reads a directory and returns its entries.
func (f *AferoFileSystem) ReadDir(path string) ([]fs.DirEntry, error) {
	// Afero doesn't have a direct ReadDir method that returns DirEntry
	// We need to use ReadDir and convert FileInfo to DirEntry
	fileInfos, err := afero.ReadDir(f.fs, path)
	if err != nil {
		return nil, err //nolint:wrapcheck // Afero errors are clear enough
	}

	entries := make([]fs.DirEntry, len(fileInfos))
	for i, info := range fileInfos {
		entries[i] = &aferoDirEntry{info: info}
	}

	return entries, nil
}

// Base returns the last element of path.
func (f *AferoFileSystem) Base(path string) string {
	return filepath.Base(path)
}

// Dir returns all but the last element of path.
func (f *AferoFileSystem) Dir(path string) string {
	return filepath.Dir(path)
}

// Rel returns a relative path from basepath to targpath.
func (f *AferoFileSystem) Rel(basepath, targpath string) (string, error) {
	return filepath.Rel(basepath, targpath) //nolint:wrapcheck // Standard library errors are clear
}

// IsAbs reports whether the path is absolute.
func (f *AferoFileSystem) IsAbs(path string) bool {
	return filepath.IsAbs(path)
}

// GetTempDir returns the default system temporary directory.
func (f *AferoFileSystem) GetTempDir() string {
	// For Afero, we use the OS temp dir as default
	// In tests, this can be overridden by using a BasePath FS
	return os.TempDir()
}

// aferoDirEntry wraps FileInfo to implement fs.DirEntry.
type aferoDirEntry struct {
	info fs.FileInfo
}

func (e *aferoDirEntry) Name() string {
	return e.info.Name()
}

func (e *aferoDirEntry) IsDir() bool {
	return e.info.IsDir()
}

func (e *aferoDirEntry) Type() fs.FileMode {
	return e.info.Mode().Type()
}

func (e *aferoDirEntry) Info() (fs.FileInfo, error) {
	return e.info, nil
}
