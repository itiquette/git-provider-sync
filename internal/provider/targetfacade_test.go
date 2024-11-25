// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package provider

import (
	"context"
	"errors"
	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
	"testing"

	git "github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockGitRemote struct {
	mock.Mock
}

func (m *MockGitRemote) Remote(name string) (model.Remote, error) {
	args := m.Called(name)
	return args.Get(0).(model.Remote), args.Error(1) //nolint
}

func (m *MockGitRemote) DeleteRemote(name string) error {
	args := m.Called(name)

	return args.Error(0) //nolint
}

func (m *MockGitRemote) CreateRemote(name, url string, fetch bool) error {
	args := m.Called(name, url, fetch)

	return args.Error(0) //nolint
}

type MockGitProvider struct {
	mock.Mock
}

// IsValidProjectName implements interfaces.GitProvider.
func (m *MockGitProvider) IsValidProjectName(_ context.Context, _ string) bool {
	panic("unimplemented")
}

func (m *MockGitProvider) Name() string {
	args := m.Called()

	return args.String(0)
}

func (m *MockGitProvider) ProjectInfos(ctx context.Context, cfg config.ProviderConfig, full bool) ([]model.ProjectInfo, error) {
	args := m.Called(ctx, cfg, full)

	return args.Get(0).([]model.ProjectInfo), args.Error(1) //nolint
}

func (m *MockGitProvider) CreateProject(ctx context.Context, cfg config.ProviderConfig, opt model.CreateProjectOption) (string, error) {
	args := m.Called(ctx, cfg, opt)

	return args.String(0), args.Error(1)
}

func (m *MockGitProvider) UnprotectProject(ctx context.Context, branch string, projectID string) error {
	args := m.Called(ctx, branch, projectID)

	return args.Error(0) //nolint
}

func (m *MockGitProvider) ProtectProject(ctx context.Context, owner string, branch string, projectID string) error {
	args := m.Called(ctx, owner, branch, projectID)

	return args.Error(0) //nolint
}

func (m *MockGitProvider) SetDefaultBranch(ctx context.Context, owner string, repo string, branch string) error {
	args := m.Called(ctx, owner, repo, branch)

	return args.Error(0) //nolint
}

type MockTargetWriter struct {
	mock.Mock
}

