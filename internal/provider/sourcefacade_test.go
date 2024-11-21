// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package provider

import (
	"context"
	"errors"
	"testing"

	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"

	mocks "itiquette/git-provider-sync/generated/mocks/mockgogit"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func testContext() context.Context {
	ctx := context.Background()
	input := model.CLIOption{CleanupName: true}
	//ctx, _ = model.CreateTmpDir(ctx, "", "testadir")

	return model.WithCLIOption(ctx, input)
}
func TestClone(t *testing.T) {
	require := require.New(t)
	ctx := testContext()

	tests := []struct {
		name         string
		projectinfos []model.ProjectInfo
		mockSetup    func(*mocks.SourceReader)
		wantErr      bool
	}{
		{
			name: "Successful clone of multiple repositories",
			projectinfos: []model.ProjectInfo{
				{HTTPSURL: "https://github.com/user/repo1.git", OriginalName: "repo1"},
				{HTTPSURL: "https://github.com/user/repo2.git", OriginalName: "repo2"},
			},
			mockSetup: func(m *mocks.SourceReader) {
				m.EXPECT().Clone(mock.Anything, mock.Anything).Return(model.Repository{}, nil).Twice()
			},
			wantErr: false,
		},
		{
			name: "Failure to clone repository",
			projectinfos: []model.ProjectInfo{
				{HTTPSURL: "https://github.com/user/repo1.git", OriginalName: "repo1"},
			},
			mockSetup: func(m *mocks.SourceReader) {
				m.EXPECT().Clone(mock.Anything, mock.Anything).Return(model.Repository{}, errors.New("clone failed"))
			},
			wantErr: true,
		},
		{
			name:         "Empty projectinfos list",
			projectinfos: []model.ProjectInfo{},
			mockSetup:    func(_ *mocks.SourceReader) {},
			wantErr:      false,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			mockReader := new(mocks.SourceReader)
			tabletest.mockSetup(mockReader)

			repos, err := Clone(ctx, mockReader, config.ProviderConfig{}, tabletest.projectinfos)

			if tabletest.wantErr {
				require.Error(err)
			} else {
				require.NoError(err)
				require.Len(repos, len(tabletest.projectinfos))
			}

			mockReader.AssertExpectations(t)
		})
	}
}

func TestFetchMetainfo(t *testing.T) {
	require := require.New(t)

	tests := []struct {
		name       string
		config     config.ProviderConfig
		mockSetup  func(*mocks.GitProvider)
		wantLength int
		wantErr    bool
	}{
		{
			name:   "Successful fetch of metainfo",
			config: config.ProviderConfig{},
			mockSetup: func(m *mocks.GitProvider) {
				m.On("Name").Return("GitHub")
				m.On("ProjectInfos", mock.Anything, mock.AnythingOfType("model.ProviderConfig"), true).
					Return([]model.ProjectInfo{
						{OriginalName: "repo1"},
						{OriginalName: "repo2"},
					}, nil)
			},
			wantLength: 2,
			wantErr:    false,
		},
		{
			name:   "Failure to fetch metainfo",
			config: config.ProviderConfig{},
			mockSetup: func(m *mocks.GitProvider) {
				m.On("Name").Return("GitLab")
				m.On("ProjectInfos", mock.Anything, mock.AnythingOfType("model.ProviderConfig"), true).
					Return(nil, errors.New("failed to fetch metainfo"))
			},
			wantLength: 0,
			wantErr:    true,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			mockProvider := new(mocks.GitProvider)
			tabletest.mockSetup(mockProvider)

			ctx := context.Background()
			projectinfos, err := FetchProjectInfo(ctx, tabletest.config, mockProvider)

			if tabletest.wantErr {
				require.Error(err)
			} else {
				require.NoError(err)
				require.Len(projectinfos, tabletest.wantLength)
			}

			mockProvider.AssertExpectations(t)
		})
	}
}
