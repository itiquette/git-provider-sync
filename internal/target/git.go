// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package target

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"itiquette/git-provider-sync/internal/configuration"
	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"

	"github.com/go-git/go-git/v5"
	gogitconfig "github.com/go-git/go-git/v5/config"

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
	logger.Debug().Str("url", option.URL).Str("target", option.TargetPath).Msg("Git:Clone")

	var gitGoCloneOptions git.CloneOptions

	var err error

	var auth transport.AuthMethod

	if strings.EqualFold(option.GitInfo.Type, model.SSHKEY) {
		auth, err = ssh.NewPublicKeysFromFile("git", option.GitInfo.SSHPrivateKeyPath, option.GitInfo.SSHPrivateKeyPW)
		if err != nil {
			return model.Repository{}, fmt.Errorf("generate publickeys failed: %w", err)
		}
	}

	if strings.EqualFold(option.GitInfo.Type, model.HTTPS) || len(option.GitInfo.Type) == 0 {
		auth = &http.BasicAuth{Username: "anyUser", Password: option.GitInfo.ProviderToken}
	}

	gitGoCloneOptions = newGitGoCloneOption(option.URL, option.Mirror, auth)

	cloneRepository, err := git.PlainClone(option.TargetPath, false, &gitGoCloneOptions)
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
func (g Git) Pull(ctx context.Context, option model.PullOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Git:Pull")
	option.DebugLog(logger).Msg("Git:Pull")

	targetRepository, err := git.PlainOpen(option.TargetPath)
	if err != nil {
		return fmt.Errorf("failed to open repository directory %s: %w", option.TargetPath, err)
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

	if strings.EqualFold(option.GitInfo.Type, model.SSHKEY) {
		auth, err = ssh.NewPublicKeysFromFile("git", option.GitInfo.SSHPrivateKeyPath, option.GitInfo.SSHPrivateKeyPW)
		if err != nil {
			return fmt.Errorf("generate publickeys failed: %w", err)
		}
	}

	if strings.EqualFold(option.GitInfo.Type, model.HTTPS) || len(option.GitInfo.Type) == 0 {
		auth = &http.BasicAuth{Username: "anyUser", Password: option.GitInfo.ProviderToken}
	}

	gitPullOptions = newGitGoPullOption(configuration.ORIGIN, "", auth)

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
func (g Git) Push(ctx context.Context, option model.PushOption, _ model.GitInfo, targetGitInfo model.GitInfo) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Git:Push")
	option.DebugLog(logger).Msg("Git:Push")

	g.logRemotes(logger)

	var auth transport.AuthMethod

	var err error

	if strings.EqualFold(targetGitInfo.Type, model.SSHKEY) {
		auth, err = ssh.NewPublicKeysFromFile("git", targetGitInfo.SSHPrivateKeyPath, targetGitInfo.SSHPrivateKeyPW)
		if err != nil {
			return fmt.Errorf("generate publickeys failed: %w", err)
		}
	}

	if strings.EqualFold(targetGitInfo.Type, model.HTTPS) || len(targetGitInfo.Type) == 0 {
		auth = &http.BasicAuth{Username: "anyUser", Password: targetGitInfo.ProviderToken}
	}

	gitOptions := newGitGoPushOption(option.Target, option.RefSpecs, option.Prune, auth)

	logger.Info().Str("target", option.Target).Msg("Pushing to:")

	if err := g.goGitRepository.Push(&gitOptions); err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			logger.Debug().Msgf("Repository %s already up-to-date, ignoring Push", g.name)
			g.updateSyncRunMetainfo(ctx, "uptodate")

			return nil
		}

		return fmt.Errorf("failed to push: %w", err)
	}

	return nil
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
