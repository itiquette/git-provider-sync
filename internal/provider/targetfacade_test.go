// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package provider

import (
	"context"
	"errors"
	mocks "itiquette/git-provider-sync/generated/mocks/mockgogit"
	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
	"regexp"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPush(t *testing.T) {
	require := require.New(t)
	ctx := testContext()

	tests := map[string]struct {
		config     config.ProviderConfig
		repository interfaces.GitRepository
		mockSetup  func(*mocks.GitProvider, *mocks.GitRepository)
		wantErr    bool
	}{
		"push success": {
			config: targetProviderConfig(),
			mockSetup: func(mockClient *mocks.GitProvider, mockRepo *mocks.GitRepository) {
				mockClient.EXPECT().Metainfos(mock.Anything, mock.Anything, mock.Anything).Return([]model.RepositoryMetainfo{{HTTPSURL: "https://url.c"}}, nil)
				mockClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(nil)
				mockRepo.EXPECT().Metainfo().Return(model.RepositoryMetainfo{
					OriginalName:  "nasename",
					HTTPSURL:      "http://a.url",
					DefaultBranch: "defbranch",
					Description:   "desc",
					Visibility:    "public",
				})
				mockRepo.EXPECT().Remote(mock.Anything).Return(model.Remote{URL: "https://up.url"}, nil)
			},
		},
		// "repository exists error": {
		//     config: targetProviderConfig(),
		//     mockSetup: func(mockClient *mocks.GitProvider, mockRepo *mocks.GitRepository) {
		//         mockClient.EXPECT().Metainfos(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("failed to check repository"))
		//         mockRepo.EXPECT().Metainfo().Return(model.RepositoryMetainfo{})
		//         mockRepo.EXPECT().Remote(mock.Anything).Return(model.Remote{URL: "https://up.url"}, errors.New("an err"))
		//     },
		//     wantErr: true,
		// },
		// "push error": {
		//     config: targetProviderConfig(),
		//     mockSetup: func(mockClient *mocks.GitProvider, mockRepo *mocks.GitRepository) {
		//         mockClient.EXPECT().Metainfos(mock.Anything, mock.Anything, mock.Anything).Return([]model.RepositoryMetainfo{{HTTPSURL: "https://url.c"}}, nil)
		//         mockRepo.EXPECT().Metainfo().Return(model.RepositoryMetainfo{})
		//         mockRepo.EXPECT().Remote(mock.Anything).Return(model.Remote{URL: "https://up.url"}, nil)
		//     },
		//     wantErr: true,
		// },
	}

	for name, tabletest := range tests {
		t.Run(name, func(t *testing.T) {
			mockClient := new(mocks.GitProvider)
			mockRepo := new(mocks.GitRepository)
			tabletest.mockSetup(mockClient, mockRepo)

			sourceGitOption := config.GitOption{}

			err := Push(ctx, tabletest.config, mockClient, mockGitCore{}, mockRepo, sourceGitOption)

			if tabletest.wantErr {
				require.Error(err)
			} else {
				require.NoError(err)
			}

			mockClient.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
		})
	}
}
func TestGetPushOption(t *testing.T) {
	require := require.New(t)
	tests := map[string]struct {
		config     config.ProviderConfig
		repository *mocks.GitRepository
		forcePush  bool
		expected   model.PushOption
	}{
		"archive provider": {
			config: config.ProviderConfig{
				ProviderType: config.ARCHIVE,
				Additional: map[string]string{
					"archivetargetdir": "/archive",
				},
			},
			repository: mockRepositoryWithName("repo"),
			forcePush:  false,
			expected:   model.NewPushOption("/archive/repo.tar.gz", false, false, config.HTTPClientOption{}),
		},
		"directory provider": {
			config: config.ProviderConfig{
				ProviderType: config.DIRECTORY,
				Additional: map[string]string{
					"directorytargetdir": "/target",
				},
			},
			repository: mockRepositoryWithName("repo"),
			forcePush:  false,
			expected:   model.NewPushOption("/target", false, false, config.HTTPClientOption{}),
		},
		"git provider with force push": {
			config: config.ProviderConfig{
				ProviderType: "gitlab",
				Domain:       "gitlab.com",
				HTTPClient:   config.HTTPClientOption{Token: "token"},
				User:         "user",
			},
			repository: mockRepositoryWithName("repo"),
			forcePush:  true,
			expected:   model.NewPushOption("https://gitlab.com/user/repo", false, true, config.HTTPClientOption{}),
		},
	}

	for name, tabletests := range tests {
		t.Run(name, func(_ *testing.T) {
			ctx := context.Background()
			ctx = model.WithCLIOption(ctx, model.CLIOption{})
			result := getPushOption(ctx, tabletests.config, tabletests.repository, tabletests.forcePush)
			require.Equal(tabletests.expected.Target, removeTimestamp(result.Target))
		})
	}
}
func TestBuildDescription(t *testing.T) {
	tests := map[string]struct {
		remote     model.Remote
		repository *mocks.GitRepository
		expected   string
	}{
		"with description": {
			remote: model.Remote{URL: "https://example.com/repo.git"},
			repository: func() *mocks.GitRepository {
				r := new(mocks.GitRepository)
				r.On("Metainfo").Return(model.RepositoryMetainfo{Description: "Test repo"})

				return r
			}(),
			expected: "Git Provider Sync cloned this from: https://example.com/repo.git: Test repo",
		},
		"without description": {
			remote: model.Remote{URL: "https://example.com/repo.git"},
			repository: func() *mocks.GitRepository {
				r := new(mocks.GitRepository)
				r.EXPECT().Metainfo().Return(model.RepositoryMetainfo{})

				return r
			}(),
			expected: "Git Provider Sync cloned this from: https://example.com/repo.git: ",
		},
	}

	for name, tabletest := range tests {
		t.Run(name, func(t *testing.T) {
			result := buildDescription(tabletest.remote, tabletest.repository, nil)
			require.Equal(t, tabletest.expected, result)
		})
	}
}

func TestIsArchiveOrDirectory(t *testing.T) {
	tests := map[string]struct {
		provider string
		expected bool
	}{
		"archive provider":   {provider: config.ARCHIVE, expected: true},
		"directory provider": {provider: config.DIRECTORY, expected: true},
		"git provider":       {provider: "gitlab", expected: false},
		"case insensitive":   {provider: "ArChIvE", expected: true},
		"empty string":       {provider: "", expected: false},
	}

	for name, tabletest := range tests {
		t.Run(name, func(t *testing.T) {
			result := isArchiveOrDirectory(tabletest.provider)
			require.Equal(t, tabletest.expected, result)
		})
	}
}

func TestRepositoryExists(t *testing.T) {
	tests := map[string]struct {
		metainfos      []model.RepositoryMetainfo
		repositoryName string
		expected       bool
	}{
		"repository exists": {
			metainfos: []model.RepositoryMetainfo{
				{OriginalName: "repo1"},
				{OriginalName: "repo2"},
			},
			repositoryName: "repo1",
			expected:       true,
		},
		"repository does not exist": {
			metainfos: []model.RepositoryMetainfo{
				{OriginalName: "repo1"},
				{OriginalName: "repo2"},
			},
			repositoryName: "repo3",
			expected:       false,
		},
		"case insensitive": {
			metainfos: []model.RepositoryMetainfo{
				{OriginalName: "Repo1"},
			},
			repositoryName: "repo1",
			expected:       true,
		},
	}

	for name, tabletest := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockProvider := new(mocks.GitProvider)
			mockProvider.EXPECT().Metainfos(ctx, mock.Anything, false).Return(tabletest.metainfos, nil)

			result := repositoryExists(ctx, config.ProviderConfig{}, mockProvider, tabletest.repositoryName)
			require.Equal(t, tabletest.expected, result)
		})
	}
}

