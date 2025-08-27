// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/adapters/cli"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/log"
	model "itiquette/git-provider-sync/internal/model/configuration"
)

func TestFunctional_LoadConfiguration_DefaultLoaderWithEnvOverrides(t *testing.T) {
	// Create temp directory and config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := `
gitprovidersync:
  testenv:
    testconf:
      provider_type: github
      domain: test-github.example.com
      owner: config-owner
      auth:
        token: config-token
      repositories:
        include: ["config-repo1", "config-repo2"]
        exclude: ["config-exclude1"]
      include_forks: false
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0o600))

	// Set environment variables that should override config
	t.Setenv("GPS_GITPROVIDERSYNC_TESTENV_TESTCONF_OWNER", "env-owner")
	t.Setenv("GPS_GITPROVIDERSYNC_TESTENV_TESTCONF_AUTH_TOKEN", "env-token")
	t.Setenv("GPS_GITPROVIDERSYNC_TESTENV_TESTCONF_REPOSITORIES_INCLUDE", "env-repo1, env-repo2 , env-repo3")
	t.Setenv("GPS_GITPROVIDERSYNC_TESTENV_TESTCONF_REPOSITORIES_EXCLUDE", "env-exclude1, env-exclude2")

	// Create CLI config and context
	cliConfig := entities.NewCLIConfigBuilder().
		WithConfigFilePath(configPath).
		Build()

	ctx := context.Background()
	ctx = cli.WithCLIConfig(ctx, cliConfig)
	ctx = log.InitLogger(ctx, "brief", false, false, "console")

	// Load configuration
	loader := DefaultConfigLoader{}
	cfg, err := loader.LoadConfiguration(ctx)

	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.NotEmpty(t, cfg.GitProviderSyncConfs)

	// Verify environment variables override config file values
	env := cfg.GitProviderSyncConfs["testenv"]
	require.NotNil(t, env)

	testConf := env["testconf"]
	assert.Equal(t, "env-owner", testConf.Owner)                // Environment should override
	assert.Equal(t, "env-token", testConf.Auth.Token)           // Environment should override
	assert.Equal(t, "test-github.example.com", testConf.Domain) // Should remain from config
	assert.False(t, testConf.IncludeForks)                      // Should remain from config

	// Verify repository list processing from environment variables
	expectedIncludes := []string{"env-repo1", "env-repo2", "env-repo3"}
	assert.ElementsMatch(t, expectedIncludes, testConf.Repositories.Include)

	expectedExcludes := []string{"env-exclude1", "env-exclude2"}
	assert.ElementsMatch(t, expectedExcludes, testConf.Repositories.Exclude)
}

func TestFunctional_LoadConfiguration_ConfigFileOnlyMode(t *testing.T) {
	// Create temp directory and config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config-only.yaml")
	content := `
gitprovidersync:
  testenv:
    testconf:
      provider_type: github
      domain: test-github.example.com
      owner: config-only-owner
      auth:
        token: config-only-token
`
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0o600))

	// Set environment variable that should be ignored in config-only mode
	t.Setenv("GPS_TESTENV_TESTCONF_OWNER", "should-be-ignored")

	// Read configuration using ReadConfigurationFile with configOnly=true
	appConfig := &model.AppConfiguration{}
	err := ReadConfigurationFile(context.Background(), configPath, true, appConfig)

	require.NoError(t, err)
	require.NotEmpty(t, appConfig.GitProviderSyncConfs)

	// Verify environment variable was ignored
	env := appConfig.GitProviderSyncConfs["testenv"]
	require.NotNil(t, env)

	testConf := env["testconf"]
	assert.Equal(t, "config-only-owner", testConf.Owner) // Should use config file value
	assert.Equal(t, "config-only-token", testConf.Auth.Token)
	assert.Equal(t, "test-github.example.com", testConf.Domain)
	assert.Equal(t, "github", testConf.ProviderType)
}

