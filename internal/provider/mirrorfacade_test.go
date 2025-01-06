// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

//nolint
package provider

import (
	"context"
	"errors"
	"testing"

	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"

	git "github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockGitRemote struct {
	mock.Mock
}

func (m *MockGitRemote) Remote(name string) (model.Remote, error) {
	args := m.Called(name)
	return args.Get(0).(model.Remote), args.Error(1)
}

func (m *MockGitRemote) DeleteRemote(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockGitRemote) CreateRemote(name, url string, fetch bool) error {
	args := m.Called(name, url, fetch)
	return args.Error(0)
}

type MockGitProvider struct {
	mock.Mock
}

func (m *MockGitProvider) Name() string {
	panic("unimplemented")
}

func (m *MockGitProvider) ProjectInfos(_ context.Context, _ model.ProviderOption, _ bool) ([]model.ProjectInfo, error) {
	panic("unimplemented")
}

func (m *MockGitProvider) IsValidProjectName(ctx context.Context, name string) bool {
	args := m.Called(ctx, name)
	return args.Bool(0)
}

func (m *MockGitProvider) ProjectExists(ctx context.Context, owner, repo string) (bool, string) {
	args := m.Called(ctx, owner, repo)
	return args.Bool(0), args.String(1)
}

func (m *MockGitProvider) CreateProject(ctx context.Context, opt model.CreateProjectOption) (string, error) {
	args := m.Called(ctx, opt)
	return args.String(0), args.Error(1)
}

func (m *MockGitProvider) UnprotectProject(ctx context.Context, branch string, projectID string) error {
	args := m.Called(ctx, branch, projectID)
	return args.Error(0)
}

func (m *MockGitProvider) ProtectProject(ctx context.Context, owner string, branch string, projectID string) error {
	args := m.Called(ctx, owner, branch, projectID)
	return args.Error(0)
}

func (m *MockGitProvider) SetDefaultBranch(ctx context.Context, owner string, repo string, branch string) error {
	args := m.Called(ctx, owner, repo, branch)
	return args.Error(0)
}

type MockMirrorWriter struct {
	mock.Mock
}

// Pull implements interfaces.MirrorWriter.
func (m *MockMirrorWriter) Pull(ctx context.Context, opt model.PullOption) error {
	panic("unimplemented")
}

func (m *MockMirrorWriter) Push(ctx context.Context, repo interfaces.GitRepository, opt model.PushOption) error {
	args := m.Called(ctx, repo, opt)
	return args.Error(0)
}

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateRemote(name string, url string, fetch bool) error {
	args := m.Called(name, url, fetch)
	return args.Error(0)
}

func (m *MockRepository) DeleteRemote(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockRepository) GoGitRepository() *git.Repository {
	args := m.Called()
	return args.Get(0).(*git.Repository)
}

func (m *MockRepository) Remote(name string) (model.Remote, error) {
	args := m.Called(name)
	return args.Get(0).(model.Remote), args.Error(1)
}

func (m *MockRepository) ProjectInfo() *model.ProjectInfo {
	args := m.Called()
	return args.Get(0).(*model.ProjectInfo)
}

func (m *MockRepository) Name(ctx context.Context) string {
	args := m.Called(ctx)
	return args.String(0)
}

type testRepository struct {
	projectInfo model.ProjectInfo
	remoteFunc  func(string) (model.Remote, error)
}

func (r testRepository) ProjectInfo() *model.ProjectInfo {
	return &r.projectInfo
}

func (r testRepository) CreateRemote(_ string, _ string, _ bool) error {
	return nil
}

func (r testRepository) DeleteRemote(_ string) error {
	return nil
}

func (r testRepository) GoGitRepository() *git.Repository {
	return nil
}

func (r testRepository) Remote(name string) (model.Remote, error) {
	return r.remoteFunc(name)
}

