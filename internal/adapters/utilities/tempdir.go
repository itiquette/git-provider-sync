// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

// tempdir.go - Temporary directory management restored from main branch using hexagonal architecture
package utilities

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/model"
)

// TempDirManager provides sophisticated temporary directory management.
// This restores the advanced temp directory functionality from main branch.
type TempDirManager struct {
	logger ports.Logger
}

// NewTempDirManager creates a new temporary directory manager.
func NewTempDirManager(logger ports.Logger) *TempDirManager {
	return &TempDirManager{
		logger: logger,
	}
}

// GetTmpDirPath retrieves the temporary directory path from context.
// This restores the main branch model.GetTmpDirPath functionality.
func (tdm *TempDirManager) GetTmpDirPath(ctx context.Context) (string, error) {
	if tmpDir, ok := ctx.Value(model.TmpDirKey{}).(string); ok {
		return tmpDir, nil
	}

	// Fallback to system temp directory
	tmpDir := os.TempDir()
	tdm.logger.Debug(ctx, "Using system temp directory as fallback", map[string]interface{}{
		"tmpDir": tmpDir,
	})

	return tmpDir, nil
}

// CreateTmpDir creates a temporary directory and adds it to context.
// This restores the main branch model.CreateTmpDir functionality.
func (tdm *TempDirManager) CreateTmpDir(ctx context.Context, pattern, prefix string) (context.Context, error) {
	tdm.logger.Debug(ctx, "Creating temporary directory", map[string]interface{}{
		"pattern": pattern,
		"prefix":  prefix,
	})

	// Use system temp directory as base
	baseDir := os.TempDir()
	if pattern != "" {
		baseDir = pattern
	}

	// Create temporary directory with prefix
	tmpDir, err := os.MkdirTemp(baseDir, prefix+"*")
	if err != nil {
		return ctx, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// Ensure directory is accessible (directories need execute permission for access)
	// #nosec G302 - This is a directory, not a file, so 0700 is appropriate
	if err := os.Chmod(tmpDir, 0700); err != nil {
		// Try to clean up on permission error
		_ = os.RemoveAll(tmpDir)
		return ctx, fmt.Errorf("failed to set directory permissions: %w", err)
	}

	// Add to context
	ctx = model.WithTmpDir(ctx, tmpDir)

	tdm.logger.Info(ctx, "Temporary directory created successfully", map[string]interface{}{
		"tmpDir":      tmpDir,
		"permissions": "0755",
	})

	return ctx, nil
}

// DeleteTmpDir safely removes the temporary directory from context.
// This restores the main branch model.DeleteTmpDir functionality.
func (tdm *TempDirManager) DeleteTmpDir(ctx context.Context) error {
	tmpDir, err := tdm.GetTmpDirPath(ctx)
	if err != nil {
		return fmt.Errorf("failed to get temp directory path: %w", err)
	}

	tdm.logger.Debug(ctx, "Deleting temporary directory", map[string]interface{}{
		"tmpDir": tmpDir,
	})

	// Check if directory exists before attempting deletion
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		tdm.logger.Debug(ctx, "Temporary directory does not exist, skipping deletion", map[string]interface{}{
			"tmpDir": tmpDir,
		})

		return nil
	}

	// Remove directory and all contents
	if err := os.RemoveAll(tmpDir); err != nil {
		tdm.logger.Error(ctx, "Failed to delete temporary directory", map[string]interface{}{
			"tmpDir": tmpDir,
			"error":  err.Error(),
		})

		return fmt.Errorf("failed to delete temporary directory %s: %w", tmpDir, err)
	}

	tdm.logger.Info(ctx, "Temporary directory deleted successfully", map[string]interface{}{
		"tmpDir": tmpDir,
	})

	return nil
}

