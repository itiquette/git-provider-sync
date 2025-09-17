// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package gogit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage"
	gitfilesystem "github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-git/go-git/v5/storage/memory"

	"itiquette/git-provider-sync/internal/adapters/filesystem"
	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// Adapter implements the GitOperations interface using go-git.
type Adapter struct {
	config     ports.GitConfig
	fileSystem ports.FileSystem
}

// New creates a new go-git adapter.
func New(config ports.GitConfig) *Adapter {
	return &Adapter{
		config:     config,
		fileSystem: filesystem.NewOSFileSystem(),
	}
}

// IsGitDir checks if a directory contains git repository files (like config, HEAD, etc.)
func isGitDir(path string) bool {
	// Check for common git directory markers
	markers := []string{"config", "HEAD", "objects", "refs"}
	for _, marker := range markers {
		if _, err := os.Stat(filepath.Join(path, marker)); err != nil {
			return false
		}
	}

	return true
}

// CreateStorer creates an appropriate storer based on the configuration.
func (a *Adapter) createStorer(path string) storage.Storer {
	switch a.config.StorageMode {
	case ports.StorageModeFilesystem:
		fs := osfs.New(path)

		return gitfilesystem.NewStorage(fs, nil)
	case ports.StorageModeMemory:
		fallthrough
	default:
		return memory.NewStorage()
	}
}

// CreateFilesystem creates an appropriate filesystem based on the configuration.
func (a *Adapter) createFilesystem(path string) billy.Filesystem {
	switch a.config.StorageMode {
	case ports.StorageModeFilesystem:
		return osfs.New(path)
	case ports.StorageModeMemory:
		fallthrough
	default:
		return memfs.New()
	}
}

// Clone clones a repository using configurable storage.
func (a *Adapter) Clone(ctx context.Context, options ports.CloneOptions) (ports.GitRepository, error) {
	auth, err := a.buildAuth(options.Auth)
	if err != nil {
		return nil, fmt.Errorf("failed to build auth: %w", err)
	}

	cloneOptions := &git.CloneOptions{
		URL:           options.URL,
		SingleBranch:  options.SingleBranch,
		Depth:         options.Depth,
		Auth:          auth,
		Progress:      nil, // Progress callback would need interface adaptation
		Tags:          a.convertTagMode(options.Tags),
		ReferenceName: a.convertBranch(options.Branch),
	}

	if options.Mirror {
		cloneOptions.Mirror = true
	}

	var (
		storer storage.Storer
		repo   *git.Repository
	)

	if options.Bare {
		// For bare repositories, don't use a worktree
		storer = a.createStorer(options.Path)
		repo, err = git.CloneContext(ctx, storer, nil, cloneOptions)
	} else {
		// For non-bare repositories, store git data in .git subdirectory
		gitPath := filepath.Join(options.Path, ".git")
		storer = a.createStorer(gitPath)
		worktree := a.createFilesystem(options.Path)
		repo, err = git.CloneContext(ctx, storer, worktree, cloneOptions)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	return &Repository{
		repo: repo,
		path: options.Path,
		url:  options.URL,
	}, nil
}

// Open opens an existing repository using configurable storage.
func (a *Adapter) Open(_ context.Context, path string) (ports.GitRepository, error) {
	// For filesystem mode, try to open existing repository
	if a.config.StorageMode == ports.StorageModeFilesystem {
		return a.openFilesystemRepository(path)
	}

	// For in-memory mode (default), create new repository since we can't "open" from path
	return a.createInMemoryRepository(path)
}

// OpenFilesystemRepository opens a repository from filesystem storage.
func (a *Adapter) openFilesystemRepository(path string) (*Repository, error) {
	storer, worktree := a.determineStorageLayout(path)

	repo, err := git.Open(storer, worktree)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	url := a.extractRemoteURL(repo)

	return &Repository{
		repo: repo,
		path: path,
		url:  url,
	}, nil
}

// DetermineStorageLayout determines the correct storer and worktree based on repository type.
func (a *Adapter) determineStorageLayout(path string) (storage.Storer, billy.Filesystem) {
	// First try to open as a non-bare repository with .git subdirectory
	gitPath := filepath.Join(path, ".git")
	if _, err := os.Stat(gitPath); err == nil {
		// Non-bare repository with .git subdirectory
		return a.createStorer(gitPath), a.createFilesystem(path)
	}

	if strings.HasSuffix(path, ".git") || isGitDir(path) {
		// Bare repository
		return a.createStorer(path), nil
	}

	// Try as a bare repository in the given path
	return a.createStorer(path), a.createFilesystem(path)
}

// ExtractRemoteURL extracts the remote URL from the repository.
func (a *Adapter) extractRemoteURL(repo *git.Repository) string {
	remotes, err := repo.Remotes()
	if err != nil || len(remotes) == 0 {
		return ""
	}

	urls := remotes[0].Config().URLs
	if len(urls) == 0 {
		return ""
	}

	return urls[0]
}

// CreateInMemoryRepository creates a new in-memory repository.
func (a *Adapter) createInMemoryRepository(path string) (*Repository, error) {
	storer := a.createStorer(path)
	worktree := a.createFilesystem(path)

	repo, err := git.Init(storer, worktree)
	if err != nil {
		return nil, fmt.Errorf("failed to create in-memory repository: %w", err)
	}

	return &Repository{
		repo: repo,
		path: path,
		url:  "",
	}, nil
}

// Init initializes a new repository using configurable storage.
func (a *Adapter) Init(_ context.Context, path string, options ports.InitOptions) (ports.GitRepository, error) {
	var (
		storer storage.Storer
		repo   *git.Repository
		err    error
	)

	if options.Bare {
		// For bare repositories, don't use a worktree
		storer = a.createStorer(path)
		repo, err = git.Init(storer, nil)
	} else {
		// For non-bare repositories, store git data in .git subdirectory
		gitPath := filepath.Join(path, ".git")
		storer = a.createStorer(gitPath)
		worktree := a.createFilesystem(path)
		repo, err = git.Init(storer, worktree)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize repository: %w", err)
	}

	// Set default branch if specified
	if options.DefaultBranch != "" {
		headRef := plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.ReferenceName("refs/heads/"+options.DefaultBranch))

		err = repo.Storer.SetReference(headRef)
		if err != nil {
			return nil, fmt.Errorf("failed to set default branch: %w", err)
		}
	}

	return &Repository{
		repo: repo,
		path: path,
	}, nil
}

