// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package gitbinary

import (
	"context"
	"fmt"
	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
)

type Service struct {
	authService     *authService
	executorService ExecutorService
	branchService   *operation
	binaryPath      string
}

func NewService() (*Service, error) {
	auth := NewAuthService()

	binaryPath, err := ValidateGitBinary()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGitBinaryNotFound, err)
	}

	if len(binaryPath) == 0 {
		return nil, ErrEmptyBinaryPath
	}

	executorService := NewExecutorService(binaryPath)

	return &Service{
		authService:     auth,
		executorService: executorService,
		branchService:   NewOperation(executorService),
		binaryPath:      binaryPath,
	}, nil
}

func (g *Service) Clone(ctx context.Context, opt model.CloneOption) (model.Repository, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Clone")
	opt.DebugLog(logger).Msg("Clone")

	env := SetupSSHCommandEnv(opt.SSHClient.SSHCommand, opt.SSHClient.RewriteSSHURLFrom, opt.SSHClient.RewriteSSHURLTo)

	tmpDirPath, err := model.GetTmpDirPath(ctx)
	if err != nil {
		return model.Repository{}, fmt.Errorf("%w: %w", ErrTmpDirPath, err)
	}

	destinationDir := filepath.Join(tmpDirPath, opt.Name)
	parentDir := filepath.Dir(destinationDir)

	cloneURL := g.prepareCloneURL(ctx, opt)

	if err := g.executorService.RunGitCommand(ctx, env, parentDir, "clone", cloneURL, destinationDir); err != nil {
		if strings.Contains(err.Error(), "Permission denied (publickey)") {
			return model.Repository{}, ErrPermissionDenied
		}

		return model.Repository{}, fmt.Errorf("%w: %w", ErrCloneRepository, err)
	}

	if err := g.branchService.Fetch(ctx, destinationDir); err != nil {
		logger.Warn().Err(err).Msg("fetch after clone failed")
	}

	return g.finalizeClone(ctx, destinationDir, cloneURL, opt.Git.Type)
}

func (g *Service) prepareCloneURL(ctx context.Context, opt model.CloneOption) string {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering prepareCloneURL")
	opt.DebugLog(logger).Msg("prepareCloneURL")

	url := opt.URL
	if !strings.EqualFold(opt.Git.Type, gpsconfig.SSHAGENT) {
		url = g.authService.AddBasicAuthToURL(ctx, opt.URL, "anyuser", opt.HTTPClient.Token)
	}

	return url
}

func (g *Service) finalizeClone(ctx context.Context, destinationDir, cloneURL, gitType string) (model.Repository, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering finalizeClone")
	logger.Debug().
		Str("destinationDir", destinationDir).
		Str("cloneURL", cloneURL).
		Str("gitType", gitType).
		Msg("finalizeClone")

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

func (g *Service) updateRepoConfig(ctx context.Context, repo *git.Repository, cloneURL string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering updateRepoConfig")
	logger.Debug().Str("cloneURL", cloneURL).Msg("updateRepoConfig")

	url := g.authService.RemoveBasicAuthFromURL(ctx, cloneURL)
	cfg, _ := repo.Config()
	cfg.Remotes["origin"].URLs = []string{url}

	if err := repo.SetConfig(cfg); err != nil {
		return fmt.Errorf("%w: %w", ErrSetRepositoryConfig, err)
	}

	return nil
}

func (g *Service) Pull(ctx context.Context, pullDirPath string, opt model.PullOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Pull")
	opt.DebugLog(logger).Str("pullDirPath", pullDirPath).Msg("Pull")

	env := SetupSSHCommandEnv(opt.SSHClient.SSHCommand, opt.SSHClient.RewriteSSHURLFrom, opt.SSHClient.RewriteSSHURLTo)

	if err := g.executorService.RunGitCommand(ctx, env, pullDirPath, "pull"); err != nil {
		return fmt.Errorf("%w: %w", ErrPullRepository, err)
	}

	return g.branchService.Fetch(ctx, pullDirPath)
}

func (g *Service) Push(ctx context.Context, _ interfaces.GitRepository, opt model.PushOption, _ gpsconfig.GitOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Push")
	opt.DebugLog(logger).Msg("Push")

	env := SetupSSHCommandEnv(opt.SSHClient.SSHCommand, opt.SSHClient.RewriteSSHURLFrom, opt.SSHClient.RewriteSSHURLTo)
	args := append([]string{"push", opt.Target}, opt.RefSpecs...)

	return g.executorService.RunGitCommand(ctx, env, "", args...) //nolint
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

func SetupSSHCommandEnv(sshcommand, rewriteurlfrom, rewriteurlto string) []string {
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
