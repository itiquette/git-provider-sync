// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"itiquette/git-provider-sync/internal/adapters/cli"
	"itiquette/git-provider-sync/internal/adapters/configuration/dto"
	"itiquette/git-provider-sync/internal/domain/entities"

	"github.com/stretchr/testify/require"
)

const testConfigTemplate = `
gitprovidersync:
  defaultenv:
    testconf:
      provider_type: github
      domain: github.com
      owner: testuser
      auth:
        token: testtoken
`

func TestReadConfigFile_MergedSources_Success(t *testing.T) {
	require := require.New(t)
	tempDir := t.TempDir()
	cwd, _ := os.Getwd()

	// Copy testdata to temp directory for isolation
	testdataPath := filepath.Join(cwd, "testdata")
	tempTestdataPath := filepath.Join(tempDir, "testdata")
	err := copyDir(testdataPath, tempTestdataPath)
	require.NoError(err)

	// Use t.Setenv for proper test isolation
	t.Setenv("XDG_CONFIG_HOME", tempTestdataPath)
	t.Setenv("GPS_TESTCONFIG_HOME", tempTestdataPath)
	t.Setenv("GPS_GITPROVIDERSYNC_ENV1_CONFXDG_MIRRORS_ATARGET_OWNER", "envgroup")
	t.Setenv("GPS_GITPROVIDERSYNC_ENV1_CONF1_MIRRORS_ANOTHERTARGET_DOMAIN", "envdomain")
	t.Setenv("GPS_GITPROVIDERSYNC_ENV1_CONFENV_PROVIDER_TYPE", "envconfprovider")
	t.Setenv("GPS_GITPROVIDERSYNC_ENV1_CONFENV_DOMAIN", "confenvdomain")
	t.Setenv("GPS_GITPROVIDERSYNC_ENV1_CONFENV_OWNER", "envconfuser")
	t.Setenv("GPS_GITPROVIDERSYNC_ENV1_CONFENV_REPOSITORIES_INCLUDE", "envconfrepo")
	t.Setenv("GPS_GITPROVIDERSYNC_ENV1_CONFENV_MIRRORS_ATARGET_PROVIDER_TYPE", "envconftarget")
	t.Setenv("GPS_GITPROVIDERSYNC_ENV1_CONFENV_MIRRORS_ATARGET_DOMAIN", "envconfdomain")
	t.Setenv("GPS_GITPROVIDERSYNC_ENV1_CONFENV_MIRRORS_ATARGET_GROUP", "envconfgroup")

	appConfiguration := &dto.AppConfiguration{}

	err = ReadConfigurationFile(context.Background(), filepath.Join(tempTestdataPath, "testdto.yaml"), false, appConfiguration)
	if err != nil {
		fmt.Println(err)
	}

	// A xdg file only defined conf
	require.Equal("xdgconfdomain", appConfiguration.GitProviderSyncConfs["env1"]["confxdg"].GetDomain())

	// A local file only defined conf
	require.Equal("localconfdomain", appConfiguration.GitProviderSyncConfs["env1"]["conflocal"].GetDomain())

	// A dotenv file only defined conf
	require.Equal("dotenvdomain", appConfiguration.GitProviderSyncConfs["env1"]["confdotenv"].Domain)

	// A env var only defined conf
	require.Equal("confenvdomain", appConfiguration.GitProviderSyncConfs["env1"]["confenv"].Domain)

	// Local confile prop without overriding
	// Local conffile, which overrides a xdg prop
	require.Equal("conf1domain", appConfiguration.GitProviderSyncConfs["env1"]["conf1"].GetDomain())
	require.Equal("gitea", appConfiguration.GitProviderSyncConfs["env1"]["conf2"].ProviderType)

	// A prop was overridden from xdg to local then by .env file
	require.Equal("dotenvprovider", appConfiguration.GitProviderSyncConfs["env1"]["conf1"].Mirrors["atarget"].ProviderType)

	// A prop was overridden from xdg to local then by .env then by env var
	require.Equal("envdomain", appConfiguration.GitProviderSyncConfs["env1"]["conf1"].Mirrors["anothertarget"].Domain)
}
func TestLoadConfiguration_InvalidConfig(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			// Create CLI config using domain entities
			cliConfig := entities.NewCLIConfigBuilder().
				WithConfigFilePath(tabletest.configFilePath).
				WithConfigFileOnly(true).
				Build()

			ctx := cli.WithCLIConfig(context.Background(), cliConfig)

			var configLoaderInstance ConfigLoader = DefaultConfigLoader{}

			_, err := configLoaderInstance.LoadConfiguration(ctx)
			require.Error(t, err)
			require.Contains(t, err.Error(), tabletest.expectedError)
		})
	}
}

