// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package gitlab

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mocks "itiquette/git-provider-sync/generated/mocks/mockgogit"
	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
)

func TestAPIClient_CreateProject(t *testing.T) {
	tests := []struct {
		name    string
		opt     model.CreateProjectOption
		want    string
		wantErr bool
		mock    func(*mocks.ProjectServicer)
	}{
		{
			name: "successful project creation",
			opt:  model.CreateProjectOption{RepositoryName: "test-project"},
			want: "123",
			mock: func(m *mocks.ProjectServicer) {
				m.EXPECT().CreateProject(mock.Anything, mock.Anything).Return("123", nil)
			},
		},
		{
			name:    "failed project creation",
			opt:     model.CreateProjectOption{RepositoryName: "test-project"},
			wantErr: true,
			mock: func(m *mocks.ProjectServicer) {
				m.EXPECT().CreateProject(mock.Anything, mock.Anything).Return("", errors.New("API error"))
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			mockProjectService := new(mocks.ProjectServicer)
			tabletest.mock(mockProjectService)

			api := APIClient{projectService: mockProjectService}

			got, err := api.CreateProject(context.Background(), tabletest.opt)
			if tabletest.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tabletest.want, got)
			mockProjectService.AssertExpectations(t)
		})
	}
}

func TestAPIClient_ProjectInfos(t *testing.T) {
	tests := []struct {
		name      string
		opt       model.ProviderOption
		filtering bool
		want      []model.ProjectInfo
		wantErr   bool
		mockProj  func(*mocks.ProjectServicer)
		mockFilt  func(*mocks.FilterServicer)
	}{
		{
			name:      "without filtering",
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
			name:      "with filtering",
			filtering: true,
			want:      []model.ProjectInfo{{OriginalName: "project1"}},
			mockProj: func(m *mocks.ProjectServicer) {
				m.EXPECT().GetProjectInfos(mock.Anything, mock.Anything).
					Return([]model.ProjectInfo{{OriginalName: "project1"}, {OriginalName: "project2"}}, nil)
			},
			mockFilt: func(m *mocks.FilterServicer) {
				m.EXPECT().FilterProjectinfos(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return([]model.ProjectInfo{{OriginalName: "project1"}}, nil)
			},
		},
		{
			name:      "project service error",
			filtering: false,
			wantErr:   true,
			mockProj: func(m *mocks.ProjectServicer) {
				m.EXPECT().GetProjectInfos(mock.Anything, mock.Anything).
					Return(nil, errors.New("failed"))
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

			got, err := api.ProjectInfos(context.Background(), tabletest.opt, tabletest.filtering)
			if tabletest.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tabletest.want, got)
			mockProjectService.AssertExpectations(t)

			if tabletest.filtering {
				mockFilterService.AssertExpectations(t)
			}
		})
	}
}

func TestAPIClient_NewGitLabAPIClient(t *testing.T) {
	tests := []struct {
		name    string
		opt     model.GitProviderClientOption
		wantErr bool
	}{
		{
			name: "success with default domain",
			opt: model.GitProviderClientOption{
				AuthCfg: config.AuthConfig{Token: "test-token"},
			},
		},
		{
			name: "success with custom domain",
			opt: model.GitProviderClientOption{
				Domain: "gitlab.example.com",
				AuthCfg: config.AuthConfig{
					Token:      "test-token",
					HTTPScheme: "https",
				},
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			client, err := NewGitLabAPIClient(context.Background(), http.DefaultClient, tabletest.opt)
			if tabletest.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, client.raw)
			require.NotNil(t, client.projectService)
			require.NotNil(t, client.protectionService)
			require.NotNil(t, client.filterService)
		})
	}
}

