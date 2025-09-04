// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package ports

import (
	"io/fs"
)

// FileSystem provides file system operations needed by the domain
// Keeps the domain pure and testable without OS dependencies.
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
}