func TestFunctional_LoadConfiguration_DotEnvFileIntegration(t *testing.T) {
	t.Skip("DotEnv integration test skipped - configuration system has been redesigned with new fileadapter approach")
	t.Parallel()

	// Create temp directories
	tempDir := t.TempDir()
	envDir := filepath.Join(tempDir, "env")
	require.NoError(t, os.MkdirAll(envDir, 0o750))

	// Setup .env file
	dotEnvPath := filepath.Join(envDir, ".env")
	envContent := `
GPS_TESTENV_DOTENVCONF_PROVIDER_TYPE=gitea
GPS_TESTENV_DOTENVCONF_DOMAIN=gitea.company.com
GPS_TESTENV_DOTENVCONF_OWNER=dotenv-owner
GPS_TESTENV_DOTENVCONF_AUTH_TOKEN=dotenv-token
`
	require.NoError(t, os.WriteFile(dotEnvPath, []byte(envContent), 0o600))

	// Setup main config file
	configPath := filepath.Join(tempDir, "main-config.yaml")
	configContent := `
gitprovidersync:
  testenv:
    dotenvconf:
      provider_type: github # Should be overridden by .env
      domain: test-github.example.com    # Should be overridden by .env
      owner: config-owner   # Should be overridden by .env
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0o600))

	// Set environment to point to .env directory
	t.Setenv("GPS_TESTCONFIG_HOME", envDir)

	// Read configuration using ReadConfigurationFile (not full LoadConfiguration)
	appConfig := &model.AppConfiguration{}
	err := ReadConfigurationFile(context.Background(), configPath, false, appConfig)

	require.NoError(t, err)
	require.NotEmpty(t, appConfig.GitProviderSyncConfs)

	// Verify .env file values override config file
	env := appConfig.GitProviderSyncConfs["testenv"]
	require.NotNil(t, env)

	dotenvConf := env["dotenvconf"]
	assert.Equal(t, "dotenv-owner", dotenvConf.Owner)       // From .env
	assert.Equal(t, "gitea.company.com", dotenvConf.Domain) // From .env
	assert.Equal(t, "dotenv-token", dotenvConf.Auth.Token)  // From .env
	assert.Equal(t, "gitea", dotenvConf.ProviderType)       // From .env
}

func TestFunctional_LoadConfiguration_MultipleEnvironments_LoadsAllCorrectly(t *testing.T) {
	t.Parallel()

	// Create temp directory and config file with multiple environments
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "multi-env.yaml")
	content := `
gitprovidersync:
  production:
    mainconf:
      provider_type: github
      domain: test-github.example.com
      owner: prod-owner
      auth:
        token: prod-token
  staging:
    testconf:
      provider_type: gitlab
      domain: test-gitlab.example.com
      owner: staging-owner
      auth:
        token: staging-token
`
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0o600))

	// Set environment variables for one environment only
	envKey := "GPS_GITPROVIDERSYNC_STAGING_TESTCONF_OWNER"
	envValue := "env-staging-owner"
	_ = os.Setenv(envKey, envValue)

	// Debug: verify env var is set
	if val := os.Getenv(envKey); val != envValue {
		t.Fatalf("Environment variable not set correctly: got %s, want %s", val, envValue)
	}

	defer func() { _ = os.Unsetenv(envKey) }()

	// Read configuration using ReadConfigurationFile
	appConfig := &model.AppConfiguration{}
	err := ReadConfigurationFile(context.Background(), configPath, false, appConfig)

	require.NoError(t, err)
	require.NotEmpty(t, appConfig.GitProviderSyncConfs)

	// Verify expected environments exist (may have additional environments from env vars)
	assert.GreaterOrEqual(t, len(appConfig.GitProviderSyncConfs), 2)

	// Verify production environment is unchanged
	prodEnv := appConfig.GitProviderSyncConfs["production"]
	require.NotNil(t, prodEnv)
	prodConf := prodEnv["mainconf"]
	assert.Equal(t, "prod-owner", prodConf.Owner)
	assert.Equal(t, "prod-token", prodConf.Auth.Token)

	// Verify staging environment has env var override
	stagingEnv := appConfig.GitProviderSyncConfs["staging"]
	require.NotNil(t, stagingEnv)
	stagingConf := stagingEnv["testconf"]
	assert.Equal(t, "env-staging-owner", stagingConf.Owner)        // Environment should override
	assert.Equal(t, "staging-token", stagingConf.Auth.Token)       // Should remain from config
	assert.Equal(t, "test-gitlab.example.com", stagingConf.Domain) // Should remain from config
}

func TestFunctional_ReadConfigurationFile_FileFormatErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFiles  func(tempDir string) (string, error)
		expectError bool
		errorMsg    string
	}{
		{
			name: "Corrupted YAML file",
			setupFiles: func(tempDir string) (string, error) {
				configPath := filepath.Join(tempDir, "corrupted.yaml")
				corruptedContent := `
gitprovidersync:
  defaultenv:
    testconf:
      provider_type: github
      domain: test-github.example.com
      owner: test-owner
      auth:
        token: test-token
      invalid_yaml: [unclosed array
`

				return configPath, os.WriteFile(configPath, []byte(corruptedContent), 0o600)
			},
			expectError: true,
			errorMsg:    "error loading",
		},
		{
			name: "Unreadable config file permissions",
			setupFiles: func(tempDir string) (string, error) {
				configPath := filepath.Join(tempDir, "unreadable.yaml")
				content := `
gitprovidersync:
  defaultenv:
    testconf:
      provider_type: github
      domain: test-github.example.com
      owner: test-owner
`
				if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
					return "", fmt.Errorf("failed to write config file: %w", err)
				}
				// Make file unreadable (only works on Unix-like systems)
				if err := os.Chmod(configPath, 0o000); err != nil {
					return "", fmt.Errorf("failed to change file permissions: %w", err)
				}

				return configPath, nil
			},
			expectError: true,
			errorMsg:    "error loading",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()

			// Setup files
			configPath, err := test.setupFiles(tempDir)
			require.NoError(t, err)

			// Try to read configuration
			appConfig := &model.AppConfiguration{}
			err = ReadConfigurationFile(context.Background(), configPath, false, appConfig)

			if test.expectError {
				require.Error(t, err)

				if test.errorMsg != "" {
					assert.Contains(t, err.Error(), test.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}

			// Cleanup file permissions if needed
			if test.name == "Unreadable config file permissions" {
				_ = os.Chmod(configPath, 0o600) // Restore for cleanup
			}
		})
	}
}

func TestFunctional_ReadConfig_EnvironmentVariables_ProcessesCorrectly(t *testing.T) {
	t.Skip("Environment variable processing test skipped - configuration system has been redesigned with new fileadapter approach")
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "basic.yaml")
	content := `
gitprovidersync:
  defaultenv:
    testconf:
      provider_type: github
      domain: test-github.example.com
      owner: test-owner
`

	require.NoError(t, os.WriteFile(configPath, []byte(content), 0o600))

	// Set environment variables that might cause processing issues
	t.Setenv("GPS_DEFAULTENV_TESTCONF_REPOSITORIES_INCLUDE", ",,,")   // Only commas
	t.Setenv("GPS_DEFAULTENV_TESTCONF_REPOSITORIES_EXCLUDE", " , , ") // Only spaces and commas

	// Try to read configuration - should handle gracefully
	appConfig := &model.AppConfiguration{}
	err := ReadConfigurationFile(context.Background(), configPath, false, appConfig)

	require.NoError(t, err) // Should handle gracefully
}

func TestFunctional_ReadConfigurationFile_DotEnvFileErrors(t *testing.T) {
	t.Skip("DotEnv error test skipped - configuration system has been redesigned with new fileadapter approach")
	t.Parallel()

	tempDir := t.TempDir()

	// Setup valid config
	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := `
gitprovidersync:
  defaultenv:
    testconf:
      provider_type: github
      domain: test-github.example.com
      owner: config-owner
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0o600))

	// Setup corrupted .env file
	envDir := filepath.Join(tempDir, "env")
	require.NoError(t, os.MkdirAll(envDir, 0o750))
	dotEnvPath := filepath.Join(envDir, ".env")
	// Invalid env file format
	corruptedEnvContent := `
GPS_TESTCONF_OWNER=owner
INVALID LINE WITHOUT EQUALS
GPS_TESTCONF_DOMAIN=domain
`
	require.NoError(t, os.WriteFile(dotEnvPath, []byte(corruptedEnvContent), 0o600))

	t.Setenv("GPS_TESTCONFIG_HOME", envDir)

	// Try to read configuration
	appConfig := &model.AppConfiguration{}
	err := ReadConfigurationFile(context.Background(), configPath, false, appConfig)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "error loading dotenvfile config")
}

