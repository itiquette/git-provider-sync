// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package gitlab

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mocks "itiquette/git-provider-sync/generated/mocks/mockgogit"
	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
)

func TestAPIClient_CreateProject(t *testing.T) {
	require := require.New(t)
	tests := []struct {
		name    string
		cfg     config.ProviderConfig
		opt     model.CreateProjectOption
		want    string
		wantErr bool
		mock    func(*mocks.ProjectServicer)
	}{
		{
			name: "successful project creation",
			cfg:  config.ProviderConfig{},
			opt:  model.CreateProjectOption{RepositoryName: "test-project"},
			want: "123",
			mock: func(m *mocks.ProjectServicer) {
				m.EXPECT().CreateProject(mock.Anything, mock.Anything, mock.Anything).
					Return("123", nil)
			},
		},
		{
			name:    "failed project creation - API error",
			cfg:     config.ProviderConfig{},
			opt:     model.CreateProjectOption{RepositoryName: "test-project"},
			wantErr: true,
			mock: func(m *mocks.ProjectServicer) {
				m.EXPECT().CreateProject(mock.Anything, mock.Anything, mock.Anything).
					Return("", errors.New("API error"))
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			mockProjectService := new(mocks.ProjectServicer)
			tabletest.mock(mockProjectService)

			api := APIClient{
				projectService: mockProjectService,
			}

			got, err := api.CreateProject(context.Background(), tabletest.cfg, tabletest.opt)
			if tabletest.wantErr {
				require.Error(err)

				return
			}

			require.NoError(err)
			require.Equal(tabletest.want, got)
			mockProjectService.AssertExpectations(t)
		})
	}
}

func TestAPIClient_ProjectInfos(t *testing.T) {
	require := require.New(t)
	tests := []struct {
		name      string
		cfg       config.ProviderConfig
		filtering bool
		want      []model.ProjectInfo
		wantErr   bool
		mockProj  func(*mocks.ProjectServicer)
		mockFilt  func(*mocks.FilterServicer)
	}{
		{
			name:      "successful retrieval without filtering",
			filtering: false,
			want: []model.ProjectInfo{
				{OriginalName: "project1"},
				{OriginalName: "project2"},
			},
			mockProj: func(m *mocks.ProjectServicer) {
				m.EXPECT().GetProjectInfos(mock.Anything, mock.Anything).
					Return([]model.ProjectInfo{{OriginalName: "project1"}, {OriginalName: "project2"}}, nil)
			},
		},
		{
			name:      "successful retrieval with filtering",
			filtering: true,
			want: []model.ProjectInfo{
				{OriginalName: "project1"},
			},
			mockProj: func(m *mocks.ProjectServicer) {
				m.EXPECT().GetProjectInfos(mock.Anything, mock.Anything).
					Return([]model.ProjectInfo{{OriginalName: "project1"}, {OriginalName: "project2"}}, nil)
			},
			mockFilt: func(m *mocks.FilterServicer) {
				m.EXPECT().FilterProjectinfos(mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).
					Return([]model.ProjectInfo{{OriginalName: "project1"}}, nil)
			},
		},
		{
			name:      "project service error",
			filtering: false,
			wantErr:   true,
			mockProj: func(m *mocks.ProjectServicer) {
				m.EXPECT().GetProjectInfos(mock.Anything, mock.Anything).
					Return(nil, errors.New("failed to get projects"))
			},
		},
		{
			name:      "filter service error",
			filtering: true,
			wantErr:   true,
			mockProj: func(m *mocks.ProjectServicer) {
				m.EXPECT().GetProjectInfos(mock.Anything, mock.Anything).
					Return([]model.ProjectInfo{{OriginalName: "project1"}}, nil)
			},
			mockFilt: func(m *mocks.FilterServicer) {
				m.EXPECT().FilterProjectinfos(mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).
					Return(nil, errors.New("filter error"))
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			mockProjectService := new(mocks.ProjectServicer)
			mockFilterService := new(mocks.FilterServicer)

			tabletest.mockProj(mockProjectService)

			if tabletest.mockFilt != nil {
				tabletest.mockFilt(mockFilterService)
			}

			api := APIClient{
				projectService: mockProjectService,
				filterService:  mockFilterService,
			}

			got, err := api.ProjectInfos(context.Background(), tabletest.cfg, tabletest.filtering)
			if tabletest.wantErr {
				require.Error(err)

				return
			}

			require.NoError(err)
			require.Equal(tabletest.want, got)
			mockProjectService.AssertExpectations(t)

			if tabletest.filtering {
				mockFilterService.AssertExpectations(t)
			}
		})
	}
}

