// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package directory provides filesystem-based repository operations.
package directory

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// Adapter implements the GitOperations interface for directory-based operations.
// This adapter provides a simpler interface for basic directory management
// without full git functionality, useful for file-based repositories or backups.
type Adapter struct {
	config ports.GitConfig
}

// New creates a new directory adapter.
func New(config ports.GitConfig) *Adapter {
	return &Adapter{
		config: config,
	}
}

// Clone creates a directory copy from source to destination.
// For directory adapter, this means copying files rather than git cloning.
func (a *Adapter) Clone(_ /* ctx */ context.Context, options ports.CloneOptions) (ports.GitRepository, error) { //nolint:ireturn
	// Ensure destination directory exists
	err := os.MkdirAll(options.Path, 0750)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// For file:// URLs, copy the directory
	if strings.HasPrefix(options.URL, "file://") {
		sourcePath := strings.TrimPrefix(options.URL, "file://")

		err = a.copyDirectory(sourcePath, options.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to copy directory: %w", err)
		}
	}

	return &Repository{
		path: options.Path,
		url:  options.URL,
	}, nil
}

// Open opens an existing directory as a repository.
func (a *Adapter) Open(_ /* ctx */ context.Context, path string) (ports.GitRepository, error) { //nolint:ireturn
	// Check if directory exists
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to access directory: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("%w: %s", domain.ErrPathNotDirectory, path)
	}

	return &Repository{
		path: path,
	}, nil
}

// Init creates a new directory.
func (a *Adapter) Init(_ /* ctx */ context.Context, path string, _ ports.InitOptions) (ports.GitRepository, error) { //nolint:ireturn
	err := os.MkdirAll(path, 0750)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	return &Repository{
		path: path,
	}, nil
}

// SupportsURL checks if directory adapter supports the given URL.
func (a *Adapter) SupportsURL(url string) bool {
	// Directory adapter supports file:// URLs and Unix local paths
	return strings.HasPrefix(url, "file://") ||
		strings.HasPrefix(url, "/")
}

// GetName returns the name of this adapter.
func (a *Adapter) GetName() string {
	return "directory"
}

// Cleanup cleans up resources at the given path.
func (a *Adapter) Cleanup(_ context.Context, path string) error {
	// For directory adapter, cleanup involves removing the directory if it exists
	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", path, err)
	}

	return nil
}

