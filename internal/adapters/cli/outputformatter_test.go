// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/sync"
	model "itiquette/git-provider-sync/internal/model/configuration"
)

func TestNewOutputFormatter(t *testing.T) {
	t.Parallel()

	formatter := NewOutputFormatter()
	assert.NotNil(t, formatter)
	assert.IsType(t, &OutputFormatter{}, formatter)
}

func TestSupportedFormats(t *testing.T) {
	t.Parallel()

	formatter := &OutputFormatter{}
	formats := formatter.SupportedFormats()

	expected := []string{"console", "json", "plain"}
	assert.Equal(t, expected, formats)
}

func TestFormatConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		format       string
		config       model.AppConfiguration
		expectError  bool
		expectOutput []string
	}{
		{
			name:   "console format",
			format: "console",
			config: model.AppConfiguration{
				GitProviderSyncConfs: map[string]model.Environment{
					"test": {
						"source": model.SyncConfig{
							BaseConfig: model.BaseConfig{
								ProviderType: "github",
								Domain:       "github.com",
								Owner:        "testowner",
								OwnerType:    "user",
							},
						},
					},
				},
			},
			expectError:  false,
			expectOutput: []string{"Git Provider Sync Configuration", "Environment: test", "github"},
		},
		{
			name:   "json format",
			format: "json",
			config: model.AppConfiguration{
				GitProviderSyncConfs: map[string]model.Environment{
					"test": {
						"source": model.SyncConfig{
							BaseConfig: model.BaseConfig{
								ProviderType: "github",
								Owner:        "testowner",
							},
						},
					},
				},
			},
			expectError:  false,
			expectOutput: []string{"{", "github", "testowner"},
		},
		{
			name:   "plain format",
			format: "plain",
			config: model.AppConfiguration{
				GitProviderSyncConfs: map[string]model.Environment{
					"test": {
						"source": model.SyncConfig{
							BaseConfig: model.BaseConfig{
								ProviderType: "github",
								Domain:       "github.com",
								Owner:        "testowner",
								OwnerType:    "user",
							},
						},
					},
				},
			},
			expectError:  false,
			expectOutput: []string{"ENVIRONMENT", "test", "source", "github"},
		},
		{
			name:        "invalid format",
			format:      "invalid",
			config:      model.AppConfiguration{},
			expectError: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			formatter := &OutputFormatter{}

			var buf bytes.Buffer

			err := formatter.FormatConfiguration(testCase.config, testCase.format, &buf)

			if testCase.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "unsupported output format")

				return
			}

			require.NoError(t, err)

			output := buf.String()
			for _, expected := range testCase.expectOutput {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestFormatConfigurationWithAuth(t *testing.T) {
	t.Parallel()

	config := model.AppConfiguration{
		GitProviderSyncConfs: map[string]model.Environment{
			"prod": {
				"github-source": model.SyncConfig{
					BaseConfig: model.BaseConfig{
						ProviderType: "github",
						Domain:       "github.com",
						Owner:        "myorg",
						OwnerType:    "organization",
						Auth: model.AuthConfig{
							Protocol: "https",
							Token:    "secret-token",
							ProxyURL: "http://proxy:8080",
						},
					},
				},
			},
		},
	}

	formatter := &OutputFormatter{}

	var buf bytes.Buffer

	err := formatter.FormatConfiguration(config, "console", &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Environment: prod")
	assert.Contains(t, output, "Authentication:")
	assert.Contains(t, output, "Protocol: https")
	assert.Contains(t, output, "Token: <*****>") // Should be masked
	assert.Contains(t, output, "Proxy URL: http://proxy:8080")
}

func TestFormatConfigurationWithMirrors(t *testing.T) {
	t.Parallel()

	config := model.AppConfiguration{
		GitProviderSyncConfs: map[string]model.Environment{
			"test": {
				"source": model.SyncConfig{
					BaseConfig: model.BaseConfig{
						ProviderType: "github",
						Owner:        "testowner",
						OwnerType:    "user",
					},
					Mirrors: map[string]model.MirrorConfig{
						"backup": {
							BaseConfig: model.BaseConfig{
								ProviderType: "gitlab",
								Domain:       "gitlab.com",
								Owner:        "backup-org",
								OwnerType:    "organization",
							},
							Settings: model.MirrorSettings{
								ForcePush:  true,
								Visibility: "private",
								Disabled:   false,
							},
						},
					},
				},
			},
		},
	}

	formatter := &OutputFormatter{}

	var buf bytes.Buffer

	err := formatter.FormatConfiguration(config, "console", &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Mirror Configurations:")
	assert.Contains(t, output, "Mirror: backup")
	assert.Contains(t, output, "Type: gitlab")
	assert.Contains(t, output, "Settings:")
	assert.Contains(t, output, "Force Push: true")
	assert.Contains(t, output, "Visibility: private")
}

func TestFormatConfigurationWithRepositories(t *testing.T) {
	t.Parallel()

	config := model.AppConfiguration{
		GitProviderSyncConfs: map[string]model.Environment{
			"test": {
				"source": model.SyncConfig{
					BaseConfig: model.BaseConfig{
						ProviderType: "github",
						Owner:        "testowner",
						OwnerType:    "user",
					},
					Repositories: model.RepositoriesOption{
						Include: []string{"repo-*", "important-*"},
						Exclude: []string{"*-archived", "temp-*"},
					},
				},
			},
		},
	}

	formatter := &OutputFormatter{}

	var buf bytes.Buffer

	err := formatter.FormatConfiguration(config, "console", &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Repositories:")
	assert.Contains(t, output, "Include:")
	assert.Contains(t, output, "repo-*")
	assert.Contains(t, output, "Exclude:")
	assert.Contains(t, output, "*-archived")
}

func TestFormatSyncResults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		format       string
		results      *sync.Results
		expectError  bool
		expectOutput []string
	}{
		{
			name:   "console format",
			format: "console",
			results: &sync.Results{
				TotalSources:      2,
				TotalMirrors:      1,
				TotalRepositories: 10,
				SuccessfulSyncs:   8,
				FailedSyncs:       1,
				SkippedSyncs:      1,
				DurationSeconds:   45.67,
				DryRun:            true,
				Results: []sync.Result{
					{
						Environment:     "prod",
						Source:          "github-source",
						Mirror:          "gitlab-mirror",
						Repository:      "test-repo",
						Action:          "CREATED",
						Status:          "SUCCESS",
						DurationSeconds: 2.3,
					},
				},
			},
			expectError:  false,
			expectOutput: []string{"Repository State Changes", "Total Sources: 2", "DRY RUN", "To mirrors:"},
		},
		{
			name:   "json format",
			format: "json",
			results: &sync.Results{
				TotalSources:    1,
				SuccessfulSyncs: 1,
				DurationSeconds: 30.0,
			},
			expectError:  false,
			expectOutput: []string{"{", "total_sources", "successful_syncs"},
		},
		{
			name:   "plain format",
			format: "plain",
			results: &sync.Results{
				Results: []sync.Result{
					{
						Environment:     "test",
						Source:          "source1",
						Repository:      "repo1",
						Mirror:          "mirror1",
						Status:          "SUCCESS",
						Action:          "UPDATED",
						DurationSeconds: 1.5,
					},
				},
			},
			expectError:  false,
			expectOutput: []string{"STATUS", "REPOSITORY", "SOURCE", "MIRROR", "repo1", "source1", "mirror1"},
		},
		{
			name:        "invalid format",
			format:      "invalid",
			results:     &sync.Results{},
			expectError: true,
		},
		{
			name:        "invalid results type",
			format:      "console",
			results:     nil,
			expectError: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			formatter := &OutputFormatter{}

			var dataBuf, progressBuf bytes.Buffer

			var err error
			if testCase.results == nil {
				err = formatter.FormatSyncResults("invalid", testCase.format, &dataBuf, &progressBuf)
			} else {
				err = formatter.FormatSyncResults(testCase.results, testCase.format, &dataBuf, &progressBuf)
			}

			if testCase.expectError {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			// Check both data and progress output
			dataOutput := dataBuf.String()
			progressOutput := progressBuf.String()
			combinedOutput := dataOutput + progressOutput

			for _, expected := range testCase.expectOutput {
				assert.Contains(t, combinedOutput, expected)
			}
		})
	}
}

func TestOutputFormatter_EmptyConfigCheckers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "isEmptyAuthConfig",
			testFunc: func(t *testing.T) {
				t.Helper()
				formatter := &OutputFormatter{}

				// Empty auth config
				emptyAuth := model.AuthConfig{}
				assert.True(t, formatter.isEmptyAuthConfig(emptyAuth))

				// Non-empty auth config
				nonEmptyAuth := model.AuthConfig{
					Protocol: "https",
					Token:    "token",
				}
				assert.False(t, formatter.isEmptyAuthConfig(nonEmptyAuth))
			},
		},
		{
			name: "isEmptyRepositoriesOption",
			testFunc: func(t *testing.T) {
				t.Helper()
				formatter := &OutputFormatter{}

				// Empty repositories option
				emptyRepos := model.RepositoriesOption{}
				assert.True(t, formatter.isEmptyRepositoriesOption(emptyRepos))

				// Non-empty repositories option
				nonEmptyRepos := model.RepositoriesOption{
					Include: []string{"repo-*"},
				}
				assert.False(t, formatter.isEmptyRepositoriesOption(nonEmptyRepos))
			},
		},
		{
			name: "isEmptyMirrorSettings",
			testFunc: func(t *testing.T) {
				t.Helper()
				formatter := &OutputFormatter{}

				// Empty mirror settings
				emptySettings := model.MirrorSettings{}
				assert.True(t, formatter.isEmptyMirrorSettings(emptySettings))

				// Non-empty mirror settings
				nonEmptySettings := model.MirrorSettings{
					ForcePush: true,
				}
				assert.False(t, formatter.isEmptyMirrorSettings(nonEmptySettings))
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testCase.testFunc(t)
		})
	}
}

