// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"itiquette/git-provider-sync/internal/adapters/configuration/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SetupTestEnvironment creates a test environment with files and environment variables.
func setupTestEnvironment(t *testing.T, setupFiles map[string]string, setupEnv map[string]string) string {
	t.Helper()
	tempDir := t.TempDir()

	// Create test files using absolute paths
	for path, content := range setupFiles {
		fullPath := filepath.Join(tempDir, path)
		dir := filepath.Dir(fullPath)

		// Handle XDG paths
		if filepath.Base(path) == "dto.yaml" {
			// is a user config, create proper XDG structure
			xdgPath := filepath.Join(dir, "gitprovidersync")
			require.NoError(t, os.MkdirAll(xdgPath, 0750))
			fullPath = filepath.Join(xdgPath, "dto.yaml")
		} else {
			require.NoError(t, os.MkdirAll(dir, 0750))
		}

		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0600))
	}

	// Set environment variables
	for key, value := range setupEnv {
		switch key {
		case "XDG_CONFIG_HOME":
			// Make XDG path absolute
			t.Setenv(key, filepath.Join(tempDir, value))
		case "GPS_TESTCONFIG_HOME":
			// Make test config path absolute
			if value == "." {
				t.Setenv(key, tempDir)
			} else {
				t.Setenv(key, filepath.Join(tempDir, value))
			}
		default:
			t.Setenv(key, value)
		}
	}

	return tempDir
}

// VerifyTestResult checks the loaded configuration against expected values.
func verifyTestResult(t *testing.T, err error, appConfig *dto.AppConfiguration, expectedOwner, description string) {
	t.Helper()

	if expectedOwner == "" {
		// Should fail to load
		require.Error(t, err, description)
	} else {
		// Should succeed and have expected value
		require.NoError(t, err, description)
		require.NotNil(t, appConfig.GitProviderSyncConfs["test"], "Config should have 'test' environment")
		require.NotNil(t, appConfig.GitProviderSyncConfs["test"]["source"], "Config should have 'source' in test env")

		actualOwner := appConfig.GitProviderSyncConfs["test"]["source"].Owner
		assert.Equal(t, expectedOwner, actualOwner, description)
	}
}

// TestConfigurationPrecedence verifies that configuration sources are loaded in the correct precedence order
// Order (highest to lowest): CLI flags > Environment variables > Project config > User config > System config
//
//nolint:paralleltest // Cannot use t.Parallel() with t.Setenv()
func TestConfigurationPrecedence(t *testing.T) {
	// Cannot use t.Parallel() when subtests use t.Setenv()
	tests := []struct {
		name           string
		setupFiles     map[string]string // path -> content
		setupEnv       map[string]string
		configFile     string // CLI flag equivalent
		configFileOnly bool   // CLI flag equivalent
		expectedOwner  string
		description    string
	}{
		{
			name: "user config only",
			setupFiles: map[string]string{
				"user/dto.yaml": `
gitprovidersync:
  test:
    source:
      provider_type: github
      owner: user-config-owner
      owner_type: user
`,
			},
			setupEnv: map[string]string{
				"XDG_CONFIG_HOME": "user", // Point to user config dir
			},
			expectedOwner: "user-config-owner",
			description:   "Should load from user config when it's the only source",
		},
		{
			name: "project config overrides user config",
			setupFiles: map[string]string{
				"user/dto.yaml": `
gitprovidersync:
  test:
    source:
      provider_type: github
      owner: user-config-owner
      owner_type: user
`,
				"project.yaml": `
gitprovidersync:
  test:
    source:
      provider_type: github
      owner: project-config-owner
      owner_type: user
`,
			},
			setupEnv: map[string]string{
				"XDG_CONFIG_HOME": "user",
			},
			configFile:    "project.yaml",
			expectedOwner: "project-config-owner",
			description:   "Project config should override user config",
		},
		{
			name: "env var overrides project and user config",
			setupFiles: map[string]string{
				"user/dto.yaml": `
gitprovidersync:
  test:
    source:
      provider_type: github
      owner: user-config-owner
      owner_type: user
`,
				"project.yaml": `
gitprovidersync:
  test:
    source:
      provider_type: github
      owner: project-config-owner
      owner_type: user
`,
			},
			setupEnv: map[string]string{
				"XDG_CONFIG_HOME":                       "user",
				"GPS_GITPROVIDERSYNC_TEST_SOURCE_OWNER": "env-var-owner",
			},
			configFile:    "project.yaml",
			expectedOwner: "env-var-owner",
			description:   "Environment variable should override all config files",
		},
		{
			name: "dotenv overrides project config",
			setupFiles: map[string]string{
				"project.yaml": `
gitprovidersync:
  test:
    source:
      provider_type: github
      owner: project-config-owner
      owner_type: user
`,
				".env": `GITPROVIDERSYNC_TEST_SOURCE_OWNER=dotenv-owner`,
			},
			setupEnv: map[string]string{
				"GPS_TESTCONFIG_HOME": ".", // Tell config loader to look in current dir for .env
			},
			configFile:    "project.yaml",
			expectedOwner: "dotenv-owner",
			description:   ".env file should override project config",
		},
		{
			name: "config-file-only ignores env and user config",
			setupFiles: map[string]string{
				"user/dto.yaml": `
gitprovidersync:
  test:
    source:
      provider_type: github
      owner: user-config-owner
      owner_type: user
`,
				"specific.yaml": `
gitprovidersync:
  test:
    source:
      provider_type: github
      owner: specific-config-owner
      owner_type: user
`,
			},
			setupEnv: map[string]string{
				"XDG_CONFIG_HOME":                       "user",
				"GPS_GITPROVIDERSYNC_TEST_SOURCE_OWNER": "env-var-owner",
			},
			configFile:     "specific.yaml",
			configFileOnly: true,
			expectedOwner:  "specific-config-owner",
			description:    "With --config-file-only, should only load specified file",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Cannot use t.Parallel() with t.Setenv()
			tempDir := setupTestEnvironment(t, testCase.setupFiles, testCase.setupEnv)

			// Load configuration using absolute paths
			appConfig := &dto.AppConfiguration{}
			configPath := testCase.configFile

			if configPath == "" {
				configPath = "gitprovidersync.yaml"
			}

			// Make config path absolute
			configPath = filepath.Join(tempDir, configPath)

			err := ReadConfigurationFile(context.Background(), configPath, testCase.configFileOnly, appConfig)

			// Check result
			verifyTestResult(t, err, appConfig, testCase.expectedOwner, testCase.description)
		})
	}
}