// SupportsURL checks if go-git supports the given URL.
func (a *Adapter) SupportsURL(url string) bool {
	// Go-git supports HTTP, HTTPS, SSH, and file URLs
	return strings.HasPrefix(url, "http://") ||
		strings.HasPrefix(url, "https://") ||
		strings.HasPrefix(url, "git@") ||
		strings.HasPrefix(url, "ssh://") ||
		strings.HasPrefix(url, "file://") ||
		strings.HasPrefix(url, "/")
}

// GetName returns the name of this git operations implementation.
func (a *Adapter) GetName() string {
	return "go-git"
}

// Cleanup cleans up resources at the given path.
func (a *Adapter) Cleanup(_ context.Context, _ string) error {
	// For go-git, cleanup mainly involves removing temporary directories
	// Actual implementation would depend on what needs to be cleaned up
	return nil
}

// CreateTmpDir implements the ports.GitOperations interface.
func (a *Adapter) CreateTmpDir(ctx context.Context, dir, prefix string) (context.Context, error) {
	ctxWithTmp, err := filesystem.CreateTmpDir(ctx, a.fileSystem, dir, prefix)
	if err != nil {
		return ctx, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	return ctxWithTmp, nil
}

// GetTmpDirPath implements the ports.GitOperations interface.
func (a *Adapter) GetTmpDirPath(ctx context.Context) (string, error) {
	path, err := filesystem.GetTmpDirPath(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get temporary directory path: %w", err)
	}

	return path, nil
}

// DeleteTmpDir implements the ports.GitOperations interface.
func (a *Adapter) DeleteTmpDir(ctx context.Context) error {
	if err := filesystem.DeleteTmpDir(ctx, a.fileSystem); err != nil {
		return fmt.Errorf("failed to delete temporary directory: %w", err)
	}

	return nil
}

// Repository implements the GitRepository interface using go-git.
type Repository struct {
	repo *git.Repository
	path string
	url  string
}

// Path returns the local path of the repository.
func (r *Repository) Path() string {
	return r.path
}

// URL returns the remote URL of the repository.
func (r *Repository) URL() string {
	return r.url
}

// Name returns the name of the repository (basename of path).
func (r *Repository) Name() string {
	return filepath.Base(r.path)
}

// IsBare returns true if the repository is bare.
func (r *Repository) IsBare() bool {
	worktree, err := r.repo.Worktree()

	return err != nil || worktree == nil
}

// IsClean returns true if the repository has no uncommitted changes.
func (r *Repository) IsClean() bool {
	if r.IsBare() {
		return true
	}

	worktree, err := r.repo.Worktree()
	if err != nil {
		return false
	}

	status, err := worktree.Status()
	if err != nil {
		return false
	}

	return status.IsClean()
}

// HasChanges returns true if the repository has uncommitted changes.
func (r *Repository) HasChanges() bool {
	return !r.IsClean()
}

// CurrentBranch returns the name of the current branch.
func (r *Repository) CurrentBranch() (string, error) {
	head, err := r.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	if head.Name().IsBranch() {
		return head.Name().Short(), nil
	}

	return head.Hash().String()[:7], nil
}

// ListBranches lists all branches in the repository.
func (r *Repository) ListBranches(_ context.Context) ([]ports.BranchInfo, error) {
	refs, err := r.repo.References()
	if err != nil {
		return nil, fmt.Errorf("failed to list references: %w", err)
	}

	var branches []ports.BranchInfo

	currentBranch, _ := r.CurrentBranch()

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsBranch() || ref.Name().IsRemote() {
			branch := ports.BranchInfo{
				Name:      ref.Name().Short(),
				Hash:      ref.Hash().String(),
				IsRemote:  ref.Name().IsRemote(),
				IsCurrent: ref.Name().Short() == currentBranch,
			}
			branches = append(branches, branch)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate references: %w", err)
	}

	return branches, nil
}

// CreateBranch creates a new branch.
func (r *Repository) CreateBranch(_ context.Context, name, source string) error {
	var sourceRef *plumbing.Reference

	var err error

	if source != "" {
		sourceRef, err = r.repo.Reference(plumbing.ReferenceName("refs/heads/"+source), true)
		if err != nil {
			return fmt.Errorf("failed to find source branch %s: %w", source, err)
		}
	} else {
		sourceRef, err = r.repo.Head()
		if err != nil {
			return fmt.Errorf("failed to get HEAD: %w", err)
		}
	}

	newRef := plumbing.NewHashReference(plumbing.ReferenceName("refs/heads/"+name), sourceRef.Hash())

	err = r.repo.Storer.SetReference(newRef)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	return nil
}

// CheckoutBranch checks out a branch.
func (r *Repository) CheckoutBranch(_ context.Context, name string) error {
	if r.IsBare() {
		return domain.ErrCannotCheckoutBare
	}

	worktree, err := r.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/heads/" + name),
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	return nil
}

// DeleteBranch deletes a branch.
func (r *Repository) DeleteBranch(_ context.Context, name string, _ bool) error {
	ref := plumbing.ReferenceName("refs/heads/" + name)

	err := r.repo.Storer.RemoveReference(ref)
	if err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	return nil
}

// SetDefaultBranch sets the default branch.
func (r *Repository) SetDefaultBranch(_ context.Context, name string) error {
	headRef := plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.ReferenceName("refs/heads/"+name))

	err := r.repo.Storer.SetReference(headRef)
	if err != nil {
		return fmt.Errorf("failed to set default branch: %w", err)
	}

	return nil
}