func TestReadConfigFileCommaDelimitedRepositoryLists(t *testing.T) {
	// Set up environment with comma-separated repository lists
	t.Setenv("GPS_GITPROVIDERSYNC_DEFAULTENV_ITIQUETTECONF_REPOSITORIES_INCLUDE", "repo1, repo2, repo3,repo4")
	t.Setenv("GPS_GITPROVIDERSYNC_DEFAULTENV_ITIQUETTECONF_REPOSITORIES_EXCLUDE", " excluded1,excluded2 , excluded3")

	// Create app configuration directly
	appConfiguration := &dto.AppConfiguration{}

	// Read the configuration using the existing minimal_dto.yaml
	err := ReadConfigurationFile(context.Background(), "testdata/minimal_dto.yaml", false, appConfiguration)
	if err != nil {
		t.Fatalf("Failed to read configuration: %v", err)
	}

	// Debug the loaded configuration
	t.Logf("Loaded configurations: %v", appConfiguration.GitProviderSyncConfs)

	// Verify the config contains our environment
	if _, exists := appConfiguration.GitProviderSyncConfs["defaultenv"]; !exists {
		t.Fatal("Expected 'defaultenv' environment to exist")
	}

	if _, exists := appConfiguration.GitProviderSyncConfs["defaultenv"]["itiquetteconf"]; !exists {
		t.Fatal("Expected 'itiquetteconf' configuration to exist")
	}

	// Access the configuration
	repoConfig := appConfiguration.GitProviderSyncConfs["defaultenv"]["itiquetteconf"]

	// Debug print the repositories
	t.Logf("Repositories: %+v", repoConfig.Repositories)

	// Convert to string for checking inclusion
	includeStr := fmt.Sprintf("%v", repoConfig.Repositories.Include)
	excludeStr := fmt.Sprintf("%v", repoConfig.Repositories.Exclude)

	// Check include repositories
	for _, repo := range []string{"repo1", "repo2", "repo3", "repo4"} {
		if !strings.Contains(includeStr, repo) {
			t.Errorf("Include repositories should contain '%s', but got: %s", repo, includeStr)
		}
	}

	// Check exclude repositories
	for _, repo := range []string{"excluded1", "excluded2", "excluded3"} {
		if !strings.Contains(excludeStr, repo) {
			t.Errorf("Exclude repositories should contain '%s', but got: %s", repo, excludeStr)
		}
	}
}

// Test additional DefaultConfigLoader methods and edge cases.
func TestDefaultConfigLoader_LoadConfiguration_Success(t *testing.T) {
	t.Parallel()

	// Create CLI config using domain entities pointing to minimal config
	cliConfig := entities.NewCLIConfigBuilder().
		WithConfigFilePath("testdata/minimal_dto.yaml").
		WithConfigFileOnly(true).
		Build()

	ctx := cli.WithCLIConfig(context.Background(), cliConfig)

	var configLoader ConfigLoader = DefaultConfigLoader{}

	appConfig, err := configLoader.LoadConfiguration(ctx)
	require.NoError(t, err)
	require.NotNil(t, appConfig)
	require.NotEmpty(t, appConfig.GitProviderSyncConfs)
}

func TestDefaultConfigLoader_LoadConfiguration_NoCLIContext(t *testing.T) {
	t.Parallel()

	// Use empty context without CLI configuration
	ctx := context.Background()

	var configLoader ConfigLoader = DefaultConfigLoader{}

	// should work because the loader uses defaults when CLI config is not found
	appConfig, err := configLoader.LoadConfiguration(ctx)

	// With defaults, this should fail because minimal_dto.yaml doesn't exist in current directory
	require.Error(t, err)
	require.Nil(t, appConfig)
}

