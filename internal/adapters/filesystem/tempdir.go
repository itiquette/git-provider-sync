// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package filesystem provides file system related adapters for the hexagonal architecture.
package filesystem

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Static errors for err113 compliance.
var (
	ErrTempDirNotFound    = errors.New("temporary directory path not found in context or is empty")
	ErrInvalidTempDirPath = errors.New("invalid temporary directory path")
)

// TempDirKey is the key type for storing and retrieving the temporary directory path in the context.
// Using a custom type for the key helps prevent collisions with other context values.
type TempDirKey struct{}

// GetTmpDirPath retrieves the temporary directory path from context.
func GetTmpDirPath(ctx context.Context) (string, error) {
	tmpDir, ok := ctx.Value(TempDirKey{}).(string)
	if !ok || tmpDir == "" {
		return "", ErrTempDirNotFound
	}

	return tmpDir, nil
}

// CreateTmpDir creates a temporary directory and stores it in context.
func CreateTmpDir(ctx context.Context, dir, prefix string) (context.Context, error) {
	tmpDir, err := os.MkdirTemp(dir, prefix+".*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory (dir: %s, prefix: %s): %w", dir, prefix, err)
	}

	return context.WithValue(ctx, TempDirKey{}, tmpDir), nil
}

// DeleteTmpDir deletes the temporary directory from context.
// Validates directory before deletion.
func DeleteTmpDir(ctx context.Context) error {
	tmpDir, err := GetTmpDirPath(ctx)
	if err != nil {
		return fmt.Errorf("failed to get temporary directory path: %w", err)
	}

	if !filepath.IsAbs(tmpDir) || !isSubdirectoryOfTemp(tmpDir) {
		return fmt.Errorf("%w: %s", ErrInvalidTempDirPath, tmpDir)
	}

	if err := os.RemoveAll(tmpDir); err != nil {
		return fmt.Errorf("failed to delete temporary directory %s: %w", tmpDir, err)
	}

	return nil
}

// isSubdirectoryOfTemp checks if the given path is a subdirectory of the system's temporary directory.
// Safety measure to prevent accidental deletion of non-temporary directories.
func isSubdirectoryOfTemp(path string) bool {
	tempDir, err := filepath.Abs(os.TempDir())
	if err != nil {
		return false
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	// Use filepath.Rel for proper path validation - prevents path traversal
	relPath, err := filepath.Rel(tempDir, absPath)
	if err != nil {
		return false
	}

	// Ensure path doesn't escape temp directory
	return !strings.HasPrefix(relPath, "..") && !strings.HasPrefix(relPath, "/")
}