// ListRemotes lists all remotes.
func (r *Repository) ListRemotes(_ context.Context) ([]ports.RemoteInfo, error) {
	remotes, err := r.repo.Remotes()
	if err != nil {
		return nil, fmt.Errorf("failed to list remotes: %w", err)
	}

	remoteInfos := make([]ports.RemoteInfo, 0, len(remotes))

	for _, remote := range remotes {
		config := remote.Config()
		info := ports.RemoteInfo{
			Name: config.Name,
		}

		if len(config.URLs) > 0 {
			info.URL = config.URLs[0]
			info.FetchURL = config.URLs[0]
			info.PushURL = config.URLs[0]
		}

		remoteInfos = append(remoteInfos, info)
	}

	return remoteInfos, nil
}

// AddRemote adds a new remote.
func (r *Repository) AddRemote(_ context.Context, name, url string) error {
	_, err := r.repo.CreateRemote(&config.RemoteConfig{
		Name: name,
		URLs: []string{url},
	})
	if err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}

	return nil
}

// RemoveRemote removes a remote.
func (r *Repository) RemoveRemote(_ context.Context, name string) error {
	err := r.repo.DeleteRemote(name)
	if err != nil {
		return fmt.Errorf("failed to remove remote: %w", err)
	}

	return nil
}