func TestAPIClient_ProtectProject(t *testing.T) {
	require := require.New(t)
	tests := []struct {
		name          string
		defaultBranch string
		projectIDStr  string
		wantErr       bool
		mock          func(*mocks.ProtectionServicer)
	}{
		{
			name:          "successful protection",
			defaultBranch: "main",
			projectIDStr:  "123",
			mock: func(m *mocks.ProtectionServicer) {
				m.EXPECT().Protect(mock.Anything, "main", "123").Return(nil)
			},
		},
		{
			name:          "failed protection",
			defaultBranch: "",
			projectIDStr:  "123",
			wantErr:       true,
			mock: func(m *mocks.ProtectionServicer) {
				m.EXPECT().Protect(mock.Anything, "", "123").
					Return(errors.New("invalid branch"))
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			mockProtectionService := new(mocks.ProtectionServicer)
			tabletest.mock(mockProtectionService)

			api := APIClient{
				protectionService: mockProtectionService,
			}

			err := api.ProtectProject(context.Background(), "", tabletest.defaultBranch, tabletest.projectIDStr)
			if tabletest.wantErr {
				require.Error(err)

				return
			}

			require.NoError(err)
			mockProtectionService.AssertExpectations(t)
		})
	}
}

func TestAPIClient_SetDefaultBranch(t *testing.T) {
	require := require.New(t)
	tests := []struct {
		name        string
		owner       string
		projectName string
		branch      string
		wantErr     bool
		mock        func(*mocks.ProjectServicer)
	}{
		{
			name:        "successful default branch set",
			owner:       "test-owner",
			projectName: "test-project",
			branch:      "main",
			mock: func(m *mocks.ProjectServicer) {
				m.EXPECT().SetDefaultBranch(mock.Anything, "test-owner", "test-project", "main").
					Return(nil)
			},
		},
		{
			name:        "failure setdefaultbranch",
			owner:       "test-owner",
			projectName: "test-project",
			branch:      "main",
			wantErr:     true,
			mock: func(m *mocks.ProjectServicer) {
				m.EXPECT().SetDefaultBranch(mock.Anything, "test-owner", "test-project", "main").
					Return(errors.New("API error"))
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			mockProjectService := new(mocks.ProjectServicer)
			tabletest.mock(mockProjectService)

			api := APIClient{
				projectService: mockProjectService,
			}

			err := api.SetDefaultBranch(context.Background(), tabletest.owner, tabletest.projectName, tabletest.branch)
			if tabletest.wantErr {
				require.Error(err)

				return
			}

			require.NoError(err)
			mockProjectService.AssertExpectations(t)
		})
	}
}

func TestAPIClient_UnprotectProject(t *testing.T) {
	require := require.New(t)
	tests := []struct {
		name          string
		defaultBranch string
		projectIDStr  string
		wantErr       bool
		mock          func(*mocks.ProtectionServicer)
	}{
		{
			name:          "successful unprotection",
			defaultBranch: "main",
			projectIDStr:  "123",
			mock: func(m *mocks.ProtectionServicer) {
				m.EXPECT().Unprotect(mock.Anything, "main", "123").
					Return(nil)
			},
		},
		{
			name:          "failure default branch",
			defaultBranch: "",
			projectIDStr:  "123",
			wantErr:       true,
			mock: func(m *mocks.ProtectionServicer) {
				m.EXPECT().Unprotect(mock.Anything, "", "123").
					Return(errors.New("default branch cannot be empty"))
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			mockProtectionService := new(mocks.ProtectionServicer)
			tabletest.mock(mockProtectionService)

			api := APIClient{
				protectionService: mockProtectionService,
			}

			err := api.UnprotectProject(context.Background(), tabletest.defaultBranch, tabletest.projectIDStr)
			if tabletest.wantErr {
				require.Error(err)

				return
			}

			require.NoError(err)
			mockProtectionService.AssertExpectations(t)
		})
	}
}

