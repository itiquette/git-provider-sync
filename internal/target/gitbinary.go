// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package target

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"

	"github.com/go-git/go-git/v5"
)

var (
	ErrGitBinaryNotFound   = errors.New("failed to find a Git executable")
	ErrEmptyBinaryPath     = errors.New("failed to find Git binary path")
	ErrPermissionDenied    = errors.New("failed with permission denied (publickey). Provide correct key in your ssh-agent")
	ErrSetRepositoryConfig = errors.New("failed to set repository config")
	ErrPullRepository      = errors.New("failed to pull repository")
	ErrGetRemoteBranches   = errors.New("failed to get remote branches")
	ErrTmpDirPath          = errors.New("failed to get tmpdirpath")
)

type GitBinary struct {
	gitBinaryPath string
}

func NewGitBinary() (GitBinary, error) {
	binaryPath, err := ValidateGitBinary()
	if err != nil {
		return GitBinary{}, fmt.Errorf("%w: %w", ErrGitBinaryNotFound, err)
	}

	if len(binaryPath) == 0 {
		return GitBinary{}, ErrEmptyBinaryPath
	}

	return GitBinary{
		gitBinaryPath: binaryPath,
	}, nil
}

func (g GitBinary) Clone(ctx context.Context, option model.CloneOption) (model.Repository, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("GitBinary:Clone")
	option.DebugLog(logger).Msg("GitBinary:CloneOption")

	env := setupSSHCommandEnv(option.SSHClient.SSHCommand, option.SSHClient.RewriteSSHURLFrom, option.SSHClient.RewriteSSHURLTo)

	tmpDirPath, err := model.GetTmpDirPath(ctx)
	if err != nil {
		return model.Repository{}, fmt.Errorf("%w: %w", ErrTmpDirPath, err)
	}

	destinationDir := filepath.Join(tmpDirPath, option.Name)
	parentDir := filepath.Dir(destinationDir)

	url := option.URL
	if !strings.EqualFold(option.Git.Type, gpsconfig.SSHAGENT) {
		url = addBasicAuthToURL(option.URL, "anyuser", option.HTTPClient.Token)
	}

	if err := g.runGitCommand(ctx, env, parentDir, "clone", url, destinationDir); err != nil {
		if strings.Contains(err.Error(), "Permission denied (publickey)") {
			return model.Repository{}, ErrPermissionDenied
		}

		return model.Repository{}, fmt.Errorf("%w: %w", ErrCloneRepository, err)
	}

	g.fetch(ctx, destinationDir) //nolint

	repo, err := git.PlainOpen(destinationDir)
	if err != nil {
		return model.Repository{}, fmt.Errorf("%w: %w", ErrOpenRepository, err)
	}

	if !strings.EqualFold(option.Git.Type, gpsconfig.SSHAGENT) {
		url = removeBasicAuthFromURL(url)

		cfg, _ := repo.Config()
		cfg.Remotes["origin"].URLs = []string{url}

		err = repo.SetConfig(cfg)
		if err != nil {
			return model.Repository{}, fmt.Errorf("%w: %w", ErrSetRepositoryConfig, err)
		}
	}

	return model.NewRepository(repo) //nolint
}

func (g GitBinary) Pull(ctx context.Context, pullDirPath string, option model.PullOption) error {
	logger := log.Logger(ctx)
	option.DebugLog(logger).Msg("GitBinary:Pull")

	env := setupSSHCommandEnv(option.SSHClient.SSHCommand, option.SSHClient.RewriteSSHURLFrom, option.SSHClient.RewriteSSHURLTo)

	if err := g.runGitCommand(ctx, env, pullDirPath, "pull"); err != nil {
		return fmt.Errorf("%w: %w", ErrPullRepository, err)
	}

	return g.fetch(ctx, pullDirPath)
}

func (g GitBinary) Push(ctx context.Context, _ interfaces.GitRepository, option model.PushOption, _ gpsconfig.ProviderConfig, _ gpsconfig.GitOption) error {
	logger := log.Logger(ctx)
	option.DebugLog(logger).Msg("GitBinary:Push")

	env := setupSSHCommandEnv(option.SSHClient.SSHCommand, option.SSHClient.RewriteSSHURLFrom, option.SSHClient.RewriteSSHURLTo)

	args := append([]string{"push", option.Target}, option.RefSpecs...)

	return g.runGitCommand(ctx, env, "", args...)
}

func (g GitBinary) fetch(ctx context.Context, workingDirPath string) error {
	commands := [][]string{
		{"fetch", "--all", "--prune"},
		{"pull", "--all"},
		{"pull", "--all"},
	}

	for _, cmd := range commands {
		if err := g.runGitCommand(ctx, nil, workingDirPath, cmd...); err != nil {
			return err
		}
	}

	return g.createTrackingBranches(ctx, workingDirPath)
}

func (g GitBinary) runGitCommand(ctx context.Context, env []string, workingDir string, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, g.gitBinaryPath, args...) //nolint:gosec

	cmd.Env = append(os.Environ(), env...)
	if len(workingDir) != 0 {
		cmd.Dir = workingDir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing '%s %s': %w. err: %s", g.gitBinaryPath, strings.Join(args, " "), err, output)
	}

	log.Logger(ctx).Debug().Msgf("Git command output: %s", output)

	return nil
}

func addBasicAuthToURL(urlStr, username, password string) string {
	parsedURL, _ := url.Parse(urlStr)
	parsedURL.User = url.UserPassword(username, password)

	return parsedURL.String()
}

func removeBasicAuthFromURL(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	parsedURL.User = nil

	return parsedURL.String()
}

func (g GitBinary) createTrackingBranches(ctx context.Context, repoPath string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("GitBinary:createTrackingBranches")

	output, err := g.runGitCommandWithOutput(ctx, repoPath, "branch", "-r")
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGetRemoteBranches, err)
	}

	for _, branch := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		branch = strings.TrimSpace(branch)
		if strings.Contains(branch, "->") {
			continue
		}

		localBranch := strings.TrimPrefix(branch, "origin/")
		if err := g.runGitCommand(ctx, nil, repoPath, "branch", "--track", localBranch, branch); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				logger.Debug().Msgf("Could not create tracking branch for %s: %s", branch, err.Error())
			}
		} else {
			logger.Debug().Msgf("Created tracking branch for %s", branch)
		}
	}

	return nil
}

func (g GitBinary) runGitCommandWithOutput(ctx context.Context, workingDir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, g.gitBinaryPath, args...) //nolint:gosec
	cmd.Dir = workingDir

	return cmd.Output() //nolint
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