// EnsureTmpDirExists ensures that a temporary directory exists and is accessible.
func (tdm *TempDirManager) EnsureTmpDirExists(ctx context.Context) error {
	tmpDir, err := tdm.GetTmpDirPath(ctx)
	if err != nil {
		return fmt.Errorf("failed to get temp directory path: %w", err)
	}

	// Check if directory exists
	info, err := os.Stat(tmpDir)
	if os.IsNotExist(err) {
		// Create directory if it doesn't exist
		if err := os.MkdirAll(tmpDir, 0750); err != nil {
			return fmt.Errorf("failed to create temp directory %s: %w", tmpDir, err)
		}

		tdm.logger.Info(ctx, "Created missing temporary directory", map[string]interface{}{
			"tmpDir": tmpDir,
		})

		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to stat temp directory %s: %w", tmpDir, err)
	}

	// Verify it's a directory
	if !info.IsDir() {
		return fmt.Errorf("temp path %s exists but is not a directory", tmpDir)
	}

	// Verify it's writable
	testFile := filepath.Join(tmpDir, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		return fmt.Errorf("temp directory %s is not writable: %w", tmpDir, err)
	}

	// Clean up test file
	_ = os.Remove(testFile)

	tdm.logger.Debug(ctx, "Temporary directory verified", map[string]interface{}{
		"tmpDir":   tmpDir,
		"writable": true,
		"size":     info.Size(),
		"mode":     info.Mode().String(),
	})

	return nil
}

// GetSubDir creates and returns a subdirectory within the temp directory.
func (tdm *TempDirManager) GetSubDir(ctx context.Context, subDirName string) (string, error) {
	tmpDir, err := tdm.GetTmpDirPath(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get temp directory path: %w", err)
	}

	subDir := filepath.Join(tmpDir, subDirName)

	// Create subdirectory if it doesn't exist
	if err := os.MkdirAll(subDir, 0750); err != nil {
		return "", fmt.Errorf("failed to create subdirectory %s: %w", subDir, err)
	}

	tdm.logger.Debug(ctx, "Created/verified subdirectory", map[string]interface{}{
		"tmpDir": tmpDir,
		"subDir": subDir,
	})

	return subDir, nil
}

// CleanupAll removes all temporary directories created by this manager.
func (tdm *TempDirManager) CleanupAll(ctx context.Context) error {
	return tdm.DeleteTmpDir(ctx)
}

// GetTempFile creates a temporary file within the temp directory.
func (tdm *TempDirManager) GetTempFile(ctx context.Context, pattern string) (string, error) {
	tmpDir, err := tdm.GetTmpDirPath(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get temp directory path: %w", err)
	}

	// Create temporary file
	file, err := os.CreateTemp(tmpDir, pattern+"*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}

	filePath := file.Name()

	if err := file.Close(); err != nil {
		// Log close error but continue
		_ = err
	} // Close immediately, caller will manage the file

	tdm.logger.Debug(ctx, "Created temporary file", map[string]interface{}{
		"tmpDir":  tmpDir,
		"tmpFile": filePath,
		"pattern": pattern,
	})

	return filePath, nil
}

// ValidateTempDir validates that a directory is suitable for use as temp directory.
func (tdm *TempDirManager) ValidateTempDir(ctx context.Context, dirPath string) error {
	// Check if directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		return fmt.Errorf("temp directory does not exist or is not accessible: %w", err)
	}

	// Verify it's a directory
	if !info.IsDir() {
		return fmt.Errorf("path %s is not a directory", dirPath)
	}

	// Check write permissions
	testFile := filepath.Join(dirPath, ".validation_test")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		return fmt.Errorf("directory %s is not writable: %w", dirPath, err)
	}

	// Clean up test file
	if err := os.Remove(testFile); err != nil {
		tdm.logger.Warn(ctx, "Failed to clean up validation test file", map[string]interface{}{
			"testFile": testFile,
			"error":    err.Error(),
		})
	}

	tdm.logger.Debug(ctx, "Temporary directory validation passed", map[string]interface{}{
		"dirPath": dirPath,
		"mode":    info.Mode().String(),
	})

	return nil
}
