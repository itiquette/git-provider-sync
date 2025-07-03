// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package gogit

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/filesystem"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// Adapter implements the GitOperations interface using go-git.
type Adapter struct {
	config ports.GitConfig
}

// New creates a new go-git adapter.
func New(config ports.GitConfig) *Adapter {
	return &Adapter{
		config: config,
	}
}

// Clone clones a repository using go-git.
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

	var repo *git.Repository

	if options.Bare {
		// For bare repositories, clone to filesystem storage
		fs := osfs.New(options.Path)
		storage := filesystem.NewStorage(fs, nil)
		repo, err = git.CloneContext(ctx, storage, nil, cloneOptions)
	} else {
		repo, err = git.PlainCloneContext(ctx, options.Path, false, cloneOptions)
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

// Open opens an existing repository.
func (a *Adapter) Open(ctx context.Context, path string) (ports.GitRepository, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Try to get remote URL
	remotes, err := repo.Remotes()

	var url string

	if err == nil && len(remotes) > 0 {
		urls := remotes[0].Config().URLs
		if len(urls) > 0 {
			url = urls[0]
		}
	}

	return &Repository{
		repo: repo,
		path: path,
		url:  url,
	}, nil
}

// Init initializes a new repository.
func (a *Adapter) Init(ctx context.Context, path string, options ports.InitOptions) (ports.GitRepository, error) {
	var repo *git.Repository

	var err error

	if options.Bare {
		fs := osfs.New(path)
		storage := filesystem.NewStorage(fs, nil)
		repo, err = git.Init(storage, nil)
	} else {
		repo, err = git.PlainInit(path, false)
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
	// go-git supports HTTP, HTTPS, SSH, and file URLs
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
func (a *Adapter) Cleanup(ctx context.Context, path string) error {
	// For go-git, cleanup mainly involves removing temporary directories
	// The actual implementation would depend on what needs to be cleaned up
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
func (r *Repository) ListBranches(ctx context.Context) ([]ports.BranchInfo, error) {
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
func (r *Repository) CreateBranch(ctx context.Context, name, source string) error {
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
func (r *Repository) CheckoutBranch(ctx context.Context, name string) error {
	if r.IsBare() {
		return errors.New("cannot checkout branch in bare repository")
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
func (r *Repository) DeleteBranch(ctx context.Context, name string, force bool) error {
	ref := plumbing.ReferenceName("refs/heads/" + name)

	err := r.repo.Storer.RemoveReference(ref)
	if err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	return nil
}

// SetDefaultBranch sets the default branch.
func (r *Repository) SetDefaultBranch(ctx context.Context, name string) error {
	headRef := plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.ReferenceName("refs/heads/"+name))

	err := r.repo.Storer.SetReference(headRef)
	if err != nil {
		return fmt.Errorf("failed to set default branch: %w", err)
	}

	return nil
}

// ListRemotes lists all remotes.
func (r *Repository) ListRemotes(ctx context.Context) ([]ports.RemoteInfo, error) {
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
func (r *Repository) AddRemote(ctx context.Context, name, url string) error {
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
func (r *Repository) RemoveRemote(ctx context.Context, name string) error {
	err := r.repo.DeleteRemote(name)
	if err != nil {
		return fmt.Errorf("failed to remove remote: %w", err)
	}

	return nil
}

// UpdateRemote updates a remote URL.
func (r *Repository) UpdateRemote(ctx context.Context, name, url string) error {
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
		return errors.New("cannot pull in bare repository")
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
func (r *Repository) GetCommit(ctx context.Context, ref string) (ports.CommitInfo, error) {
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

// ListCommits lists commits based on options.
func (r *Repository) ListCommits(ctx context.Context, options ports.ListCommitsOptions) ([]ports.CommitInfo, error) {
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
			return errors.New("reached max count")
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
func (r *Repository) ListTags(ctx context.Context) ([]ports.TagInfo, error) {
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
func (r *Repository) CreateTag(ctx context.Context, name, message, ref string) error {
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
func (r *Repository) DeleteTag(ctx context.Context, name string) error {
	tagRef := plumbing.ReferenceName("refs/tags/" + name)

	err := r.repo.Storer.RemoveReference(tagRef)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	return nil
}

// Status returns the status of the working directory.
func (r *Repository) Status(ctx context.Context) (ports.StatusResult, error) {
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
func (r *Repository) Diff(ctx context.Context, options ports.DiffOptions) (string, error) {
	// For simplicity, return a basic diff implementation
	// A full implementation would use go-git's diff capabilities
	return "Diff functionality not yet implemented in go-git adapter", nil
}

// Close closes the repository.
func (r *Repository) Close() error {
	// go-git repositories don't need explicit closing
	return nil
}

// Helper methods

func (a *Adapter) buildAuth(authOptions ports.AuthOptions) (transport.AuthMethod, error) {
	switch authOptions.Type {
	case ports.AuthTypeNone:
		// No authentication required - return a no-op auth method
		return &noAuth{}, nil
	case ports.AuthTypeSSHAgent:
		// SSH agent auth is not implemented in this adapter
		return nil, errors.New("SSH agent auth not supported by go-git adapter")
	case ports.AuthTypeBasic:
		return &http.BasicAuth{
			Username: authOptions.Username,
			Password: authOptions.Password,
		}, nil
	case ports.AuthTypeToken:
		return &http.BasicAuth{
			Username: authOptions.Username,
			Password: authOptions.Token,
		}, nil
	case ports.AuthTypeSSHKey:
		if authOptions.SSHKeyPath != "" {
			publicKeys, err := ssh.NewPublicKeysFromFile(authOptions.Username, authOptions.SSHKeyPath, authOptions.Passphrase)
			if err != nil {
				return nil, fmt.Errorf("failed to load SSH key from file: %w", err)
			}

			return publicKeys, nil
		}

		if len(authOptions.SSHKey) > 0 {
			publicKeys, err := ssh.NewPublicKeys(authOptions.Username, authOptions.SSHKey, authOptions.Passphrase)
			if err != nil {
				return nil, fmt.Errorf("failed to load SSH key from bytes: %w", err)
			}

			return publicKeys, nil
		}

		return nil, errors.New("SSH key auth requires either SSHKeyPath or SSHKey")
	default:
		return nil, fmt.Errorf("unsupported auth type: %v", authOptions.Type)
	}
}

func (r *Repository) buildAuth(authOptions ports.AuthOptions) (transport.AuthMethod, error) {
	adapter := &Adapter{}

	return adapter.buildAuth(authOptions)
}

func (a *Adapter) convertTagMode(mode ports.TagMode) git.TagMode {
	switch mode {
	case ports.TagModeAll:
		return git.AllTags
	case ports.TagModeFollow:
		return git.TagFollowing
	case ports.TagModeNone:
		return git.NoTags
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

// noAuth implements transport.AuthMethod for cases where no authentication is needed.
type noAuth struct{}

// Name returns the name of the authentication method.
func (n *noAuth) Name() string {
	return "none"
}

// String returns a string representation of the auth method.
func (n *noAuth) String() string {
	return "no authentication"
}