// TestXDGConfigFallback verifies that XDG config falls back correctly when XDG_CONFIG_HOME is not set.
func TestXDGConfigFallback(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()
	tests := []struct {
		name         string
		xdgHome      string
		homeDir      string
		expectedPath string
	}{
		{
			name:         "XDG_CONFIG_HOME is set",
			xdgHome:      "/custom/xdg",
			homeDir:      "/home/user",
			expectedPath: "/custom/xdg/gitprovidersync/dto.yaml",
		},
		{
			name:         "XDG_CONFIG_HOME not set, HOME is set",
			xdgHome:      "",
			homeDir:      "/home/user",
			expectedPath: "/home/user/.config/gitprovidersync/dto.yaml",
		},
		{
			name:         "Neither XDG_CONFIG_HOME nor HOME set",
			xdgHome:      "",
			homeDir:      "",
			expectedPath: "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Set test environment (t.Setenv handles cleanup)
			if testCase.xdgHome != "" {
				t.Setenv("XDG_CONFIG_HOME", testCase.xdgHome)
			} else {
				t.Setenv("XDG_CONFIG_HOME", "")
			}

			if testCase.homeDir != "" {
				t.Setenv("HOME", testCase.homeDir)
			} else {
				t.Setenv("HOME", "")
			}

			// Test
			path := getUserConfigPath()
			assert.Equal(t, testCase.expectedPath, path)
		})
	}
}

// TestGetUserConfigPath verifies the XDG config path resolution.
func TestGetUserConfigPath(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()
	t.Run("with XDG_CONFIG_HOME", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "/custom/config")

		path := getUserConfigPath()
		assert.Equal(t, "/custom/config/gitprovidersync/dto.yaml", path)
	})

	t.Run("without XDG_CONFIG_HOME but with HOME", func(t *testing.T) {
		// Cannot use t.Parallel() with t.Setenv()
		// Make sure XDG_CONFIG_HOME is not set
		t.Setenv("XDG_CONFIG_HOME", "")
		// HOME is typically always set in test environments
		// Using os.Getenv for read-only access is safe
		if home := os.Getenv("HOME"); home != "" {
			expectedPath := filepath.Join(home, ".config", "gitprovidersync", "dto.yaml")
			path := getUserConfigPath()
			assert.Equal(t, expectedPath, path)
		}
	})
}
