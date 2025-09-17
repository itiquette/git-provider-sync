// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package archive

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/constants"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// Static errors for err113 compliance.
var (
	ErrCreateBranchNotSupported     = errors.New("creating branches is not supported for archive repositories")
	ErrOnlyMainBranchExists         = errors.New("only 'main' branch exists in archive repositories")
	ErrDeleteBranchNotSupported     = errors.New("deleting branches is not supported for archive repositories")
	ErrSetDefaultBranchNotSupported = errors.New("setting default branch is not supported for archive repositories")
	ErrCommitNotFoundInArchive      = errors.New("commit not found in archive repository")
	ErrAddRemoteNotSupported        = errors.New("adding remotes is not supported for archive repositories")
	ErrArchiveReposNoRemotes        = errors.New("archive repositories don't have remotes")
	ErrRemoveRemoteNotSupported     = errors.New("removing remotes is not supported for archive repositories")
	ErrUpdateRemoteNotSupported     = errors.New("updating remotes is not supported for archive repositories")
	ErrCreateTagNotSupported        = errors.New("creating tags is not supported for archive repositories")
	ErrDeleteTagNotSupported        = errors.New("deleting tags is not supported for archive repositories")
	ErrDiffNotSupported             = errors.New("diff is not supported for archive repositories")
)

// Repository represents an archive-based repository
// is a simplified repository that provides basic operations
// For archive-extracted content.
type Repository struct {
	path   string
	config ports.GitConfig
	fs     ports.FileSystem
}

// Path returns the repository path.
func (r *Repository) Path() string {
	return r.path
}

// URL returns empty string for archive repositories.
func (r *Repository) URL() string {
	return ""
}

// Name returns the directory name.
func (r *Repository) Name() string {
	return filepath.Base(r.path)
}

// IsBare returns false as archive repositories are not bare.
func (r *Repository) IsBare() bool {
	return false
}

// IsClean returns true if the directory exists and is readable.
func (r *Repository) IsClean() bool {
	_, err := r.fs.Stat(r.path)

	return err == nil
}

// HasChanges always returns false for archive repositories.
func (r *Repository) HasChanges() bool {
	return false
}

// GetCurrentCommit returns a dummy commit for archive repositories
// Archive repositories don't have real git history.
func (r *Repository) GetCurrentCommit(_ context.Context) (ports.CommitInfo, error) {
	// Return a dummy commit representing the archive extraction time
	return ports.CommitInfo{
		Hash:    "archive-extraction",
		Message: "Archive extraction",
		Author: ports.PersonInfo{
			Name:  "Archive Adapter",
			Email: "archive@git-provider-sync",
			When:  time.Now(),
		},
		Committer: ports.PersonInfo{
			Name:  "Archive Adapter",
			Email: "archive@git-provider-sync",
			When:  time.Now(),
		},
		Parents: []string{},
	}, nil
}

// ListBranches returns a single "main" branch for archive repositories.
func (r *Repository) ListBranches(_ context.Context) ([]ports.BranchInfo, error) {
	return []ports.BranchInfo{
		{
			Name:      "main",
			Hash:      "archive-extraction",
			IsRemote:  false,
			IsCurrent: true,
			Upstream:  "",
		},
	}, nil
}

// GetBranches returns a single "main" branch for archive repositories (legacy method).
func (r *Repository) GetBranches(_ context.Context) ([]string, error) {
	return []string{"main"}, nil
}

// CurrentBranch returns "main" as the current branch.
func (r *Repository) CurrentBranch() (string, error) {
	return constants.DefaultBranch, nil
}

// GetCurrentBranch returns "main" as the current branch (legacy method).
func (r *Repository) GetCurrentBranch(_ context.Context) (string, error) {
	return constants.DefaultBranch, nil
}

// CreateBranch is not supported for archive repositories.
func (r *Repository) CreateBranch(_ context.Context, _, _ string) error {
	return ErrCreateBranchNotSupported
}