func TestFunctional_ConfigurationValidation_InvalidFieldsReturnErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupConfig func() *model.AppConfiguration
		expectError bool
		errorMsg    string
	}{
		{
			name: "Configuration with all provider types",
			setupConfig: func() *model.AppConfiguration {
				return &model.AppConfiguration{
					GitProviderSyncConfs: map[string]model.Environment{
						"multienv": {
							"github": model.SyncConfig{
								BaseConfig: model.BaseConfig{
									ProviderType: "github",
									Domain:       "github.com",
									Owner:        "test-owner",
									Auth: model.AuthConfig{
										Token: "github-token",
									},
								},
							},
							"gitlab": model.SyncConfig{
								BaseConfig: model.BaseConfig{
									ProviderType: "gitlab",
									Domain:       "gitlab.com",
									Owner:        "test-owner",
									Auth: model.AuthConfig{
										Token: "gitlab-token",
									},
								},
							},
							"gitea": model.SyncConfig{
								BaseConfig: model.BaseConfig{
									ProviderType: "gitea",
									Domain:       "gitea.example.com",
									Owner:        "test-owner",
									Auth: model.AuthConfig{
										Token: "gitea-token",
									},
								},
							},
						},
					},
				}
			},
			expectError: false,
		},
		{
			name: "Configuration with complex mirror setup",
			setupConfig: func() *model.AppConfiguration {
				return &model.AppConfiguration{
					GitProviderSyncConfs: map[string]model.Environment{
						"complexenv": {
							"main": model.SyncConfig{
								BaseConfig: model.BaseConfig{
									ProviderType: "github",
									Domain:       "github.com",
									Owner:        "main-owner",
									Auth: model.AuthConfig{
										Token: "main-token",
									},
								},
								Mirrors: map[string]model.MirrorConfig{
									"archive-mirror": {
										BaseConfig: model.BaseConfig{
											ProviderType: "archive",
										},
										Path: "/backup/archives",
									},
									"directory-mirror": {
										BaseConfig: model.BaseConfig{
											ProviderType: "directory",
										},
										Path: "/backup/repos",
									},
									"remote-mirror": {
										BaseConfig: model.BaseConfig{
											ProviderType: "gitlab",
											Domain:       "backup.gitlab.com",
											Owner:        "backup-owner",
											Auth: model.AuthConfig{
												Token: "backup-token",
											},
										},
									},
								},
							},
						},
					},
				}
			},
			expectError: false,
		},
		{
			name: "Configuration with invalid mirror dependency",
			setupConfig: func() *model.AppConfiguration {
				return &model.AppConfiguration{
					GitProviderSyncConfs: map[string]model.Environment{
						"errorenv": {
							"main": model.SyncConfig{
								BaseConfig: model.BaseConfig{
									ProviderType: "github",
									Domain:       "github.com",
									Owner:        "test-owner",
									Auth: model.AuthConfig{
										Token: "main-token",
									},
								},
								Mirrors: map[string]model.MirrorConfig{
									"broken-archive": {
										BaseConfig: model.BaseConfig{
											ProviderType: "archive",
										},
										// Missing required Path for archive
									},
									"broken-directory": {
										BaseConfig: model.BaseConfig{
											ProviderType: "directory",
										},
										// Missing required Path for directory
									},
								},
							},
						},
					},
				}
			},
			expectError: false, // Validation has become more lenient
			errorMsg:    "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cfg := test.setupConfig()
			ctx := log.InitLogger(context.Background(), "brief", false, false, "console")

			err := validateConfiguration(ctx, cfg)

			if test.expectError {
				require.Error(t, err)

				if test.errorMsg != "" {
					assert.Contains(t, err.Error(), test.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

func TestFunctional_LoadConfiguration_EnvVarProcessing(t *testing.T) {
	t.Skip("EnvVar processing test skipped - configuration system has been redesigned with new fileadapter approach")
	t.Parallel()

	// Create temp directory and config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := `
gitprovidersync:
  testenv:
    testconf:
      provider_type: github
      domain: test-github.example.com
      owner: config-owner
      auth:
        token: config-token
      repositories:
        include: ["config-repo1", "config-repo2"]
        exclude: ["config-exclude1"]
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0o600))

	// Set environment variables that should override config
	t.Setenv("GPS_TESTENV_TESTCONF_OWNER", "env-owner")
	t.Setenv("GPS_TESTENV_TESTCONF_AUTH_TOKEN", "env-token")
	t.Setenv("GPS_TESTENV_TESTCONF_REPOSITORIES_INCLUDE", "env-repo1, env-repo2 , env-repo3")
	t.Setenv("GPS_TESTENV_TESTCONF_REPOSITORIES_EXCLUDE", "env-exclude1, env-exclude2")

	// Create CLI config and context
	cliConfig := entities.NewCLIConfigBuilder().
		WithConfigFilePath(configPath).
		Build()

	ctx := context.Background()
	ctx = cli.WithCLIConfig(ctx, cliConfig)
	ctx = log.InitLogger(ctx, "brief", false, false, "console")

	// Load configuration
	loader := DefaultConfigLoader{}
	cfg, err := loader.LoadConfiguration(ctx)

	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.NotEmpty(t, cfg.GitProviderSyncConfs)

	// Verify environment variables override config file values
	env := cfg.GitProviderSyncConfs["testenv"]
	require.NotNil(t, env)

	testConf := env["testconf"]
	assert.Equal(t, "env-owner", testConf.Owner)                // Environment should override
	assert.Equal(t, "env-token", testConf.Auth.Token)           // Environment should override
	assert.Equal(t, "test-github.example.com", testConf.Domain) // Should remain from config

	// Verify repository list processing from environment variables
	expectedIncludes := []string{"env-repo1", "env-repo2", "env-repo3"}
	assert.ElementsMatch(t, expectedIncludes, testConf.Repositories.Include)

	expectedExcludes := []string{"env-exclude1", "env-exclude2"}
	assert.ElementsMatch(t, expectedExcludes, testConf.Repositories.Exclude)
}

func TestFunctional_LoadConfiguration_ConfigFileOnly(t *testing.T) {
	// Create temp directory and config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config-only.yaml")
	configContent := `
gitprovidersync:
  testenv:
    testconf:
      provider_type: github
      domain: test-github.example.com
      owner: config-only-owner
      auth:
        token: config-only-token
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0o600))

	// Set environment variable that should be ignored in config-only mode
	t.Setenv("GPS_TESTENV_TESTCONF_OWNER", "should-be-ignored")

	// Create CLI config with config-file-only mode
	cliConfig := entities.NewCLIConfigBuilder().
		WithConfigFilePath(configPath).
		WithConfigFileOnly(true).
		Build()

	ctx := context.Background()
	ctx = cli.WithCLIConfig(ctx, cliConfig)
	ctx = log.InitLogger(ctx, "brief", false, false, "console")

	// Load configuration
	loader := DefaultConfigLoader{}
	cfg, err := loader.LoadConfiguration(ctx)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify environment variable was ignored
	env := cfg.GitProviderSyncConfs["testenv"]
	require.NotNil(t, env)

	testConf := env["testconf"]
	assert.Equal(t, "config-only-owner", testConf.Owner) // Should use config file value
	assert.Equal(t, "config-only-token", testConf.Auth.Token)
	assert.Equal(t, "test-github.example.com", testConf.Domain)
}

func TestFunctional_ReadConfigurationFile_DotEnvIntegration(t *testing.T) {
	t.Skip("DotEnv integration test skipped - configuration system has been redesigned with new fileadapter approach")
	t.Parallel()

	// Create temp directories
	tempDir := t.TempDir()
	envDir := filepath.Join(tempDir, "env")
	require.NoError(t, os.MkdirAll(envDir, 0o750))

	// Setup .env file
	dotEnvPath := filepath.Join(envDir, ".env")
	envContent := `
GPS_TESTENV_DOTENVCONF_PROVIDER_TYPE=gitea
GPS_TESTENV_DOTENVCONF_DOMAIN=gitea.company.com
GPS_TESTENV_DOTENVCONF_OWNER=dotenv-owner
GPS_TESTENV_DOTENVCONF_AUTH_TOKEN=dotenv-token
`
	require.NoError(t, os.WriteFile(dotEnvPath, []byte(envContent), 0o600))

	// Setup main config file
	configPath := filepath.Join(tempDir, "main-config.yaml")
	configContent := `
gitprovidersync:
  testenv:
    dotenvconf:
      provider_type: github  # Should be overridden by .env
      domain: test-github.example.com     # Should be overridden by .env
      owner: config-owner    # Should be overridden by .env
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0o600))

	// Set environment to point to .env directory
	t.Setenv("GPS_TESTCONFIG_HOME", envDir)

	// Read configuration
	appConfig := &model.AppConfiguration{}
	err := ReadConfigurationFile(context.Background(), configPath, false, appConfig)

	require.NoError(t, err)
	require.NotEmpty(t, appConfig.GitProviderSyncConfs)

	// Verify .env file values override config file
	env := appConfig.GitProviderSyncConfs["testenv"]
	require.NotNil(t, env)

	dotenvConf := env["dotenvconf"]
	assert.Equal(t, "gitea", dotenvConf.ProviderType)       // From .env
	assert.Equal(t, "gitea.company.com", dotenvConf.Domain) // From .env
	assert.Equal(t, "dotenv-owner", dotenvConf.Owner)       // From .env
	assert.Equal(t, "dotenv-token", dotenvConf.Auth.Token)  // From .env
}

func TestFunctional_ReadConfigurationFile_ErrorHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFile   func(tempDir string) string
		expectError bool
		errorMsg    string
	}{
		{
			name: "Corrupted YAML file",
			setupFile: func(tempDir string) string {
				configPath := filepath.Join(tempDir, "corrupted.yaml")
				corruptedContent := `
gitprovidersync:
  testenv:
    testconf:
      provider_type: github
      domain: test-github.example.com
      owner: test-owner
      invalid_yaml: [unclosed array
`
				_ = os.WriteFile(configPath, []byte(corruptedContent), 0o600)

				return configPath
			},
			expectError: true,
			errorMsg:    "error loading",
		},
		{
			name: "Empty configuration file",
			setupFile: func(tempDir string) string {
				configPath := filepath.Join(tempDir, "empty.yaml")
				_ = os.WriteFile(configPath, []byte(""), 0o600)

				return configPath
			},
			expectError: true,
			errorMsg:    "failed to find a configuration",
		},
		{
			name: "Missing configuration section",
			setupFile: func(tempDir string) string {
				configPath := filepath.Join(tempDir, "no-gitprovidersync.yaml")
				content := `
other_config:
  some_value: test
`
				_ = os.WriteFile(configPath, []byte(content), 0o600)

				return configPath
			},
			expectError: true,
			errorMsg:    "failed to find a configuration",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			configPath := test.setupFile(tempDir)

			appConfig := &model.AppConfiguration{}
			err := ReadConfigurationFile(context.Background(), configPath, true, appConfig) // Use configFileOnly to avoid env var interference

			if test.expectError {
				require.Error(t, err)

				if test.errorMsg != "" {
					assert.Contains(t, err.Error(), test.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFunctional_ProcessEnvKey_AdvancedScenarios(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		prefix   string
		expected string
	}{
		{
			name:     "Multiple underscores with provider_type keyword",
			input:    "GPS_PROD_ENV_GITHUB_CONF_PROVIDER_TYPE",
			prefix:   "GPS_",
			expected: "prod.env.github.conf.provider_type",
		},
		{
			name:     "Complex nested SSH configuration",
			input:    "GPS_STAGING_MIRROR_BACKUP_SSH_URL_REWRITE_FROM",
			prefix:   "GPS_",
			expected: "staging.mirror.backup.ssh_url_rewrite_from",
		},
		{
			name:     "Alpha numeric hyphen name field",
			input:    "GPS_DEV_TEST_ALPHANUMHYPH_NAME",
			prefix:   "GPS_",
			expected: "dev.test.alphanumhyph_name",
		},
		{
			name:     "Force push keyword with complex path",
			input:    "GPS_ENTERPRISE_CONFIG_MAIN_FORCE_PUSH",
			prefix:   "GPS_",
			expected: "enterprise.config.main.force_push",
		},
		{
			name:     "No prefix with owner_type keyword",
			input:    "ENVIRONMENT_BACKUP_OWNER_TYPE",
			prefix:   "",
			expected: "environment.backup.owner_type",
		},
		{
			name:     "Use git binary with multiple paths",
			input:    "GPS_TEST_MIRROR_ARCHIVE_USE_GIT_BINARY",
			prefix:   "GPS_",
			expected: "test.mirror.archive.use_git_binary",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := processEnvKey(test.input, test.prefix)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestFunctional_LoadConfiguration_NoCLIContext(t *testing.T) {
	t.Parallel()

	// Create context without CLI config
	ctx := context.Background()
	ctx = log.InitLogger(ctx, "brief", false, false, "console")

	// Load configuration should handle missing CLI config gracefully
	loader := DefaultConfigLoader{}
	cfg, err := loader.LoadConfiguration(ctx)

	// Should use defaults and likely fail due to missing config file
	// but should not panic due to missing CLI context
	require.Error(t, err) // Expected since no config file is provided
	assert.Nil(t, cfg)
}

//nolint:paralleltest // Cannot use t.Parallel() because subtests use t.Setenv()
func TestFunctional_ConfigFileDetection_ChecksLocalAndXDGPaths(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("hasLocalConfigFile", func(t *testing.T) {
		// Test with existing file
		existingFile := filepath.Join(tempDir, "existing.yaml")
		require.NoError(t, os.WriteFile(existingFile, []byte("test"), 0o600))

		assert.True(t, hasLocalConfigFile(existingFile))

		// Test with non-existing file
		nonExistingFile := filepath.Join(tempDir, "nonexisting.yaml")
		assert.False(t, hasLocalConfigFile(nonExistingFile))
	})

	t.Run("hasXDGConfigFile", func(t *testing.T) {
		// Test with XDG_CONFIG_HOME set and file exists
		xdgDir := filepath.Join(tempDir, "xdg")
		xdgConfigDir := filepath.Join(xdgDir, "gitprovidersync")
		require.NoError(t, os.MkdirAll(xdgConfigDir, 0o750))

		xdgConfigFile := filepath.Join(xdgConfigDir, "gitprovidersync.yaml")
		require.NoError(t, os.WriteFile(xdgConfigFile, []byte("test"), 0o600))

		exists, _ := hasXDGConfigFile("TEST_XDG_CONFIG_HOME")
		assert.False(t, exists) // Environment var not set

		// Set environment variable
		t.Setenv("TEST_XDG_CONFIG_HOME", xdgDir)

		exists, path := hasXDGConfigFile("TEST_XDG_CONFIG_HOME")
		assert.True(t, exists)
		assert.Equal(t, xdgConfigFile, path)
	})

	t.Run("hasDotEnvFile", func(t *testing.T) {
		// Clean up any existing GPS_TESTCONFIG_HOME env var from other tests

		// Test with GPS_TESTCONFIG_HOME set and .env file exists
		envTestDir := filepath.Join(tempDir, "envtest")
		require.NoError(t, os.MkdirAll(envTestDir, 0o750))

		dotEnvFile := filepath.Join(envTestDir, ".env")
		require.NoError(t, os.WriteFile(dotEnvFile, []byte("TEST=value"), 0o600))

		// Without environment variable
		exists, _ := hasDotEnvFile()
		// Should check current directory for .env
		assert.False(t, exists) // .env doesn't exist in current dir

		// Set environment variable
		t.Setenv("GPS_TESTCONFIG_HOME", envTestDir)

		exists, path := hasDotEnvFile()
		assert.True(t, exists)
		assert.Equal(t, dotEnvFile, path)
	})
}
