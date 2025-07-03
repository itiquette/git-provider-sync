// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package archive

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/domain/constants"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// Repository represents an archive-based repository.
// This is a simplified repository that provides basic operations
// for archive-extracted content.
type Repository struct {
	path   string
	config ports.GitConfig
}

// Path returns the repository path.
func (r *Repository) Path() string {
	return r.path
}

// URL returns an empty string as archive repositories don't have URLs.
func (r *Repository) URL() string {
	return ""
}

// Name returns the repository name (directory name).
func (r *Repository) Name() string {
	return filepath.Base(r.path)
}

// IsBare returns false as archive repositories are not bare.
func (r *Repository) IsBare() bool {
	return false
}

// IsClean returns true if the directory exists and is readable.
func (r *Repository) IsClean() bool {
	_, err := os.Stat(r.path)

	return err == nil
}

// HasChanges always returns false for archive repositories.
func (r *Repository) HasChanges() bool {
	return false
}

// GetCurrentCommit returns a dummy commit for archive repositories.
// Archive repositories don't have real git history.
func (r *Repository) GetCurrentCommit(ctx context.Context) (ports.CommitInfo, error) {
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
func (r *Repository) ListBranches(ctx context.Context) ([]ports.BranchInfo, error) {
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
func (r *Repository) GetBranches(ctx context.Context) ([]string, error) {
	return []string{"main"}, nil
}

// CurrentBranch returns "main" as the current branch.
func (r *Repository) CurrentBranch() (string, error) {
	return constants.DefaultBranch, nil
}

// GetCurrentBranch returns "main" as the current branch (legacy method).
func (r *Repository) GetCurrentBranch(ctx context.Context) (string, error) {
	return constants.DefaultBranch, nil
}

// CreateBranch is not supported for archive repositories.
func (r *Repository) CreateBranch(ctx context.Context, name, source string) error {
	return errors.New("creating branches is not supported for archive repositories")
}

// CheckoutBranch is not supported for archive repositories.
func (r *Repository) CheckoutBranch(ctx context.Context, branchName string) error {
	if branchName != "main" {
		return errors.New("only 'main' branch exists in archive repositories")
	}

	return nil
}

// DeleteBranch is not supported for archive repositories.
func (r *Repository) DeleteBranch(ctx context.Context, name string, force bool) error {
	return errors.New("deleting branches is not supported for archive repositories")
}

// SetDefaultBranch is not supported for archive repositories.
func (r *Repository) SetDefaultBranch(ctx context.Context, name string) error {
	return errors.New("setting default branch is not supported for archive repositories")
}

// ListCommits returns a single dummy commit.
func (r *Repository) ListCommits(ctx context.Context, options ports.ListCommitsOptions) ([]ports.CommitInfo, error) {
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

	return ports.CommitInfo{}, errors.New("commit not found in archive repository")
}

// AddRemote is not supported for archive repositories.
func (r *Repository) AddRemote(ctx context.Context, name, url string) error {
	return errors.New("adding remotes is not supported for archive repositories")
}

// ListRemotes returns an empty list for archive repositories.
func (r *Repository) ListRemotes(ctx context.Context) ([]ports.RemoteInfo, error) {
	return []ports.RemoteInfo{}, nil
}

// GetRemote returns an error as archive repositories don't have remotes.
func (r *Repository) GetRemote(ctx context.Context, name string) (ports.RemoteInfo, error) {
	return ports.RemoteInfo{}, errors.New("archive repositories don't have remotes")
}

// RemoveRemote is not supported for archive repositories.
func (r *Repository) RemoveRemote(ctx context.Context, name string) error {
	return errors.New("removing remotes is not supported for archive repositories")
}

// UpdateRemote is not supported for archive repositories.
func (r *Repository) UpdateRemote(ctx context.Context, name, url string) error {
	return errors.New("updating remotes is not supported for archive repositories")
}

// ListTags returns an empty list as archive repositories don't have tags.
func (r *Repository) ListTags(ctx context.Context) ([]ports.TagInfo, error) {
	return []ports.TagInfo{}, nil
}

// CreateTag is not supported for archive repositories.
func (r *Repository) CreateTag(ctx context.Context, name, message, ref string) error {
	return errors.New("creating tags is not supported for archive repositories")
}

// DeleteTag is not supported for archive repositories.
func (r *Repository) DeleteTag(ctx context.Context, name string) error {
	return errors.New("deleting tags is not supported for archive repositories")
}

// Status returns the status of files in the archive directory.
func (r *Repository) Status(ctx context.Context) (ports.StatusResult, error) {
	var status ports.StatusResult

	// Check if the repository directory exists and has content
	entries, err := os.ReadDir(r.path)
	if err != nil {
		return status, err
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
func (r *Repository) Diff(ctx context.Context, options ports.DiffOptions) (string, error) {
	return "", errors.New("diff is not supported for archive repositories")
}

// GetStatus returns the status of files in the archive directory (legacy method).
func (r *Repository) GetStatus(ctx context.Context) (ports.StatusResult, error) {
	return r.Status(ctx)
}

// Fetch is not supported for archive repositories.
func (r *Repository) Fetch(ctx context.Context, options ports.FetchOptions) error {
	return errors.New("fetch is not supported for archive repositories")
}

// Pull is not supported for archive repositories.
func (r *Repository) Pull(ctx context.Context, options ports.PullOptions) error {
	return errors.New("pull is not supported for archive repositories")
}

// Push is not supported for archive repositories.
func (r *Repository) Push(ctx context.Context, options ports.PushOptions) error {
	return errors.New("push is not supported for archive repositories")
}

// Add is not supported for archive repositories.
func (r *Repository) Add(ctx context.Context, files []string) error {
	return errors.New("adding files is not supported for archive repositories")
}

// Commit is not supported for archive repositories.
func (r *Repository) Commit(ctx context.Context, message string) error {
	return errors.New("committing is not supported for archive repositories")
}

// GetConfig returns the git configuration for this repository.
func (r *Repository) GetConfig() ports.GitConfig {
	return r.config
}

// GetRemoteURL returns an empty string as archive repositories don't have remotes.
func (r *Repository) GetRemoteURL(ctx context.Context, remoteName string) (string, error) {
	return "", errors.New("archive repositories don't have remotes")
}

// SetRemoteURL is not supported for archive repositories.
func (r *Repository) SetRemoteURL(ctx context.Context, remoteName, url string) error {
	return errors.New("setting remote URLs is not supported for archive repositories")
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
func (r *Repository) HasUncommittedChanges(ctx context.Context) (bool, error) {
	return false, nil
}

// GetFileContent reads the content of a file in the archive directory.
func (r *Repository) GetFileContent(ctx context.Context, filePath string) ([]byte, error) {
	fullPath := filepath.Join(r.path, filePath)

	// Security check: ensure the path is within the repository
	absRepoPath, err := filepath.Abs(r.path)
	if err != nil {
		return nil, err
	}

	absFilePath, err := filepath.Abs(fullPath)
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(absFilePath, absRepoPath+string(filepath.Separator)) && absFilePath != absRepoPath {
		return nil, errors.New("file path is outside repository")
	}

	// #nosec G304 - Path is validated above for security
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// WriteFile writes content to a file in the archive directory.
func (r *Repository) WriteFile(ctx context.Context, filePath string, content []byte) error {
	fullPath := filepath.Join(r.path, filePath)

	// Security check: ensure the path is within the repository
	absRepoPath, err := filepath.Abs(r.path)
	if err != nil {
		return err
	}

	absFilePath, err := filepath.Abs(fullPath)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(absFilePath, absRepoPath+string(filepath.Separator)) && absFilePath != absRepoPath {
		return errors.New("file path is outside repository")
	}

	// Ensure parent directory exists
	err = os.MkdirAll(filepath.Dir(fullPath), 0750)
	if err != nil {
		return err
	}

	if err := os.WriteFile(fullPath, content, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ListFiles returns a list of files in the archive directory.
func (r *Repository) ListFiles(ctx context.Context, path string) ([]string, error) {
	searchPath := r.path
	if path != "" {
		searchPath = filepath.Join(r.path, path)
	}

	var files []string

	err := filepath.Walk(searchPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relPath, err := filepath.Rel(r.path, filePath)
			if err != nil {
				return err
			}

			files = append(files, relPath)
		}

		return nil
	})

	return files, err
}

// Close performs any necessary cleanup for the repository.
func (r *Repository) Close() error {
	// Archive repositories don't need special cleanup
	return nil
}
