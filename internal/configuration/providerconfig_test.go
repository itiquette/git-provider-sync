// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestProviderConfig_String(t *testing.T) {
	tests := []struct {
		name     string
		config   ProviderConfig
		expected string
	}{
		{
			name: "Complete config",
			config: ProviderConfig{
				Provider:         "github",
				Domain:           "github.com",
				Token:            "secret",
				User:             "user1",
				Scheme:           "https",
				Group:            "group1",
				Exclude:          map[string]string{"repo": "excluded"},
				Include:          map[string]string{"repo": "included"},
				Providerspecific: map[string]string{"key": "value"},
			},
			expected: "ProviderConfig: Provider: github, Domain: github.com, Token: <****>, User: user1, Scheme: https, GitInfo: GitInfo: Type: , SSHPrivateKeyPath: , SSHPrivateKeyPW: ****,  Group: group1, Exclude: map[repo:excluded], Include: map[repo:included], ProviderSpecific: map[key:value]",
		},
		{
			name: "Minimal config",
			config: ProviderConfig{
				Provider: "gitlab",
				Domain:   "gitlab.com",
			},
			expected: "ProviderConfig: Provider: gitlab, Domain: gitlab.com, Token: <****>, User: , Scheme: , GitInfo: GitInfo: Type: , SSHPrivateKeyPath: , SSHPrivateKeyPW: ****,  Group: , Exclude: map[], Include: map[], ProviderSpecific: map[]",
		},
		{
			name: "Config with empty token",
			config: ProviderConfig{
				Provider: "bitbucket",
				Domain:   "bitbucket.org",
				Token:    "",
			},
			expected: "ProviderConfig: Provider: bitbucket, Domain: bitbucket.org, Token: <****>, User: , Scheme: , GitInfo: GitInfo: Type: , SSHPrivateKeyPath: , SSHPrivateKeyPW: ****,  Group: , Exclude: map[], Include: map[], ProviderSpecific: map[]",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			result := tabletest.config.String()
			require.Equal(t, tabletest.expected, result)
		})
	}
}

func TestProviderConfig_DebugLog(t *testing.T) {
	tests := []struct {
		name     string
		config   ProviderConfig
		expected []string
	}{
		{
			name: "GitHub config",
			config: ProviderConfig{
				Provider: "github",
				Domain:   "github.com",
				User:     "user1",
				Group:    "group1",
			},
			expected: []string{"provider", "github", "domain", "github.com", "user,group", "user1", "group1"},
		},
		{
			name: "Directory config",
			config: ProviderConfig{
				Provider:         DIRECTORY,
				Providerspecific: map[string]string{"directorytargetdir": "/path/to/dir"},
			},
			expected: []string{"provider", DIRECTORY, "target directory", "/path/to/dir"},
		},
		{
			name: "Archive config",
			config: ProviderConfig{
				Provider:         ARCHIVE,
				Providerspecific: map[string]string{"archivetargetdir": "/path/to/archive"},
			},
			expected: []string{"provider", ARCHIVE, "target directory", "/path/to/archive"},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := zerolog.New(&buf)
			tabletest.config.DebugLog(&logger).Msg(tabletest.name)

			logOutput := buf.String()
			for _, exp := range tabletest.expected {
				require.Contains(t, logOutput, exp)
			}
		})
	}
}

func TestProviderConfig_Methods(t *testing.T) {
	config := ProviderConfig{
		Provider: "github",
		Domain:   "github.com",
		User:     "user1",
		Group:    "group1",
		Include:  map[string]string{"repositories": "repo1, repo2, repo3"},
		Exclude:  map[string]string{"repositories": "excluded1, excluded2"},
		Providerspecific: map[string]string{
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
		repos := config.IncludedRepositories()
		require.Equal(t, []string{"repo1", "repo2", "repo3"}, repos)
	})

	t.Run("ExcludedRepositories", func(t *testing.T) {
		repos := config.ExcludedRepositories()
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
		Configurations: map[string]ProvidersConfig{
			"test": {
				SourceProvider: ProviderConfig{
					Provider: "github",
					Domain:   "github.com",
				},
				ProviderTargets: map[string]ProviderConfig{
					"target1": {
						Provider: "gitlab",
						Domain:   "gitlab.com",
					},
					"target2": {
						Provider:         DIRECTORY,
						Providerspecific: map[string]string{"directorytargetdir": "/path/to/dir"},
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
