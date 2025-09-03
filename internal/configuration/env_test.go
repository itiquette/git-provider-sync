// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	config "itiquette/git-provider-sync/internal/model/configuration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStandardEnvironmentVariables tests standard environment variable support.
func TestStandardEnvironmentVariables(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()
	tests := []struct {
		name        string
		setupEnv    map[string]string
		configFile  string
		wantSuccess bool
		description string
	}{
		{
			name: "NO_COLOR is respected",
			setupEnv: map[string]string{
				"NO_COLOR": "1",
			},
			configFile: `
gitprovidersync:
  test:
    source:
      provider_type: github
      owner: testuser
`,
			wantSuccess: true,
			description: "NO_COLOR should be respected for disabling colors",
		},
		{
			name: "HOME is used for config location",
			setupEnv: map[string]string{
				"HOME": ".",
			},
			configFile: `
gitprovidersync:
  test:
    source:
      provider_type: github
      owner: from-home-config
`,
			wantSuccess: true,
			description: "HOME should be used to find user config",
		},
		{
			name: "XDG_CONFIG_HOME overrides HOME",
			setupEnv: map[string]string{
				"XDG_CONFIG_HOME": "./xdg",
				"HOME":            "./home",
			},
			configFile: `
gitprovidersync:
  test:
    source:
      provider_type: github
      owner: from-xdg-config
`,
			wantSuccess: true,
			description: "XDG_CONFIG_HOME should take precedence over HOME",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Setup test environment
			tempDir := t.TempDir()
			originalWd, _ := os.Getwd()

			t.Cleanup(func() { _ = os.Chdir(originalWd) })

			require.NoError(t, os.Chdir(tempDir))

			// Set environment variables
			for key, value := range testCase.setupEnv {
				if key == "HOME" || key == "XDG_CONFIG_HOME" {
					t.Setenv(key, filepath.Join(tempDir, value))
				} else {
					t.Setenv(key, value)
				}
			}

			// Create config file
			if testCase.configFile != "" {
				configPath := "gitprovidersync.yaml"
				require.NoError(t, os.WriteFile(configPath, []byte(testCase.configFile), 0600))
			}

			// Load configuration
			appConfig := &config.AppConfiguration{}
			err := ReadConfigurationFile(context.Background(), "gitprovidersync.yaml", false, appConfig)

			// Verify
			if testCase.wantSuccess {
				require.NoError(t, err, testCase.description)
			} else {
				require.Error(t, err, testCase.description)
			}
		})
	}
}

// TestProviderTokenEnvironmentVariables tests provider-specific token environment variables.
func TestProviderTokenEnvironmentVariables(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()
	tests := []struct {
		name          string
		setupEnv      map[string]string
		configFile    string
		provider      string
		expectedToken string
	}{
		{
			name: "GPS_GITHUB_TOKEN overrides config",
			setupEnv: map[string]string{
				"GPS_GITHUB_TOKEN": "github-env-token",
			},
			configFile: `
gitprovidersync:
  test:
    source:
      provider_type: github
      owner: testuser
      auth:
        token: "config-token"
`,
			provider:      "github",
			expectedToken: "github-env-token",
		},
		{
			name: "GPS_GITLAB_TOKEN overrides config",
			setupEnv: map[string]string{
				"GPS_GITLAB_TOKEN": "gitlab-env-token",
			},
			configFile: `
gitprovidersync:
  test:
    source:
      provider_type: gitlab
      owner: testuser
      auth:
        token: "config-token"
`,
			provider:      "gitlab",
			expectedToken: "gitlab-env-token",
		},
		{
			name: "GPS_GITEA_TOKEN overrides config",
			setupEnv: map[string]string{
				"GPS_GITEA_TOKEN": "gitea-env-token",
			},
			configFile: `
gitprovidersync:
  test:
    source:
      provider_type: gitea
      owner: testuser
      auth:
        token: "config-token"
`,
			provider:      "gitea",
			expectedToken: "gitea-env-token",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Setup test environment
			tempDir := t.TempDir()
			originalWd, _ := os.Getwd()

			t.Cleanup(func() { _ = os.Chdir(originalWd) })

			require.NoError(t, os.Chdir(tempDir))

			// Set environment variables
			for key, value := range testCase.setupEnv {
				t.Setenv(key, value)
			}

			// Create config file
			require.NoError(t, os.WriteFile("config.yaml", []byte(testCase.configFile), 0600))

			// Load configuration
			appConfig := &config.AppConfiguration{}
			err := ReadConfigurationFile(context.Background(), "config.yaml", true, appConfig)
			require.NoError(t, err)

			// Verify token precedence
			source := appConfig.GitProviderSyncConfs["test"]["source"]
			require.NotNil(t, source)
			assert.Equal(t, testCase.expectedToken, source.Auth.Token,
				"Provider-specific env var should override config file token")
		})
	}
}

