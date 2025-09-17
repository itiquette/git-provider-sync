// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package validation

import (
	"fmt"
	"io/fs"
	"os"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// StatPath gets file info for a path using the provided filesystem.
func statPath(fileSystem ports.FileSystem, path string) (fs.FileInfo, error) {
	absPath, err := fileSystem.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", path, err)
	}

	info, err := fileSystem.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path %s: %w", absPath, err)
	}

	return info, nil
}

// IsReadable checks if a path is readable using the provided filesystem.
func isReadable(fileSystem ports.FileSystem, path string) bool {
	// Path is validated before this call
	file, err := fileSystem.Open(path)
	if err != nil {
		return false
	}

	defer func() {
		if err := file.Close(); err != nil {
			// Log close error
			_ = err
		}
	}()

	return true
}

// IsWritable checks if a path is writable using the provided filesystem.
func isWritable(fileSystem ports.FileSystem, path string) bool {
	// For directories, try to create a temp file
	info, err := fileSystem.Stat(path)
	if err != nil {
		return false
	}

	if info.IsDir() {
		// Try to create a temporary file in the directory
		tempFileName, tempFile, err := fileSystem.CreateTemp(path, "write_test_")
		if err != nil {
			return false
		}

		defer func() {
			if err := fileSystem.Remove(tempFileName); err != nil {
				// Log remove error
				_ = err
			}
		}()
		defer func() {
			if err := tempFile.Close(); err != nil {
				// Log close error
				_ = err
			}
		}()

		return true
	}

	// For files, check if we can open for writing
	// Path is validated before this call
	file, err := fileSystem.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return false
	}

	defer func() {
		if err := file.Close(); err != nil {
			// Log close error
			_ = err
		}
	}()

	return true
}
