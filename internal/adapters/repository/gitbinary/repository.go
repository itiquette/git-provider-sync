// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package gitbinary

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// GitRepository implements the ports.GitRepository interface using git binary.
type GitRepository struct {
	path    string
	adapter *Adapter
}

// NewGitRepository creates a new git repository wrapper.
func NewGitRepository(path string, adapter *Adapter) ports.GitRepository {
	return &GitRepository{
		path:    path,
		adapter: adapter,
	}
}

// GitRepositoryInfo interface implementation

func (r *GitRepository) Path() string {
	return r.path
}

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

func (r *GitRepository) Name() string {
	return filepath.Base(r.path)
}

func (r *GitRepository) IsBare() bool {
	// Check if this is a bare repository
	return r.adapter.mirrorSvc.executorSvc.RunGitCommand(context.Background(), []string{}, r.path, "rev-parse", "--is-bare-repository") == nil
}

func (r *GitRepository) IsClean() bool {
	status, err := r.Status(context.Background())
	if err != nil {
		return false
	}

	return status.IsClean
}

func (r *GitRepository) HasChanges() bool {
	return !r.IsClean()
}

func (r *GitRepository) Close() error {
	// Git binary doesn't need explicit closing
	return nil
}

// GitBranchOperations interface implementation

func (r *GitRepository) CurrentBranch() (string, error) {
	// This would need implementation in operations service
	return r.adapter.mirrorSvc.operationsSvc.GetCurrentBranch(context.Background(), r.path)
}

func (r *GitRepository) ListBranches(ctx context.Context) ([]ports.BranchInfo, error) {
	branches, err := r.adapter.mirrorSvc.operationsSvc.GetBranches(ctx, r.path)
	if err != nil {
		return nil, err
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

func (r *GitRepository) CreateBranch(ctx context.Context, name, source string) error {
	return r.adapter.mirrorSvc.operationsSvc.CreateBranch(ctx, r.path, name)
}

func (r *GitRepository) CheckoutBranch(ctx context.Context, name string) error {
	return r.adapter.mirrorSvc.executorSvc.RunGitCommand(ctx, []string{}, r.path, "checkout", name)
}

func (r *GitRepository) DeleteBranch(ctx context.Context, name string, force bool) error {
	return r.adapter.mirrorSvc.operationsSvc.DeleteBranch(ctx, r.path, name, force)
}

func (r *GitRepository) SetDefaultBranch(ctx context.Context, name string) error {
	return r.adapter.mirrorSvc.executorSvc.RunGitCommand(ctx, []string{}, r.path, "symbolic-ref", "HEAD", "refs/heads/"+name)
}

// GitRemoteOperations interface implementation

func (r *GitRepository) ListRemotes(ctx context.Context) ([]ports.RemoteInfo, error) {
	remotes, err := r.adapter.mirrorSvc.operationsSvc.GetRemotes(ctx, r.path)
	if err != nil {
		return nil, err
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

func (r *GitRepository) AddRemote(ctx context.Context, name, url string) error {
	return r.adapter.mirrorSvc.operationsSvc.AddRemote(ctx, r.path, name, url)
}

func (r *GitRepository) RemoveRemote(ctx context.Context, name string) error {
	return r.adapter.mirrorSvc.operationsSvc.RemoveRemote(ctx, r.path, name)
}

func (r *GitRepository) UpdateRemote(ctx context.Context, name, url string) error {
	// Remove and re-add the remote
	if err := r.RemoveRemote(ctx, name); err != nil {
		return err
	}

	return r.AddRemote(ctx, name, url)
}

// GitSyncOperations interface implementation

func (r *GitRepository) Fetch(ctx context.Context, options ports.FetchOptions) error {
	// Use operations service for fetch
	return r.adapter.mirrorSvc.operationsSvc.Fetch(ctx, r.path)
}

func (r *GitRepository) Pull(ctx context.Context, options ports.PullOptions) error {
	config := MirrorConfig{
		AuthConfig: r.adapter.convertAuthOptions(options.Auth),
	}

	return r.adapter.mirrorSvc.Pull(ctx, r.path, config)
}

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

func (r *GitRepository) GetCommit(ctx context.Context, ref string) (ports.CommitInfo, error) {
	// This would need implementation in operations service
	return ports.CommitInfo{}, fmt.Errorf("GetCommit not yet implemented for git binary")
}

func (r *GitRepository) ListCommits(ctx context.Context, options ports.ListCommitsOptions) ([]ports.CommitInfo, error) {
	// This would need implementation in operations service
	return []ports.CommitInfo{}, fmt.Errorf("ListCommits not yet implemented for git binary")
}

// GitTagOperations interface implementation

func (r *GitRepository) ListTags(ctx context.Context) ([]ports.TagInfo, error) {
	tags, err := r.adapter.mirrorSvc.operationsSvc.GetTags(ctx, r.path)
	if err != nil {
		return nil, err
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

func (r *GitRepository) CreateTag(ctx context.Context, name, message, ref string) error {
	args := []string{"tag"}
	if message != "" {
		args = append(args, "-m", message)
	}
	args = append(args, name)
	if ref != "" {
		args = append(args, ref)
	}

	return r.adapter.mirrorSvc.executorSvc.RunGitCommand(ctx, []string{}, r.path, args...)
}

func (r *GitRepository) DeleteTag(ctx context.Context, name string) error {
	return r.adapter.mirrorSvc.executorSvc.RunGitCommand(ctx, []string{}, r.path, "tag", "-d", name)
}

// GitStatusOperations interface implementation

func (r *GitRepository) Status(ctx context.Context) (ports.StatusResult, error) {
	return r.adapter.mirrorSvc.operationsSvc.GetStatus(ctx, r.path)
}

func (r *GitRepository) Diff(ctx context.Context, options ports.DiffOptions) (string, error) {
	// This would need to capture output from git command
	// For now, return placeholder
	_ = options // Mark as used to satisfy linter
	return "", fmt.Errorf("Diff not yet implemented for git binary")
}

// Ensure GitRepository implements ports.GitRepository interface
var _ ports.GitRepository = (*GitRepository)(nil)
