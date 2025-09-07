// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package shared

import (
	"fmt"
	"os"
	"strings"
)

// RemoveAllInTempDir safely removes a path only if it's within a temporary directory.
// Returns an error if the path is outside temp directories.
func RemoveAllInTempDir(path string) error {
	// Only allow removal in temp directories
	if !strings.HasPrefix(path, "/tmp/") &&
		!strings.HasPrefix(path, os.TempDir()) {
		return fmt.Errorf("refusing to remove %s: not in temp directory", path)
	}

	return os.RemoveAll(path) //nolint:wrapcheck // OS errors are clear enough
}