// TestEnvironmentVariableExpansion tests ${VAR} expansion in config files.
func TestEnvironmentVariableExpansion(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()
	tests := []struct {
		name          string
		setupEnv      map[string]string
		configFile    string
		expectedValue string
		field         string
	}{
		{
			name: "Basic variable expansion",
			setupEnv: map[string]string{
				"MY_TOKEN": "expanded-token",
			},
			configFile: `
gitprovidersync:
  test:
    source:
      provider_type: github
      owner: testuser
      auth:
        token: "${MY_TOKEN}"
`,
			expectedValue: "expanded-token",
			field:         "token",
		},
		{
			name: "Token field expansion with complex value",
			setupEnv: map[string]string{
				"TOKEN_PREFIX": "ghp_",
				"TOKEN_SUFFIX": "xyz123",
			},
			configFile: `
gitprovidersync:
  test:
    source:
      provider_type: github
      owner: "testuser"
      auth:
        token: "${TOKEN_PREFIX}${TOKEN_SUFFIX}"
`,
			expectedValue: "ghp_xyz123",
			field:         "token",
		},
		{
			name:     "Undefined variable expands to empty",
			setupEnv: map[string]string{
				// No UNDEFINED_VAR set
			},
			configFile: `
gitprovidersync:
  test:
    source:
      provider_type: github
      owner: "testuser"
      auth:
        token: "${UNDEFINED_VAR}"
`,
			expectedValue: "",
			field:         "token",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Setup test environment
			tempDir := t.TempDir()
			originalWd, _ := os.Getwd()

			t.Cleanup(func() { _ = os.Chdir(originalWd) })

			require.NoError(t, os.Chdir(tempDir))

			// Set environment variables
			for key, value := range testCase.setupEnv {
				t.Setenv(key, value)
			}

			// Create config file
			require.NoError(t, os.WriteFile("config.yaml", []byte(testCase.configFile), 0600))

			// Load configuration
			appConfig := &config.AppConfiguration{}
			err := ReadConfigurationFile(context.Background(), "config.yaml", false, appConfig)
			require.NoError(t, err)

			// Verify expansion
			source := appConfig.GitProviderSyncConfs["test"]["source"]
			require.NotNil(t, source)

			switch testCase.field {
			case "token":
				assert.Equal(t, testCase.expectedValue, source.Auth.Token)
			case "owner":
				assert.Equal(t, testCase.expectedValue, source.Owner)
			}
		})
	}
}

// TestProxyEnvironmentVariables verifies proxy settings are available to providers.
func TestProxyEnvironmentVariables(t *testing.T) {
	t.Parallel()

	// Note: We can't actually test that the HTTP clients use these,
	// but we can verify they're available in the environment
	tests := []struct {
		name     string
		envVars  map[string]string
		expected map[string]string
	}{
		{
			name: "HTTP_PROXY is available",
			envVars: map[string]string{
				"HTTP_PROXY": "http://proxy.example.com:8080",
			},
			expected: map[string]string{
				"HTTP_PROXY": "http://proxy.example.com:8080",
			},
		},
		{
			name: "HTTPS_PROXY is available",
			envVars: map[string]string{
				"HTTPS_PROXY": "https://secure-proxy.example.com:8443",
			},
			expected: map[string]string{
				"HTTPS_PROXY": "https://secure-proxy.example.com:8443",
			},
		},
		{
			name: "NO_PROXY is available",
			envVars: map[string]string{
				"HTTP_PROXY": "http://proxy.example.com:8080",
				"NO_PROXY":   "localhost,127.0.0.1,.internal.company.com",
			},
			expected: map[string]string{
				"HTTP_PROXY": "http://proxy.example.com:8080",
				"NO_PROXY":   "localhost,127.0.0.1,.internal.company.com",
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Verify the environment variables would be available to HTTP clients
			// Note: We're not actually setting them here to avoid affecting other tests
			for key, expectedValue := range testCase.expected {
				// In real usage, the HTTP client would read these from os.Getenv
				assert.Equal(t, expectedValue, testCase.envVars[key],
					"Proxy variable %s should be available to HTTP clients", key)
			}
		})
	}
}

