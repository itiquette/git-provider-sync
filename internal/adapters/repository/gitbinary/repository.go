// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package gitbinary

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// GitRepository implements the ports.GitRepository interface using git binary.
type GitRepository struct {
	path    string
	adapter *Adapter
}

// NewGitRepository creates a new git repository wrapper.
func NewGitRepository(path string, adapter *Adapter) ports.GitRepository { //nolint:ireturn
	return &GitRepository{
		path:    path,
		adapter: adapter,
	}
}

// GitRepositoryInfo interface implementation

// Path returns the repository path.
func (r *GitRepository) Path() string {
	return r.path
}

// URL returns the repository origin URL.
func (r *GitRepository) URL() string {
	// Get origin remote URL
	remotes, err := r.ListRemotes(context.Background())
	if err != nil || len(remotes) == 0 {
		return ""
	}

	for _, remote := range remotes {
		if remote.Name == "origin" {
			return remote.URL
		}
	}

	return remotes[0].URL
}

// Name returns the repository name.
func (r *GitRepository) Name() string {
	return filepath.Base(r.path)
}

// IsBare returns whether the repository is bare.
func (r *GitRepository) IsBare() bool {
	// Check if this is a bare repository
	return r.adapter.mirrorSvc.executorSvc.RunGitCommand(context.Background(), []string{}, r.path, "rev-parse", "--is-bare-repository") == nil
}

// IsClean returns true if the repository working directory is clean.
func (r *GitRepository) IsClean() bool {
	status, err := r.Status(context.Background())
	if err != nil {
		return false
	}

	return status.IsClean
}

// HasChanges returns true if the repository has uncommitted changes.
func (r *GitRepository) HasChanges() bool {
	return !r.IsClean()
}

// Close closes the repository and cleans up resources.
func (r *GitRepository) Close() error {
	// Git binary doesn't need explicit closing
	return nil
}

// GitBranchOperations interface implementation

// CurrentBranch returns the name of the current branch.
func (r *GitRepository) CurrentBranch() (string, error) {
	// would need implementation in operations service
	branch, err := r.adapter.mirrorSvc.operationsSvc.GetCurrentBranch(context.Background(), r.path)
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	return branch, nil
}

// ListBranches lists all branches in the repository.
func (r *GitRepository) ListBranches(ctx context.Context) ([]ports.BranchInfo, error) {
	branches, err := r.adapter.mirrorSvc.operationsSvc.GetBranches(ctx, r.path)
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}

	result := make([]ports.BranchInfo, len(branches))
	for i, branch := range branches {
		result[i] = ports.BranchInfo{
			Name:      branch,
			IsRemote:  strings.HasPrefix(branch, "origin/"),
			IsCurrent: false, // Would need to implement current branch detection
		}
	}

	return result, nil
}

// CreateBranch creates a new branch with the given name.
func (r *GitRepository) CreateBranch(ctx context.Context, name, _ string) error {
	if err := r.adapter.mirrorSvc.operationsSvc.CreateBranch(ctx, r.path, name); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	return nil
}

// CheckoutBranch checks out the specified branch.
func (r *GitRepository) CheckoutBranch(ctx context.Context, name string) error {
	if err := r.adapter.mirrorSvc.executorSvc.RunGitCommand(ctx, []string{}, r.path, "checkout", name); err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	return nil
}

// DeleteBranch deletes the specified branch.
func (r *GitRepository) DeleteBranch(ctx context.Context, name string, force bool) error {
	if err := r.adapter.mirrorSvc.operationsSvc.DeleteBranch(ctx, r.path, name, force); err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	return nil
}

// SetDefaultBranch sets the default branch for the repository.
func (r *GitRepository) SetDefaultBranch(ctx context.Context, name string) error {
	if err := r.adapter.mirrorSvc.executorSvc.RunGitCommand(ctx, []string{}, r.path, "symbolic-ref", "HEAD", "refs/heads/"+name); err != nil {
		return fmt.Errorf("failed to set default branch: %w", err)
	}

	return nil
}

// GitRemoteOperations interface implementation

// ListRemotes lists all remote repositories.
func (r *GitRepository) ListRemotes(ctx context.Context) ([]ports.RemoteInfo, error) {
	remotes, err := r.adapter.mirrorSvc.operationsSvc.GetRemotes(ctx, r.path)
	if err != nil {
		return nil, fmt.Errorf("failed to get remotes: %w", err)
	}

	result := make([]ports.RemoteInfo, len(remotes))
	for i, remote := range remotes {
		result[i] = ports.RemoteInfo{
			Name:     remote.Name,
			URL:      remote.URL,
			FetchURL: remote.URL,
			PushURL:  remote.URL,
		}
	}

	return result, nil
}

