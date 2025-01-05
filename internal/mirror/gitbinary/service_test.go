// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package gitbinary

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
)

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

func createTestRepo(t *testing.T, tmpDir string) *git.Repository {
	t.Helper()

	repoPath := filepath.Join(tmpDir, "repo")
	require.NoError(t, os.MkdirAll(repoPath, 0755))

	repo, err := git.PlainInit(repoPath, false)
	require.NoError(t, err)

	cfg, err := repo.Config()
	require.NoError(t, err)

	cfg.Remotes = make(map[string]*config.RemoteConfig)
	cfg.Remotes["origin"] = &config.RemoteConfig{
		Name: "origin",
		URLs: []string{"https://example.com/repo.git"},
	}
	require.NoError(t, repo.SetConfig(cfg))

	worktree, err := repo.Worktree()
	require.NoError(t, err)

	testFile := filepath.Join(repoPath, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0600))

	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	_, err = worktree.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	return repo
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
				t.Setenv("PATH", "/usr/bin:/usr/local/bin")
			},
		},
		// { TO-DO: fix so that this can be tested
		// 	name: "git binary not found",
		// 	setupEnv: func() {
		// 		t.Setenv("PATH", "/nonexistent")
		// 	},
		// 	wantErr:     true,
		// 	expectedErr: ErrGitBinaryNotFound,
		// },
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

				return
			}

			require.NoError(t, err)
			require.NotNil(t, service)
			require.NotEmpty(t, service.binaryPath)
		})
	}
}

//TO-DO
// func TestService_Clone(t *testing.T) {
// 	tmpDir := t.TempDir()

// 	tests := []struct {
// 		name        string
// 		opt         model.CloneOption
// 		setup       func(*mockExecutorService)
// 		wantErr     bool
// 		errType     error
// 		checkOutput func(*testing.T, model.Repository)
// 	}{
// 		{
// 			name: "successful http clone",
// 			opt: model.CloneOption{
// 				URL:  "https://example.com/repo.git",
// 				Name: "repo",
// 				SourceCfg: gpsconfig.SyncConfig{
// 					ProviderType: "http",
// 				},
// 				AuthCfg: gpsconfig.AuthConfig{
// 					Token: "token123",
// 				},
// 			},
// 			setup: func(m *mockExecutorService) {
// 				m.outputData = []byte("origin/main\norigin/develop")
// 			},
// 		},
// 		{
// 			name: "successful ssh clone",
// 			opt: model.CloneOption{
// 				URL:  "git@example.com:repo.git",
// 				Name: "repo",
// 				SourceCfg: gpsconfig.SyncConfig{
// 					ProviderType: gpsconfig.SSHAGENT,
// 				},
// 				AuthCfg: gpsconfig.AuthConfig{
// 					SSHCommand: "ssh -i key",
// 				},
// 			},
// 			setup: func(m *mockExecutorService) {
// 				m.outputData = []byte("origin/main")
// 			},
// 		},
// 		{
// 			name: "clone permission denied",
// 			opt: model.CloneOption{
// 				URL:  "https://example.com/repo.git",
// 				Name: "repo",
// 			},
// 			setup: func(m *mockExecutorService) {
// 				m.runErr = errors.New("Permission denied (publickey)")
// 			},
// 			wantErr: true,
// 			errType: ErrPermissionDenied,
// 		},
// 	}

// 	for _, tabletest := range tests {
// 		t.Run(tabletest.name, func(t *testing.T) {
// 			mock := &mockExecutorService{}
// 			if tabletest.setup != nil {
// 				tabletest.setup(mock)
// 			}

// 			svc := &Service{
// 				authService:     NewAuthService(),
// 				executorService: mock,
// 				branchService:   NewOperation(mock),
// 			}

// 			ctx := context.WithValue(context.Background(), model.TmpDirKey{}, tmpDir)
// 			repo, err := svc.Clone(ctx, tabletest.opt)

// 			if tabletest.wantErr {
// 				require.Error(t, err)

// 				if tabletest.errType != nil {
// 					require.ErrorIs(t, err, tabletest.errType)
// 				}

// 				return
// 			}

// 			require.NoError(t, err)
// 			require.NotNil(t, repo)

// 			if tabletest.checkOutput != nil {
// 				tabletest.checkOutput(t, repo)
// 			}
// 		})
// 	}
// }