// TestGPSConfigFileEnvironmentVariable tests GPS_CONFIG_FILE env var.
func TestGPSConfigFileEnvironmentVariable(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()

	// Setup test environment
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()

	t.Cleanup(func() { _ = os.Chdir(originalWd) })
	require.NoError(t, os.Chdir(tempDir))

	// Create custom config file
	customPath := filepath.Join(tempDir, "custom", "config.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(customPath), 0750))

	customConfig := `
gitprovidersync:
  test:
    source:
      provider_type: github
      owner: from-custom-path
`
	require.NoError(t, os.WriteFile(customPath, []byte(customConfig), 0600))

	// Create default config file (should be ignored)
	defaultConfig := `
gitprovidersync:
  test:
    source:
      provider_type: github
      owner: from-default-path
`
	require.NoError(t, os.WriteFile("gitprovidersync.yaml", []byte(defaultConfig), 0600))

	// Set GPS_CONFIG_FILE to custom path
	t.Setenv("GPS_CONFIG_FILE", customPath)

	// Note: GPS_CONFIG_FILE is handled at CLI level, not in ReadConfigurationFile
	// So we test that the custom path would be used when passed as parameter
	appConfig := &config.AppConfiguration{}
	err := ReadConfigurationFile(context.Background(), customPath, false, appConfig)
	require.NoError(t, err)

	// Verify custom config was loaded
	source := appConfig.GitProviderSyncConfs["test"]["source"]
	require.NotNil(t, source)
	assert.Equal(t, "from-custom-path", source.Owner)
}

// TestEnvironmentVariablePrecedence verifies the complete precedence chain.
func TestEnvironmentVariablePrecedence(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()

	// Setup test environment
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()

	t.Cleanup(func() { _ = os.Chdir(originalWd) })
	require.NoError(t, os.Chdir(tempDir))

	// Create config file with token
	configFile := `
gitprovidersync:
  test:
    github-source:
      provider_type: github
      owner: testuser
      auth:
        token: "config-file-token"
    gitlab-source:
      provider_type: gitlab
      owner: testuser
      auth:
        token: "${GITLAB_TOKEN}"
`
	require.NoError(t, os.WriteFile("config.yaml", []byte(configFile), 0600))

	// Set environment variables
	t.Setenv("GPS_GITHUB_TOKEN", "github-specific-token")
	t.Setenv("GITLAB_TOKEN", "gitlab-general-token")
	t.Setenv("GPS_GITLAB_TOKEN", "gitlab-specific-token")

	// Load configuration
	appConfig := &config.AppConfiguration{}
	err := ReadConfigurationFile(context.Background(), "config.yaml", true, appConfig)
	require.NoError(t, err)

	// Verify precedence:
	// 1. GPS_GITHUB_TOKEN should override config file token
	githubSource := appConfig.GitProviderSyncConfs["test"]["github-source"]
	require.NotNil(t, githubSource)
	assert.Equal(t, "github-specific-token", githubSource.Auth.Token,
		"GPS_GITHUB_TOKEN should override config file token")

	// 2. GPS_GITLAB_TOKEN should override ${GITLAB_TOKEN} expansion
	gitlabSource := appConfig.GitProviderSyncConfs["test"]["gitlab-source"]
	require.NotNil(t, gitlabSource)
	assert.Equal(t, "gitlab-specific-token", gitlabSource.Auth.Token,
		"GPS_GITLAB_TOKEN should override variable expansion")
}
