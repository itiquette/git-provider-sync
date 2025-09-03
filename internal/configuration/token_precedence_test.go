// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	config "itiquette/git-provider-sync/internal/model/configuration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTokenPrecedenceOrder verifies that tokens are loaded in the correct precedence order:
// 1. Provider-specific env vars (GPS_GITHUB_TOKEN, etc)
// 2. Environment variable expansion in config (token: "${MY_TOKEN}")
// 3. Token file specified in config (token_file: "/path/to/file").
func TestTokenPrecedenceOrder(t *testing.T) {
	// Cannot run parallel with t.Setenv
	tests := []struct {
		name          string
		setupEnv      map[string]string
		configContent string
		tokenFile     map[string]string // filename -> content
		expectedToken string
		description   string
	}{
		{
			name: "provider-specific env var takes highest precedence",
			setupEnv: map[string]string{
				"GPS_GITHUB_TOKEN": "from_provider_env",
				"MY_TOKEN":         "from_expansion",
			},
			configContent: `
gitprovidersync:
  test:
    source:
      provider_type: github
      domain: github.com
      owner: test
      owner_type: user
      auth:
        token: "${MY_TOKEN}"
        token_file: "token.txt"
`,
			tokenFile: map[string]string{
				"token.txt": "from_file",
			},
			expectedToken: "from_provider_env",
			description:   "GPS_GITHUB_TOKEN should override both expansion and file",
		},
		{
			name: "env expansion takes precedence over token file",
			setupEnv: map[string]string{
				"MY_TOKEN": "from_expansion",
			},
			configContent: `
gitprovidersync:
  test:
    source:
      provider_type: github
      domain: github.com
      owner: test
      owner_type: user
      auth:
        token: "${MY_TOKEN}"
        token_file: "token.txt"
`,
			tokenFile: map[string]string{
				"token.txt": "from_file",
			},
			expectedToken: "from_expansion",
			description:   "Environment expansion should override token_file",
		},
		{
			name:     "token file used when no env vars",
			setupEnv: map[string]string{},
			configContent: `
gitprovidersync:
  test:
    source:
      provider_type: github
      domain: github.com
      owner: test
      owner_type: user
      auth:
        token_file: "token.txt"
`,
			tokenFile: map[string]string{
				"token.txt": "from_file_only",
			},
			expectedToken: "from_file_only",
			description:   "Token file should be used when no env vars are set",
		},
		{
			name: "provider-specific tokens for different providers",
			setupEnv: map[string]string{
				"GPS_GITHUB_TOKEN": "github_token_123",
				"GPS_GITLAB_TOKEN": "gitlab_token_456",
			},
			configContent: `
gitprovidersync:
  test:
    github-source:
      provider_type: github
      domain: github.com
      owner: test
      owner_type: user
      mirrors:
        gitlab-mirror:
          provider_type: gitlab
          domain: gitlab.com
          owner: test
          owner_type: user
`,
			expectedToken: "github_token_123", // We'll check the source token
			description:   "Each provider should get its own token",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Setup temp directory
			tempDir := t.TempDir()

			// Set environment variables
			for k, v := range test.setupEnv {
				t.Setenv(k, v)
			}

			// Create token files if specified
			for filename, content := range test.tokenFile {
				fullPath := filepath.Join(tempDir, filename)
				err := os.WriteFile(fullPath, []byte(content), 0600)
				require.NoError(t, err)
				// Replace token file path in config with full path
				test.configContent = strings.ReplaceAll(test.configContent,
					`token_file: "`+filename+`"`,
					`token_file: "`+fullPath+`"`)
			}

			// Write config file
			configFile := filepath.Join(tempDir, "config.yaml")
			err := os.WriteFile(configFile, []byte(test.configContent), 0600)
			require.NoError(t, err)

			// Load configuration
			appConfig := &config.AppConfiguration{}
			err = ReadConfigurationFile(context.Background(), configFile, true, appConfig)
			require.NoError(t, err, test.description)

			// Get the first source to check token
			var actualToken string

			for _, env := range appConfig.GitProviderSyncConfs {
				for _, source := range env {
					actualToken = source.Auth.Token

					break
				}

				break
			}

			assert.Equal(t, test.expectedToken, actualToken, test.description)
		})
	}
}

// TestProviderSpecificTokensForMirrors verifies that mirrors get correct provider-specific tokens.
func TestProviderSpecificTokensForMirrors(t *testing.T) {
	// Set different tokens for each provider
	t.Setenv("GPS_GITHUB_TOKEN", "github_token_123")
	t.Setenv("GPS_GITLAB_TOKEN", "gitlab_token_456")
	t.Setenv("GPS_GITEA_TOKEN", "gitea_token_789")

	configContent := `
gitprovidersync:
  test:
    github-source:
      provider_type: github
      domain: github.com
      owner: test
      owner_type: user
      mirrors:
        gitlab-mirror:
          provider_type: gitlab
          domain: gitlab.com
          owner: backup
          owner_type: user
        gitea-mirror:
          provider_type: gitea
          domain: gitea.com
          owner: backup
          owner_type: user
`

	// Write config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err)

	// Load configuration
	appConfig := &config.AppConfiguration{}
	err = ReadConfigurationFile(context.Background(), configFile, true, appConfig)
	require.NoError(t, err)

	// Verify source token
	source := appConfig.GitProviderSyncConfs["test"]["github-source"]
	assert.Equal(t, "github_token_123", source.Auth.Token, "GitHub source should have GitHub token")

	// Verify mirror tokens
	gitlabMirror := source.Mirrors["gitlab-mirror"]
	assert.Equal(t, "gitlab_token_456", gitlabMirror.Auth.Token, "GitLab mirror should have GitLab token")

	giteaMirror := source.Mirrors["gitea-mirror"]
	assert.Equal(t, "gitea_token_789", giteaMirror.Auth.Token, "Gitea mirror should have Gitea token")
}
