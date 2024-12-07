// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

//nolint:wrapcheck
package gitbinary

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"

	"github.com/go-git/go-git/v5/config"

	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
)

// Mock implementation of ExecutorService.
type mockExecutorService struct {
	runCalled    bool
	runErr       error
	runArgs      []string
	runEnv       []string
	runDir       string
	outputCalled bool
	outputData   []byte
	outputErr    error
}

func (m *mockExecutorService) RunGitCommand(_ context.Context, env []string, dir string, args ...string) error {
	m.runCalled = true
	m.runEnv = env
	m.runDir = dir
	m.runArgs = args

	return m.runErr
}

func (m *mockExecutorService) RunGitCommandWithOutput(_ context.Context, dir string, args ...string) ([]byte, error) {
	m.outputCalled = true
	m.runDir = dir
	m.runArgs = args

	return m.outputData, m.outputErr
}

//nolint:unparam
func createTmpGitBareRepo(tmpDir, tmpRepo string) (*git.Repository, error) {
	tempPath := filepath.Join(tmpDir, "temp-"+tmpRepo)

	if err := os.MkdirAll(tempPath, 0777); err != nil {
		return nil, err
	}

	// Initialize temporary repo
	tempRep, err := git.PlainInitWithOptions(tempPath, &git.PlainInitOptions{
		Bare: false,
		InitOptions: git.InitOptions{
			DefaultBranch: plumbing.NewBranchReferenceName("main"),
		},
	})
	if err != nil {
		return nil, err
	}

	// Create remotes
	if err := setupRemotes(tempRep, tmpRepo, "https://origin.dot/"+tmpRepo+".git"); err != nil {
		return nil, err
	}

	// Create and commit content
	if err := createInitialCommit(tempRep, tempPath); err != nil {
		return nil, err
	}

	return tempRep, nil
}

func setupRemotes(repo *git.Repository, _, url string) error {
	remotes := []struct {
		name string
		url  string
	}{
		{gpsconfig.ORIGIN, url},
		{gpsconfig.GPSUPSTREAM, "http://gpsupstream.dot/anotherrepo.git"},
	}

	for _, remote := range remotes {
		_, err := repo.CreateRemote(&config.RemoteConfig{
			Name: remote.name,
			URLs: []string{remote.url},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func createInitialCommit(repo *git.Repository, repoPath string) error {
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	filename := filepath.Join(repoPath, "example-git-file")
	if err := os.WriteFile(filename, []byte("hello world!"), 0600); err != nil {
		return err
	}

	if _, err := worktree.Add("example-git-file"); err != nil {
		return err
	}

	_, err = worktree.Commit("chore: add example file", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Laval",
			Email: "laval@cavora.chi",
			When:  time.Now(),
		},
	})

	return err
}
func TestNewService(t *testing.T) {
	tests := []struct {
		name        string
		setupEnv    func()
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful creation",
			setupEnv: func() {
				// Ensure git is available in PATH for the test
				t.Setenv("PATH", "/usr/bin:/usr/local/bin")
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			if tabletest.setupEnv != nil {
				tabletest.setupEnv()
			}

			service, err := NewService()

			if tabletest.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tabletest.expectedErr)
				require.Nil(t, service)
			} else {
				require.NoError(t, err)
				require.NotNil(t, service)
				require.NotEmpty(t, service.binaryPath)
			}
		})
	}
}

func TestService_Clone(t *testing.T) {
	tmpDir := t.TempDir()
	_, _ = createTmpGitBareRepo(tmpDir, "test-repo")

	tests := []struct {
		name          string
		opt           model.CloneOption
		setupExecutor *mockExecutorService
		wantErr       bool
		expectedErr   error
		validateCalls func(*testing.T, *mockExecutorService)
	}{
		{
			name: "successful clone with HTTP",
			opt: model.CloneOption{
				URL:  filepath.Join(tmpDir, "temp-test-repo"),
				Name: "temp-test-repo",
				Git: gpsconfig.GitOption{
					Type: "http",
				},
				HTTPClient: gpsconfig.HTTPClientOption{
					Token: "test-token",
				},
			},
			setupExecutor: &mockExecutorService{
				outputData: []byte("git version 2.34.1"),
			},
			validateCalls: func(t *testing.T, m *mockExecutorService) {
				t.Helper()
				require.True(t, m.runCalled)
				require.Contains(t, m.runArgs, "git version 2.34.1")
			},
		},
		{
			name: "clone error",
			opt: model.CloneOption{
				URL:  "invalid-url",
				Name: "test-repo",
			},
			setupExecutor: &mockExecutorService{
				runErr: ErrCloneRepository,
			},
			wantErr:     true,
			expectedErr: ErrCloneRepository,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			service := &Service{
				executorService: tabletest.setupExecutor,
				authService:     NewAuthService(),
				branchService:   NewOperation(tabletest.setupExecutor),
			}

			// Set temporary directory for the test
			ctx := context.WithValue(context.Background(), model.TmpDirKey{}, tmpDir)

			repo, err := service.Clone(ctx, tabletest.opt)

			if tabletest.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tabletest.expectedErr)
				require.Empty(t, repo)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, repo)
			}

			if tabletest.validateCalls != nil {
				tabletest.validateCalls(t, tabletest.setupExecutor)
			}
		})
	}
}

