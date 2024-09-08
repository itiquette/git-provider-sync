// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"itiquette/git-provider-sync/internal/log"
)

// TmpDirKey is the key type for storing and retrieving the temporary directory path in the context.
// Using a custom type for the key helps prevent collisions with other context values.
type TmpDirKey struct{}

// GetTmpDirPath retrieves the temporary directory path from the context.
// It ensures that the path exists and is not empty before returning.
//
// Parameters:
//   - ctx: The context containing the temporary directory path.
//
// Returns:
//   - string: The path to the temporary directory.
//   - error: An error if the path is not found or is empty.
//
// Usage:
//
//	path, err := GetTmpDirPath(ctx)
//	if err != nil {
//	    log.Printf("Failed to get temp dir: %v", err)
//	    return
//	}
//	// Use the temporary directory path
func GetTmpDirPath(ctx context.Context) (string, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GetTmpDirPath: retrieving path")

	tmpDir, ok := ctx.Value(TmpDirKey{}).(string)
	if !ok || tmpDir == "" {
		return "", errors.New("temporary directory path not found in context or is empty")
	}

	return tmpDir, nil
}

// CreateTmpDir creates a new temporary directory and stores its path in the context.
// The directory is created using os.MkdirTemp, ensuring a unique name.
//
// Parameters:
//   - ctx: The parent context.
//   - dir: The parent directory in which to create the temporary directory. If empty, os.TempDir() is used.
//   - prefix: The prefix for the temporary directory name.
//
// Returns:
//   - context.Context: A new context containing the temporary directory path.
//   - error: An error if the directory creation fails.
//
// Usage:
//
//	newCtx, err := CreateTmpDir(ctx, "", "myapp")
//	if err != nil {
//	    log.Printf("Failed to create temp dir: %v", err)
//	    return
//	}
//	// Use newCtx in subsequent operations
func CreateTmpDir(ctx context.Context, dir, prefix string) (context.Context, error) {
	tmpDir, err := os.MkdirTemp(dir, prefix+".*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory (dir: %s, prefix: %s): %w", dir, prefix, err)
	}

	return context.WithValue(ctx, TmpDirKey{}, tmpDir), nil
}

// DeleteTmpDir safely deletes the temporary directory specified in the context.
// It performs several safety checks to ensure the directory is valid before deletion.
//
// Parameters:
//   - ctx: The context containing the temporary directory path.
//
// Returns:
//   - error: An error if the deletion fails or if the directory is invalid.
//
// Usage:
//
//	if err := DeleteTmpDir(ctx); err != nil {
//	    log.Printf("Failed to delete temp dir: %v", err)
//	    return
//	}
func DeleteTmpDir(ctx context.Context) error {
	logger := log.Logger(ctx)

	tmpDir, err := GetTmpDirPath(ctx)
	if err != nil {
		return fmt.Errorf("failed to get temporary directory path: %w", err)
	}

	if !filepath.IsAbs(tmpDir) || !isSubdirectoryOfTemp(tmpDir) {
		return fmt.Errorf("invalid temporary directory path: %s", tmpDir)
	}

	logger.Debug().Str("tmpDir", tmpDir).Msg("Deleting temporary directory")

	if err := os.RemoveAll(tmpDir); err != nil {
		return fmt.Errorf("failed to delete temporary directory %s: %w", tmpDir, err)
	}

	return nil
}

// isSubdirectoryOfTemp checks if the given path is a subdirectory of the system's temporary directory.
// This is a safety measure to prevent accidental deletion of non-temporary directories.
//
// Parameters:
//   - path: The path to check.
//
// Returns:
//   - bool: True if the path is a subdirectory of the system's temporary directory, false otherwise.
func isSubdirectoryOfTemp(path string) bool {
	tempDir := os.TempDir()

	return strings.HasPrefix(path, tempDir)
}