func TestPush(t *testing.T) {
	ctx := testContext()
	tests := []struct {
		name              string
		mirrorConfig      gpsconfig.MirrorConfig
		syncConfig        gpsconfig.SyncConfig
		setupMocks        func(*MockGitProvider, *MockMirrorWriter, *MockRepository)
		expectedErr       error
		expectedErrString string
	}{
		{
			name: "successful push",
			mirrorConfig: gpsconfig.MirrorConfig{
				BaseConfig: gpsconfig.BaseConfig{
					Owner: "testuser",
				},
				Settings: gpsconfig.MirrorSettings{
					Disabled: false,
				},
			},
			syncConfig: gpsconfig.SyncConfig{
				BaseConfig: gpsconfig.BaseConfig{
					ProviderType: "github",
				},
			},
			setupMocks: func(provider *MockGitProvider, writer *MockMirrorWriter, repo *MockRepository) {
				repo.On("ProjectInfo").Return(&model.ProjectInfo{
					DefaultBranch: "main",
					OriginalName:  "test-repo",
				})
				provider.On("ProjectExists", mock.Anything, "testuser", "test-repo").
					Return(true, "123")
				writer.On("Push", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				provider.On("SetDefaultBranch", mock.Anything, "testuser", "test-repo", "main").Return(nil)
			},
		},
		{
			name: "push with protection toggle",
			mirrorConfig: gpsconfig.MirrorConfig{
				BaseConfig: gpsconfig.BaseConfig{
					Owner: "testuser",
				},
				Settings: gpsconfig.MirrorSettings{
					Disabled: true,
				},
			},
			setupMocks: func(provider *MockGitProvider, writer *MockMirrorWriter, repo *MockRepository) {
				repo.On("ProjectInfo").Return(&model.ProjectInfo{
					DefaultBranch: "main",
					OriginalName:  "test-repo",
				})
				provider.On("ProjectExists", mock.Anything, "testuser", "test-repo").
					Return(true, "123")
				provider.On("UnprotectProject", mock.Anything, "main", "123").Return(nil)
				writer.On("Push", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				provider.On("SetDefaultBranch", mock.Anything, "testuser", "test-repo", "main").Return(nil)
				provider.On("ProtectProject", mock.Anything, "testuser", "main", "123").Return(nil)
			},
		},
		{
			name: "push failure",
			mirrorConfig: gpsconfig.MirrorConfig{
				BaseConfig: gpsconfig.BaseConfig{
					Owner: "testuser",
				},
			},
			setupMocks: func(provider *MockGitProvider, writer *MockMirrorWriter, repo *MockRepository) {
				repo.On("ProjectInfo").Return(&model.ProjectInfo{
					DefaultBranch: "main",
					OriginalName:  "test-repo",
				})
				provider.On("ProjectExists", mock.Anything, "testuser", "test-repo").
					Return(true, "123")
				writer.On("Push", mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("push failed"))
			},
			expectedErr:       ErrPushChanges,
			expectedErrString: "push failed",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			require := require.New(t)
			provider := new(MockGitProvider)
			writer := new(MockMirrorWriter)
			repo := new(MockRepository)
			tabletest.setupMocks(provider, writer, repo)

			err := Push(ctx, tabletest.syncConfig, tabletest.mirrorConfig, provider, writer, repo)

			if tabletest.expectedErr != nil {
				require.Error(err)
				require.ErrorIs(err, tabletest.expectedErr)

				if tabletest.expectedErrString != "" {
					require.Contains(err.Error(), tabletest.expectedErrString)
				}
			} else {
				require.NoError(err)
			}

			provider.AssertExpectations(t)
			writer.AssertExpectations(t)
			repo.AssertExpectations(t)
		})
	}
}

func TestGetPushOption(t *testing.T) {
	ctx := testContext()
	tests := []struct {
		name         string
		mirrorConfig gpsconfig.MirrorConfig
		repository   testRepository
		forcePush    bool
		want         model.PushOption
	}{
		{
			name: "archive provider type",
			mirrorConfig: gpsconfig.MirrorConfig{
				BaseConfig: gpsconfig.BaseConfig{
					ProviderType: "archive",
				},
				Path: "/archive/path",
			},
			repository: testRepository{
				projectInfo: model.ProjectInfo{
					OriginalName: "test-repo",
				},
			},
			want: model.PushOption{
				Target:  "/archive/path/test-repo",
				Force:   false,
				AuthCfg: gpsconfig.AuthConfig{},
			},
		},
		{
			name: "git provider with force push",
			mirrorConfig: gpsconfig.MirrorConfig{
				BaseConfig: gpsconfig.BaseConfig{
					ProviderType: "github",
					Domain:       "github.com",
					Owner:        "testuser",
					Auth: gpsconfig.AuthConfig{
						Token: "test-token",
					},
				},
			},
			repository: testRepository{
				projectInfo: model.ProjectInfo{
					OriginalName: "test-repo",
				},
			},
			forcePush: true,
			want: model.PushOption{
				Target:  "https://any:test-token@github.com/testuser/test-repo",
				Force:   true,
				AuthCfg: gpsconfig.AuthConfig{Token: "test-token"},
			},
		},
		{
			name: "domain with trailing slash",
			mirrorConfig: gpsconfig.MirrorConfig{
				BaseConfig: gpsconfig.BaseConfig{
					ProviderType: "github",
					Domain:       "github.com/",
					Owner:        "testuser",
				},
			},
			repository: testRepository{
				projectInfo: model.ProjectInfo{
					OriginalName: "test-repo",
				},
			},
			want: model.PushOption{
				Target:  "https://any:@github.com/testuser/test-repo",
				Force:   false,
				AuthCfg: gpsconfig.AuthConfig{},
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			require := require.New(t)
			result := getPushOption(ctx, tabletest.mirrorConfig, tabletest.repository, tabletest.forcePush)
			require.Contains(result.Target, tabletest.want.Target)
			require.Equal(tabletest.want.Force, result.Force)
			require.Equal(tabletest.want.AuthCfg, result.AuthCfg)
		})
	}
}

type testGitProvider struct {
	createProjectFunc func(context.Context, model.CreateProjectOption) (string, error)
}

func (t testGitProvider) Name() string {
	panic("unimplemented")
}

func (t testGitProvider) ProjectInfos(_ context.Context, _ model.ProviderOption, _ bool) ([]model.ProjectInfo, error) {
	panic("unimplemented")
}

func (t testGitProvider) IsValidProjectName(_ context.Context, _ string) bool {
	return true
}

func (t testGitProvider) ProjectExists(_ context.Context, _ string, _ string) (bool, string) {
	return true, "123"
}

func (t testGitProvider) CreateProject(ctx context.Context, opt model.CreateProjectOption) (string, error) {
	return t.createProjectFunc(ctx, opt)
}

func (t testGitProvider) ProtectProject(_ context.Context, _ string, _ string, _ string) error {
	return nil
}

func (t testGitProvider) SetDefaultBranch(_ context.Context, _ string, _ string, _ string) error {
	return nil
}

func (t testGitProvider) UnprotectProject(_ context.Context, _ string, _ string) error {
	return nil
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name               string
		mirrorConfig       gpsconfig.MirrorConfig
		sourceProviderType string
		repository         testRepository
		provider           testGitProvider
		wantProjectID      string
		wantErr            bool
		expectedError      error
		expectedErrMsg     string
	}{
		{
			name: "successful creation",
			mirrorConfig: gpsconfig.MirrorConfig{
				BaseConfig: gpsconfig.BaseConfig{
					ProviderType: "gitlab",
				},
				Settings: gpsconfig.MirrorSettings{
					Visibility:        "private",
					DescriptionPrefix: "test description",
				},
			},
			repository: testRepository{
				projectInfo: model.ProjectInfo{
					OriginalName:  "test-repo",
					DefaultBranch: "main",
					Description:   "repo description",
				},
				remoteFunc: func(string) (model.Remote, error) {
					return model.Remote{URL: "https://github.com/original/repo.git"}, nil
				},
			},
			provider: testGitProvider{
				createProjectFunc: func(context.Context, model.CreateProjectOption) (string, error) {
					return "123", nil
				},
			},
			wantProjectID: "123",
		},
		{
			name: "missing gpsupstream remote",
			mirrorConfig: gpsconfig.MirrorConfig{
				BaseConfig: gpsconfig.BaseConfig{
					ProviderType: "gitlab",
				},
			},
			repository: testRepository{
				remoteFunc: func(string) (model.Remote, error) {
					return model.Remote{}, errors.New("remote not found")
				},
			},
			wantErr:        true,
			expectedErrMsg: "failed to get gpsupstream remote",
		},
		{
			name: "empty gpsupstream URL",
			mirrorConfig: gpsconfig.MirrorConfig{
				BaseConfig: gpsconfig.BaseConfig{
					ProviderType: "gitlab",
				},
			},
			repository: testRepository{
				remoteFunc: func(string) (model.Remote, error) {
					return model.Remote{URL: ""}, nil
				},
			},
			wantErr:        true,
			expectedErrMsg: "failed to get gpsupstream remote",
		},
		{
			name: "create project failure",
			mirrorConfig: gpsconfig.MirrorConfig{
				BaseConfig: gpsconfig.BaseConfig{
					ProviderType: "gitlab",
				},
				Settings: gpsconfig.MirrorSettings{
					Visibility: "private",
				},
			},
			repository: testRepository{
				projectInfo: model.ProjectInfo{
					OriginalName:  "test-repo",
					DefaultBranch: "main",
				},
				remoteFunc: func(string) (model.Remote, error) {
					return model.Remote{URL: "https://github.com/test/repo.git"}, nil
				},
			},
			provider: testGitProvider{
				createProjectFunc: func(context.Context, model.CreateProjectOption) (string, error) {
					return "", errors.New("creation failed")
				},
			},
			wantErr:        true,
			expectedError:  ErrCreateRepository,
			expectedErrMsg: "creation failed",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			require := require.New(t)

			projectID, err := create(testContext(), tabletest.mirrorConfig, tabletest.provider, tabletest.sourceProviderType, tabletest.repository)

			if tabletest.wantErr {
				require.Error(err)

				if tabletest.expectedError != nil {
					require.ErrorIs(err, tabletest.expectedError)
				}

				if tabletest.expectedErrMsg != "" {
					require.Contains(err.Error(), tabletest.expectedErrMsg)
				}
			} else {
				require.NoError(err)
				require.Equal(tabletest.wantProjectID, projectID)
			}
		})
	}
}

