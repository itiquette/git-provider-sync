// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package ports

import (
	"io"
	"io/fs"
)

// FileSystem provides file system operations needed by the domain
// Keeps the domain pure and testable without OS dependencies.
// Many methods provide complete abstraction over filesystem operations,
// enabling both OS and in-memory implementations for testing.
//
//nolint:interfacebloat // Comprehensive filesystem abstraction requires many methods
type FileSystem interface {
	// Exists checks if a path exists in the file system
	Exists(path string) (bool, error)

	// MkdirAll creates a directory and all necessary parents
	MkdirAll(path string, perm fs.FileMode) error

	// RemoveAll removes a path and all its contents
	RemoveAll(path string) error

	// Stat returns file information for the given path
	Stat(path string) (fs.FileInfo, error)

	// TempDir creates a new temporary directory
	TempDir(dir, pattern string) (string, error)

	// Join joins path elements
	Join(elem ...string) string

	// Clean cleans a path
	Clean(path string) string

	// SanitizePath removes path traversal sequences and converts absolute paths to relative paths
	// Security sanitization to prevent directory traversal attacks
	SanitizePath(path string) string

	// ReadFile reads the contents of a file
	ReadFile(path string) ([]byte, error)

	// WriteFile writes data to a file
	WriteFile(path string, data []byte, perm fs.FileMode) error

	// Walk walks the file tree rooted at root
	Walk(root string, walkFn func(path string, info fs.FileInfo, err error) error) error

	// Chmod changes file permissions
	Chmod(path string, mode fs.FileMode) error

	// Open opens a file for reading
	Open(path string) (io.ReadCloser, error)

	// Create creates or truncates a file for writing
	Create(path string) (io.WriteCloser, error)

	// OpenFile opens a file with specific flags and permissions
	OpenFile(path string, flag int, perm fs.FileMode) (io.ReadWriteCloser, error)

	// Remove removes a file or empty directory
	Remove(path string) error

	// CreateTemp creates a new temporary file in the directory dir
	CreateTemp(dir, pattern string) (string, io.WriteCloser, error)

	// Abs returns the absolute path
	Abs(path string) (string, error)

	// ReadDir reads a directory and returns its entries
	ReadDir(path string) ([]fs.DirEntry, error)

	// Base returns the last element of path
	Base(path string) string

	// Dir returns all but the last element of path
	Dir(path string) string

	// Rel returns a relative path from basepath to targpath
	Rel(basepath, targpath string) (string, error)

	// IsAbs reports whether the path is absolute
	IsAbs(path string) bool

	// GetTempDir returns the default system temporary directory
	GetTempDir() string
}
