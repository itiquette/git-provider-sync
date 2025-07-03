// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

// Package filesystem provides file system related adapters for the hexagonal architecture.
// This restores the sophisticated temporary directory management from main branch.
package filesystem

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// TempDirKey is the key type for storing and retrieving the temporary directory path in the context.
// Using a custom type for the key helps prevent collisions with other context values.
type TempDirKey struct{}

// TempDirManager provides sophisticated temporary directory management.
// This restores the temp directory functionality from main branch.
type TempDirManager struct {
	logger ports.Logger
}

// NewTempDirManager creates a new temporary directory manager.
func NewTempDirManager(logger ports.Logger) *TempDirManager {
	return &TempDirManager{
		logger: logger,
	}
}

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
// This restores the GetTmpDirPath functionality from main branch.
func (tdm *TempDirManager) GetTmpDirPath(ctx context.Context) (string, error) {
	tdm.logger.Debug(ctx, "Retrieving temporary directory path from context", nil)

	tmpDir, ok := ctx.Value(TempDirKey{}).(string)
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
// This restores the CreateTmpDir functionality from main branch.
func (tdm *TempDirManager) CreateTmpDir(ctx context.Context, dir, prefix string) (context.Context, error) {
	tdm.logger.Debug(ctx, "Creating temporary directory", map[string]interface{}{
		"dir":    dir,
		"prefix": prefix,
	})

	tmpDir, err := os.MkdirTemp(dir, prefix+".*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory (dir: %s, prefix: %s): %w", dir, prefix, err)
	}

	tdm.logger.Info(ctx, "Created temporary directory", map[string]interface{}{
		"tmp_dir": tmpDir,
	})

	return context.WithValue(ctx, TempDirKey{}, tmpDir), nil
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
// This restores the DeleteTmpDir functionality from main branch.
func (tdm *TempDirManager) DeleteTmpDir(ctx context.Context) error {
	tdm.logger.Debug(ctx, "Attempting to delete temporary directory", nil)

	tmpDir, err := tdm.GetTmpDirPath(ctx)
	if err != nil {
		return fmt.Errorf("failed to get temporary directory path: %w", err)
	}

	if !filepath.IsAbs(tmpDir) || !tdm.isSubdirectoryOfTemp(tmpDir) {
		return fmt.Errorf("invalid temporary directory path: %s", tmpDir)
	}

	tdm.logger.Debug(ctx, "Deleting temporary directory", map[string]interface{}{
		"tmp_dir": tmpDir,
	})

	if err := os.RemoveAll(tmpDir); err != nil {
		return fmt.Errorf("failed to delete temporary directory %s: %w", tmpDir, err)
	}

	tdm.logger.Info(ctx, "Successfully deleted temporary directory", map[string]interface{}{
		"tmp_dir": tmpDir,
	})

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
func (tdm *TempDirManager) isSubdirectoryOfTemp(path string) bool {
	tempDir := os.TempDir()

	return strings.HasPrefix(path, tempDir)
}

// Convenience functions for backward compatibility with main branch patterns

// GetTmpDirPath retrieves the temporary directory path from context (static version).
// This maintains compatibility with main branch usage patterns.
func GetTmpDirPath(ctx context.Context) (string, error) {
	tmpDir, ok := ctx.Value(TempDirKey{}).(string)
	if !ok || tmpDir == "" {
		return "", errors.New("temporary directory path not found in context or is empty")
	}

	return tmpDir, nil
}

// CreateTmpDir creates a temporary directory and stores it in context (static version).
// This maintains compatibility with main branch usage patterns.
func CreateTmpDir(ctx context.Context, dir, prefix string) (context.Context, error) {
	tmpDir, err := os.MkdirTemp(dir, prefix+".*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory (dir: %s, prefix: %s): %w", dir, prefix, err)
	}

	return context.WithValue(ctx, TempDirKey{}, tmpDir), nil
}

// DeleteTmpDir deletes the temporary directory from context (static version).
// This maintains compatibility with main branch usage patterns.
func DeleteTmpDir(ctx context.Context) error {
	tmpDir, err := GetTmpDirPath(ctx)
	if err != nil {
		return fmt.Errorf("failed to get temporary directory path: %w", err)
	}

	if !filepath.IsAbs(tmpDir) || !isSubdirectoryOfTemp(tmpDir) {
		return fmt.Errorf("invalid temporary directory path: %s", tmpDir)
	}

	if err := os.RemoveAll(tmpDir); err != nil {
		return fmt.Errorf("failed to delete temporary directory %s: %w", tmpDir, err)
	}

	return nil
}

// isSubdirectoryOfTemp static version for compatibility.
func isSubdirectoryOfTemp(path string) bool {
	tempDir := os.TempDir()

	return strings.HasPrefix(path, tempDir)
}