func TestSetGPSUpstreamRemoteFromOrigin(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*MockGitRemote)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful setup",
			setupMock: func(m *MockGitRemote) {
				m.On("Remote", "origin").Return(model.Remote{URL: "git@github.com:test/repo.git"}, nil)
				m.On("DeleteRemote", "gpsupstream").Return(nil)
				m.On("CreateRemote", "gpsupstream", "git@github.com:test/repo.git", true).Return(nil)
				m.On("Remote", "gpsupstream").Return(model.Remote{URL: "git@github.com:test/repo.git"}, nil)
			},
		},
		{
			name: "origin remote error",
			setupMock: func(m *MockGitRemote) {
				m.On("Remote", "origin").Return(model.Remote{}, errors.New("origin not found"))
			},
			wantErr: true,
			errMsg:  "failed to get origin remote",
		},
		{
			name: "delete remote error",
			setupMock: func(m *MockGitRemote) {
				m.On("Remote", "origin").Return(model.Remote{URL: "git@github.com:test/repo.git"}, nil)
				m.On("DeleteRemote", "gpsupstream").Return(errors.New("delete failed"))
			},
			wantErr: true,
			errMsg:  "failed to delete gpsupstream remote",
		},
		{
			name: "create remote error",
			setupMock: func(m *MockGitRemote) {
				m.On("Remote", "origin").Return(model.Remote{URL: "git@github.com:test/repo.git"}, nil)
				m.On("DeleteRemote", "gpsupstream").Return(nil)
				m.On("CreateRemote", "gpsupstream", "git@github.com:test/repo.git", true).Return(errors.New("create failed"))
			},
			wantErr: true,
			errMsg:  "failed to create gpsupstream remote",
		},
		{
			name: "url mismatch error",
			setupMock: func(m *MockGitRemote) {
				m.On("Remote", "origin").Return(model.Remote{URL: "git@github.com:test/repo.git"}, nil)
				m.On("DeleteRemote", "gpsupstream").Return(nil)
				m.On("CreateRemote", "gpsupstream", "git@github.com:test/repo.git", true).Return(nil)
				m.On("Remote", "gpsupstream").Return(model.Remote{URL: "different-url"}, nil)
			},
			wantErr: true,
			errMsg:  "mismatch in gpsupstream vs origin remote",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			require := require.New(t)
			remote := new(MockGitRemote)
			tabletest.setupMock(remote)

			err := SetGPSUpstreamRemoteFromOrigin(context.Background(), remote)

			if tabletest.wantErr {
				require.Error(err)
				require.Contains(err.Error(), tabletest.errMsg)
			} else {
				require.NoError(err)
			}

			remote.AssertExpectations(t)
		})
	}
}

