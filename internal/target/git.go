// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

// Package target provides Git operations for repository management.
package target

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	gogitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/rs/zerolog"
)

// Common errors.
var (
	ErrBranchCheckout         = errors.New("failed to checkout branch")
	ErrCloneRepository        = errors.New("failed to clone repository")
	ErrCreateNewRepository    = errors.New("failed to create new repository")
	ErrFetchBranches          = errors.New("failed to fetch branches")
	ErrGetAuthMethod          = errors.New("failed to get auth method")
	ErrGetWorktree            = errors.New("failed to get worktree")
	ErrHeadSet                = errors.New("failed to set HEAD reference")
	ErrOpenRepository         = errors.New("failed to open repository")
	ErrOpenRepositoryWorktree = errors.New("failed to open repository worktree")
	ErrRemoteCreation         = errors.New("failed to set remote in target repository")
	ErrRepoInitialization     = errors.New("failed to initialize target repository")
	ErrRepoPush               = errors.New("failed to push to target repository")
	ErrRepoPull               = errors.New("failed to pull updates")
	ErrUncleanWorkspace       = errors.New("workspace is unclean, aborting")
	ErrUnsupportedGitType     = errors.New("unsupported git option type")
	ErrReadSSHKey             = errors.New("failed to read ssh key")
)

// Git represents operations against a Git repository.
type Git struct {
	goGitRepository *git.Repository
	name            string
}

// Clone clones a repository according to given clone options.
func (g Git) Clone(ctx context.Context, option model.CloneOption) (model.Repository, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Git:Clone")
	logger.Debug().Str("url", option.URL).Msg("Git:Clone")

	auth, err := g.getAuthMethod(option.Git, option.HTTPClient.Token)
	if err != nil {
		return model.Repository{}, fmt.Errorf("%w: %w", ErrGetAuthMethod, err)
	}

	var memFS billy.Filesystem

	memStorage := memory.NewStorage()

	if option.PlainRepo {
		memFS = memfs.New()
	}

	cloneOptions := newGitGoCloneOption(option.URL, option.Mirror, auth)

	repo, err := git.Clone(memStorage, memFS, &cloneOptions)
	if err != nil {
		return model.Repository{}, fmt.Errorf("%w: %w", ErrCloneRepository, err)
	}

	newRepo, err := model.NewRepository(repo)
	if err != nil {
		return model.Repository{}, fmt.Errorf("%w: %w", ErrCreateNewRepository, err)
	}

	return newRepo, nil
}

// Pull performs a git pull operation on the repository.
func (g Git) Pull(ctx context.Context, pullDirPath string, option model.PullOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Git:Pull")
	option.DebugLog(logger).Msg("Git:Pull")

	repo, err := git.PlainOpen(pullDirPath)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrOpenRepository, pullDirPath, err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOpenRepositoryWorktree, err)
	}

	if status, _ := worktree.Status(); !status.IsClean() {
		return fmt.Errorf("%w: %s", ErrUncleanWorkspace, pullDirPath)
	}

	auth, err := g.getAuthMethod(option.GitOption, option.HTTPClientOption.Token)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGetAuthMethod, err)
	}

	pullOptions := newGitGoPullOption(gpsconfig.ORIGIN, "", auth)

	if err := worktree.Pull(&pullOptions); err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			logger.Debug().Msgf("Repository %s already up-to-date, ignoring Pull", g.name)
			g.updateSyncRunMetainfo(ctx, "uptodate")

			return nil
		}

		return fmt.Errorf("%w: %w", ErrRepoPull, err)
	}

	return g.fetchBranches(ctx, repo)
}

// Push pushes an existing repository to a target provider.
func (g Git) Push(ctx context.Context, repository interfaces.GitRepository, option model.PushOption, _ gpsconfig.ProviderConfig, targetGitOption gpsconfig.GitOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Git:Push")
	option.DebugLog(logger).Msg("Git:Push")
	g.logRemotes(logger)

	auth, err := g.getAuthMethod(targetGitOption, option.HTTPClient.Token)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGetAuthMethod, err)
	}

	pushOptions := newGitGoPushOption(option.Target, option.RefSpecs, option.Prune, auth)
	logger.Info().Str("target", option.Target).Msg("Pushing to:")

	if err := repository.GoGitRepository().Push(&pushOptions); err != nil {
		return g.handlePushError(ctx, err)
	}

	return nil
}