// CheckoutBranch is not supported for archive repositories.
func (r *Repository) CheckoutBranch(_ context.Context, branchName string) error {
	if branchName != "main" {
		return ErrOnlyMainBranchExists
	}

	return nil
}

// DeleteBranch is not supported for archive repositories.
func (r *Repository) DeleteBranch(_ context.Context, _ string, _ bool) error {
	return ErrDeleteBranchNotSupported
}

// SetDefaultBranch is not supported for archive repositories.
func (r *Repository) SetDefaultBranch(_ context.Context, _ string) error {
	return ErrSetDefaultBranchNotSupported
}

// ListCommits returns a single dummy commit.
func (r *Repository) ListCommits(ctx context.Context, _ ports.ListCommitsOptions) ([]ports.CommitInfo, error) {
	commit, err := r.GetCurrentCommit(ctx)
	if err != nil {
		return nil, err
	}

	return []ports.CommitInfo{commit}, nil
}

// GetCommits returns a single dummy commit (legacy method).
func (r *Repository) GetCommits(ctx context.Context, options ports.ListCommitsOptions) ([]ports.CommitInfo, error) {
	return r.ListCommits(ctx, options)
}

// GetCommit returns the dummy commit if hash matches.
func (r *Repository) GetCommit(ctx context.Context, hash string) (ports.CommitInfo, error) {
	if hash == "archive-extraction" || hash == "HEAD" {
		return r.GetCurrentCommit(ctx)
	}

	return ports.CommitInfo{}, ErrCommitNotFoundInArchive
}

// AddRemote is not supported for archive repositories.
func (r *Repository) AddRemote(_ context.Context, _, _ string) error {
	return ErrAddRemoteNotSupported
}

// ListRemotes returns an empty list for archive repositories.
func (r *Repository) ListRemotes(_ context.Context) ([]ports.RemoteInfo, error) {
	return []ports.RemoteInfo{}, nil
}

// GetRemote returns an error as archive repositories don't have remotes.
func (r *Repository) GetRemote(_ context.Context, _ string) (ports.RemoteInfo, error) {
	return ports.RemoteInfo{}, ErrArchiveReposNoRemotes
}

// RemoveRemote is not supported for archive repositories.
func (r *Repository) RemoveRemote(_ context.Context, _ string) error {
	return ErrRemoveRemoteNotSupported
}

// UpdateRemote is not supported for archive repositories.
func (r *Repository) UpdateRemote(_ context.Context, _, _ string) error {
	return ErrUpdateRemoteNotSupported
}

// ListTags returns an empty list for archive repositories.
func (r *Repository) ListTags(_ context.Context) ([]ports.TagInfo, error) {
	return []ports.TagInfo{}, nil
}

// CreateTag is not supported for archive repositories.
func (r *Repository) CreateTag(_ context.Context, _, _, _ string) error {
	return ErrCreateTagNotSupported
}

// DeleteTag is not supported for archive repositories.
func (r *Repository) DeleteTag(_ context.Context, _ string) error {
	return ErrDeleteTagNotSupported
}

// Status returns the status of files in the archive directory.
func (r *Repository) Status(_ context.Context) (ports.StatusResult, error) {
	var status ports.StatusResult

	// Check if the repository directory exists and has content
	entries, err := r.fs.ReadDir(r.path)
	if err != nil {
		return status, fmt.Errorf("failed to read directory: %w", err)
	}

	// Count files as "untracked" since archive repos don't have git tracking
	for _, entry := range entries {
		if !entry.IsDir() {
			status.Untracked = append(status.Untracked, entry.Name())
		}
	}

	return status, nil
}

// Diff is not supported for archive repositories.
func (r *Repository) Diff(_ context.Context, _ ports.DiffOptions) (string, error) {
	return "", ErrDiffNotSupported
}

// GetStatus returns the status of files in the archive directory (legacy method).
func (r *Repository) GetStatus(ctx context.Context) (ports.StatusResult, error) {
	return r.Status(ctx)
}

// Fetch is not supported for archive repositories.
func (r *Repository) Fetch(_ context.Context, _ ports.FetchOptions) error {
	return domain.ErrUnsupportedOperation
}