func TestIsArchiveOrDirectory(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		want     bool
	}{
		{"archive provider", "archive", true},
		{"directory provider", "directory", true},
		{"git provider", "gitlab", false},
		{"empty provider", "", false},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			require := require.New(t)
			result := isArchiveOrDirectory(tabletest.provider)
			require.Equal(tabletest.want, result)
		})
	}
}

func TestGetProjectPath(t *testing.T) {
	tests := []struct {
		name           string
		config         gpsconfig.MirrorConfig
		repositoryName string
		want           string
	}{
		{
			name: "owner path",
			config: gpsconfig.MirrorConfig{
				BaseConfig: gpsconfig.BaseConfig{
					Owner: "test-owner",
				},
			},
			repositoryName: "repo",
			want:           "test-owner/repo",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			require := require.New(t)
			result := getProjectPath(tabletest.repositoryName, tabletest.config)
			require.Equal(tabletest.want, result)
		})
	}
}

func TestBuildDescription(t *testing.T) {
	tests := []struct {
		name            string
		remote          model.Remote
		repository      interfaces.GitRepository
		userDescription string
		want            string
	}{
		{
			name:   "with user description",
			remote: model.Remote{URL: "git@github.com:test/repo.git"},
			repository: &testRepository{
				projectInfo: model.ProjectInfo{Description: "repo description"},
			},
			userDescription: "custom description",
			want:            "custom descriptionrepo description",
		},
		{
			name:   "without user description",
			remote: model.Remote{URL: "git@github.com:test/repo.git"},
			repository: &testRepository{
				projectInfo: model.ProjectInfo{Description: "repo description"},
			},
			want: "Git Provider Sync cloned this from: git@github.com:test/repo.git: repo description",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			require := require.New(t)
			result := buildDescription(tabletest.userDescription, tabletest.remote, tabletest.repository)
			require.Equal(tabletest.want, result)
		})
	}
}