func TestFormatConfiguration_EmptyAndMinimalConfig_HandlesCorrectly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config model.AppConfiguration
	}{
		{
			name: "empty configuration",
			config: model.AppConfiguration{
				GitProviderSyncConfs: map[string]model.Environment{},
			},
		},
		{
			name: "minimal configuration",
			config: model.AppConfiguration{
				GitProviderSyncConfs: map[string]model.Environment{
					"minimal": {
						"source": model.SyncConfig{
							BaseConfig: model.BaseConfig{
								ProviderType: "github",
							},
						},
					},
				},
			},
		},
		{
			name: "configuration with ssh auth",
			config: model.AppConfiguration{
				GitProviderSyncConfs: map[string]model.Environment{
					"ssh": {
						"source": model.SyncConfig{
							BaseConfig: model.BaseConfig{
								ProviderType: "github",
								Owner:        "test",
								OwnerType:    "user",
								Auth: model.AuthConfig{
									Protocol:          "ssh",
									SSHCommand:        "ssh -i ~/.ssh/id_rsa",
									SSHURLRewriteFrom: "https://github.com/",
									SSHURLRewriteTo:   "git@github.com:",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			formatter := &OutputFormatter{}

			// Test all formats
			formats := []string{"console", "json", "plain"}
			for _, format := range formats {
				var buf bytes.Buffer

				err := formatter.FormatConfiguration(testCase.config, format, &buf)
				require.NoError(t, err, "Format %s should not error", format)

				// Should produce some output
				if len(testCase.config.GitProviderSyncConfs) > 0 {
					assert.NotEmpty(t, buf.String(), "Format %s should produce output", format)
				}
			}
		})
	}
}

func TestFormatSyncResultsErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		results     interface{}
		expectError string
	}{
		{
			name:        "wrong type - string",
			results:     "not a sync result",
			expectError: "invalid sync results type",
		},
		{
			name:        "wrong type - number",
			results:     42,
			expectError: "invalid sync results type",
		},
		{
			name:        "wrong type - map",
			results:     map[string]string{"key": "value"},
			expectError: "invalid sync results type",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			formatter := &OutputFormatter{}

			var dataBuf, progressBuf bytes.Buffer

			err := formatter.FormatSyncResults(testCase.results, "console", &dataBuf, &progressBuf)
			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.expectError)
		})
	}
}