func TestGetProjectPath(t *testing.T) {
	tests := map[string]struct {
		config         config.ProviderConfig
		repositoryName string
		expected       string
	}{
		"group repository": {
			config:         config.ProviderConfig{Group: "mygroup"},
			repositoryName: "repo",
			expected:       "mygroup/repo",
		},
		"user repository": {
			config:         config.ProviderConfig{User: "myuser"},
			repositoryName: "repo",
			expected:       "myuser/repo",
		},
	}

	for name, tabletest := range tests {
		t.Run(name, func(t *testing.T) {
			result := getProjectPath(tabletest.config, tabletest.repositoryName)
			require.Equal(t, tabletest.expected, result)
		})
	}
}

func mockRepositoryWithName(name string) *mocks.GitRepository {
	r := new(mocks.GitRepository)
	r.EXPECT().Metainfo().Return(model.RepositoryMetainfo{OriginalName: name})

	return r
}

func removeTimestamp(input string) string {
	re := regexp.MustCompile(`_\d{8}_\d{6}_\d{13}`)

	return re.ReplaceAllString(input, "")
}

func testContext() context.Context {
	ctx := context.Background()
	input := model.CLIOption{CleanupName: true}
	ctx, _ = model.CreateTmpDir(ctx, "", "testadir")

	return model.WithCLIOption(ctx, input)
}
func targetProviderConfig() config.ProviderConfig {
	return config.ProviderConfig{Group: "d", Domain: "https://a.gitprovider.com", ProviderType: "gitlab", HTTPClient: config.HTTPClientOption{Token: "s"}}
}

type mockGitCore struct{}

func (mockGitCore) Clone(_ context.Context, _ model.CloneOption) (model.Repository, error) {
	return model.Repository{}, nil
}

func (mockGitCore) Push(_ context.Context, _ model.PushOption, _ config.GitOption, _ config.GitOption) error {
	return nil
}

var ErrTest = errors.New("testerr")

const (
	TESTDOMAIN = "test.se"
	TESTUSER   = "user"
	BASENAME   = "basename"
)
