// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package directory

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// validateDirectoryPath validates that a path is safe for directory operations.
// Returns an error if the path points to a critical system directory.
func validateDirectoryPath(path string) error {
	// Normalize the path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Clean the path to remove any . or .. components
	cleanPath := filepath.Clean(absPath)

	// Never allow root directory
	if cleanPath == "/" {
		return errors.New("refusing to operate on root directory")
	}

	// Critical system directories that should never be deleted
	criticalPaths := []string{
		"/bin",
		"/boot",
		"/dev",
		"/etc",
		"/lib",
		"/lib64",
		"/proc",
		"/root",
		"/sbin",
		"/sys",
		"/usr",
		"/var",
	}

	for _, critical := range criticalPaths {
		if path == critical || strings.HasPrefix(path, critical+"/") {
			return fmt.Errorf("refusing to operate on system directory: %s", path)
		}
	}

	// Warn but allow home directories - user might legitimately backup here
	// But not the home root itself
	if path == "/home" {
		return errors.New("refusing to operate on /home root directory")
	}

	// Check for common important directories
	dangerousPaths := []string{
		"/Applications", // macOS
		"/System",       // macOS
		"/Library",      // macOS
		"/opt",          // Common installation directory
		"/mnt",          // Mount points
		"/media",        // Mount points
		"/srv",          // Server data
	}

	for _, dangerous := range dangerousPaths {
		if path == dangerous {
			return fmt.Errorf("refusing to operate on important directory: %s", path)
		}
	}

	return nil
}