func TestFormatConfigurationComplexScenarios(t *testing.T) {
	t.Parallel()

	// Complex configuration with multiple environments, mirrors, auth, etc.
	complexConfig := model.AppConfiguration{
		GitProviderSyncConfs: map[string]model.Environment{
			"production": {
				"github-main": model.SyncConfig{
					BaseConfig: model.BaseConfig{
						ProviderType: "github",
						Domain:       "github.com",
						Owner:        "mycompany",
						OwnerType:    "organization",
						UseGitBinary: true,
						Auth: model.AuthConfig{
							Protocol: "https",
							Token:    "github-token",
							ProxyURL: "http://corporate-proxy:8080",
						},
					},
					IncludeForks:    true,
					ActiveFromLimit: "2023-01-01",
					Repositories: model.RepositoriesOption{
						Include: []string{"production-*", "core-*"},
						Exclude: []string{"*-deprecated", "temp-*"},
					},
					Mirrors: map[string]model.MirrorConfig{
						"gitlab-backup": {
							BaseConfig: model.BaseConfig{
								ProviderType: "gitlab",
								Domain:       "gitlab.internal.com",
								Owner:        "backup-group",
								OwnerType:    "group",
								Auth: model.AuthConfig{
									Protocol: "https",
									Token:    "gitlab-token",
								},
							},
							Settings: model.MirrorSettings{
								ForcePush:         true,
								Visibility:        "internal",
								DescriptionPrefix: "[BACKUP] ",
								AlphaNumHyphName:  true,
							},
						},
					},
				},
			},
			"development": {
				"github-dev": model.SyncConfig{
					BaseConfig: model.BaseConfig{
						ProviderType: "github",
						Domain:       "github.com",
						Owner:        "dev-team",
						OwnerType:    "organization",
					},
				},
			},
		},
	}

	formatter := &OutputFormatter{}

	t.Run("console format complex", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		err := formatter.FormatConfiguration(complexConfig, "console", &buf)
		require.NoError(t, err)

		output := buf.String()

		// Check all major sections are present
		expectedSections := []string{
			"Git Provider Sync Configuration",
			"Environment: production",
			"Environment: development",
			"Sync Configuration: github-main",
			"Provider Type: github",
			"Authentication:",
			"Include Forks: true",
			"Mirror Configurations:",
			"Mirror: gitlab-backup",
			"Settings:",
			"Force Push: true",
		}

		for _, section := range expectedSections {
			assert.Contains(t, output, section)
		}
	})

	t.Run("json format complex", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		err := formatter.FormatConfiguration(complexConfig, "json", &buf)
		require.NoError(t, err)

		output := buf.String()

		// Should be valid JSON with expected keys
		assert.Contains(t, output, "gitprovidersync")
		assert.Contains(t, output, "production")
		assert.Contains(t, output, "development")
		assert.Contains(t, output, "github-main")

		// Check that it's properly formatted JSON
		assert.True(t, strings.HasPrefix(output, "{"))
		assert.True(t, strings.HasSuffix(strings.TrimSpace(output), "}"))
	})

	t.Run("plain format complex", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		err := formatter.FormatConfiguration(complexConfig, "plain", &buf)
		require.NoError(t, err)

		output := buf.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")

		// Should have header + 2 data lines
		assert.Len(t, lines, 3)
		assert.Contains(t, lines[0], "ENVIRONMENT")

		// Check that both environments are present (order may vary)
		dataLines := strings.Join(lines[1:], "\n")
		assert.Contains(t, dataLines, "production")
		assert.Contains(t, dataLines, "development")
	})
}

