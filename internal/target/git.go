// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package target

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"

	"github.com/go-git/go-git/v5"
	gogitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/rs/zerolog"
)

// Git represents operations against a Git repository.
// It encapsulates a go-git Repository and provides methods for common Git operations.
type Git struct {
	goGitRepository *git.Repository
	name            string
}

// Clone clones a repository to a target according to given clone options.
// It creates a new repository at the specified target path and returns a model.Repository.
//
// Parameters:
// - ctx: The context for the operation, which can be used for cancellation and passing values.
// - option: The CloneOption containing details about the clone operation, including the source URL and target path.
//
// Returns:
// - A model.Repository representing the cloned repository.
// - An error if the cloning process fails at any step.
func (g Git) Clone(ctx context.Context, option model.CloneOption) (model.Repository, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Git:Clone")
	logger.Debug().Str("url", option.URL).Msg("Git:Clone")

	auth, err := g.getAuthMethod(option.Git, option.HTTPClient.Token)
	if err != nil {
		return model.Repository{}, fmt.Errorf("failed to get auth method: %w", err)
	}

	memStorage := memory.NewStorage()
	cloneOptions := newGitGoCloneOption(option.URL, option.Mirror, auth)

	cloneRepository, err := git.Clone(memStorage, nil, &cloneOptions)
	if err != nil {
		return model.Repository{}, fmt.Errorf("failed to clone repository: %w", err)
	}

	repo, err := model.NewRepository(cloneRepository)
	if err != nil {
		return model.Repository{}, fmt.Errorf("failed to create new repository: %w", err)
	}

	return repo, nil
}

// Pull performs a git pull operation on the repository.
// It updates the current branch with changes from the remote repository.
//
// Parameters:
// - ctx: The context for the operation, which can be used for cancellation and passing values.
// - option: The PullOption containing details about the pull operation, including the target path.
//
// Returns an error if the pull operation fails or if the workspace is unclean.
func (g Git) Pull(ctx context.Context, targetDirPath string, option model.PullOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Git:Pull")
	option.DebugLog(logger).Msg("Git:Pull")

	targetRepository, err := git.PlainOpen(targetDirPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := targetRepository.Worktree()
	if err != nil {
		return fmt.Errorf("failed to open repository worktree: %w", err)
	}

	if status, _ := worktree.Status(); !status.IsClean() {
		return fmt.Errorf("workspace at %s is unclean, aborting", option.TargetPath)
	}

	var gitPullOptions git.PullOptions

	var auth transport.AuthMethod

	if strings.EqualFold(option.GitOption.Type, config.SSHKEY) {
		auth, err = ssh.NewPublicKeysFromFile("git", option.GitOption.SSHPrivateKeyPath, option.GitOption.SSHPrivateKeyPW)
		if err != nil {
			return fmt.Errorf("generate publickeys failed: %w", err)
		}
	}

	fmt.Println(option.HTTPClientOption.Token)

	if strings.EqualFold(option.GitOption.Type, config.HTTPS) || len(option.GitOption.Type) == 0 {
		auth = &http.BasicAuth{Username: "anyUser", Password: option.HTTPClientOption.Token}
	}

	gitPullOptions = newGitGoPullOption(config.ORIGIN, "", auth)

	if err := worktree.Pull(&gitPullOptions); err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			logger.Debug().Msgf("Repository %s already up-to-date, ignoring Pull", g.name)
			g.updateSyncRunMetainfo(ctx, "uptodate")

			return nil
		}

		return fmt.Errorf("failed to pull repository: %w", err)
	}

	if err := g.fetchBranches(ctx, targetRepository); err != nil {
		return fmt.Errorf("failed to fetch branches: %w", err)
	}

	return nil
}