// CreateTmpDir implements the ports.GitOperations interface.
func (a *Adapter) CreateTmpDir(ctx context.Context, dir, prefix string) (context.Context, error) {
	ctxWithTmp, err := entities.CreateTmpDir(ctx, dir, prefix)
	if err != nil {
		return ctx, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	return ctxWithTmp, nil
}

// GetTmpDirPath implements the ports.GitOperations interface.
func (a *Adapter) GetTmpDirPath(ctx context.Context) (string, error) {
	path, err := entities.GetTmpDirPath(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get temporary directory path: %w", err)
	}

	return path, nil
}

// DeleteTmpDir implements the ports.GitOperations interface.
func (a *Adapter) DeleteTmpDir(ctx context.Context) error {
	if err := entities.DeleteTmpDir(ctx); err != nil {
		return fmt.Errorf("failed to delete temporary directory: %w", err)
	}

	return nil
}

// Repository implements the GitRepository interface for directories.
type Repository struct {
	path string
	url  string
}

// Path returns the directory path.
func (r *Repository) Path() string {
	return r.path
}

// URL returns the source URL if available.
func (r *Repository) URL() string {
	return r.url
}

// Name returns the directory name.
func (r *Repository) Name() string {
	return filepath.Base(r.path)
}

// IsBare always returns false for directory repositories.
func (r *Repository) IsBare() bool {
	return false
}

// IsClean always returns true for directory repositories.
func (r *Repository) IsClean() bool {
	return true
}

// HasChanges always returns false for directory repositories.
func (r *Repository) HasChanges() bool {
	return false
}

// CurrentBranch returns "main" as default for directory repositories.
func (r *Repository) CurrentBranch() (string, error) {
	return "main", nil
}

// ListBranches returns a single "main" branch for directory repositories.
func (r *Repository) ListBranches(_ context.Context) ([]ports.BranchInfo, error) {
	return []ports.BranchInfo{
		{
			Name:      "main",
			Hash:      "0000000000000000000000000000000000000000",
			IsRemote:  false,
			IsCurrent: true,
		},
	}, nil
}

// CreateBranch is not supported for directory repositories.
func (r *Repository) CreateBranch(_ context.Context, _, _ string) error {
	return domain.ErrBranchOpsNotSupported
}

// CheckoutBranch is not supported for directory repositories.
func (r *Repository) CheckoutBranch(_ context.Context, _ string) error {
	return domain.ErrBranchOpsNotSupported
}

// DeleteBranch is not supported for directory repositories.
func (r *Repository) DeleteBranch(_ context.Context, _ string, _ bool) error {
	return domain.ErrBranchOpsNotSupported
}

// SetDefaultBranch is not supported for directory repositories.
func (r *Repository) SetDefaultBranch(_ context.Context, _ string) error {
	return domain.ErrBranchOpsNotSupported
}

// ListRemotes returns empty list for directory repositories.
func (r *Repository) ListRemotes(_ context.Context) ([]ports.RemoteInfo, error) {
	if r.url != "" {
		return []ports.RemoteInfo{
			{
				Name:     "origin",
				URL:      r.url,
				FetchURL: r.url,
				PushURL:  r.url,
			},
		}, nil
	}

	return []ports.RemoteInfo{}, nil
}

// AddRemote is not supported for directory repositories.
func (r *Repository) AddRemote(_ context.Context, _, _ string) error {
	return domain.ErrRemoteOpsNotSupported
}

// RemoveRemote is not supported for directory repositories.
func (r *Repository) RemoveRemote(_ context.Context, _ string) error {
	return domain.ErrRemoteOpsNotSupported
}

// UpdateRemote is not supported for directory repositories.
func (r *Repository) UpdateRemote(_ context.Context, _, _ string) error {
	return domain.ErrRemoteOpsNotSupported
}

// Fetch syncs directory from source if available.
func (r *Repository) Fetch(_ context.Context, _ ports.FetchOptions) error {
	if r.url == "" {
		return domain.ErrNoSourceURLConfigured
	}

	if strings.HasPrefix(r.url, "file://") {
		sourcePath := strings.TrimPrefix(r.url, "file://")
		adapter := &Adapter{}

		return adapter.copyDirectory(sourcePath, r.path)
	}

	return fmt.Errorf("%w: %s", domain.ErrFetchNotSupportedForURLType, r.url)
}

// Pull is the same as fetch for directory repositories.
func (r *Repository) Pull(ctx context.Context, options ports.PullOptions) error {
	fetchOptions := ports.FetchOptions{
		Remote: options.Remote,
	}

	return r.Fetch(ctx, fetchOptions)
}

// Push copies directory to destination if supported.
func (r *Repository) Push(_ context.Context, _ ports.PushOptions) error {
	return domain.ErrPushOpsNotSupported
}

// GetCommit returns a dummy commit for directory repositories.
func (r *Repository) GetCommit(_ context.Context, _ string) (ports.CommitInfo, error) {
	return ports.CommitInfo{
		Hash:    "0000000000000000000000000000000000000000",
		Message: "Directory snapshot",
		Author: ports.PersonInfo{
			Name:  "Directory System",
			Email: "system@directory",
		},
		Committer: ports.PersonInfo{
			Name:  "Directory System",
			Email: "system@directory",
		},
	}, nil
}

// ListCommits returns empty list for directory repositories.
func (r *Repository) ListCommits(_ context.Context, _ ports.ListCommitsOptions) ([]ports.CommitInfo, error) {
	return []ports.CommitInfo{}, nil
}

// ListTags returns empty list for directory repositories.
func (r *Repository) ListTags(_ context.Context) ([]ports.TagInfo, error) {
	return []ports.TagInfo{}, nil
}

// CreateTag is not supported for directory repositories.
func (r *Repository) CreateTag(_ context.Context, _, _, _ string) error {
	return domain.ErrTagOpsNotSupported
}

// DeleteTag is not supported for directory repositories.
func (r *Repository) DeleteTag(_ context.Context, _ string) error {
	return domain.ErrTagOpsNotSupported
}

// Status returns basic directory status.
func (r *Repository) Status(_ context.Context) (ports.StatusResult, error) {
	// Check if directory exists and is accessible
	_, err := os.Stat(r.path)
	if err != nil {
		return ports.StatusResult{}, fmt.Errorf("failed to access directory: %w", err)
	}

	return ports.StatusResult{
		IsClean: true,
	}, nil
}

// Diff is not supported for directory repositories.
func (r *Repository) Diff(_ context.Context, _ ports.DiffOptions) (string, error) {
	return "", domain.ErrDiffOpsNotSupported
}

// Close is a no-op for directory repositories.
func (r *Repository) Close() error {
	return nil
}

// Helper methods

// copyDirectory recursively copies a directory from src to dst.
func (a *Adapter) copyDirectory(src, dst string) error {
	err := filepath.WalkDir(src, func(path string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path from %s to %s: %w", src, path, err)
		}

		dstPath := filepath.Join(dst, relPath)

		if dirEntry.IsDir() {
			// Create directory - get mode from DirEntry
			info, err := dirEntry.Info()
			if err != nil {
				return fmt.Errorf("failed to get dir info: %w", err)
			}

			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		return a.copyFile(path, dstPath)
	})
	if err != nil {
		return fmt.Errorf("failed to copy directory from %s to %s: %w", src, dst, err)
	}

	return nil
}

// copyFile copies a single file from src to dst.
func (a *Adapter) copyFile(src, dst string) error {
	// Read source file
	// #nosec G304 - Source path comes from trusted directory walking
	srcData, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Get source file info for permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	// Create destination directory if needed
	dstDir := filepath.Dir(dst)

	err = os.MkdirAll(dstDir, 0750)
	if err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Write destination file
	err = os.WriteFile(dst, srcData, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}
