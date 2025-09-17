// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package testutil

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"testing"

	"github.com/spf13/afero"

	"itiquette/git-provider-sync/internal/adapters/filesystem"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// ErrorFileSystem is a filesystem that can simulate errors for testing.
type ErrorFileSystem struct {
	ports.FileSystem
	failOn map[string]error
}

// NewErrorFileSystem creates a new filesystem that can simulate errors.
func NewErrorFileSystem(t *testing.T) *ErrorFileSystem {
	t.Helper()

	memFS := afero.NewMemMapFs()

	return &ErrorFileSystem{
		FileSystem: filesystem.NewAferoFileSystem(memFS),
		failOn:     make(map[string]error),
	}
}

// SetError configures the filesystem to return an error for a specific path and operation.
func (fs *ErrorFileSystem) SetError(path string, err error) {
	fs.failOn[path] = err
}

// MkdirAll creates a directory hierarchy, or returns configured error.
func (fs *ErrorFileSystem) MkdirAll(path string, perm os.FileMode) error {
	if err, ok := fs.failOn[path]; ok {
		return err
	}

	return fs.FileSystem.MkdirAll(path, perm) //nolint:wrapcheck // Mock filesystem for testing
}

// Create creates a file, or returns configured error.
func (fs *ErrorFileSystem) Create(name string) (io.WriteCloser, error) {
	if err, ok := fs.failOn[name]; ok {
		return nil, err
	}

	return fs.FileSystem.Create(name) //nolint:wrapcheck // Mock filesystem for testing
}

// Open opens a file, or returns configured error.
func (fs *ErrorFileSystem) Open(name string) (io.ReadCloser, error) {
	if err, ok := fs.failOn[name]; ok {
		return nil, err
	}

	return fs.FileSystem.Open(name) //nolint:wrapcheck // Mock filesystem for testing
}

// OpenFile opens a file with flags, or returns configured error.
func (fs *ErrorFileSystem) OpenFile(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
	if err, ok := fs.failOn[name]; ok {
		return nil, err
	}

	return fs.FileSystem.OpenFile(name, flag, perm) //nolint:wrapcheck // Mock filesystem for testing
}

// Stat returns file info, or returns configured error.
func (fs *ErrorFileSystem) Stat(name string) (fs.FileInfo, error) {
	if err, ok := fs.failOn[name]; ok {
		return nil, err
	}

	return fs.FileSystem.Stat(name) //nolint:wrapcheck // Mock filesystem for testing
}

// WriteFile writes a file, or returns configured error.
func (fs *ErrorFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	if err, ok := fs.failOn[name]; ok {
		return err
	}

	return fs.FileSystem.WriteFile(name, data, perm) //nolint:wrapcheck // Mock filesystem for testing
}

// ReadFile reads a file, or returns configured error.
func (fs *ErrorFileSystem) ReadFile(name string) ([]byte, error) {
	if err, ok := fs.failOn[name]; ok {
		return nil, err
	}

	return fs.FileSystem.ReadFile(name) //nolint:wrapcheck // Mock filesystem for testing
}

// NewTestFileSystem creates a new in-memory filesystem for testing.
func NewTestFileSystem(t *testing.T) ports.FileSystem {
	t.Helper()

	memFS := afero.NewMemMapFs()

	return filesystem.NewAferoFileSystem(memFS)
}

// Common test errors.
var (
	ErrPermissionDenied = errors.New("permission denied")
	ErrDiskFull         = errors.New("no space left on device")
	ErrFileExists       = errors.New("file already exists")
	ErrNotDirectory     = errors.New("not a directory")
)