func TestProcessEnvKey_EnvironmentVariables_ConvertsToConfigKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		prefix   string
		expected string
	}{
		{
			name:     "provider_type keyword",
			input:    "GPS_GITPROVIDERSYNC_ENV1_CONF_PROVIDER_TYPE",
			prefix:   "GPS_",
			expected: "gitprovidersync.env1.conf.provider_type",
		},
		{
			name:     "owner_type keyword",
			input:    "GPS_GITPROVIDERSYNC_ENV1_CONF_OWNER_TYPE",
			prefix:   "GPS_",
			expected: "gitprovidersync.env1.conf.owner_type",
		},
		{
			name:     "use_git_binary keyword",
			input:    "GPS_GITPROVIDERSYNC_ENV1_CONF_USE_GIT_BINARY",
			prefix:   "GPS_",
			expected: "gitprovidersync.env1.conf.use_git_binary",
		},
		{
			name:     "regular field",
			input:    "GPS_GITPROVIDERSYNC_ENV1_CONF_DOMAIN",
			prefix:   "GPS_",
			expected: "gitprovidersync.env1.conf.domain",
		},
		{
			name:     "no prefix",
			input:    "GITPROVIDERSYNC_ENV1_CONF_PROVIDER_TYPE",
			prefix:   "",
			expected: "gitprovidersync.env1.conf.provider_type",
		},
		{
			name:     "force_push keyword",
			input:    "GPS_GITPROVIDERSYNC_ENV1_CONF_FORCE_PUSH",
			prefix:   "GPS_",
			expected: "gitprovidersync.env1.conf.force_push",
		},
		{
			name:     "ssh_command keyword",
			input:    "GPS_GITPROVIDERSYNC_ENV1_CONF_SSH_COMMAND",
			prefix:   "GPS_",
			expected: "gitprovidersync.env1.conf.ssh_command",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := processEnvKey(test.input, test.prefix)
			require.Equal(t, test.expected, result)
		})
	}
}