// AddRemote adds a new remote repository.
func (r *GitRepository) AddRemote(ctx context.Context, name, url string) error {
	if err := r.adapter.mirrorSvc.operationsSvc.AddRemote(ctx, r.path, name, url); err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}

	return nil
}

// RemoveRemote removes a remote repository.
func (r *GitRepository) RemoveRemote(ctx context.Context, name string) error {
	if err := r.adapter.mirrorSvc.operationsSvc.RemoveRemote(ctx, r.path, name); err != nil {
		return fmt.Errorf("failed to remove remote: %w", err)
	}

	return nil
}

// UpdateRemote updates a remote repository URL.
func (r *GitRepository) UpdateRemote(ctx context.Context, name, url string) error {
	// Remove and re-add the remote
	if err := r.RemoveRemote(ctx, name); err != nil {
		return err
	}

	return r.AddRemote(ctx, name, url)
}

// GitSyncOperations interface implementation

// Fetch fetches updates from remote repository.
func (r *GitRepository) Fetch(ctx context.Context, _ ports.FetchOptions) error {
	// Use operations service for fetch
	if err := r.adapter.mirrorSvc.operationsSvc.Fetch(ctx, r.path); err != nil {
		return fmt.Errorf("failed to fetch: %w", err)
	}

	return nil
}

// Pull fetches updates from the remote and merges them into the local branch.
func (r *GitRepository) Pull(ctx context.Context, options ports.PullOptions) error {
	config := MirrorConfig{
		AuthConfig: r.adapter.convertAuthOptions(options.Auth),
	}

	return r.adapter.mirrorSvc.Pull(ctx, r.path, config)
}

// Push pushes local commits to the remote repository.
func (r *GitRepository) Push(ctx context.Context, options ports.PushOptions) error {
	config := MirrorConfig{
		ForcePush:  options.Force,
		AuthConfig: r.adapter.convertAuthOptions(options.Auth),
	}

	// Create minimal repository entity for push (placeholder)
	repo := r.adapter.mirrorSvc.createRepositoryEntity(ctx, r.path, config)

	return r.adapter.mirrorSvc.Push(ctx, repo, config)
}

// GitCommitOperations interface implementation

// GetCommit retrieves commit information for a specific reference.
func (r *GitRepository) GetCommit(_ context.Context, _ string) (ports.CommitInfo, error) {
	// would need implementation in operations service
	return ports.CommitInfo{}, domain.ErrNotYetImplemented
}

// ListCommits lists commits in the repository.
func (r *GitRepository) ListCommits(_ context.Context, _ ports.ListCommitsOptions) ([]ports.CommitInfo, error) {
	// would need implementation in operations service
	return []ports.CommitInfo{}, domain.ErrNotYetImplemented
}

// GitTagOperations interface implementation

// ListTags returns all tags in the repository.
func (r *GitRepository) ListTags(ctx context.Context) ([]ports.TagInfo, error) {
	tags, err := r.adapter.mirrorSvc.operationsSvc.GetTags(ctx, r.path)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	result := make([]ports.TagInfo, len(tags))
	for i, tag := range tags {
		result[i] = ports.TagInfo{
			Name: tag,
			// Hash, Message, Timestamp, Tagger would need additional git commands
		}
	}

	return result, nil
}

// CreateTag creates a new tag in the repository.
func (r *GitRepository) CreateTag(ctx context.Context, name, message, ref string) error {
	args := []string{"tag"}
	if message != "" {
		args = append(args, "-m", message)
	}

	args = append(args, name)
	if ref != "" {
		args = append(args, ref)
	}

	if err := r.adapter.mirrorSvc.executorSvc.RunGitCommand(ctx, []string{}, r.path, args...); err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	return nil
}

// DeleteTag deletes a tag from the repository.
func (r *GitRepository) DeleteTag(ctx context.Context, name string) error {
	if err := r.adapter.mirrorSvc.executorSvc.RunGitCommand(ctx, []string{}, r.path, "tag", "-d", name); err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	return nil
}

// GitStatusOperations interface implementation

// Status returns the current status of the repository working tree.
func (r *GitRepository) Status(ctx context.Context) (ports.StatusResult, error) {
	status, err := r.adapter.mirrorSvc.operationsSvc.GetStatus(ctx, r.path)
	if err != nil {
		return ports.StatusResult{}, fmt.Errorf("failed to get status: %w", err)
	}

	return status, nil
}

// Diff returns the diff between two commits.
func (r *GitRepository) Diff(_ context.Context, options ports.DiffOptions) (string, error) {
	// would need to capture output from git command
	// For now, return placeholder
	_ = options // Mark as used to satisfy linter

	return "", domain.ErrNotYetImplemented
}

var _ ports.GitRepository = (*GitRepository)(nil)