func TestAPIClient_ProjectExists(t *testing.T) {
	tests := []struct {
		name   string
		owner  string
		repo   string
		exists bool
		id     string
		mock   func(*mocks.ProjectServicer)
	}{
		{
			name:   "exists",
			owner:  "owner",
			repo:   "repo",
			exists: true,
			id:     "123",
			mock: func(m *mocks.ProjectServicer) {
				m.EXPECT().Exists(mock.Anything, "owner", "repo").Return(true, "123", nil)
			},
		},
		{
			name:   "does not exist",
			owner:  "owner",
			repo:   "repo",
			exists: false,
			id:     "",
			mock: func(m *mocks.ProjectServicer) {
				m.EXPECT().Exists(mock.Anything, "owner", "repo").Return(false, "", nil)
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			mockProjectService := new(mocks.ProjectServicer)
			tabletest.mock(mockProjectService)

			api := APIClient{projectService: mockProjectService}
			exists, id := api.ProjectExists(context.Background(), tabletest.owner, tabletest.repo)

			require.Equal(t, tabletest.exists, exists)
			require.Equal(t, tabletest.id, id)
			mockProjectService.AssertExpectations(t)
		})
	}
}
func TestAPIClient_ProtectProject(t *testing.T) {
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
				m.EXPECT().Protect(mock.Anything, "", "123").Return(errors.New("invalid branch"))
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			mockService := new(mocks.ProtectionServicer)
			tabletest.mock(mockService)

			api := APIClient{protectionService: mockService}
			err := api.ProtectProject(context.Background(), "", tabletest.defaultBranch, tabletest.projectIDStr)

			if tabletest.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			mockService.AssertExpectations(t)
		})
	}
}

func TestAPIClient_SetDefaultBranch(t *testing.T) {
	tests := []struct {
		name        string
		owner       string
		projectName string
		branch      string
		wantErr     bool
		mock        func(*mocks.ProjectServicer)
	}{
		{
			name:        "successful set",
			owner:       "owner",
			projectName: "project",
			branch:      "main",
			mock: func(m *mocks.ProjectServicer) {
				m.EXPECT().SetDefaultBranch(mock.Anything, "owner", "project", "main").Return(nil)
			},
		},
		{
			name:        "failure",
			owner:       "owner",
			projectName: "project",
			branch:      "main",
			wantErr:     true,
			mock: func(m *mocks.ProjectServicer) {
				m.EXPECT().SetDefaultBranch(mock.Anything, "owner", "project", "main").Return(errors.New("API error"))
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			mockService := new(mocks.ProjectServicer)
			tabletest.mock(mockService)

			api := APIClient{projectService: mockService}
			err := api.SetDefaultBranch(context.Background(), tabletest.owner, tabletest.projectName, tabletest.branch)

			if tabletest.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			mockService.AssertExpectations(t)
		})
	}
}

func TestAPIClient_UnprotectProject(t *testing.T) {
	tests := []struct {
		name          string
		defaultBranch string
		projectIDStr  string
		wantErr       bool
		mock          func(*mocks.ProtectionServicer)
	}{
		{
			name:          "success",
			defaultBranch: "main",
			projectIDStr:  "123",
			mock: func(m *mocks.ProtectionServicer) {
				m.EXPECT().Unprotect(mock.Anything, "main", "123").Return(nil)
			},
		},
		{
			name:          "failure",
			defaultBranch: "",
			projectIDStr:  "123",
			wantErr:       true,
			mock: func(m *mocks.ProtectionServicer) {
				m.EXPECT().Unprotect(mock.Anything, "", "123").Return(errors.New("error"))
			},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			mockService := new(mocks.ProtectionServicer)
			tabletest.mock(mockService)

			api := APIClient{protectionService: mockService}
			err := api.UnprotectProject(context.Background(), tabletest.defaultBranch, tabletest.projectIDStr)

			if tabletest.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			mockService.AssertExpectations(t)
		})
	}
}

func TestAPIClient_IsValidProjectName(t *testing.T) {
	tests := []struct {
		name     string
		projName string
		want     bool
	}{
		{"valid name", "valid-project-name", true},
		{"valid with numbers", "project-123", true},
		{"invalid with underscore", "project_name_123", false},
		{"invalid with slash", "invalid/project", false},
		{"invalid with asterisk", "invalid*name", false},
		{"reserved name", "preview", false},
		{"empty name", "", false},
		{"too long", strings.Repeat("a", 259), false},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			api := APIClient{}
			got := api.IsValidProjectName(context.Background(), tabletest.projName)
			require.Equal(t, tabletest.want, got)
		})
	}
}

func TestAPIClient_Name(t *testing.T) {
	api := APIClient{}
	require.Equal(t, "gitlab", api.Name())
}
