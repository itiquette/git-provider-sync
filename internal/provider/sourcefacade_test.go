// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package provider

import (
	"context"
	"errors"
	mocks "itiquette/git-provider-sync/generated/mocks/mockgogit"
	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func testContext() context.Context {
	ctx := context.Background()
	input := model.CLIOption{ASCIIName: true}

	return model.WithCLIOpt(ctx, input)
}

func TestClone(t *testing.T) {
	ctx := testContext()
	tests := []struct {
		name         string
		projectinfos []model.ProjectInfo
		syncCfg      config.SyncConfig
		mockSetup    func(*mocks.SourceReader)
		wantErr      bool
	}{
		{
			name: "successful multiple clone",
			projectinfos: []model.ProjectInfo{
				{HTTPSURL: "https://github.com/user/repo1.git", OriginalName: "repo1"},
				{HTTPSURL: "https://github.com/user/repo2.git", OriginalName: "repo2"},
			},
			syncCfg: config.SyncConfig{
				BaseConfig: config.BaseConfig{},
			},
			mockSetup: func(srcR *mocks.SourceReader) {
				srcR.EXPECT().Clone(mock.Anything, mock.Anything).Return(model.Repository{}, nil).Twice()
			},
		},
		{
			name: "clone failure",
			projectinfos: []model.ProjectInfo{
				{HTTPSURL: "https://github.com/user/repo1.git", OriginalName: "repo1"},
			},
			syncCfg: config.SyncConfig{
				BaseConfig: config.BaseConfig{},
			},
			mockSetup: func(srcR *mocks.SourceReader) {
				srcR.EXPECT().Clone(mock.Anything, mock.Anything).Return(model.Repository{}, errors.New("clone failed"))
			},
			wantErr: true,
		},
		{
			name:         "empty projectinfos",
			projectinfos: []model.ProjectInfo{},
			syncCfg: config.SyncConfig{
				BaseConfig: config.BaseConfig{},
			},
			mockSetup: func(*mocks.SourceReader) {},
		},
		{
			name: "clone with ascii name setting",
			projectinfos: []model.ProjectInfo{
				{HTTPSURL: "https://github.com/user/repo1.git", OriginalName: "repo1"},
			},
			syncCfg: config.SyncConfig{
				BaseConfig: config.BaseConfig{},
				Mirrors: map[string]config.MirrorConfig{
					"test": {
						BaseConfig: config.BaseConfig{},
						Settings:   config.MirrorSettings{ASCIIName: true},
					},
				},
			},
			mockSetup: func(srcR *mocks.SourceReader) {
				srcR.EXPECT().Clone(mock.Anything, mock.Anything).Return(model.Repository{}, nil)
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			mockReader := new(mocks.SourceReader)
			tabletest.mockSetup(mockReader)

			repos, err := Clone(ctx, mockReader, tabletest.syncCfg, tabletest.projectinfos)
			if tabletest.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Len(t, repos, len(tabletest.projectinfos))
			mockReader.AssertExpectations(t)
		})
	}
}

func TestFetchProjectInfo(t *testing.T) {
	tests := []struct {
		name      string
		syncCfg   config.SyncConfig
		mockSetup func(*mocks.GitProvider)
		wantLen   int
		wantErr   bool
	}{
		{
			name: "successful fetch",
			syncCfg: config.SyncConfig{
				BaseConfig: config.BaseConfig{
					Owner:     "owner",
					OwnerType: "user",
				},
				IncludeForks: true,
			},
			mockSetup: func(gitP *mocks.GitProvider) {
				gitP.On("Name").Return("GitHub")
				gitP.On("ProjectInfos", mock.Anything, mock.MatchedBy(func(opt model.ProviderOption) bool {
					return opt.IncludeForks == true &&
						opt.Owner == "owner" &&
						opt.OwnerType == "user"
				}), true).Return([]model.ProjectInfo{
					{OriginalName: "repo1"},
					{OriginalName: "repo2"},
				}, nil)
			},
			wantLen: 2,
		},
		{
			name: "fetch failure",
			syncCfg: config.SyncConfig{
				BaseConfig: config.BaseConfig{},
			},
			mockSetup: func(gitP *mocks.GitProvider) {
				gitP.On("Name").Return("GitLab")
				gitP.On("ProjectInfos", mock.Anything, mock.Anything, true).
					Return(nil, errors.New("fetch failed"))
			},
			wantErr: true,
		},
		{
			name: "fetch with repository filters",
			syncCfg: config.SyncConfig{
				BaseConfig: config.BaseConfig{
					Owner:     "owner",
					OwnerType: "user",
				},
				Repositories: config.RepositoriesOption{
					Include: "repo1",
					Exclude: "repo2",
				},
			},
			mockSetup: func(gitP *mocks.GitProvider) {
				gitP.On("Name").Return("GitHub")
				gitP.On("ProjectInfos", mock.Anything, mock.MatchedBy(func(opt model.ProviderOption) bool {
					included := opt.IncludedRepositories
					excluded := opt.ExcludedRepositories

					return len(included) == 1 && included[0] == "repo1" &&
						len(excluded) == 1 && excluded[0] == "repo2"
				}), true).Return([]model.ProjectInfo{{OriginalName: "repo1"}}, nil)
			},
			wantLen: 1,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			mockProvider := new(mocks.GitProvider)
			tabletest.mockSetup(mockProvider)

			projectinfos, err := FetchProjectInfo(context.Background(), tabletest.syncCfg, mockProvider)
			if tabletest.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Len(t, projectinfos, tabletest.wantLen)
			mockProvider.AssertExpectations(t)
		})
	}
}