func TestReadConfigurationFile_ErrorCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		configFile     string
		configFileOnly bool
		envs           map[string]string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "empty configuration",
			configFile:     "",
			configFileOnly: true,
			expectError:    true,
			errorContains:  "config file path cannot be empty",
		},
		{
			name:           "nonexistent configuration file",
			configFile:     "nonexistent.yaml",
			configFileOnly: true, // Use configFileOnly to avoid picking up env vars
			expectError:    true,
			errorContains:  "error loading specified config file",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Set environment variables if provided
			for key, value := range test.envs {
				t.Setenv(key, value)
			}

			appConfiguration := &dto.AppConfiguration{}
			err := ReadConfigurationFile(context.Background(), test.configFile, test.configFileOnly, appConfiguration)

			if test.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.errorContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestReadConfigurationFile_ConfigFileOnlyMode(t *testing.T) {
	tempDir := t.TempDir()

	// Create a temporary config file
	configContent := testConfigTemplate
	configFile := filepath.Join(tempDir, "test.yaml")
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	// Set environment variables that should be ignored in config-file-only mode
	t.Setenv("GPS_GITPROVIDERSYNC_DEFAULTENV_TESTCONF_DOMAIN", "ignored.com")

	appConfiguration := &dto.AppConfiguration{}
	err = ReadConfigurationFile(context.Background(), configFile, true, appConfiguration)
	require.NoError(t, err)

	// Verify that environment variables were ignored
	require.Equal(t, "github.com", appConfiguration.GitProviderSyncConfs["defaultenv"]["testconf"].GetDomain())
}

func TestProviderSpecificTokenEnvironmentVariables(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()
	tests := []struct {
		name        string
		envVars     map[string]string
		configYAML  string
		expectToken map[string]string // provider -> expected token
	}{
		{
			name: "GitHub token from environment",
			envVars: map[string]string{
				"GPS_GITHUB_TOKEN": "ghp_test123",
			},
			configYAML: `
gitprovidersync:
  production:
    github-source:
      provider_type: github
      domain: github.com
      owner: testuser
      owner_type: user
`,
			expectToken: map[string]string{
				"github-source": "ghp_test123",
			},
		},
		{
			name: "GitLab token from environment",
			envVars: map[string]string{
				"GPS_GITLAB_TOKEN": "glpat_test456",
			},
			configYAML: `
gitprovidersync:
  production:
    gitlab-mirror:
      provider_type: gitlab
      domain: gitlab.com
      owner: testuser
      owner_type: user
`,
			expectToken: map[string]string{
				"gitlab-mirror": "glpat_test456",
			},
		},
		{
			name: "Multiple provider tokens",
			envVars: map[string]string{
				"GPS_GITHUB_TOKEN": "ghp_test123",
				"GPS_GITLAB_TOKEN": "glpat_test456",
				"GPS_GITEA_TOKEN":  "gitea_test789",
			},
			configYAML: `
gitprovidersync:
  production:
    github-source:
      provider_type: github
      domain: github.com
      owner: testuser
      owner_type: user
      mirrors:
        gitlab-backup:
          provider_type: gitlab
          domain: gitlab.com
          owner: backup
          owner_type: user
        gitea-backup:
          provider_type: gitea
          domain: gitea.com
          owner: backup
          owner_type: user
`,
			expectToken: map[string]string{
				"github-source": "ghp_test123",
				"gitlab-backup": "glpat_test456",
				"gitea-backup":  "gitea_test789",
			},
		},
		{
			name: "Environment expansion in config",
			envVars: map[string]string{
				"GITHUB_TOKEN": "ghp_expanded",
			},
			configYAML: `
gitprovidersync:
  production:
    github-source:
      provider_type: github
      domain: github.com
      owner: testuser
      owner_type: user
      auth:
        token: "${GITHUB_TOKEN}"
`,
			expectToken: map[string]string{
				"github-source": "ghp_expanded",
			},
		},
		{
			name: "Provider env var takes precedence over config token",
			envVars: map[string]string{
				"GPS_GITHUB_TOKEN": "ghp_from_env",
			},
			configYAML: `
gitprovidersync:
  production:
    github-source:
      provider_type: github
      domain: github.com
      owner: testuser
      owner_type: user
      auth:
        token: "ghp_from_config"
`,
			expectToken: map[string]string{
				"github-source": "ghp_from_env", // Provider env var has highest precedence
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create temp config file
			tempDir := t.TempDir()
			configFile := filepath.Join(tempDir, "dto.yaml")
			err := os.WriteFile(configFile, []byte(test.configYAML), 0600)
			require.NoError(t, err)

			// Set environment variables
			for key, value := range test.envVars {
				t.Setenv(key, value)
			}

			// Load configuration
			appConfig := &dto.AppConfiguration{}
			err = ReadConfigurationFile(context.Background(), configFile, false, appConfig)
			require.NoError(t, err)

			// Check tokens
			for sourceName, expectedToken := range test.expectToken {
				if sourceName == "gitlab-backup" || sourceName == "gitea-backup" {
					// These are mirrors
					source := appConfig.GitProviderSyncConfs["production"]["github-source"]
					mirror := source.Mirrors[sourceName]
					require.Equal(t, expectedToken, mirror.Auth.Token,
						"Mirror %s should have token %s", sourceName, expectedToken)
				} else {
					// These are sources
					source := appConfig.GitProviderSyncConfs["production"][sourceName]
					require.Equal(t, expectedToken, source.Auth.Token,
						"Source %s should have token %s", sourceName, expectedToken)
				}
			}
		})
	}
}

func TestReadConfigurationFile_WithDotEnv(t *testing.T) {
	// Test that .env file processing doesn't break the configuration loading
	tempDir := t.TempDir()

	// Create .env file
	dotEnvContent := `# Test .env file
GPS_GITPROVIDERSYNC_DEFAULTENV_TESTCONF_DOMAIN=dotenv.com`
	dotEnvFile := filepath.Join(tempDir, ".env")
	err := os.WriteFile(dotEnvFile, []byte(dotEnvContent), 0600)
	require.NoError(t, err)

	// Create minimal config file
	configContent := testConfigTemplate
	configFile := filepath.Join(tempDir, "test.yaml")
	err = os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	// Set GPS_TESTCONFIG_HOME to temp directory
	t.Setenv("GPS_TESTCONFIG_HOME", tempDir)

	appConfiguration := &dto.AppConfiguration{}
	err = ReadConfigurationFile(context.Background(), configFile, false, appConfiguration)
	require.NoError(t, err)

	// Verify configuration was loaded successfully
	require.NotEmpty(t, appConfiguration.GitProviderSyncConfs)
	require.Equal(t, "github", appConfiguration.GitProviderSyncConfs["defaultenv"]["testconf"].ProviderType)
}

func TestReadConfigurationFile_RepositoryLists_ParsesVariousFormats(t *testing.T) {
	// Test various repository list formats
	tests := []struct {
		name            string
		includeValue    string
		excludeValue    string
		expectedInclude []string
		expectedExclude []string
	}{
		{
			name:            "comma separated with spaces",
			includeValue:    "repo1, repo2, repo3",
			excludeValue:    "excl1, excl2",
			expectedInclude: []string{"repo1", "repo2", "repo3"},
			expectedExclude: []string{"excl1", "excl2"},
		},
		{
			name:            "comma separated no spaces",
			includeValue:    "repo1,repo2,repo3",
			excludeValue:    "excl1,excl2",
			expectedInclude: []string{"repo1", "repo2", "repo3"},
			expectedExclude: []string{"excl1", "excl2"},
		},
		{
			name:            "with empty entries",
			includeValue:    "repo1, , repo3",
			excludeValue:    "excl1, ,excl2",
			expectedInclude: []string{"repo1", "repo3"},
			expectedExclude: []string{"excl1", "excl2"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("GPS_GITPROVIDERSYNC_DEFAULTENV_TESTCONF_REPOSITORIES_INCLUDE", test.includeValue)
			t.Setenv("GPS_GITPROVIDERSYNC_DEFAULTENV_TESTCONF_REPOSITORIES_EXCLUDE", test.excludeValue)

			// Create minimal config file
			tempDir := t.TempDir()
			configContent := testConfigTemplate
			configFile := filepath.Join(tempDir, "test.yaml")
			err := os.WriteFile(configFile, []byte(configContent), 0600)
			require.NoError(t, err)

			appConfiguration := &dto.AppConfiguration{}
			err = ReadConfigurationFile(context.Background(), configFile, false, appConfiguration)
			require.NoError(t, err)

			// Check that repositories were processed correctly
			repoConfig := appConfiguration.GitProviderSyncConfs["defaultenv"]["testconf"]
			includeStr := fmt.Sprintf("%v", repoConfig.Repositories.Include)
			excludeStr := fmt.Sprintf("%v", repoConfig.Repositories.Exclude)

			for _, repo := range test.expectedInclude {
				require.Contains(t, includeStr, repo, "Include should contain %s", repo)
			}

			for _, repo := range test.expectedExclude {
				require.Contains(t, excludeStr, repo, "Exclude should contain %s", repo)
			}
		})
	}
}

// CopyDir recursively copies a directory from src to dst for test isolation.
func copyDir(src, dst string) error {
	if err := filepath.WalkDir(src, func(path string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		dstPath := filepath.Join(dst, relPath)

		if dirEntry.IsDir() {
			// Create directory
			info, err := dirEntry.Info()
			if err != nil {
				return fmt.Errorf("failed to get dir info: %w", err)
			}
			if err := os.MkdirAll(dstPath, info.Mode()); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

			return nil
		}

		// Copy file
		return copyFile(path, dstPath)
	}); err != nil {
		return fmt.Errorf("failed to walk source directory: %w", err)
	}

	return nil
}

// CopyFile copies a single file from src to dst for test isolation.
func copyFile(src, dst string) error {
	// Create destination directory if needed
	dstDir := filepath.Dir(dst)

	err := os.MkdirAll(dstDir, 0750)
	if err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Open source file
	// #nosec G304 - src path comes from trusted test data
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}

	defer func() { _ = srcFile.Close() }()

	// Create destination file
	// #nosec G304 - dst path is controlled by test
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}

	defer func() { _ = dstFile.Close() }()

	// Copy file contents
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Get source file info for permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	// Set destination file permissions
	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	return nil
}
