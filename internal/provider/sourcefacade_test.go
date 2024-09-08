// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package provider

import (
	"context"
	"errors"
	"testing"

	"itiquette/git-provider-sync/internal/configuration"
	"itiquette/git-provider-sync/internal/model"

	mocks "itiquette/git-provider-sync/generated/mocks/mockgogit"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestClone(t *testing.T) {
	require := require.New(t)
	ctx := testContext()

	tests := []struct {
		name      string
		metainfos []model.RepositoryMetainfo
		mockSetup func(*mocks.SourceReader)
		wantErr   bool
	}{
		{
			name: "Successful clone of multiple repositories",
			metainfos: []model.RepositoryMetainfo{
				{HTTPSURL: "https://github.com/user/repo1.git", OriginalName: "repo1"},
				{HTTPSURL: "https://github.com/user/repo2.git", OriginalName: "repo2"},
			},
			mockSetup: func(m *mocks.SourceReader) {
				m.On("Clone", mock.Anything, mock.Anything).Return(model.Repository{}, nil).Twice()
			},
			wantErr: false,
		},
		{
			name: "Failure to clone repository",
			metainfos: []model.RepositoryMetainfo{
				{HTTPSURL: "https://github.com/user/repo1.git", OriginalName: "repo1"},
			},
			mockSetup: func(m *mocks.SourceReader) {
				m.On("Clone", mock.Anything, mock.Anything).Return(model.Repository{}, errors.New("clone failed"))
			},
			wantErr: true,
		},
		{
			name:      "Empty metainfos list",
			metainfos: []model.RepositoryMetainfo{},
			mockSetup: func(_ *mocks.SourceReader) {},
			wantErr:   false,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			mockReader := new(mocks.SourceReader)
			tabletest.mockSetup(mockReader)

			repos, err := Clone(ctx, mockReader, tabletest.metainfos)

			if tabletest.wantErr {
				require.Error(err)
			} else {
				require.NoError(err)
				require.Len(repos, len(tabletest.metainfos))
			}

			mockReader.AssertExpectations(t)
		})
	}
}

func TestFetchMetainfo(t *testing.T) {
	require := require.New(t)

	tests := []struct {
		name       string
		config     configuration.ProviderConfig
		mockSetup  func(*mocks.GitProvider)
		wantLength int
		wantErr    bool
	}{
		{
			name: "Successful fetch of metainfo",
			config: configuration.ProviderConfig{
				Domain: "https://github.com",
			},
			mockSetup: func(m *mocks.GitProvider) {
				m.On("Name").Return("GitHub")
				m.On("Metainfos", mock.Anything, mock.AnythingOfType("configuration.ProviderConfig"), true).
					Return([]model.RepositoryMetainfo{
						{OriginalName: "repo1"},
						{OriginalName: "repo2"},
					}, nil)
			},
			wantLength: 2,
			wantErr:    false,
		},
		{
			name: "Failure to fetch metainfo",
			config: configuration.ProviderConfig{
				Domain: "https://gitlab.com",
			},
			mockSetup: func(m *mocks.GitProvider) {
				m.On("Name").Return("GitLab")
				m.On("Metainfos", mock.Anything, mock.AnythingOfType("configuration.ProviderConfig"), true).
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
			metainfos, err := FetchMetainfo(ctx, tabletest.config, mockProvider)

			if tabletest.wantErr {
				require.Error(err)
			} else {
				require.NoError(err)
				require.Len(metainfos, tabletest.wantLength)
			}

			mockProvider.AssertExpectations(t)
		})
	}
}