func TestService_Pull(t *testing.T) {
	tmpDir := t.TempDir()
	_, _ = createTmpGitBareRepo(tmpDir, "test-repo")

	tests := []struct {
		name          string
		pullDirPath   string
		opt           model.PullOption
		setupExecutor *mockExecutorService
		wantErr       bool
		expectedErr   error
		validateCalls func(*testing.T, *mockExecutorService)
	}{
		{
			name:        "successful pull",
			pullDirPath: filepath.Join(tmpDir, "temp-test-repo"),
			opt: model.PullOption{
				SSHClient: gpsconfig.SSHClientOption{},
			},
			setupExecutor: &mockExecutorService{
				outputData: []byte("Already up to date."),
			},
			validateCalls: func(t *testing.T, m *mockExecutorService) {
				t.Helper()
				require.True(t, m.runCalled)
				strings.Contains(string(m.outputData), "Already")
			},
		},
		{
			name:        "pull error",
			pullDirPath: "/invalid/repo",
			opt:         model.PullOption{},
			setupExecutor: &mockExecutorService{
				runErr: ErrPullRepository,
			},
			wantErr:     true,
			expectedErr: ErrPullRepository,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			service := &Service{
				executorService: tabletest.setupExecutor,
				branchService:   NewOperation(tabletest.setupExecutor),
			}

			err := service.Pull(context.Background(), tabletest.pullDirPath, tabletest.opt)

			if tabletest.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tabletest.expectedErr)
			} else {
				require.NoError(t, err)
			}

			if tabletest.validateCalls != nil {
				tabletest.validateCalls(t, tabletest.setupExecutor)
			}
		})
	}
}

func TestService_Push(t *testing.T) {
	tmpDir := t.TempDir()
	_, _ = createTmpGitBareRepo(tmpDir, "test-repo")

	tests := []struct {
		name          string
		opt           model.PushOption
		gitOpt        gpsconfig.GitOption
		setupExecutor *mockExecutorService
		wantErr       bool
		expectedErr   error
		validateCalls func(*testing.T, *mockExecutorService)
	}{
		{
			name: "successful push",
			opt: model.PushOption{
				Target:    filepath.Join(tmpDir, "temp-test-repo"),
				RefSpecs:  []string{"refs/heads/main:refs/heads/main"},
				SSHClient: gpsconfig.SSHClientOption{},
			},
			setupExecutor: &mockExecutorService{
				outputData: []byte("Everything up-to-date"),
			},
			validateCalls: func(t *testing.T, m *mockExecutorService) {
				t.Helper()
				require.True(t, m.runCalled)
				require.Contains(t, m.runArgs, "push")
				require.Contains(t, m.runArgs, "refs/heads/main:refs/heads/main")
			},
		},
		{
			name: "push error",
			opt: model.PushOption{
				Target: "origin",
			},
			setupExecutor: &mockExecutorService{
				runErr: ErrPushRepository,
			},
			wantErr:     true,
			expectedErr: ErrPushRepository,
		},
		{
			name: "push with command output",
			opt: model.PushOption{
				Target:   "origin",
				RefSpecs: []string{"HEAD:refs/for/main"},
			},
			setupExecutor: &mockExecutorService{
				outputData: []byte("New branch created"),
			},
			validateCalls: func(t *testing.T, m *mockExecutorService) {
				t.Helper()
				require.True(t, m.runCalled)
				require.Equal(t, []string{"push", "origin", "HEAD:refs/for/main"}, m.runArgs)
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			service := &Service{
				executorService: tabletest.setupExecutor,
			}

			err := service.Push(context.Background(), nil, tabletest.opt, tabletest.gitOpt)

			if tabletest.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tabletest.expectedErr)
			} else {
				require.NoError(t, err)
			}

			if tabletest.validateCalls != nil {
				tabletest.validateCalls(t, tabletest.setupExecutor)
			}
		})
	}
}

func TestSetupSSHCommandEnv(t *testing.T) {
	tests := []struct {
		name           string
		sshCommand     string
		rewriteFrom    string
		rewriteTo      string
		expectedEnv    []string
		expectedLength int
	}{
		{
			name:           "empty ssh command",
			sshCommand:     "",
			expectedEnv:    []string{},
			expectedLength: 0,
		},
		{
			name:        "with ssh command and rewrite rules",
			sshCommand:  "ssh -i /path/to/key",
			rewriteFrom: "git@github.com:",
			rewriteTo:   "ssh://git@github.com/",
			expectedEnv: []string{
				"GIT_SSH_COMMAND=ssh -i /path/to/key",
				"GIT_CONFIG_COUNT=1",
				"GIT_CONFIG_KEY_0=url.ssh://git@github.com/.insteadOf",
				"GIT_CONFIG_VALUE_0=git@github.com:",
			},
			expectedLength: 4,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			env := SetupSSHCommandEnv(tabletest.sshCommand, tabletest.rewriteFrom, tabletest.rewriteTo)
			require.Len(t, env, tabletest.expectedLength)
			require.Equal(t, tabletest.expectedEnv, env)
		})
	}
}
