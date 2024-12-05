// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package gitlib

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
)

// Mock for MetadataHandler.
type mockMetadataHandler struct {
	updateCalled bool
	updateStatus string
	updatePath   string
}

func (m *mockMetadataHandler) UpdateSyncMetadata(_ context.Context, status, path string) {
	m.updateCalled = true
	m.updateStatus = status
	m.updatePath = path
}

// Mock for AuthService.
type mockAuthService struct {
	auth transport.AuthMethod
	err  error
}

func (m *mockAuthService) GetAuthMethod(context.Context, gpsconfig.GitOption, gpsconfig.HTTPClientOption, gpsconfig.SSHClientOption) (transport.AuthMethod, error) {
	return m.auth, m.err
}

func TestService_Clone(t *testing.T) {
	tmpDir := t.TempDir()
	_, _ = createTmpGitBareRepo(tmpDir, "test-repo")

	tests := []struct {
		name        string
		opt         model.CloneOption
		setupAuth   *mockAuthService
		wantErr     bool
		expectedErr error
		setup       func(string) error
	}{
		{
			name: "successful clone",
			opt: model.CloneOption{
				URL:         filepath.Join(tmpDir, "temp-test-repo"),
				Mirror:      false,
				NonBareRepo: true,
				Git:         gpsconfig.GitOption{},
			},
			setupAuth: &mockAuthService{
				auth: &http.BasicAuth{
					Username: "test",
					Password: "test",
				},
			},
		},
		{
			name: "auth error",
			opt: model.CloneOption{
				URL: "https://github.com/test/repo.git",
			},
			setupAuth: &mockAuthService{
				err: ErrAuthMethod,
			},
			wantErr:     true,
			expectedErr: ErrAuthMethod,
		},
		{
			name: "clone error - invalid URL",
			opt: model.CloneOption{
				URL: "",
			},
			setupAuth: &mockAuthService{
				auth: &http.BasicAuth{},
			},
			wantErr:     true,
			expectedErr: ErrCloneRepository,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			serv := Service{
				authService: tabletest.setupAuth,
				Ops:         *NewOperation(),
				metadata:    &mockMetadataHandler{},
			}

			repo, err := serv.Clone(context.Background(), tabletest.opt)

			if tabletest.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tabletest.expectedErr)
				require.Empty(t, repo)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, repo)
			}
		})
	}
}

func TestService_Pull(t *testing.T) {
	tmpDir := t.TempDir()
	_, _ = createTmpGitBareRepo(tmpDir, "test-repo")
	_, _ = createTmpGitRepo(tmpDir, "test-repo2", filepath.Join(tmpDir, "temp-test-repo"))

	tests := []struct {
		name        string
		opt         model.PullOption
		setupAuth   *mockAuthService
		wantErr     bool
		expectedErr error
		verifyState func(*testing.T, string)
	}{
		{
			name: "successful pull - up to date",
			opt: model.PullOption{
				URL:       filepath.Join(tmpDir, "temp-test-repo"),
				GitOption: gpsconfig.GitOption{},
			},
			setupAuth: &mockAuthService{
				auth: &http.BasicAuth{
					Username: "test",
					Password: "test",
				},
			},
			verifyState: func(t *testing.T, state string) {
				t.Helper()
				require.Equal(t, "uptodate", state)
			},
		},
		{
			name: "auth error",
			opt: model.PullOption{
				GitOption: gpsconfig.GitOption{},
			},
			setupAuth: &mockAuthService{
				err: ErrAuthMethod,
			},
			wantErr:     true,
			expectedErr: ErrAuthMethod,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			serv := &Service{
				authService: tabletest.setupAuth,
				Ops:         *NewOperation(),
				metadata:    NewMetadataHandler(),
			}

			err := serv.Pull(context.Background(), tabletest.opt, filepath.Join(tmpDir, "temp-test-repo2"))

			if tabletest.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tabletest.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestService_Push(t *testing.T) {
	tmpDir := t.TempDir()
	repo, _ := createTmpGitBareRepo(tmpDir, "test-repo")
	_, _ = createTmpGitRepo(tmpDir, "test-repo2", filepath.Join(tmpDir, "temp-test-repo"))

	tests := []struct {
		name        string
		opt         model.PushOption
		gitOpt      gpsconfig.GitOption
		setupAuth   *mockAuthService
		wantErr     bool
		expectedErr error
		verifyState func(*testing.T, string)
	}{
		{
			name: "successful push - up to date",
			opt: model.PushOption{
				Target:   filepath.Join(tmpDir, "temp-test-repo2"),
				RefSpecs: []string{"refs/heads/main:refs/heads/main"},
			},
			gitOpt: gpsconfig.GitOption{},
			setupAuth: &mockAuthService{
				auth: &http.BasicAuth{
					Username: "test",
					Password: "test",
				},
			},
			verifyState: func(t *testing.T, m string) {
				t.Helper()
				require.Equal(t, "uptodate", m)
			},
		},
		{
			name: "auth error",
			opt: model.PushOption{
				Target: filepath.Join(tmpDir, "tempt-test-repo2"),
			},
			gitOpt: gpsconfig.GitOption{},
			setupAuth: &mockAuthService{
				err: ErrAuthMethod,
			},
			wantErr:     true,
			expectedErr: ErrAuthMethod,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			metadata := &mockMetadataHandler{}
			serv := &Service{
				authService: tabletest.setupAuth,
				Ops:         *NewOperation(),
				metadata:    metadata,
			}

			modelRepo, err := model.NewRepository(repo)
			require.NoError(t, err)

			err = serv.Push(context.Background(), modelRepo, tabletest.opt, tabletest.gitOpt)

			if tabletest.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tabletest.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestService_prepareRepository(t *testing.T) {
	tmpDir := t.TempDir()
	_, _ = createTmpGitBareRepo(tmpDir, "test-repo")
	_, _ = createTmpGitBareRepo(tmpDir, "test-repo-unclean")

	tests := []struct {
		name        string
		setup       func(string) error
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful preparation",
		},
		{
			name:        "repository not found",
			wantErr:     true,
			expectedErr: ErrOpenRepository,
		},
		{
			name: "unclean worktree",
			setup: func(_ string) error {
				return os.WriteFile(
					filepath.Join(tmpDir, "temp-test-repo-unclean", "untracked.txt"),
					[]byte("untracked content"),
					0600,
				)
			},
			wantErr:     true,
			expectedErr: ErrUncleanWorkspace,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			testDir := filepath.Join(tmpDir, "temp-test-repo")
			if tabletest.wantErr {
				testDir = filepath.Join(tmpDir, "a")
			}

			if tabletest.setup != nil {
				testDir = filepath.Join(tmpDir, "temp-test-repo-unclean")
				require.NoError(t, tabletest.setup(testDir))
			}

			serv := NewService()
			repo, worktree, err := serv.prepareRepository(context.Background(), testDir)

			if tabletest.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tabletest.expectedErr)
				require.Nil(t, repo)
				require.Nil(t, worktree)
			} else {
				require.NoError(t, err)
				require.NotNil(t, repo)
				require.NotNil(t, worktree)
			}
		})
	}
}
