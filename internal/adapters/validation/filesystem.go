// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package validation

import (
	"fmt"
	"os"
	"path/filepath"
)

// statPath gets file info for a path.
func statPath(path string) (os.FileInfo, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", path, err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path %s: %w", absPath, err)
	}

	return info, nil
}

// isReadable checks if a path is readable.
func isReadable(path string) bool {
	// #nosec G304 - Path is validated before this call
	file, err := os.Open(path)
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

// isWritable checks if a path is writable.
func isWritable(path string) bool {
	// For directories, try to create a temp file
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	if info.IsDir() {
		// Try to create a temporary file in the directory
		tempFile, err := os.CreateTemp(path, "write_test_")
		if err != nil {
			return false
		}

		defer func() {
			if err := os.Remove(tempFile.Name()); err != nil {
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
	// #nosec G304 - Path is validated before this call
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
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
