// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"

	"github.com/stretchr/testify/require"
)

func TestReadConfigFileMergedOptionsInOrderXDGLocalDotEnvEnvVarSuccess(t *testing.T) {
	require := require.New(t)
	cwd, _ := os.Getwd()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(cwd, "testdata"))
	t.Setenv("GPS_TESTCONFIG_HOME", filepath.Join(cwd, "testdata"))
	t.Setenv("GPS_GITPROVIDERSYNC_ENV1_CONFXDG_MIRRORS_ATARGET_OWNER", "envgroup")
	t.Setenv("GPS_GITPROVIDERSYNC_ENV1_CONF1_MIRRORS_ANOTHERTARGET_DOMAIN", "envdomain")

	t.Setenv("GPS_GITPROVIDERSYNC_ENV1_CONFENV_PROVIDER_TYPE", "envconfprovider")
	t.Setenv("GPS_GITPROVIDERSYNC_ENV1_CONFENV_DOMAIN", "confenvdomain")
	t.Setenv("GPS_GITPROVIDERSYNC_ENV1_CONFENV_OWNER", "envconfuser")
	t.Setenv("GPS_GITPROVIDERSYNC_ENV1_CONFENV_REPOSITORIES_INCLUDE", "envconfrepo")

	t.Setenv("GPS_GITPROVIDERSYNC_ENV1_CONFENV_MIRRORS_ATARGET_PROVIDER_TYPE", "envconftarget")
	t.Setenv("GPS_GITPROVIDERSYNC_ENV1_CONFENV_MIRRORS_ATARGET_DOMAIN", "envconfdomain")
	t.Setenv("GPS_GITPROVIDERSYNC_ENV1_CONFENV_MIRRORS_ATARGET_GROUP", "envconfgroup")

	appConfiguration := &config.AppConfiguration{}

	err := ReadConfigurationFile("testdata/testconfig.yaml", false, appConfiguration)
	if err != nil {
		fmt.Println(err)
	}

	//	require.Len(appConfiguration.GitProviderSyncConfs["env1"], 6)

	// a xdg file only defined conf
	require.Equal("xdgconfdomain", appConfiguration.GitProviderSyncConfs["env1"]["confxdg"].GetDomain())

	// a local file only defined conf
	require.Equal("localconfdomain", appConfiguration.GitProviderSyncConfs["env1"]["conflocal"].GetDomain())

	// a dotenv file only defined conf
	require.Equal("dotenvdomain", appConfiguration.GitProviderSyncConfs["env1"]["confdotenv"].Domain)

	// a env var only defined conf
	require.Equal("confenvdomain", appConfiguration.GitProviderSyncConfs["env1"]["confenv"].Domain)

	// local confile prop without overriding
	// local conffile, which overrides a xdg prop
	require.Equal("conf1domain", appConfiguration.GitProviderSyncConfs["env1"]["conf1"].GetDomain())
	require.Equal("gitea", appConfiguration.GitProviderSyncConfs["env1"]["conf2"].ProviderType)

	// a prop was overridden from xdg to local then by .env file
	require.Equal("dotenvprovider", appConfiguration.GitProviderSyncConfs["env1"]["conf1"].Mirrors["atarget"].ProviderType)

	// a prop was overridden from xdg to local then by .env then by env var
	require.Equal("envdomain", appConfiguration.GitProviderSyncConfs["env1"]["conf1"].Mirrors["anothertarget"].Domain)
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