// UpdateRemote updates a remote URL.
func (r *Repository) UpdateRemote(_ context.Context, name, url string) error {
	err := r.repo.DeleteRemote(name)
	if err != nil {
		return fmt.Errorf("failed to remove existing remote: %w", err)
	}

	_, err = r.repo.CreateRemote(&config.RemoteConfig{
		Name: name,
		URLs: []string{url},
	})
	if err != nil {
		return fmt.Errorf("failed to recreate remote: %w", err)
	}

	return nil
}

// Fetch fetches from a remote.
func (r *Repository) Fetch(ctx context.Context, options ports.FetchOptions) error {
	auth, err := r.buildAuth(options.Auth)
	if err != nil {
		return fmt.Errorf("failed to build auth: %w", err)
	}

	fetchOptions := &git.FetchOptions{
		RemoteName: options.Remote,
		Auth:       auth,
		Progress:   nil, // Progress callback would need interface adaptation
		Depth:      options.Depth,
		Force:      options.Force,
	}

	if len(options.RefSpecs) > 0 {
		for _, refSpec := range options.RefSpecs {
			fetchOptions.RefSpecs = append(fetchOptions.RefSpecs, config.RefSpec(refSpec))
		}
	}

	err = r.repo.FetchContext(ctx, fetchOptions)
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("failed to fetch: %w", err)
	}

	return nil
}

// Pull pulls from a remote.
func (r *Repository) Pull(ctx context.Context, options ports.PullOptions) error {
	if r.IsBare() {
		return domain.ErrCannotPullBare
	}

	worktree, err := r.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	auth, err := r.buildAuth(options.Auth)
	if err != nil {
		return fmt.Errorf("failed to build auth: %w", err)
	}

	pullOptions := &git.PullOptions{
		RemoteName: options.Remote,
		Auth:       auth,
		Progress:   nil, // Progress callback would need interface adaptation
	}

	if options.Branch != "" {
		pullOptions.ReferenceName = plumbing.ReferenceName("refs/heads/" + options.Branch)
	}

	err = worktree.PullContext(ctx, pullOptions)
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("failed to pull: %w", err)
	}

	return nil
}

// Push pushes to a remote.
func (r *Repository) Push(ctx context.Context, options ports.PushOptions) error {
	auth, err := r.buildAuth(options.Auth)
	if err != nil {
		return fmt.Errorf("failed to build auth: %w", err)
	}

	pushOptions := &git.PushOptions{
		RemoteName: options.Remote,
		Auth:       auth,
		Progress:   nil, // Progress callback would need interface adaptation
		Force:      options.Force,
		Atomic:     options.Atomic,
	}

	if len(options.RefSpecs) > 0 {
		for _, refSpec := range options.RefSpecs {
			pushOptions.RefSpecs = append(pushOptions.RefSpecs, config.RefSpec(refSpec))
		}
	}

	err = r.repo.PushContext(ctx, pushOptions)
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("failed to push: %w", err)
	}

	return nil
}

// GetCommit gets a commit by reference.
func (r *Repository) GetCommit(_ context.Context, ref string) (ports.CommitInfo, error) {
	hash, err := r.repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return ports.CommitInfo{}, fmt.Errorf("failed to resolve revision: %w", err)
	}

	commit, err := r.repo.CommitObject(*hash)
	if err != nil {
		return ports.CommitInfo{}, fmt.Errorf("failed to get commit: %w", err)
	}

	return r.convertCommit(commit), nil
}

