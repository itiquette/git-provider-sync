// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package directory

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
func (a *Adapter) Clone(ctx context.Context, options ports.CloneOptions) (ports.GitRepository, error) {
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
func (a *Adapter) Open(ctx context.Context, path string) (ports.GitRepository, error) {
	// Check if directory exists
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to access directory: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", path)
	}

	return &Repository{
		path: path,
	}, nil
}

// Init creates a new directory.
func (a *Adapter) Init(ctx context.Context, path string, options ports.InitOptions) (ports.GitRepository, error) {
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
	// Directory adapter supports file:// URLs and local paths
	return strings.HasPrefix(url, "file://") ||
		strings.HasPrefix(url, "/") ||
		strings.Contains(url, ":\\") // Windows paths
}

// GetName returns the name of this adapter.
func (a *Adapter) GetName() string {
	return "directory"
}

// Cleanup cleans up resources at the given path.
func (a *Adapter) Cleanup(ctx context.Context, path string) error {
	// For directory adapter, cleanup involves removing the directory if it exists
	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", path, err)
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
func (r *Repository) ListBranches(ctx context.Context) ([]ports.BranchInfo, error) {
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
func (r *Repository) CreateBranch(ctx context.Context, name, source string) error {
	return errors.New("branch operations not supported for directory repositories")
}

// CheckoutBranch is not supported for directory repositories.
func (r *Repository) CheckoutBranch(ctx context.Context, name string) error {
	return errors.New("branch operations not supported for directory repositories")
}

// DeleteBranch is not supported for directory repositories.
func (r *Repository) DeleteBranch(ctx context.Context, name string, force bool) error {
	return errors.New("branch operations not supported for directory repositories")
}

// SetDefaultBranch is not supported for directory repositories.
func (r *Repository) SetDefaultBranch(ctx context.Context, name string) error {
	return errors.New("branch operations not supported for directory repositories")
}

// ListRemotes returns empty list for directory repositories.
func (r *Repository) ListRemotes(ctx context.Context) ([]ports.RemoteInfo, error) {
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
func (r *Repository) AddRemote(ctx context.Context, name, url string) error {
	return errors.New("remote operations not supported for directory repositories")
}

// RemoveRemote is not supported for directory repositories.
func (r *Repository) RemoveRemote(ctx context.Context, name string) error {
	return errors.New("remote operations not supported for directory repositories")
}

// UpdateRemote is not supported for directory repositories.
func (r *Repository) UpdateRemote(ctx context.Context, name, url string) error {
	return errors.New("remote operations not supported for directory repositories")
}

// Fetch syncs directory from source if available.
func (r *Repository) Fetch(ctx context.Context, options ports.FetchOptions) error {
	if r.url == "" {
		return errors.New("no source URL configured for fetch")
	}

	if strings.HasPrefix(r.url, "file://") {
		sourcePath := strings.TrimPrefix(r.url, "file://")
		adapter := &Adapter{}

		return adapter.copyDirectory(sourcePath, r.path)
	}

	return fmt.Errorf("fetch not supported for URL type: %s", r.url)
}

// Pull is the same as fetch for directory repositories.
func (r *Repository) Pull(ctx context.Context, options ports.PullOptions) error {
	fetchOptions := ports.FetchOptions{
		Remote: options.Remote,
	}

	return r.Fetch(ctx, fetchOptions)
}

// Push copies directory to destination if supported.
func (r *Repository) Push(ctx context.Context, options ports.PushOptions) error {
	return errors.New("push operations not supported for directory repositories")
}

// GetCommit returns a dummy commit for directory repositories.
func (r *Repository) GetCommit(ctx context.Context, ref string) (ports.CommitInfo, error) {
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
func (r *Repository) ListCommits(ctx context.Context, options ports.ListCommitsOptions) ([]ports.CommitInfo, error) {
	return []ports.CommitInfo{}, nil
}

// ListTags returns empty list for directory repositories.
func (r *Repository) ListTags(ctx context.Context) ([]ports.TagInfo, error) {
	return []ports.TagInfo{}, nil
}

// CreateTag is not supported for directory repositories.
func (r *Repository) CreateTag(ctx context.Context, name, message, ref string) error {
	return errors.New("tag operations not supported for directory repositories")
}

// DeleteTag is not supported for directory repositories.
func (r *Repository) DeleteTag(ctx context.Context, name string) error {
	return errors.New("tag operations not supported for directory repositories")
}

// Status returns basic directory status.
func (r *Repository) Status(ctx context.Context) (ports.StatusResult, error) {
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
func (r *Repository) Diff(ctx context.Context, options ports.DiffOptions) (string, error) {
	return "", errors.New("diff operations not supported for directory repositories")
}

// Close is a no-op for directory repositories.
func (r *Repository) Close() error {
	return nil
}

// Helper methods

// copyDirectory recursively copies a directory from src to dst.
func (a *Adapter) copyDirectory(src, dst string) error {
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path from %s to %s: %w", src, path, err)
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// Create directory
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
