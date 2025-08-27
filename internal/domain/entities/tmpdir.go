// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package entities

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"itiquette/git-provider-sync/internal/domain"
)

// TmpDirKey is the key type for storing and retrieving the temporary directory path in the context.
// Using a custom type for the key helps prevent collisions with other context values.
type TmpDirKey struct{}

// TmpDir represents a temporary directory with immutable properties.
// This is a value object in DDD terms - it's defined by its value, not identity.
type TmpDir struct {
	path string
}

// NewTmpDir creates a new TmpDir value object with validation.
func NewTmpDir(path string) (TmpDir, error) {
	if path == "" {
		return TmpDir{}, domain.ErrTempDirectoryNotFound
	}

	cleanPath := filepath.Clean(path)

	return TmpDir{path: cleanPath}, nil
}

// Path returns the temporary directory path.
func (t TmpDir) Path() string {
	return t.path
}

// SubDir creates a new TmpDir representing a subdirectory.
func (t TmpDir) SubDir(subPath string) TmpDir {
	cleanedSub := CleanPath(subPath)
	fullPath := filepath.Join(t.path, cleanedSub)

	return TmpDir{path: filepath.Clean(fullPath)}
}

// Exists checks if the temporary directory exists on the filesystem.
func (t TmpDir) Exists() bool {
	info, err := os.Stat(t.path)

	return err == nil && info.IsDir()
}

// String implements fmt.Stringer for debugging.
func (t TmpDir) String() string {
	return fmt.Sprintf("TmpDir{path: %q}", t.path)
}

// GetTmpDirPath retrieves the temporary directory path from the context.
// It ensures that the path exists and is not empty before returning.
func GetTmpDirPath(ctx context.Context) (string, error) {
	tmpDir, ok := ctx.Value(TmpDirKey{}).(string)
	if !ok || tmpDir == "" {
		return "", domain.ErrTempDirectoryNotFound
	}

	return tmpDir, nil
}

// CreateTmpDir creates a new temporary directory and stores its path in the context.
// The directory is created using os.MkdirTemp, ensuring a unique name.
func CreateTmpDir(ctx context.Context, dir, prefix string) (context.Context, error) {
	tmpDir, err := os.MkdirTemp(dir, prefix+".*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory (dir: %s, prefix: %s): %w", dir, prefix, err)
	}

	return context.WithValue(ctx, TmpDirKey{}, tmpDir), nil
}

// DeleteTmpDir removes the temporary directory stored in the context.
// It safely handles the case where the directory might not exist or the context value is missing.
func DeleteTmpDir(ctx context.Context) error {
	tmpDir, err := GetTmpDirPath(ctx)
	if err != nil {
		// If we can't get the temp dir path, it might not have been set or already cleaned up
		return nil //nolint:nilerr // Intentional: missing temp dir is not an error condition
	}

	if err := os.RemoveAll(tmpDir); err != nil {
		return fmt.Errorf("failed to delete temporary directory %s: %w", tmpDir, err)
	}

	return nil
}

// WithTmpDir adds a temporary directory path to the context without creating it.
// This is useful when you want to use an existing directory.
func WithTmpDir(ctx context.Context, tmpDir string) context.Context {
	return context.WithValue(ctx, TmpDirKey{}, tmpDir)
}

// GetOrCreateTmpDir gets the temporary directory from context, or creates one if it doesn't exist.
func GetOrCreateTmpDir(ctx context.Context, dir, prefix string) (context.Context, string, error) {
	tmpDir, err := GetTmpDirPath(ctx)
	if err == nil {
		return ctx, tmpDir, nil
	}

	newCtx, err := CreateTmpDir(ctx, dir, prefix)
	if err != nil {
		return ctx, "", err
	}

	tmpDir, err = GetTmpDirPath(newCtx)
	if err != nil {
		return ctx, "", err
	}

	return newCtx, tmpDir, nil
}

// GetSubTmpDir creates a subdirectory within the temporary directory.
func GetSubTmpDir(ctx context.Context, subDir string) (string, error) {
	tmpDir, err := GetTmpDirPath(ctx)
	if err != nil {
		return "", err
	}

	subPath := filepath.Join(tmpDir, subDir)
	if err := os.MkdirAll(subPath, 0750); err != nil {
		return "", fmt.Errorf("failed to create subdirectory %s: %w", subPath, err)
	}

	return subPath, nil
}

// CleanPath ensures the path is clean and doesn't contain dangerous elements.
func CleanPath(path string) string {
	cleaned := filepath.Clean(path)

	// Remove any leading separators to prevent absolute path issues
	cleaned = strings.TrimLeft(cleaned, string(filepath.Separator))

	// Remove any remaining parent directory traversals for security
	for strings.HasPrefix(cleaned, "..") {
		switch {
		case strings.HasPrefix(cleaned, "../"):
			cleaned = cleaned[3:]
		case cleaned == "..":
			return "."
		default:
			return cleaned
		}
	}

	// If path becomes empty after cleaning, return current directory
	if cleaned == "" {
		return "."
	}

	return cleaned
}