func (m *MockTargetWriter) Push(ctx context.Context, repo interfaces.GitRepository, opt model.PushOption, git config.GitOption) error {
	args := m.Called(ctx, repo, opt, git)

	return args.Error(0) //nolint
}
func TestPush(t *testing.T) {
	ctx := testContext()
	tests := []struct {
		name              string
		targetConfig      config.ProviderConfig
		sourceConfig      config.ProviderConfig
		setupMocks        func(*MockGitProvider, *MockTargetWriter, *MockRepository)
		expectedErr       error
		expectedErrString string
	}{
		{
			name: "successful push",
			targetConfig: config.ProviderConfig{
				User: "testuser",
				Project: config.ProjectOption{
					Disabled: false,
				},
			},
			sourceConfig: config.ProviderConfig{
				ProviderType: "github",
			},
			setupMocks: func(provider *MockGitProvider, writer *MockTargetWriter, repo *MockRepository) {
				repo.On("ProjectInfo").Return(model.ProjectInfo{
					DefaultBranch: "main",
					OriginalName:  "test-repo",
				})
				provider.On("ProjectInfos", mock.Anything, mock.Anything, false).
					Return([]model.ProjectInfo{{ProjectID: "123", OriginalName: "test-repo"}}, nil)
				writer.On("Push", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				provider.On("SetDefaultBranch", mock.Anything, "testuser", mock.Anything, "main").Return(nil)
			},
		},
		{
			name: "push with protection toggle",
			targetConfig: config.ProviderConfig{
				User: "testuser",
				Project: config.ProjectOption{
					Disabled: true,
				},
			},
			setupMocks: func(provider *MockGitProvider, writer *MockTargetWriter, repo *MockRepository) {
				repo.On("ProjectInfo").Return(model.ProjectInfo{
					DefaultBranch: "main",
					OriginalName:  "test-repo",
				})
				provider.On("ProjectInfos", mock.Anything, mock.Anything, false).
					Return([]model.ProjectInfo{{ProjectID: "123", OriginalName: "test-repo"}}, nil)
				provider.On("UnprotectProject", mock.Anything, "main", "123").Return(nil)
				writer.On("Push", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				provider.On("SetDefaultBranch", mock.Anything, "testuser", mock.Anything, "main").Return(nil)
				provider.On("ProtectProject", mock.Anything, "testuser", "main", "123").Return(nil)
			},
		},
		{
			name: "push failure",
			targetConfig: config.ProviderConfig{
				User: "testuser",
			},
			setupMocks: func(provider *MockGitProvider, writer *MockTargetWriter, repo *MockRepository) {
				repo.On("ProjectInfo").Return(model.ProjectInfo{
					DefaultBranch: "main",
					OriginalName:  "test-repo",
				})
				provider.On("ProjectInfos", mock.Anything, mock.Anything, false).
					Return([]model.ProjectInfo{{ProjectID: "123", OriginalName: "test-repo"}}, nil)
				writer.On("Push", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
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
			writer := new(MockTargetWriter)
			repo := new(MockRepository)
			tabletest.setupMocks(provider, writer, repo)

			err := Push(ctx, tabletest.targetConfig, provider, writer, repo, tabletest.sourceConfig)

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

// MockRepository for testing.
type MockRepository struct {
	mock.Mock
}

// CreateRemote implements interfaces.GitRepository.
func (m *MockRepository) CreateRemote(_ string, _ string, _ bool) error {
	panic("unimplemented")
}

// DeleteRemote implements interfaces.GitRepository.
func (m *MockRepository) DeleteRemote(_ string) error {
	panic("unimplemented")
}

// GoGitRepository implements interfaces.GitRepository.
func (m *MockRepository) GoGitRepository() *git.Repository {
	panic("unimplemented")
}

// Remote implements interfaces.GitRepository.
func (m *MockRepository) Remote(_ string) (model.Remote, error) {
	panic("unimplemented")
}

func (m *MockRepository) ProjectInfo() model.ProjectInfo {
	args := m.Called()

	return args.Get(0).(model.ProjectInfo) //nolint
}

func (m *MockRepository) Name(ctx context.Context) string {
	args := m.Called(ctx)

	return args.String(0)
}

// Simple test repository implementation.
type testRepository struct {
	projectInfo model.ProjectInfo
	remoteFunc  func(string) (model.Remote, error)
}

// CreateRemote implements interfaces.GitRepository.
func (r testRepository) CreateRemote(_ string, _ string, _ bool) error {
	panic("unimplemented")
}

// DeleteRemote implements interfaces.GitRepository.
func (r testRepository) DeleteRemote(_ string) error {
	panic("unimplemented")
}

// GoGitRepository implements interfaces.GitRepository.
func (r testRepository) GoGitRepository() *git.Repository {
	panic("unimplemented")
}

func (r testRepository) ProjectInfo() model.ProjectInfo {
	return r.projectInfo
}

func (r testRepository) Remote(name string) (model.Remote, error) {
	return r.remoteFunc(name)
}

func TestGetPushOption(t *testing.T) {
	ctx := testContext()
	tests := []struct {
		name           string
		ctx            context.Context //nolint
		providerConfig config.ProviderConfig
		repository     testRepository
		forcePush      bool
		want           model.PushOption
	}{
		{
			name: "archive provider type",
			providerConfig: config.ProviderConfig{
				ProviderType: config.ARCHIVE,
				Additional: map[string]string{
					"archivetargetdir": "/archive/path",
				},
			},
			repository: testRepository{
				projectInfo: model.ProjectInfo{
					OriginalName: "test-repo",
				},
			},
			forcePush: false,
			want: model.PushOption{
				Target:     "/archive/path/test-repo",
				Force:      false,
				HTTPClient: config.HTTPClientOption{},
			},
		},
		{
			name: "directory provider type",
			providerConfig: config.ProviderConfig{
				ProviderType: config.DIRECTORY,
				Additional: map[string]string{
					"directorytargetdir": "/target/directory",
				},
			},
			repository: testRepository{
				projectInfo: model.ProjectInfo{
					OriginalName: "test-repo",
				},
			},
			forcePush: false,
			want: model.PushOption{
				Target:     "/target/directory",
				Force:      false,
				HTTPClient: config.HTTPClientOption{},
			},
		},
		{
			name: "git provider with force push",
			providerConfig: config.ProviderConfig{
				ProviderType: "github",
				Domain:       "github.com",
				User:         "testuser",
				HTTPClient: config.HTTPClientOption{
					Token: "test-token",
				},
			},
			repository: testRepository{
				projectInfo: model.ProjectInfo{
					OriginalName: "test-repo",
				},
			},
			forcePush: true,
			want: model.PushOption{
				Target:     "https://github.com/testuser/test-repo",
				Force:      true,
				HTTPClient: config.HTTPClientOption{Token: "test-token"},
			},
		},
		{
			name: "git provider with custom scheme",
			providerConfig: config.ProviderConfig{
				ProviderType: "gitlab",
				Domain:       "gitlab.com",
				User:         "testuser",
				HTTPClient: config.HTTPClientOption{
					Scheme: "git",
				},
			},
			repository: testRepository{
				projectInfo: model.ProjectInfo{
					OriginalName: "test-repo",
				},
			},
			forcePush: false,
			want: model.PushOption{
				Target:     "git://gitlab.com/testuser/test-repo",
				Force:      false,
				HTTPClient: config.HTTPClientOption{Scheme: "git"},
			},
		},
		{
			name: "with group instead of user",
			providerConfig: config.ProviderConfig{
				ProviderType: "github",
				Domain:       "github.com",
				Group:        "testgroup",
			},
			repository: testRepository{
				projectInfo: model.ProjectInfo{
					OriginalName: "test-repo",
				},
			},
			forcePush: false,
			want: model.PushOption{
				Target:     "https://github.com/testgroup/test-repo",
				Force:      false,
				HTTPClient: config.HTTPClientOption{},
			},
		},
		{
			name: "domain with trailing slash",
			providerConfig: config.ProviderConfig{
				ProviderType: "github",
				Domain:       "github.com/",
				User:         "testuser",
			},
			repository: testRepository{
				projectInfo: model.ProjectInfo{
					OriginalName: "test-repo",
				},
			},
			forcePush: false,
			want: model.PushOption{
				Target:     "https://github.com/testuser/test-repo",
				Force:      false,
				HTTPClient: config.HTTPClientOption{},
			},
		},
		{
			name: "empty provider type defaults to git URL",
			providerConfig: config.ProviderConfig{
				Domain: "github.com",
				User:   "testuser",
			},
			repository: testRepository{
				projectInfo: model.ProjectInfo{
					OriginalName: "test-repo",
				},
			},
			forcePush: false,
			want: model.PushOption{
				Target:     "https://github.com/testuser/test-repo",
				Force:      false,
				HTTPClient: config.HTTPClientOption{},
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			require := require.New(t)

			result := getPushOption(ctx, tabletest.providerConfig, tabletest.repository, tabletest.forcePush)

			if tabletest.providerConfig.ProviderType == config.ARCHIVE {
				require.Contains(result.Target, tabletest.want.Target)
			} else if tabletest.providerConfig.ProviderType == config.DIRECTORY {
				require.Contains(result.Target, tabletest.want.Target)
			} else {
				require.Equal(tabletest.want.Target, result.Target, "URLs should match")
			}

			require.Equal(tabletest.want.Force, result.Force, "Force flags should match")
			require.Equal(tabletest.want.HTTPClient, result.HTTPClient, "HTTP client options should match")
		})
	}
}

// Test implementations.
type testGitProvider struct {
	createProjectFunc func(context.Context, config.ProviderConfig, model.CreateProjectOption) (string, error)
}

// IsValidProjectName implements interfaces.GitProvider.
func (t testGitProvider) IsValidProjectName(_ context.Context, _ string) bool {
	panic("unimplemented")
}

// Name implements interfaces.GitProvider.
func (t testGitProvider) Name() string {
	panic("unimplemented")
}

// ProjectInfos implements interfaces.GitProvider.
func (t testGitProvider) ProjectInfos(_ context.Context, _ config.ProviderConfig, _ bool) ([]model.ProjectInfo, error) {
	panic("unimplemented")
}

// ProtectProject implements interfaces.GitProvider.
func (t testGitProvider) ProtectProject(_ context.Context, _ string, _ string, _ string) error {
	panic("unimplemented")
}

// SetDefaultBranch implements interfaces.GitProvider.
func (t testGitProvider) SetDefaultBranch(_ context.Context, _ string, _ string, _ string) error {
	panic("unimplemented")
}

// UnprotectProject implements interfaces.GitProvider.
func (t testGitProvider) UnprotectProject(_ context.Context, _ string, _ string) error {
	panic("unimplemented")
}

func (t testGitProvider) CreateProject(ctx context.Context, cfg config.ProviderConfig, opt model.CreateProjectOption) (string, error) {
	return t.createProjectFunc(ctx, cfg, opt)
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name               string
		targetConfig       config.ProviderConfig
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
			targetConfig: config.ProviderConfig{
				Project: config.ProjectOption{
					Visibility:  "private",
					Description: "test description",
				},
			},
			repository: testRepository{
				projectInfo: model.ProjectInfo{
					OriginalName:  "test-repo",
					DefaultBranch: "main",
					Description:   "repo description",
				},
				remoteFunc: func(_ string) (model.Remote, error) {
					return model.Remote{URL: "https://github.com/original/repo.git"}, nil
				},
			},
			provider: testGitProvider{
				createProjectFunc: func(_ context.Context, _ config.ProviderConfig, _ model.CreateProjectOption) (string, error) {
					return "123", nil
				},
			},
			wantProjectID: "123",
		},
		{
			name:         "missing gpsupstream remote",
			targetConfig: config.ProviderConfig{},
			repository: testRepository{
				remoteFunc: func(_ string) (model.Remote, error) {
					return model.Remote{}, errors.New("remote not found")
				},
			},
			wantErr:        true,
			expectedErrMsg: "failed to get gpsupstream remote",
		},
		{
			name:         "empty gpsupstream URL",
			targetConfig: config.ProviderConfig{},
			repository: testRepository{
				remoteFunc: func(_ string) (model.Remote, error) {
					return model.Remote{URL: ""}, nil
				},
			},
			wantErr:        true,
			expectedErrMsg: "failed to get gpsupstream remote",
		},
		{
			name: "visibility mapping failure",
			targetConfig: config.ProviderConfig{
				Project: config.ProjectOption{
					Visibility: "", // Empty to trigger mapping
				},
			},
			sourceProviderType: "unknown",
			repository: testRepository{
				projectInfo: model.ProjectInfo{
					Visibility: "invalid",
				},
				remoteFunc: func(_ string) (model.Remote, error) {
					return model.Remote{URL: "https://github.com/test/repo.git"}, nil
				},
			},
			wantErr:        true,
			expectedErrMsg: "failed to map visibility",
		},
		{
			name: "create project failure",
			targetConfig: config.ProviderConfig{
				Project: config.ProjectOption{
					Visibility: "private",
				},
			},
			repository: testRepository{
				projectInfo: model.ProjectInfo{
					OriginalName:  "test-repo",
					DefaultBranch: "main",
				},
				remoteFunc: func(_ string) (model.Remote, error) {
					return model.Remote{URL: "https://github.com/test/repo.git"}, nil
				},
			},
			provider: testGitProvider{
				createProjectFunc: func(_ context.Context, _ config.ProviderConfig, _ model.CreateProjectOption) (string, error) {
					return "", errors.New("creation failed")
				},
			},
			wantErr:        true,
			expectedError:  ErrCreateRepository,
			expectedErrMsg: "failed to create repository",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			require := require.New(t)

			projectID, err := create(testContext(), tabletest.targetConfig, tabletest.provider, tabletest.sourceProviderType, tabletest.repository)

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
				m.On("Remote", config.ORIGIN).Return(model.Remote{URL: "git@github.com:test/repo.git"}, nil)
				m.On("DeleteRemote", config.GPSUPSTREAM).Return(nil)
				m.On("CreateRemote", config.GPSUPSTREAM, "git@github.com:test/repo.git", true).Return(nil)
				m.On("Remote", config.GPSUPSTREAM).Return(model.Remote{URL: "git@github.com:test/repo.git"}, nil)
			},
			wantErr: false,
		},
		{
			name: "origin remote error",
			setupMock: func(m *MockGitRemote) {
				m.On("Remote", config.ORIGIN).Return(model.Remote{}, errors.New("origin not found"))
			},
			wantErr: true,
			errMsg:  "failed to get origin remote",
		},
		{
			name: "delete remote error",
			setupMock: func(m *MockGitRemote) {
				m.On("Remote", config.ORIGIN).Return(model.Remote{URL: "git@github.com:test/repo.git"}, nil)
				m.On("DeleteRemote", config.GPSUPSTREAM).Return(errors.New("delete failed"))
			},
			wantErr: true,
			errMsg:  "failed to delete gpsupstream remote",
		},
		{
			name: "create remote error",
			setupMock: func(m *MockGitRemote) {
				m.On("Remote", config.ORIGIN).Return(model.Remote{URL: "git@github.com:test/repo.git"}, nil)
				m.On("DeleteRemote", config.GPSUPSTREAM).Return(nil)
				m.On("CreateRemote", config.GPSUPSTREAM, "git@github.com:test/repo.git", true).Return(errors.New("create failed"))
			},
			wantErr: true,
			errMsg:  "failed to create gpsupstream remote",
		},
		{
			name: "url mismatch error",
			setupMock: func(m *MockGitRemote) {
				m.On("Remote", config.ORIGIN).Return(model.Remote{URL: "git@github.com:test/repo.git"}, nil)
				m.On("DeleteRemote", config.GPSUPSTREAM).Return(nil)
				m.On("CreateRemote", config.GPSUPSTREAM, "git@github.com:test/repo.git", true).Return(nil)
				m.On("Remote", config.GPSUPSTREAM).Return(model.Remote{URL: "different-url"}, nil)
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
		{"archive provider", "ARCHIVE", true},
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
		config         config.ProviderConfig
		repositoryName string
		want           string
	}{
		{
			name: "group path",
			config: config.ProviderConfig{
				Group: "test-group",
				User:  "test-user",
			},
			repositoryName: "repo",
			want:           "test-group/repo",
		},
		{
			name: "user path",
			config: config.ProviderConfig{
				User: "test-user",
			},
			repositoryName: "repo",
			want:           "test-user/repo",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			require := require.New(t)
			result := getProjectPath(tabletest.config, tabletest.repositoryName)
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
			repository: &model.Repository{
				ProjectMetaInfo: model.ProjectInfo{Description: "repo description"},
			},
			userDescription: "custom description",
			want:            "custom descriptionrepo description",
		},
		{
			name:   "without user description",
			remote: model.Remote{URL: "git@github.com:test/repo.git"},
			repository: &model.Repository{
				ProjectMetaInfo: model.ProjectInfo{Description: "repo description"},
			},
			want: "Git Provider Sync cloned this from: git@github.com:test/repo.git: repo description",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			require := require.New(t)
			result := buildDescription(tabletest.remote, tabletest.repository, tabletest.userDescription)
			require.Equal(tabletest.want, result)
		})
	}
}
