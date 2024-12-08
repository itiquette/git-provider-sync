// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

//TODO fix me
// func TestProviderConfig_String(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		config   ProviderConfig
// 		expected string
// 	}{
// 		{
// 			name: "Complete config",
// 			config: ProviderConfig{
// 				ProviderType: "github",
// 				Domain:       "github.com",
// 				HTTPClient:   HTTPClientOption{Token: "secret"},
// 				User:         "user1",
// 				Group:        "group1",
// 				Repositories: RepositoriesOption{Exclude: "excluded", Include: "included"},
// 				Additional:   map[string]string{"key": "value"},
// 			},
// 			expected: "ProviderConfig: ProviderType: github, Domain: github.com, UploadDomain: , User: user1, Group: group1, Repositories: RepositoryOption: Exclude excluded, Include: included, Git: GitOption: Type: , IncludeForks: false, UseGitBinary: false, Project: ProjectOption: Type: , HTTPClient: HTTPClientOption: ProxyURL , Token: **cret, SSHClient: SSHClientOption{ }, SyncRun: SyncRunOption{ ForcePush: false IgnoreInvalidName: false ASCIIName: false }, Additional: map[key:value]",
// 		},
// 		{
// 			name: "Minimal config",
// 			config: ProviderConfig{
// 				ProviderType: "gitlab",
// 				Domain:       "gitlab.com",
// 			},
// 			expected: "ProviderConfig: ProviderType: gitlab, Domain: gitlab.com, UploadDomain: , User: , Group: , Repositories: RepositoryOption: Exclude , Include: , Git: GitOption: Type: , IncludeForks: false, UseGitBinary: false, Project: ProjectOption: Type: , HTTPClient: HTTPClientOption: ProxyURL , Token: , SSHClient: SSHClientOption{ }, SyncRun: SyncRunOption{ ForcePush: false IgnoreInvalidName: false ASCIIName: false }, Additional: map[]",
// 		},
// 		{
// 			name: "Config with empty token",
// 			config: ProviderConfig{
// 				ProviderType: "bitbucket",
// 				Domain:       "bitbucket.org",
// 				HTTPClient:   HTTPClientOption{Token: ""},
// 			},
// 			expected: "ProviderConfig: ProviderType: bitbucket, Domain: bitbucket.org, UploadDomain: , User: , Group: , Repositories: RepositoryOption: Exclude , Include: , Git: GitOption: Type: , IncludeForks: false, UseGitBinary: false, Project: ProjectOption: Type: , HTTPClient: HTTPClientOption: ProxyURL , Token: , SSHClient: SSHClientOption{ }, SyncRun: SyncRunOption{ ForcePush: false IgnoreInvalidName: false ASCIIName: false }, Additional: map[]",
// 		},
// 	}

// 	for _, tabletest := range tests {
// 		t.Run(tabletest.name, func(t *testing.T) {
// 			result := tabletest.config.String()
// 			require.Equal(t, tabletest.expected, result)
// 		})
// 	}
// }

// func TestProviderConfig_DebugLog(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		config   ProviderConfig
// 		expected []string
// 	}{
// 		{
// 			name: "GitHub config",
// 			config: ProviderConfig{
// 				ProviderType: "github",
// 				Domain:       "github.com",
// 				User:         "user1",
// 				Group:        "group1",
// 			},
// 			expected: []string{"provider", "github", "domain", "github.com", "user_group", "user1", "group1"},
// 		},
// 		{
// 			name: "Directory config",
// 			config: ProviderConfig{
// 				ProviderType: DIRECTORY,
// 				Additional:   map[string]string{"directorytargetdir": "/path/to/dir"},
// 			},
// 			expected: []string{"provider", DIRECTORY, "target directory", "/path/to/dir"},
// 		},
// 		{
// 			name: "Archive config",
// 			config: ProviderConfig{
// 				ProviderType: ARCHIVE,
// 				Additional:   map[string]string{"archivetargetdir": "/path/to/archive"},
// 			},
// 			expected: []string{"provider", ARCHIVE, "target directory", "/path/to/archive"},
// 		},
// 	}

// 	for _, tabletest := range tests {
// 		t.Run(tabletest.name, func(t *testing.T) {
// 			var buf bytes.Buffer
// 			logger := zerolog.New(&buf)
// 			tabletest.config.DebugLog(&logger).Msg(tabletest.name)

// 			logOutput := buf.String()
// 			for _, exp := range tabletest.expected {
// 				require.Contains(t, logOutput, exp)
// 			}
// 		})
// 	}
// }

func TestProviderConfig_Methods(t *testing.T) {
	config := ProviderConfig{
		ProviderType: "github",
		Domain:       "github.com",
		User:         "user1",
		Group:        "group1",
		Repositories: RepositoriesOption{Include: "repo1,repo2,repo3", Exclude: "excluded1,excluded2"},
		Additional: map[string]string{
			"archivetargetdir":   "/path/to/archive",
			"directorytargetdir": "/path/to/directory",
		},
	}

	t.Run("IsGroup", func(t *testing.T) {
		require.True(t, config.IsGroup())
		config.Group = ""
		require.False(t, config.IsGroup())
	})

	t.Run("IncludedRepositories", func(t *testing.T) {
		repos := config.Repositories.IncludedRepositories()
		require.Equal(t, []string{"repo1", "repo2", "repo3"}, repos)
	})

	t.Run("ExcludedRepositories", func(t *testing.T) {
		repos := config.Repositories.ExcludedRepositories()
		require.Equal(t, []string{"excluded1", "excluded2"}, repos)
	})

	t.Run("ArchiveTargetDir", func(t *testing.T) {
		dir := config.ArchiveTargetDir()
		require.Equal(t, "/path/to/archive", dir)
	})

	t.Run("DirectoryTargetDir", func(t *testing.T) {
		dir := config.DirectoryTargetDir()
		require.Equal(t, "/path/to/directory", dir)
	})
}

func TestAppConfiguration_DebugLog(t *testing.T) {
	config := AppConfiguration{
		GitProviderSyncConfs: map[string]ProvidersConfig{
			"test": {
				SourceProvider: ProviderConfig{
					ProviderType: "github",
					Domain:       "github.com",
				},
				ProviderTargets: map[string]ProviderConfig{
					"target1": {
						ProviderType: "gitlab",
						Domain:       "gitlab.com",
					},
					"target2": {
						ProviderType: DIRECTORY,
						Additional:   map[string]string{"directorytargetdir": "/path/to/dir"},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	config.DebugLog(&logger)

	logOutput := buf.String()
	require.Contains(t, logOutput, "test:SourceProvider")
	require.Contains(t, logOutput, "target1:TargetProvider")
	require.Contains(t, logOutput, "target2:TargetProvider")
	require.Contains(t, logOutput, "github.com")
	require.Contains(t, logOutput, "gitlab.com")
	require.Contains(t, logOutput, "/path/to/dir")
}