// Push pushes an existing repository to a target provider according to given push options.
// It sends local changes to the remote repository.
//
// Parameters:
// - ctx: The context for the operation, which can be used for cancellation and passing values.
// - option: The PushOption containing details about the push operation, including the target and RefSpecs.
//
// Returns an error if the push operation fails.
func (g Git) Push(ctx context.Context, repository interfaces.GitRepository, option model.PushOption, _ config.ProviderConfig, targetGitOption config.GitOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Git:Push")
	option.DebugLog(logger).Msg("Git:Push")
	g.logRemotes(logger)

	auth, err := g.getAuthMethod(targetGitOption, option.HTTPClient.Token)
	if err != nil {
		return fmt.Errorf("failed to get auth method: %w", err)
	}

	gitOptions := newGitGoPushOption(option.Target, option.RefSpecs, option.Prune, auth)
	logger.Info().Str("target", option.Target).Msg("Pushing to:")

	if err := repository.GoGitRepository().Push(&gitOptions); err != nil {
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

	return fmt.Errorf("failed to push: %w", err)
}

func (g Git) getAuthMethod(gitOption config.GitOption, token string) (transport.AuthMethod, error) {
	switch {
	case strings.EqualFold(gitOption.Type, config.SSHKEY):
		var keys *ssh.PublicKeys

		var err error
		if keys, err = ssh.NewPublicKeysFromFile("git", gitOption.SSHPrivateKeyPath, gitOption.SSHPrivateKeyPW); err != nil {
			return nil, fmt.Errorf("failed to get public key from file. Path: %s. Err: %w", gitOption.SSHPrivateKeyPath, err)
		}

		return keys, nil
	case strings.EqualFold(gitOption.Type, config.HTTPS) || gitOption.Type == "":
		return &http.BasicAuth{Username: "anyUser", Password: token}, nil
	default:
		return nil, fmt.Errorf("unsupported git option type: %s", gitOption.Type)
	}
}

// Fetch fetches all branches locally.
// It updates the local repository with changes from the remote without merging them.
//
// Parameters:
// - ctx: The context for the operation, which can be used for cancellation and passing values.
// - repository: The model.Repository to fetch branches for.
//
// Returns an error if the fetch operation fails.
func (g Git) Fetch(ctx context.Context, repository model.Repository) error {
	return g.fetchBranches(ctx, repository.GoGitRepository())
}

// fetchBranches fetches all branches from the remote repository.
// This is an internal method used by Fetch and Pull operations.
func (g Git) fetchBranches(ctx context.Context, repository *git.Repository) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Git:fetchBranches")

	refSpecs := []gogitconfig.RefSpec{
		"refs/*:refs/*",
		"^refs/pull/*:refs/pull/*",
	}

	options := &git.FetchOptions{
		RefSpecs: refSpecs,
	}

	if err := repository.Fetch(options); err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			logger.Debug().Msg("Repository already up-to-date, ignoring fetchBranches")

			return nil
		}

		return fmt.Errorf("failed to fetch branches: %w", err)
	}

	return nil
}

// updateSyncRunMetainfo updates the synchronization run metadata in the context.
// It is used to track repositories that are already up-to-date during sync operations.
func (g Git) updateSyncRunMetainfo(ctx context.Context, key string) {
	if syncRunMeta, ok := ctx.Value(model.SyncRunMetainfoKey{}).(model.SyncRunMetainfo); ok {
		syncRunMeta.Fail[key] = append(syncRunMeta.Fail[key], g.name)
	}
}

// logRemotes logs the remote configurations of the repository.
// This is used for debugging purposes to show the configured remotes.
func (g Git) logRemotes(logger *zerolog.Logger) {
	if remotes, err := g.goGitRepository.Remotes(); err == nil {
		for _, remote := range remotes {
			logger.Debug().Strs("url", remote.Config().URLs).Str("name", remote.Config().Name).Msg("Remote:")
		}
	}
}

// NewGit creates a new Git instance.
// It initializes a Git struct with a go-git Repository and a name.
//
// Parameters:
// - repository: An interface that provides access to a go-git Repository.
// - name: A string identifier for the repository.
//
// Returns a new Git instance.
func NewGit(repository interfaces.GitRepository, name string) Git {
	return Git{goGitRepository: repository.GoGitRepository(), name: name}
}
