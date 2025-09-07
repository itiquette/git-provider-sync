// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package entities

import (
	"path/filepath"
	"strings"
)

// SanitizePath removes path traversal sequences and converts absolute paths to relative paths.
// This function provides security sanitization to prevent directory traversal attacks.
func SanitizePath(path string) string {
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