func TestAPIClient_IsValidProjectName(t *testing.T) {
	require := require.New(t)
	tests := []struct {
		name     string
		projName string
		want     bool
	}{
		{
			name:     "valid project name",
			projName: "valid-project-name",
			want:     true,
		},
		{
			name:     "valid project name with numbers",
			projName: "project-123",
			want:     true,
		},
		{
			name:     "valid project name with underscores",
			projName: "project_name_123",
			want:     true,
		},
		{
			name:     "invalid characters - slash",
			projName: "invalid/project",
			want:     false,
		},
		{
			name:     "invalid characters - asterisk",
			projName: "invalid*name",
			want:     false,
		},
		{
			name:     "reserved name - tree",
			projName: "tree",
			want:     false,
		},
		{
			name:     "reserved name - preview",
			projName: "preview",
			want:     false,
		},
		{
			name:     "empty name",
			projName: "",
			want:     false,
		},
		{
			name:     "too long name",
			projName: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-very-long-project-name-that-exceeds-the-maximum-length-allowed-by-gitlab-a-very-long-project-name-that-exceeds-the-maximum-length-allowed-by-gitlab-a-very-long-project-name-that-exceeds-the-maximum-length-allowed-by-gitlabaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			want:     false,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(_ *testing.T) {
			api := APIClient{}
			got := api.IsValidProjectName(context.Background(), tabletest.projName)
			require.Equal(tabletest.want, got)
		})
	}
}

func TestAPIClient_Name(t *testing.T) {
	require := require.New(t)
	tests := []struct {
		name string
		want string
	}{
		{
			name: "returns gitlab provider name",
			want: config.GITLAB,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(_ *testing.T) {
			api := APIClient{}
			got := api.Name()
			require.Equal(tabletest.want, got)
		})
	}
}

func TestNewGitLabAPIClient(t *testing.T) {
	require := require.New(t)
	tests := []struct {
		name     string
		opt      model.GitProviderClientOption
		wantErr  bool
		validate func(*testing.T, APIClient)
	}{
		{
			name: "successful client creation with default domain",
			opt: model.GitProviderClientOption{
				HTTPClient: config.HTTPClientOption{
					Token: "test-token",
				},
			},
			validate: func(_ *testing.T, client APIClient) {
				require.NotNil(client.raw)
				require.NotNil(client.projectService)
				require.NotNil(client.protectionService)
				require.NotNil(client.filterService)
			},
		},
		{
			name: "successful client creation with custom domain",
			opt: model.GitProviderClientOption{
				Domain: "gitlab.example.com",
				HTTPClient: config.HTTPClientOption{
					Token:  "test-token",
					Scheme: "https",
				},
			},
			validate: func(_ *testing.T, client APIClient) {
				require.NotNil(client.raw)
				require.NotNil(client.projectService)
				require.NotNil(client.protectionService)
				require.NotNil(client.filterService)
			},
		},
		{
			name: "successful client creation - empty token",
			opt: model.GitProviderClientOption{
				Domain: "gitlab.example.com",
				HTTPClient: config.HTTPClientOption{
					Token:  "",
					Scheme: "http",
				},
			},
			wantErr: false,
			validate: func(_ *testing.T, client APIClient) {
				require.NotNil(client.raw)
				require.NotNil(client.projectService)
				require.NotNil(client.protectionService)
				require.NotNil(client.filterService)
			},
		},
		{
			name: "successful client creation with http scheme",
			opt: model.GitProviderClientOption{
				Domain: "gitlab.example.com",
				HTTPClient: config.HTTPClientOption{
					Token:  "test-token",
					Scheme: "http",
				},
			},
			validate: func(_ *testing.T, client APIClient) {
				require.NotNil(client.raw)
				require.NotNil(client.projectService)
				require.NotNil(client.protectionService)
				require.NotNil(client.filterService)
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			client, err := NewGitLabAPIClient(context.Background(), tabletest.opt, http.DefaultClient)
			if tabletest.wantErr {
				require.Error(err)

				return
			}

			require.NoError(err)
			tabletest.validate(t, client)
		})
	}
}