func TestService_Pull(t *testing.T) {
	tmpDir := t.TempDir()
	repo := createTestRepo(t, tmpDir)

	tests := []struct {
		name    string
		opt     model.PullOption
		setup   func(*mockExecutorService)
		wantErr bool
		errType error
	}{
		{
			name: "successful pull",
			opt: model.PullOption{
				AuthCfg: gpsconfig.AuthConfig{
					SSHCommand: "ssh -i key",
				},
			},
		},
		{
			name: "pull error",
			opt:  model.PullOption{},
			setup: func(m *mockExecutorService) {
				m.runErr = errors.New("network error")
			},
			wantErr: true,
			errType: ErrPullRepository,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			mock := &mockExecutorService{}
			if tabletest.setup != nil {
				tabletest.setup(mock)
			}

			svc := &Service{
				executorService: mock,
				branchService:   NewOperation(mock),
			}
			worktree, _ := repo.Worktree()
			err := svc.Pull(context.Background(), worktree.Filesystem.Root(), tabletest.opt)

			if tabletest.wantErr {
				require.Error(t, err)

				if tabletest.errType != nil {
					require.ErrorIs(t, err, tabletest.errType)
				}

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestService_Push(t *testing.T) {
	tmpDir := t.TempDir()
	repo := createTestRepo(t, tmpDir)

	repa, _ := model.NewRepository(repo)
	repa.ProjectMetaInfo = &model.ProjectInfo{}
	repa.ProjectMetaInfo.OriginalName = "name"
	repa.ProjectMetaInfo.CleanName = "name"

	tests := []struct {
		name      string
		opt       model.PushOption
		setup     func(*mockExecutorService)
		rep       interfaces.GitRepository
		wantErr   bool
		errType   error
		checkCall func(*testing.T, *mockExecutorService)
	}{
		{
			name: "successful push",
			opt: model.PushOption{
				Target:   "origin",
				RefSpecs: []string{"HEAD:refs/for/main"},
				AuthCfg: gpsconfig.AuthConfig{
					SSHCommand: "ssh -i key",
				},
			},
			rep: repa,
			checkCall: func(t *testing.T, m *mockExecutorService) {
				t.Helper()
				require.Equal(t, []string{"push", "origin", "HEAD:refs/for/main"}, m.runArgs)
				require.Contains(t, m.runEnv, "GIT_SSH_COMMAND=ssh -i key")
			},
		},
		{
			name: "push error",
			opt: model.PushOption{
				Target: "origin",
			},
			rep: repa,
			setup: func(m *mockExecutorService) {
				m.runErr = errors.New("network error")
			},
			wantErr: true,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			mock := &mockExecutorService{}
			if tabletest.setup != nil {
				tabletest.setup(mock)
			}

			svc := &Service{
				executorService: mock,
			}

			err := svc.Push(testContext(tmpDir), tabletest.rep, tabletest.opt)

			if tabletest.wantErr {
				require.Error(t, err)

				if tabletest.errType != nil {
					require.ErrorIs(t, err, tabletest.errType)
				}

				return
			}

			require.NoError(t, err)

			if tabletest.checkCall != nil {
				tabletest.checkCall(t, mock)
			}
		})
	}
}

func testContext(tmpDir string) context.Context {
	ctx := context.Background()

	return context.WithValue(ctx, model.TmpDirKey{}, tmpDir)
}

func TestSetupSSHCommandEnv(t *testing.T) {
	tests := []struct {
		name        string
		sshCmd      string
		rewriteFrom string
		rewriteTo   string
		want        []string
	}{
		{
			name:   "empty command",
			sshCmd: "",
			want:   []string{},
		},
		{
			name:        "full config",
			sshCmd:      "ssh -i key",
			rewriteFrom: "git@github.com:",
			rewriteTo:   "ssh://git@github.com/",
			want: []string{
				"GIT_SSH_COMMAND=ssh -i key",
				"GIT_CONFIG_COUNT=1",
				"GIT_CONFIG_KEY_0=url.ssh://git@github.com/.insteadOf",
				"GIT_CONFIG_VALUE_0=git@github.com:",
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			got := SetupSSHCommandEnv(tabletest.sshCmd, tabletest.rewriteFrom, tabletest.rewriteTo)
			require.Equal(t, tabletest.want, got)
		})
	}
}