func TestFormatConfigurationErrorHandling(t *testing.T) {
	t.Parallel()

	// Create a formatter
	formatter := NewOutputFormatter()

	// Create a failing writer
	failingWriter := &failingWriter{}

	cfg := model.AppConfiguration{
		GitProviderSyncConfs: map[string]model.Environment{
			"test": {
				"source": model.SyncConfig{
					BaseConfig: model.BaseConfig{
						ProviderType: "github",
						Domain:       "github.com",
						Owner:        "testuser",
					},
				},
			},
		},
	}

	// Test console format with writer failure
	err := formatter.FormatConfiguration(cfg, "console", failingWriter)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write")
}

// failingWriter always returns an error when Write is called.
type failingWriter struct{}

func (fw *failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestFormatConfigurationEmptyEnvironments(t *testing.T) {
	t.Parallel()

	formatter := NewOutputFormatter()

	// Test with empty environments map
	cfg := model.AppConfiguration{
		GitProviderSyncConfs: map[string]model.Environment{},
	}

	var buf strings.Builder

	err := formatter.FormatConfiguration(cfg, "console", &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Git Provider Sync Configuration")
	assert.NotContains(t, output, "Environment:")
}

func TestFormatConfigurationWithComplexAuth(t *testing.T) {
	t.Parallel()

	formatter := NewOutputFormatter()

	// Test configuration with valid fields to exercise more branches
	cfg := model.AppConfiguration{
		GitProviderSyncConfs: map[string]model.Environment{
			"complex-auth": {
				"source-with-auth": model.SyncConfig{
					BaseConfig: model.BaseConfig{
						ProviderType: "gitlab",
						Domain:       "gitlab.example.com",
						Owner:        "myorg",
						Auth: model.AuthConfig{
							Token:      "glpat_secret_token",
							SSHCommand: "ssh -i ~/.ssh/id_rsa",
							ProxyURL:   "http://proxy.example.com:8080",
							Protocol:   "https",
						},
					},
					Repositories: model.RepositoriesOption{
						Include: []string{"important-*", "critical-*"},
						Exclude: []string{"temp-*", "draft-*"},
					},
					Mirrors: map[string]model.MirrorConfig{
						"backup-mirror": {
							BaseConfig: model.BaseConfig{
								ProviderType: "github",
								Domain:       "github.com",
								Owner:        "backuporg",
								Auth: model.AuthConfig{
									Token: "ghp_secret_token",
								},
							},
							Settings: model.MirrorSettings{
								Visibility:        "private",
								DescriptionPrefix: "Backup of ",
								ForcePush:         true,
								Disabled:          false,
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name   string
		format string
	}{
		{"console format with complex auth", "console"},
		{"json format with complex auth", "json"},
		{"plain format with complex auth", "plain"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var buf strings.Builder

			err := formatter.FormatConfiguration(cfg, testCase.format, &buf)
			require.NoError(t, err)

			output := buf.String()
			assert.NotEmpty(t, output)

			// Check that sensitive information is masked in text formats
			if testCase.format == formatConsole || testCase.format == formatPlain {
				assert.NotContains(t, output, "ghp_secret_token")
				assert.NotContains(t, output, "glpat_secret_token")
				// Console format uses <*****> for masking
				if testCase.format == formatConsole {
					assert.Contains(t, output, "<*****>")
				}
			}
		})
	}
}

// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

func TestOutputFormatter_writeMirrorConfigOptionalFields(t *testing.T) {
	t.Parallel()

	formatter := &OutputFormatter{}

	tests := []struct {
		name         string
		mirrorCfg    model.MirrorConfig
		expectedText []string
	}{
		{
			name: "all optional fields set",
			mirrorCfg: model.MirrorConfig{
				BaseConfig: model.BaseConfig{
					UseGitBinary: true,
				},
				Path: "/custom/path/to/mirror",
			},
			expectedText: []string{
				"Use Git Binary: true",
				"Path: /custom/path/to/mirror",
			},
		},
		{
			name: "only UseGitBinary set",
			mirrorCfg: model.MirrorConfig{
				BaseConfig: model.BaseConfig{
					UseGitBinary: true,
				},
				Path: "",
			},
			expectedText: []string{
				"Use Git Binary: true",
			},
		},
		{
			name: "only Path set",
			mirrorCfg: model.MirrorConfig{
				BaseConfig: model.BaseConfig{
					UseGitBinary: false,
				},
				Path: "/some/path",
			},
			expectedText: []string{
				"Path: /some/path",
			},
		},
		{
			name: "no optional fields set",
			mirrorCfg: model.MirrorConfig{
				BaseConfig: model.BaseConfig{
					UseGitBinary: false,
				},
				Path: "",
			},
			expectedText: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var buffer bytes.Buffer

			err := formatter.writeMirrorConfigOptionalFields(&buffer, "  ", test.mirrorCfg)

			require.NoError(t, err)

			output := buffer.String()

			// Check that all expected text is present
			for _, expectedText := range test.expectedText {
				assert.Contains(t, output, expectedText, "Output should contain: %s", expectedText)
			}

			// Check that unexpected text is not present when fields are not set
			if !test.mirrorCfg.UseGitBinary {
				assert.NotContains(t, output, "Use Git Binary: true")
			}

			if test.mirrorCfg.Path == "" {
				assert.NotContains(t, output, "Path:")
			}
		})
	}
}

func TestOutputFormatter_writeAuthOptionalFields(t *testing.T) {
	t.Parallel()

	formatter := &OutputFormatter{}

	tests := []struct {
		name         string
		authCfg      model.AuthConfig
		expectedText []string
		notExpected  []string
	}{
		{
			name: "all optional fields set",
			authCfg: model.AuthConfig{
				HTTPScheme:  "https",
				Token:       "secret-token",
				ProxyURL:    "http://proxy.example.com",
				CertDirPath: "/path/to/certs",
			},
			expectedText: []string{
				"HTTP Scheme: https",
				"Token: <*****>", // Token should be masked
				"Proxy URL: http://proxy.example.com",
				"Certificate Directory: /path/to/certs",
			},
		},
		{
			name: "only HTTPScheme set",
			authCfg: model.AuthConfig{
				HTTPScheme: "https",
			},
			expectedText: []string{
				"HTTP Scheme: https",
			},
			notExpected: []string{
				"Token:",
				"Proxy URL:",
				"Certificate Directory:",
			},
		},
		{
			name: "only Token set",
			authCfg: model.AuthConfig{
				Token: "my-token",
			},
			expectedText: []string{
				"Token: <*****>", // Should be masked
			},
			notExpected: []string{
				"HTTP Scheme:",
				"Proxy URL:",
				"Certificate Directory:",
			},
		},
		{
			name: "only ProxyURL set",
			authCfg: model.AuthConfig{
				ProxyURL: "http://company-proxy.com:8080",
			},
			expectedText: []string{
				"Proxy URL: http://company-proxy.com:8080",
			},
			notExpected: []string{
				"HTTP Scheme:",
				"Token:",
				"Certificate Directory:",
			},
		},
		{
			name: "only CertDirPath set",
			authCfg: model.AuthConfig{
				CertDirPath: "/etc/ssl/custom",
			},
			expectedText: []string{
				"Certificate Directory: /etc/ssl/custom",
			},
			notExpected: []string{
				"HTTP Scheme:",
				"Token:",
				"Proxy URL:",
			},
		},
		{
			name:         "no optional fields set",
			authCfg:      model.AuthConfig{},
			expectedText: []string{},
			notExpected: []string{
				"HTTP Scheme:",
				"Token:",
				"Proxy URL:",
				"Certificate Directory:",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var buffer bytes.Buffer

			err := formatter.writeAuthOptionalFields(&buffer, "  ", test.authCfg)

			require.NoError(t, err)

			output := buffer.String()

			// Check that all expected text is present
			for _, expectedText := range test.expectedText {
				assert.Contains(t, output, expectedText, "Output should contain: %s", expectedText)
			}

			// Check that unexpected text is not present
			for _, notExpectedText := range test.notExpected {
				assert.NotContains(t, output, notExpectedText, "Output should not contain: %s", notExpectedText)
			}
		})
	}
}

func TestOutputFormatter_writeSyncConfigMandatoryFields(t *testing.T) {
	t.Parallel()

	formatter := &OutputFormatter{}

	tests := []struct {
		name         string
		syncCfg      model.SyncConfig
		expectedText []string
	}{
		{
			name: "complete sync config",
			syncCfg: model.SyncConfig{
				BaseConfig: model.BaseConfig{
					ProviderType: "github",
					Domain:       "github.com",
					Owner:        "testuser",
					OwnerType:    "user",
				},
			},
			expectedText: []string{
				"Provider Type: github",
				"Domain: github.com",
				"Owner: testuser",
				"Owner Type: user",
			},
		},
		{
			name: "sync config with empty domain",
			syncCfg: model.SyncConfig{
				BaseConfig: model.BaseConfig{
					ProviderType: "gitlab",
					Domain:       "",
					Owner:        "testorg",
					OwnerType:    "organization",
				},
			},
			expectedText: []string{
				"Provider Type: gitlab",
				"Domain: ", // GetDomain() will handle empty domains
				"Owner: testorg",
				"Owner Type: organization",
			},
		},
		{
			name: "gitea sync config",
			syncCfg: model.SyncConfig{
				BaseConfig: model.BaseConfig{
					ProviderType: "gitea",
					Domain:       "gitea.example.com",
					Owner:        "dev-team",
					OwnerType:    "organization",
				},
			},
			expectedText: []string{
				"Provider Type: gitea",
				"Domain: gitea.example.com",
				"Owner: dev-team",
				"Owner Type: organization",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var buffer bytes.Buffer

			err := formatter.writeMandatoryFields(&buffer, "  ", test.syncCfg)

			require.NoError(t, err)

			output := buffer.String()

			// Check that all expected text is present
			for _, expectedText := range test.expectedText {
				assert.Contains(t, output, expectedText, "Output should contain: %s", expectedText)
			}
		})
	}
}

func TestOutputFormatter_writeSyncSummary(t *testing.T) {
	t.Parallel()

	formatter := &OutputFormatter{}

	tests := []struct {
		name         string
		syncResults  *sync.Results
		expectedText []string
	}{
		{
			name: "complete sync results",
			syncResults: &sync.Results{
				TotalSources:      3,
				TotalMirrors:      5,
				TotalRepositories: 15,
				SuccessfulSyncs:   12,
				FailedSyncs:       2,
				SkippedSyncs:      1,
				DurationSeconds:   45.67,
				DryRun:            false,
			},
			expectedText: []string{
				"Repository State Changes",
				"========================",
				"Total Sources: 3",
				"Total Mirrors: 5",
				"Total Repositories: 15",
				"Successful Syncs: 12",
				"Failed Syncs: 2",
				"Skipped Syncs: 1",
				"Duration: 45.67 seconds",
			},
		},
		{
			name: "dry run sync results",
			syncResults: &sync.Results{
				TotalSources:      1,
				TotalMirrors:      2,
				TotalRepositories: 5,
				SuccessfulSyncs:   0,
				FailedSyncs:       0,
				SkippedSyncs:      5,
				DurationSeconds:   2.34,
				DryRun:            true,
			},
			expectedText: []string{
				"Repository State Changes",
				"========================",
				"Total Sources: 1",
				"Total Mirrors: 2",
				"Total Repositories: 5",
				"Successful Syncs: 0",
				"Failed Syncs: 0",
				"Skipped Syncs: 5",
				"Duration: 2.34 seconds",
				"Mode: DRY RUN",
			},
		},
		{
			name: "zero duration sync",
			syncResults: &sync.Results{
				TotalSources:      0,
				TotalMirrors:      0,
				TotalRepositories: 0,
				SuccessfulSyncs:   0,
				FailedSyncs:       0,
				SkippedSyncs:      0,
				DurationSeconds:   0.00,
				DryRun:            false,
			},
			expectedText: []string{
				"Repository State Changes",
				"========================",
				"Total Sources: 0",
				"Total Mirrors: 0",
				"Total Repositories: 0",
				"Successful Syncs: 0",
				"Failed Syncs: 0",
				"Skipped Syncs: 0",
				"Duration: 0.00 seconds",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var buffer bytes.Buffer

			err := formatter.writeSyncSummary(test.syncResults, &buffer)

			require.NoError(t, err)

			output := buffer.String()

			// Check that all expected text is present
			for _, expectedText := range test.expectedText {
				assert.Contains(t, output, expectedText, "Output should contain: %s", expectedText)
			}

			// Check dry run mode is only shown when applicable
			if !test.syncResults.DryRun {
				assert.NotContains(t, output, "Mode: DRY RUN")
			}
		})
	}
}

func TestOutputFormatter_printSyncConfig(t *testing.T) {
	t.Parallel()

	formatter := &OutputFormatter{}

	tests := []struct {
		name         string
		configName   string
		syncCfg      model.SyncConfig
		level        int
		indentSize   int
		expectedText []string
	}{
		{
			name:       "basic sync config",
			configName: "source-github",
			syncCfg: model.SyncConfig{
				BaseConfig: model.BaseConfig{
					ProviderType: "github",
					Domain:       "github.com",
					Owner:        "testuser",
					OwnerType:    "user",
				},
			},
			level:      1,
			indentSize: 2,
			expectedText: []string{
				"Sync Configuration: source-github",
				"Provider Type: github",
				"Domain: github.com",
				"Owner: testuser",
				"Owner Type: user",
			},
		},
		{
			name:       "sync config with optional fields",
			configName: "source-gitlab",
			syncCfg: model.SyncConfig{
				BaseConfig: model.BaseConfig{
					ProviderType: "gitlab",
					Domain:       "gitlab.com",
					Owner:        "myorg",
					OwnerType:    "organization",
					UseGitBinary: true,
				},
				IncludeForks:    true,
				ActiveFromLimit: "2023-01-01",
			},
			level:      2,
			indentSize: 4,
			expectedText: []string{
				"Sync Configuration: source-gitlab",
				"Provider Type: gitlab",
				"Domain: gitlab.com",
				"Owner: myorg",
				"Owner Type: organization",
				"Include Forks: true",
				"Use Git Binary: true",
				"Active From Limit: 2023-01-01",
			},
		},
		{
			name:       "sync config with auth",
			configName: "source-gitea",
			syncCfg: model.SyncConfig{
				BaseConfig: model.BaseConfig{
					ProviderType: "gitea",
					Domain:       "gitea.example.com",
					Owner:        "devteam",
					OwnerType:    "organization",
					Auth: model.AuthConfig{
						Protocol: "https",
						Token:    "secret",
					},
				},
			},
			level:      0,
			indentSize: 2,
			expectedText: []string{
				"Sync Configuration: source-gitea",
				"Provider Type: gitea",
				"Domain: gitea.example.com",
				"Owner: devteam",
				"Owner Type: organization",
				"Authentication:",
				"Protocol: https",
				"Token: <*****>",
			},
		},
		{
			name:       "sync config with repositories filter",
			configName: "filtered-source",
			syncCfg: model.SyncConfig{
				BaseConfig: model.BaseConfig{
					ProviderType: "github",
					Domain:       "github.com",
					Owner:        "testuser",
					OwnerType:    "user",
				},
				Repositories: model.RepositoriesOption{
					Include: []string{"important-*", "main-repo"},
					Exclude: []string{"test-*", "temp-*"},
				},
			},
			level:      1,
			indentSize: 2,
			expectedText: []string{
				"Sync Configuration: filtered-source",
				"Provider Type: github",
				"Domain: github.com",
				"Owner: testuser",
				"Owner Type: user",
				"Repositories:",
				"Include:",
				"important-*",
				"main-repo",
				"Exclude:",
				"test-*",
				"temp-*",
			},
		},
		{
			name:       "sync config with mirrors",
			configName: "mirrored-source",
			syncCfg: model.SyncConfig{
				BaseConfig: model.BaseConfig{
					ProviderType: "github",
					Domain:       "github.com",
					Owner:        "testuser",
					OwnerType:    "user",
				},
				Mirrors: map[string]model.MirrorConfig{
					"backup": {
						BaseConfig: model.BaseConfig{
							ProviderType: "gitea",
							Domain:       "gitea.backup.com",
							Owner:        "backups",
							OwnerType:    "organization",
						},
					},
				},
			},
			level:      1,
			indentSize: 2,
			expectedText: []string{
				"Sync Configuration: mirrored-source",
				"Provider Type: github",
				"Domain: github.com",
				"Owner: testuser",
				"Owner Type: user",
				"Mirror Configurations:",
				"Mirror: backup",
				"Type: gitea",
				"Domain: gitea.backup.com",
				"Owner: backups",
				"Owner Type: organization",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var buffer bytes.Buffer

			err := formatter.printSyncConfig(test.configName, test.syncCfg, &buffer, test.level, test.indentSize)

			require.NoError(t, err)

			output := buffer.String()

			// Check that all expected text is present
			for _, expectedText := range test.expectedText {
				assert.Contains(t, output, expectedText, "Output should contain: %s", expectedText)
			}
		})
	}
}
func TestOutputFormatter_WriteMirrorSettingsFieldsErrorPath(t *testing.T) {
	t.Parallel()

	formatter := &OutputFormatter{}

	t.Run("writeMirrorSettingsFields error", func(t *testing.T) {
		t.Parallel()

		settings := model.MirrorSettings{
			AlphaNumHyphName: true,
		}

		// Test settings field write error
		err := formatter.writeMirrorSettingsFields(errorWriter{}, "  ", settings)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write ASCII Name")
	})
}

func TestOutputFormatter_MoreWriteErrors(t *testing.T) {
	t.Parallel()

	formatter := &OutputFormatter{}

	t.Run("writeAuthOptionalFields HTTP scheme error", func(t *testing.T) {
		t.Parallel()

		authCfg := model.AuthConfig{
			HTTPScheme: "https",
		}

		err := formatter.writeAuthOptionalFields(errorWriter{}, "  ", authCfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write HTTP scheme")
	})

	t.Run("writeAuthOptionalFields token error", func(t *testing.T) {
		t.Parallel()

		authCfg := model.AuthConfig{
			Token: "secret",
		}

		// Create a multi-stage error writer that succeeds first then fails
		var callCount int

		writer := &multiStageErrorWriter{callCount: &callCount, failAfter: 0} // Fail on second call (token)

		err := formatter.writeAuthOptionalFields(writer, "  ", authCfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write token")
	})

	t.Run("writeAuthOptionalFields proxy URL error", func(t *testing.T) {
		t.Parallel()

		authCfg := model.AuthConfig{
			ProxyURL: "http://proxy.com",
		}

		err := formatter.writeAuthOptionalFields(errorWriter{}, "  ", authCfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write proxy URL")
	})

	t.Run("writeAuthOptionalFields cert dir error", func(t *testing.T) {
		t.Parallel()

		authCfg := model.AuthConfig{
			CertDirPath: "/certs",
		}

		err := formatter.writeAuthOptionalFields(errorWriter{}, "  ", authCfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write certificate directory")
	})
}

// multiStageErrorWriter succeeds for a certain number of calls then starts failing.
type multiStageErrorWriter struct {
	callCount *int
	failAfter int
}

func (w *multiStageErrorWriter) Write(data []byte) (int, error) {
	*w.callCount++
	if *w.callCount > w.failAfter {
		return 0, errors.New("write error")
	}

	return len(data), nil
}
