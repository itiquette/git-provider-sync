// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"itiquette/git-provider-sync/internal/model"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestReadConfigFileMergedOptionsInOrderXDGLocalDotEnvEnvVarSuccess(t *testing.T) {
	require := require.New(t)
	cwd, _ := os.Getwd()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(cwd, "testdata"))
	t.Setenv("GPS_TESTCONFIG_HOME", filepath.Join(cwd, "testdata"))
	t.Setenv("GPS_CONFIGURATIONS_CONFXDG_TARGETS_ATARGET_GROUP", "envgroup")
	t.Setenv("GPS_CONFIGURATIONS_CONF1_TARGETS_ANOTHERTARGET_DOMAIN", "envdomain")

	t.Setenv("GPS_CONFIGURATIONS_CONFENV_SOURCE_PROVIDER", "envconfprovider")
	t.Setenv("GPS_CONFIGURATIONS_CONFENV_SOURCE_DOMAIN", "confenvdomain")
	t.Setenv("GPS_CONFIGURATIONS_CONFENV_SOURCE_USER", "envconfuser")
	t.Setenv("GPS_CONFIGURATIONS_CONFENV_SOURCE_INCLUDE_REPOSITORIES", "envconfrepo")
	t.Setenv("GPS_CONFIGURATIONS_CONFENV_TARGETS_ATARGET_PROVIDER", "envconftarget")
	t.Setenv("GPS_CONFIGURATIONS_CONFENV_TARGETS_ATARGET_DOMAIN", "envconfdomain")
	t.Setenv("GPS_CONFIGURATIONS_CONFENV_TARGETS_ATARGET_GROUP", "envconfgroup")

	appConfiguration := &AppConfiguration{}
	_ = ReadConfigurationFile(appConfiguration, "testdata/testconfig.yaml", false)

	require.Len(appConfiguration.Configurations, 6)

	// a xdg file only defined conf
	require.Equal("xdgconfdomain", appConfiguration.Configurations["confxdg"].SourceProvider.Domain)

	// a local file only defined conf
	require.Equal("localconfdomain", appConfiguration.Configurations["conflocal"].SourceProvider.Domain)

	// a dotenv file only defined conf
	require.Equal("dotenvdomain", appConfiguration.Configurations["confdotenv"].SourceProvider.Domain)

	// a env var only defined conf
	require.Equal("confenvdomain", appConfiguration.Configurations["confenv"].SourceProvider.Domain)

	// xdg spec value is read
	require.Equal("xdguser1", appConfiguration.Configurations["conf1"].SourceProvider.User)

	// local confile prop without overriding
	// local conffile, which overrides a xdg prop
	require.Equal("conf1domain", appConfiguration.Configurations["conf1"].SourceProvider.Domain)
	require.Equal("gitea", appConfiguration.Configurations["conf2"].SourceProvider.Provider)

	// a prop was overridden from xdg to local then by .env file
	require.Equal("dotenvprovider", appConfiguration.Configurations["conf1"].ProviderTargets["atarget"].Provider)

	// a prop was overridden from xdg to local then by .env then by env var
	require.Equal("envdomain", appConfiguration.Configurations["conf1"].ProviderTargets["anothertarget"].Domain)
}
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
				Group:            "group1",
				Exclude:          map[string]string{"repo": "excluded"},
				Include:          map[string]string{"repo": "included"},
				Providerspecific: map[string]string{"key": "value"},
			},
			expected: "ProviderConfig: Provider: github, Domain: github.com, Token: <****>, User: user1, Group: group1, Exclude: map[repo:excluded], Include: map[repo:included], ProviderSpecific: map[key:value]",
		},
		{
			name: "Minimal config",
			config: ProviderConfig{
				Provider: "gitlab",
				Domain:   "gitlab.com",
			},
			expected: "ProviderConfig: Provider: gitlab, Domain: gitlab.com, Token: <****>, User: , Group: , Exclude: map[], Include: map[], ProviderSpecific: map[]",
		},
		{
			name: "Config with empty token",
			config: ProviderConfig{
				Provider: "bitbucket",
				Domain:   "bitbucket.org",
				Token:    "",
			},
			expected: "ProviderConfig: Provider: bitbucket, Domain: bitbucket.org, Token: <****>, User: , Group: , Exclude: map[], Include: map[], ProviderSpecific: map[]",
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

func TestLoadConfiguration_InvalidConfig(t *testing.T) {
	tests := []struct {
		name           string
		configFilePath string
		expectedError  string
	}{
		{
			name:           "Missing required fields",
			configFilePath: "testdata/missing_fields.yaml",
			expectedError:  "failed to validate configuration",
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), model.CLIOptionKey{}, model.CLIOption{
				ConfigFilePath: tabletest.configFilePath,
				ConfigFileOnly: true,
			})

			var configLoaderInstance ConfigLoader = DefaultConfigLoader{}
			_, err := configLoaderInstance.LoadConfiguration(ctx)
			require.Error(t, err)
			require.Contains(t, err.Error(), tabletest.expectedError)
		})
	}
}

func TestSplitAndTrim(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"a,b,c", []string{"a", "b", "c"}},
		{"a, b, c", []string{"a", "b", "c"}},
		{"", []string{""}},
		{" ", []string{""}},
		{"a,,c", []string{"a", "", "c"}},
		{" a , b , c ", []string{"a", "b", "c"}},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.input, func(t *testing.T) {
			result := splitAndTrim(tabletest.input)
			require.Equal(t, tabletest.expected, result)
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

func TestReadConfigurationFile_NoConfigurations(t *testing.T) {
	appConfig := &AppConfiguration{}

	require.Panics(t, func() {
		_ = ReadConfigurationFile(appConfig, "testdata/empty_config.yaml", true)
	}, "Expected panic for empty configuration")
}

func TestHasLocalConfigFile(t *testing.T) {
	require.True(t, hasLocalConfigFile("testdata/testconfig.yaml"))
	require.False(t, hasLocalConfigFile("nonexistent.yaml"))
}

func TestHasXDGConfigFile(t *testing.T) {
	t.Run("XDG config exists", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "testdata")

		exists, path := hasXDGConfigFile("XDG_CONFIG_HOME", "/gitprovidersync/gitprovidersync.yaml")
		require.True(t, exists)
		require.Contains(t, path, "testdata/gitprovidersync/gitprovidersync.yaml")
	})

	t.Run("XDG config does not exist", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "testdata/nonexistent")

		exists, _ := hasXDGConfigFile("XDG_CONFIG_HOME", "/gitprovidersync/gitprovidersync.yaml")
		require.False(t, exists)
	})
}

func TestHasDotEnvFile(t *testing.T) {
	t.Run("Dot env file exists", func(t *testing.T) {
		t.Setenv("GPS_TESTCONFIG_HOME", "testdata")

		exists, path := hasDotEnvFile(".env")
		require.True(t, exists)
		require.Contains(t, path, "testdata/.env")
	})

	t.Run("Dot env file does not exist", func(t *testing.T) {
		t.Setenv("GPS_TESTCONFIG_HOME", "testdata/nonexistent")

		exists, _ := hasDotEnvFile(".env")
		require.False(t, exists)
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
