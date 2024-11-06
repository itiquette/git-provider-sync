// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package target

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/go-git/go-git/v5"

	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
	"os/exec"
	"strings"
)

type GitBinary struct {
	authProv   authProvider
	executor   CommandExecutor
	branchMgr  BranchManager
	binaryPath string
}

func NewGitBinary() (*GitBinary, error) {
	binaryPath, err := ValidateGitBinary()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGitBinaryNotFound, err)
	}

	if len(binaryPath) == 0 {
		return nil, ErrEmptyBinaryPath
	}

	executor := newExecService(binaryPath)

	return &GitBinary{
		authProv:   newAuthProvider(),
		executor:   executor,
		branchMgr:  newGitBranch(executor),
		binaryPath: binaryPath,
	}, nil
}

func (g *GitBinary) Clone(ctx context.Context, opt model.CloneOption) (model.Repository, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitBinary:Clone")
	opt.DebugLog(logger).Msg("GitBinary:Clone")

	env := setupSSHCommandEnv(opt.SSHClient.SSHCommand, opt.SSHClient.RewriteSSHURLFrom, opt.SSHClient.RewriteSSHURLTo)

	tmpDirPath, err := model.GetTmpDirPath(ctx)
	if err != nil {
		return model.Repository{}, fmt.Errorf("%w: %w", ErrTmpDirPath, err)
	}

	destinationDir := filepath.Join(tmpDirPath, opt.Name)
	parentDir := filepath.Dir(destinationDir)

	cloneURL := g.prepareCloneURL(ctx, opt)

	if err := g.executor.RunGitCommand(ctx, env, parentDir, "clone", cloneURL, destinationDir); err != nil {
		if strings.Contains(err.Error(), "Permission denied (publickey)") {
			return model.Repository{}, ErrPermissionDenied
		}

		return model.Repository{}, fmt.Errorf("%w: %w", ErrCloneRepository, err)
	}

	if err := g.branchMgr.Fetch(ctx, destinationDir); err != nil {
		logger.Warn().Err(err).Msg("fetch after clone failed")
	}

	return g.finalizeClone(ctx, destinationDir, cloneURL, opt.Git.Type)
}

func (g *GitBinary) prepareCloneURL(ctx context.Context, opt model.CloneOption) string {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitBinary:prepareCloneURL")
	opt.DebugLog(logger).Msg("GitBinary:prepareCloneURL")

	url := opt.URL
	if !strings.EqualFold(opt.Git.Type, gpsconfig.SSHAGENT) {
		url = addBasicAuthToURL(ctx, opt.URL, "anyuser", opt.HTTPClient.Token)
	}

	return url
}

func (g *GitBinary) finalizeClone(ctx context.Context, destinationDir, cloneURL, gitType string) (model.Repository, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitBinary:finalizeClone")
	logger.Debug().Str("destinationDir", destinationDir).Str("cloneURL", cloneURL).Str("gitType", gitType).Msg("GitBinary:finalizeClone")

	repo, err := git.PlainOpen(destinationDir)
	if err != nil {
		return model.Repository{}, fmt.Errorf("%w: %w", ErrOpenRepository, err)
	}

	if !strings.EqualFold(gitType, gpsconfig.SSHAGENT) {
		if err := g.updateRepoConfig(ctx, repo, cloneURL); err != nil {
			return model.Repository{}, err
		}
	}

	return model.NewRepository(repo) //nolint
}

func (g *GitBinary) updateRepoConfig(ctx context.Context, repo *git.Repository, cloneURL string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitBinary:updateRepoConfig")
	logger.Debug().Str("cloneURL", cloneURL).Msg("GitBinary:updateRepoConfig")

	url := removeBasicAuthFromURL(ctx, cloneURL)
	cfg, _ := repo.Config()
	cfg.Remotes["origin"].URLs = []string{url}

	if err := repo.SetConfig(cfg); err != nil {
		return fmt.Errorf("%w: %w", ErrSetRepositoryConfig, err)
	}

	return nil
}

func (g *GitBinary) Pull(ctx context.Context, pullDirPath string, opt model.PullOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitBinary:Pull")
	opt.DebugLog(logger).Str("pullDirPath", pullDirPath).Msg("GitBinary:Pull")

	env := setupSSHCommandEnv(opt.SSHClient.SSHCommand, opt.SSHClient.RewriteSSHURLFrom, opt.SSHClient.RewriteSSHURLTo)

	if err := g.executor.RunGitCommand(ctx, env, pullDirPath, "pull"); err != nil {
		return fmt.Errorf("%w: %w", ErrPullRepository, err)
	}

	return g.branchMgr.Fetch(ctx, pullDirPath) //nolint
}

func (g *GitBinary) Push(ctx context.Context, _ interfaces.GitRepository, opt model.PushOption, _ gpsconfig.GitOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitBinary:Push")
	opt.DebugLog(logger).Msg("GitBinary:Push")

	env := setupSSHCommandEnv(opt.SSHClient.SSHCommand, opt.SSHClient.RewriteSSHURLFrom, opt.SSHClient.RewriteSSHURLTo)
	args := append([]string{"push", opt.Target}, opt.RefSpecs...)

	return g.executor.RunGitCommand(ctx, env, "", args...) //nolint
}

func ValidateGitBinary() (string, error) {
	paths := []string{"git", "/usr/bin/git", "/usr/local/bin/git", "/opt/homebrew/bin/git"}
	for _, path := range paths {
		if output, err := exec.Command(path, "--version").Output(); err == nil && strings.HasPrefix(string(output), "git version") {
			return path, nil
		}
	}

	return "", ErrGitBinaryNotFound
}

func setupSSHCommandEnv(sshcommand, rewriteurlfrom, rewriteurlto string) []string {
	if sshcommand == "" {
		return []string{}
	}

	return []string{
		"GIT_SSH_COMMAND=" + sshcommand,
		"GIT_CONFIG_COUNT=1",
		"GIT_CONFIG_KEY_0=url." + rewriteurlto + ".insteadOf",
		"GIT_CONFIG_VALUE_0=" + rewriteurlfrom,
	}
}