// Pull is not supported for archive repositories.
func (r *Repository) Pull(_ context.Context, _ ports.PullOptions) error {
	return domain.ErrPullNotSupportedArchive
}

// Push is not supported for archive repositories.
func (r *Repository) Push(_ context.Context, _ ports.PushOptions) error {
	return domain.ErrPushNotSupportedArchive
}

// Add is not supported for archive repositories.
func (r *Repository) Add(_ context.Context, _ []string) error {
	return domain.ErrAddFilesNotSupportedArchive
}

// Commit is not supported for archive repositories.
func (r *Repository) Commit(_ context.Context, _ string) error {
	return domain.ErrCommitNotSupportedArchive
}

// GetConfig returns the git configuration for this repository.
func (r *Repository) GetConfig() ports.GitConfig {
	return r.config
}

// GetRemoteURL returns an empty string as archive repositories don't have remotes.
func (r *Repository) GetRemoteURL(_ context.Context, _ string) (string, error) {
	return "", ErrArchiveReposNoRemotes
}

// SetRemoteURL is not supported for archive repositories.
func (r *Repository) SetRemoteURL(_ context.Context, _, _ string) error {
	return domain.ErrRemoteURLNotSupportedArchive
}

// GetWorkingDirectory returns the repository path.
func (r *Repository) GetWorkingDirectory() string {
	return r.path
}

// GetGitDirectory returns a simulated .git directory path.
func (r *Repository) GetGitDirectory() string {
	return filepath.Join(r.path, ".archive-metadata")
}

// HasUncommittedChanges always returns false for archive repositories.
func (r *Repository) HasUncommittedChanges(_ context.Context) (bool, error) {
	return false, nil
}

// GetFileContent reads the content of a file in the archive directory.
func (r *Repository) GetFileContent(_ context.Context, filePath string) ([]byte, error) {
	fullPath := filepath.Join(r.path, filePath)

	// Security check: ensure the path is within the repository
	absRepoPath, err := filepath.Abs(r.path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute repository path: %w", err)
	}

	absFilePath, err := filepath.Abs(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute file path: %w", err)
	}

	if !strings.HasPrefix(absFilePath, absRepoPath+string(filepath.Separator)) && absFilePath != absRepoPath {
		return nil, domain.ErrFilePathOutsideRepository
	}

	// #nosec G304 - Path is validated above for security
	data, err := r.fs.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// WriteFile writes content to a file in the archive directory.
func (r *Repository) WriteFile(_ context.Context, filePath string, content []byte) error {
	fullPath := filepath.Join(r.path, filePath)

	// Security check: ensure the path is within the repository
	absRepoPath, err := filepath.Abs(r.path)
	if err != nil {
		return fmt.Errorf("failed to get absolute repository path: %w", err)
	}

	absFilePath, err := filepath.Abs(fullPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute file path: %w", err)
	}

	if !strings.HasPrefix(absFilePath, absRepoPath+string(filepath.Separator)) && absFilePath != absRepoPath {
		return domain.ErrFilePathOutsideRepository
	}

	// Ensure parent directory exists
	err = r.fs.MkdirAll(r.fs.Dir(fullPath), 0750)
	if err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	if err := r.fs.WriteFile(fullPath, content, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ListFiles returns a list of files in the archive directory.
func (r *Repository) ListFiles(_ context.Context, path string) ([]string, error) {
	searchPath := r.path
	if path != "" {
		searchPath = filepath.Join(r.path, path)
	}

	var files []string

	err := filepath.WalkDir(searchPath, func(filePath string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !dirEntry.IsDir() {
			relPath, err := filepath.Rel(r.path, filePath)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}

			files = append(files, relPath)
		}

		return nil
	})
	if err != nil {
		return files, fmt.Errorf("failed to walk directory: %w", err)
	}

	return files, nil
}

// Close performs any necessary cleanup for the repository.
func (r *Repository) Close() error {
	// Archive repositories don't need special cleanup
	return nil
}