func (g Git) handlePushError(ctx context.Context, err error) error {
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		log.Logger(ctx).Debug().Msgf("Repository %s already up-to-date, ignoring Push", g.name)
		g.updateSyncRunMetainfo(ctx, "uptodate")

		return nil
	}

	return fmt.Errorf("%w: %w", ErrRepoPush, err)
}

func (g Git) getAuthMethod(gitOption gpsconfig.GitOption, token string) (transport.AuthMethod, error) {
	switch strings.ToLower(gitOption.Type) {
	case gpsconfig.SSHKEY:
		keys, err := ssh.NewPublicKeysFromFile("git", gitOption.SSHPrivateKeyPath, gitOption.SSHPrivateKeyPW)
		if err != nil {
			return nil, fmt.Errorf("%w: key path: %s. err: %w", ErrReadSSHKey, gitOption.SSHPrivateKeyPath, err)
		}

		return keys, nil
	case gpsconfig.HTTPS, "":
		return &http.BasicAuth{Username: "anyUser", Password: token}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedGitType, gitOption.Type)
	}
}

// Fetch fetches all branches locally.
func (g Git) Fetch(ctx context.Context, repository model.Repository) error {
	return g.fetchBranches(ctx, repository.GoGitRepository())
}

func (g Git) fetchBranches(ctx context.Context, repository *git.Repository) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Git:fetchBranches")

	refSpecs := []gogitconfig.RefSpec{
		"refs/*:refs/*",
		"^refs/pull/*:refs/pull/*",
	}

	options := &git.FetchOptions{RefSpecs: refSpecs}

	if err := repository.Fetch(options); err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			logger.Debug().Msg("Repository already up-to-date, ignoring fetchBranches")

			return nil
		}

		return fmt.Errorf("%w: %w", ErrFetchBranches, err)
	}

	return nil
}

func (g Git) updateSyncRunMetainfo(ctx context.Context, key string) {
	if syncRunMeta, ok := ctx.Value(model.SyncRunMetainfoKey{}).(model.SyncRunMetainfo); ok {
		syncRunMeta.Fail[key] = append(syncRunMeta.Fail[key], g.name)
	}
}

func (g Git) logRemotes(logger *zerolog.Logger) {
	if remotes, err := g.goGitRepository.Remotes(); err == nil {
		for _, remote := range remotes {
			logger.Debug().Strs("url", remote.Config().URLs).Str("name", remote.Config().Name).Msg("Remote:")
		}
	}
}

// NewGit creates a new Git instance.
func NewGit(repository interfaces.GitRepository, name string) Git {
	return Git{goGitRepository: repository.GoGitRepository(), name: name}
}

func setRemoteAndBranch(repository interfaces.GitRepository, sourceDirPath string) error {
	targetRepo, err := git.PlainOpen(sourceDirPath)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrOpenRepository, sourceDirPath, err)
	}

	remote, err := repository.GoGitRepository().Remote(gpsconfig.ORIGIN)
	if err == nil {
		if _, err := targetRepo.CreateRemote(&gogitconfig.RemoteConfig{
			Name: gpsconfig.ORIGIN,
			URLs: remote.Config().URLs,
		}); err != nil {
			return fmt.Errorf("%w: %w", ErrRemoteCreation, err)
		}
	}

	return nil
}

func setDefaultBranch(repoPath, branchName string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrOpenRepository, repoPath, err)
	}

	branchRef := plumbing.NewBranchReferenceName(branchName)

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGetWorktree, err)
	}

	if err := worktree.Checkout(&git.CheckoutOptions{
		Branch: branchRef,
		Force:  true,
	}); err != nil {
		return fmt.Errorf("%w: %s: %w", ErrBranchCheckout, branchName, err)
	}

	headRef := plumbing.NewSymbolicReference(plumbing.HEAD, branchRef)
	if err := repo.Storer.SetReference(headRef); err != nil {
		return fmt.Errorf("%w: %w", ErrHeadSet, err)
	}

	return nil
}
