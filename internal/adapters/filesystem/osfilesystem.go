// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package filesystem

import (
	"io/fs"
	"os"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// OSFileSystem implements FileSystem port using standard OS operations.
// This adapter isolates all OS-specific file operations from the domain.
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