// ListCommits lists commits based on options
//
//nolint:cyclop // Complex commit listing logic with multiple filtering options
func (r *Repository) ListCommits(_ context.Context, options ports.ListCommitsOptions) ([]ports.CommitInfo, error) {
	var startRef plumbing.Hash

	var err error

	if options.From != "" {
		startHash, resolveErr := r.repo.ResolveRevision(plumbing.Revision(options.From))
		if resolveErr != nil {
			return nil, fmt.Errorf("failed to resolve from revision: %w", resolveErr)
		}

		startRef = *startHash
	} else {
		head, headErr := r.repo.Head()
		if headErr != nil {
			return nil, fmt.Errorf("failed to get HEAD: %w", headErr)
		}

		startRef = head.Hash()
	}

	commitIter, err := r.repo.Log(&git.LogOptions{
		From: startRef,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}
	defer commitIter.Close()

	var commits []ports.CommitInfo

	count := 0

	err = commitIter.ForEach(func(commit *object.Commit) error {
		if options.MaxCount > 0 && count >= options.MaxCount {
			return domain.ErrMaxCountReached
		}

		// Apply filters
		if options.Since != nil && commit.Committer.When.Before(*options.Since) {
			return nil
		}

		if options.Until != nil && commit.Committer.When.After(*options.Until) {
			return nil
		}

		if options.Author != "" && !strings.Contains(commit.Author.Name, options.Author) {
			return nil
		}

		if options.Message != "" && !strings.Contains(commit.Message, options.Message) {
			return nil
		}

		commits = append(commits, r.convertCommit(commit))
		count++

		return nil
	})
	if err != nil && !strings.Contains(err.Error(), "reached max count") {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return commits, nil
}

// ListTags lists all tags.
func (r *Repository) ListTags(_ context.Context) ([]ports.TagInfo, error) {
	tagRefs, err := r.repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	var tags []ports.TagInfo

	err = tagRefs.ForEach(func(ref *plumbing.Reference) error {
		tag := ports.TagInfo{
			Name: ref.Name().Short(),
			Hash: ref.Hash().String(),
		}

		// Try to get annotated tag object
		tagObj, err := r.repo.TagObject(ref.Hash())
		if err == nil {
			tag.Message = tagObj.Message
			tag.Timestamp = tagObj.Tagger.When
			tag.Tagger = ports.PersonInfo{
				Name:  tagObj.Tagger.Name,
				Email: tagObj.Tagger.Email,
				When:  tagObj.Tagger.When,
			}
		} else {
			// Lightweight tag - get commit info
			commit, commitErr := r.repo.CommitObject(ref.Hash())
			if commitErr == nil {
				tag.Timestamp = commit.Committer.When
			}
		}

		tags = append(tags, tag)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate tags: %w", err)
	}

	return tags, nil
}

// CreateTag creates a new tag.
func (r *Repository) CreateTag(_ context.Context, name, _, ref string) error {
	var hash plumbing.Hash

	var err error

	if ref != "" {
		resolvedHash, err := r.repo.ResolveRevision(plumbing.Revision(ref))
		if err != nil {
			return fmt.Errorf("failed to resolve reference: %w", err)
		}

		hash = *resolvedHash
	} else {
		head, err := r.repo.Head()
		if err != nil {
			return fmt.Errorf("failed to get HEAD: %w", err)
		}

		hash = head.Hash()
	}

	tagRef := plumbing.NewHashReference(plumbing.ReferenceName("refs/tags/"+name), hash)

	err = r.repo.Storer.SetReference(tagRef)
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	return nil
}

// DeleteTag deletes a tag.
func (r *Repository) DeleteTag(_ context.Context, name string) error {
	tagRef := plumbing.ReferenceName("refs/tags/" + name)

	err := r.repo.Storer.RemoveReference(tagRef)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	return nil
}

// Status returns the status of the working directory
//
//nolint:cyclop // Complex status checking logic with multiple file states
func (r *Repository) Status(_ context.Context) (ports.StatusResult, error) {
	if r.IsBare() {
		return ports.StatusResult{IsClean: true}, nil
	}

	worktree, err := r.repo.Worktree()
	if err != nil {
		return ports.StatusResult{}, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return ports.StatusResult{}, fmt.Errorf("failed to get status: %w", err)
	}

	result := ports.StatusResult{
		IsClean: status.IsClean(),
	}

	for file, stat := range status {
		switch stat.Staging {
		case git.Added:
			result.Added = append(result.Added, file)
		case git.Modified:
			result.Modified = append(result.Modified, file)
		case git.Deleted:
			result.Deleted = append(result.Deleted, file)
		case git.Renamed:
			result.Renamed = append(result.Renamed, file)
		case git.Untracked:
			result.Untracked = append(result.Untracked, file)
		case git.Unmodified, git.Copied, git.UpdatedButUnmerged:
		}

		if stat.Worktree == git.Modified && stat.Staging != git.Modified {
			result.Modified = append(result.Modified, file)
		}
	}

	return result, nil
}

// Diff generates a diff between two references.
func (r *Repository) Diff(_ context.Context, _ ports.DiffOptions) (string, error) {
	// For simplicity, return a basic diff implementation
	// A full implementation would use go-git's diff capabilities
	return "Diff functionality not yet implemented in go-git adapter", nil
}

// Close closes the repository.
func (r *Repository) Close() error {
	// Go-git repositories don't need explicit closing
	return nil
}

// Helper methods

func (a *Adapter) buildAuth(authOptions ports.AuthOptions) (transport.AuthMethod, error) { //nolint:ireturn
	switch authOptions.Type {
	case ports.AuthTypeNone:
		return &noAuth{}, nil
	case ports.AuthTypeSSH:
		// Generic SSH type - fallback to SSH key auth
		return a.buildSSHKeyAuth(authOptions)
	case ports.AuthTypeSSHAgent:
		return nil, domain.ErrSSHAgentNotSupported
	case ports.AuthTypeBasic:
		return a.buildBasicAuth(authOptions), nil
	case ports.AuthTypeToken:
		return a.buildTokenAuth(authOptions), nil
	case ports.AuthTypeSSHKey:
		return a.buildSSHKeyAuth(authOptions)
	default:
		return nil, fmt.Errorf("%w: %v", domain.ErrUnsupportedAuthType, authOptions.Type)
	}
}

// BuildBasicAuth creates basic authentication.
func (a *Adapter) buildBasicAuth(authOptions ports.AuthOptions) transport.AuthMethod { //nolint:ireturn
	return &http.BasicAuth{
		Username: authOptions.Username,
		Password: authOptions.Password,
	}
}

// BuildTokenAuth creates token-based authentication.
func (a *Adapter) buildTokenAuth(authOptions ports.AuthOptions) transport.AuthMethod { //nolint:ireturn
	return &http.BasicAuth{
		Username: authOptions.Username,
		Password: authOptions.Token,
	}
}

// BuildSSHKeyAuth creates SSH key authentication.
func (a *Adapter) buildSSHKeyAuth(authOptions ports.AuthOptions) (transport.AuthMethod, error) { //nolint:ireturn
	if authOptions.SSHKeyPath != "" {
		return a.buildSSHKeyFromFile(authOptions)
	}

	if len(authOptions.SSHKey) > 0 {
		return a.buildSSHKeyFromBytes(authOptions)
	}

	return nil, domain.ErrSSHKeyRequired
}

// BuildSSHKeyFromFile creates SSH authentication from a key file.
func (a *Adapter) buildSSHKeyFromFile(authOptions ports.AuthOptions) (transport.AuthMethod, error) { //nolint:ireturn
	publicKeys, err := ssh.NewPublicKeysFromFile(authOptions.Username, authOptions.SSHKeyPath, authOptions.Passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to load SSH key from file: %w", err)
	}

	return publicKeys, nil
}

// BuildSSHKeyFromBytes creates SSH authentication from key bytes.
func (a *Adapter) buildSSHKeyFromBytes(authOptions ports.AuthOptions) (transport.AuthMethod, error) { //nolint:ireturn
	publicKeys, err := ssh.NewPublicKeys(authOptions.Username, authOptions.SSHKey, authOptions.Passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to load SSH key from bytes: %w", err)
	}

	return publicKeys, nil
}

func (r *Repository) buildAuth(authOptions ports.AuthOptions) (transport.AuthMethod, error) { //nolint:ireturn
	adapter := &Adapter{}

	return adapter.buildAuth(authOptions)
}

func (a *Adapter) convertTagMode(mode ports.TagMode) git.TagMode {
	switch mode {
	case ports.TagModeDefault:
		return git.NoTags // Default to no tags
	case ports.TagModeAll:
		return git.AllTags
	case ports.TagModeNone:
		return git.NoTags
	case ports.TagModeFollowing:
		return git.TagFollowing
	case ports.TagModeFollow:
		return git.TagFollowing
	default:
		return git.NoTags
	}
}

func (a *Adapter) convertBranch(branch string) plumbing.ReferenceName {
	if branch == "" {
		return ""
	}

	return plumbing.ReferenceName("refs/heads/" + branch)
}

func (r *Repository) convertCommit(commit *object.Commit) ports.CommitInfo {
	parents := make([]string, 0, len(commit.ParentHashes))
	for _, parent := range commit.ParentHashes {
		parents = append(parents, parent.String())
	}

	return ports.CommitInfo{
		Hash:    commit.Hash.String(),
		Message: commit.Message,
		Author: ports.PersonInfo{
			Name:  commit.Author.Name,
			Email: commit.Author.Email,
			When:  commit.Author.When,
		},
		Committer: ports.PersonInfo{
			Name:  commit.Committer.Name,
			Email: commit.Committer.Email,
			When:  commit.Committer.When,
		},
		Parents:   parents,
		Timestamp: commit.Committer.When,
	}
}

// NoAuth implements transport.AuthMethod for cases where no authentication is needed.
type noAuth struct{}

// Name returns the name of the authentication method.
func (n *noAuth) Name() string {
	return "none"
}

// String returns a string representation of the auth method.
func (n *noAuth) String() string {
	return "no authentication"
}
