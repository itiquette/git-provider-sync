// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package validation

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/ports"
)

func TestValidateAppConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		config         ports.AppConfiguration
		expectValid    bool
		expectedErrors int
	}{
		{
			name: "valid configuration",
			config: ports.AppConfiguration{
				GlobalSettings: ports.GlobalSettings{
					LogLevel:     ports.LogLevelInfo,
					LogFormat:    ports.LogFormatJSON,
					MaxCacheSize: 100,
					CacheTTL:     time.Minute * 10,
				},
				Environments: map[string]ports.EnvironmentConfiguration{
					"prod": {
						Name:    "prod",
						Enabled: true,
						Source: ports.SourceConfiguration{
							ProviderType: "github",
							Domain:       "test-github.example.com",
							Owner:        "testorg",
							Authentication: ports.AuthenticationConfiguration{
								Type:  ports.AuthenticationTypeToken,
								Token: "test-token",
							},
						},
						Mirrors: map[string]ports.MirrorConfiguration{
							"backup": {
								Name:         "backup",
								ProviderType: "gitlab",
								Domain:       "test-gitlab.example.com",
								Owner:        "backuporg",
								Enabled:      true,
								Authentication: ports.AuthenticationConfiguration{
									Type:  ports.AuthenticationTypeToken,
									Token: "backup-token",
								},
							},
						},
						Options: ports.EnvironmentOptions{
							MaxConcurrency: 5,
							Timeout:        time.Minute * 30,
							RetryAttempts:  3,
							RetryDelay:     time.Second * 5,
						},
					},
				},
			},
			expectValid:    true,
			expectedErrors: 0,
		},
		{
			name: "invalid global settings",
			config: ports.AppConfiguration{
				GlobalSettings: ports.GlobalSettings{
					LogLevel:     "invalid",
					LogFormat:    "invalid",
					MaxCacheSize: -100,
					CacheTTL:     -time.Minute,
				},
				Environments: map[string]ports.EnvironmentConfiguration{},
			},
			expectValid:    false,
			expectedErrors: 4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := ValidateAppConfiguration(test.config)

			assert.Equal(t, test.expectValid, result.Valid)
			assert.Len(t, result.Results, test.expectedErrors)
		})
	}
}

func TestValidateGlobalSettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		settings    ports.GlobalSettings
		expectValid bool
		expectedLen int
	}{
		{
			name: "valid settings",
			settings: ports.GlobalSettings{
				LogLevel:     ports.LogLevelInfo,
				LogFormat:    ports.LogFormatJSON,
				MaxCacheSize: 100,
				CacheTTL:     time.Minute * 10,
			},
			expectValid: true,
			expectedLen: 0,
		},
		{
			name: "invalid log level",
			settings: ports.GlobalSettings{
				LogLevel:     "invalid",
				LogFormat:    ports.LogFormatJSON,
				MaxCacheSize: 100,
				CacheTTL:     time.Minute * 10,
			},
			expectValid: false,
			expectedLen: 1,
		},
		{
			name: "invalid log format",
			settings: ports.GlobalSettings{
				LogLevel:     ports.LogLevelInfo,
				LogFormat:    "invalid",
				MaxCacheSize: 100,
				CacheTTL:     time.Minute * 10,
			},
			expectValid: false,
			expectedLen: 1,
		},
		{
			name: "negative cache size",
			settings: ports.GlobalSettings{
				LogLevel:     ports.LogLevelInfo,
				LogFormat:    ports.LogFormatJSON,
				MaxCacheSize: -100,
				CacheTTL:     time.Minute * 10,
			},
			expectValid: false,
			expectedLen: 1,
		},
		{
			name: "negative cache TTL",
			settings: ports.GlobalSettings{
				LogLevel:     ports.LogLevelInfo,
				LogFormat:    ports.LogFormatJSON,
				MaxCacheSize: 100,
				CacheTTL:     -time.Minute,
			},
			expectValid: false,
			expectedLen: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := ValidateGlobalSettings(test.settings)

			assert.Equal(t, test.expectValid, result.Valid)
			assert.Len(t, result.Results, test.expectedLen)

			if !test.expectValid {
				// Verify error details
				for _, res := range result.Results {
					assert.False(t, res.Valid)
					assert.NotEmpty(t, res.Code)
					assert.NotEmpty(t, res.Message)
				}
			}
		})
	}
}

func TestValidateEnvironment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		env         ports.EnvironmentConfiguration
		expectValid bool
	}{
		{
			name: "valid environment",
			env: ports.EnvironmentConfiguration{
				Name:    "production",
				Enabled: true,
				Source: ports.SourceConfiguration{
					ProviderType: "github",
					Domain:       "test-github.example.com",
					Owner:        "testorg",
					Authentication: ports.AuthenticationConfiguration{
						Type:  ports.AuthenticationTypeToken,
						Token: "test-token",
					},
				},
				Mirrors: map[string]ports.MirrorConfiguration{
					"backup": {
						Name:         "backup",
						ProviderType: "gitlab",
						Domain:       "test-gitlab.example.com",
						Owner:        "backuporg",
						Enabled:      true,
						Authentication: ports.AuthenticationConfiguration{
							Type:  ports.AuthenticationTypeToken,
							Token: "backup-token",
						},
					},
				},
				Options: ports.EnvironmentOptions{
					MaxConcurrency: 5,
					Timeout:        time.Minute * 30,
					RetryAttempts:  3,
					RetryDelay:     time.Second * 5,
				},
			},
			expectValid: true,
		},
		{
			name: "empty environment name",
			env: ports.EnvironmentConfiguration{
				Name:    "",
				Enabled: true,
				Source: ports.SourceConfiguration{
					ProviderType: "github",
					Domain:       "test-github.example.com",
					Owner:        "testorg",
					Authentication: ports.AuthenticationConfiguration{
						Type:  ports.AuthenticationTypeToken,
						Token: "test-token",
					},
				},
			},
			expectValid: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := ValidateEnvironment(test.env)

			assert.Equal(t, test.expectValid, result.Valid)

			if !test.expectValid {
				assert.NotEmpty(t, result.Results)

				for _, res := range result.Results {
					if !res.Valid {
						assert.NotEmpty(t, res.Code)
						assert.NotEmpty(t, res.Message)
					}
				}
			}
		})
	}
}

func TestValidateSourceConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		source      ports.SourceConfiguration
		expectValid bool
	}{
		{
			name: "valid GitHub source",
			source: ports.SourceConfiguration{
				ProviderType: "github",
				Domain:       "test-github.example.com",
				Owner:        "testorg",
				Authentication: ports.AuthenticationConfiguration{
					Type:  ports.AuthenticationTypeToken,
					Token: "test-token",
				},
			},
			expectValid: true,
		},
		{
			name: "invalid provider type",
			source: ports.SourceConfiguration{
				ProviderType: "invalid",
				Domain:       "test-github.example.com",
				Owner:        "testorg",
				Authentication: ports.AuthenticationConfiguration{
					Type:  ports.AuthenticationTypeToken,
					Token: "test-token",
				},
			},
			expectValid: false,
		},
		{
			name: "invalid domain",
			source: ports.SourceConfiguration{
				ProviderType: "github",
				Domain:       "invalid..domain",
				Owner:        "testorg",
				Authentication: ports.AuthenticationConfiguration{
					Type:  ports.AuthenticationTypeToken,
					Token: "test-token",
				},
			},
			expectValid: false,
		},
		{
			name: "invalid owner",
			source: ports.SourceConfiguration{
				ProviderType: "github",
				Domain:       "test-github.example.com",
				Owner:        "",
				Authentication: ports.AuthenticationConfiguration{
					Type:  ports.AuthenticationTypeToken,
					Token: "test-token",
				},
			},
			expectValid: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := ValidateSourceConfiguration(test.source)

			assert.Equal(t, test.expectValid, result.Valid)

			if !test.expectValid {
				assert.NotEmpty(t, result.Results)
			}
		})
	}
}

func TestValidateMirrorConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mirror      ports.MirrorConfiguration
		expectValid bool
	}{
		{
			name: "valid remote mirror",
			mirror: ports.MirrorConfiguration{
				Name:         "backup",
				ProviderType: "gitlab",
				Domain:       "test-gitlab.example.com",
				Owner:        "backuporg",
				Enabled:      true,
				Authentication: ports.AuthenticationConfiguration{
					Type:  ports.AuthenticationTypeToken,
					Token: "backup-token",
				},
			},
			expectValid: true,
		},
		{
			name: "valid local directory mirror",
			mirror: ports.MirrorConfiguration{
				Name:         "local-backup",
				ProviderType: "directory",
				Path:         "/path/to/backup",
				Enabled:      true,
			},
			expectValid: true,
		},
		{
			name: "valid local archive mirror",
			mirror: ports.MirrorConfiguration{
				Name:         "archive-backup",
				ProviderType: "archive",
				Path:         "/path/to/backup.tar.gz",
				Enabled:      true,
			},
			expectValid: true,
		},
		{
			name: "empty mirror name",
			mirror: ports.MirrorConfiguration{
				Name:         "",
				ProviderType: "gitlab",
				Domain:       "test-gitlab.example.com",
				Owner:        "backuporg",
				Enabled:      true,
			},
			expectValid: false,
		},
		{
			name: "invalid provider type",
			mirror: ports.MirrorConfiguration{
				Name:         "backup",
				ProviderType: "invalid",
				Domain:       "test-gitlab.example.com",
				Owner:        "backuporg",
				Enabled:      true,
			},
			expectValid: false,
		},
		{
			name: "local provider missing path",
			mirror: ports.MirrorConfiguration{
				Name:         "local-backup",
				ProviderType: "directory",
				Path:         "",
				Enabled:      true,
			},
			expectValid: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := ValidateMirrorConfiguration(test.mirror)

			assert.Equal(t, test.expectValid, result.Valid)

			if !test.expectValid {
				assert.NotEmpty(t, result.Results)
			}
		})
	}
}

func TestValidateAuthentication(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		auth        ports.AuthenticationConfiguration
		expectValid bool
	}{
		{
			name: "valid token auth",
			auth: ports.AuthenticationConfiguration{
				Type:  ports.AuthenticationTypeToken,
				Token: "test-token",
			},
			expectValid: true,
		},
		{
			name: "valid basic auth",
			auth: ports.AuthenticationConfiguration{
				Type:     ports.AuthenticationTypeBasic,
				Username: "testuser",
				Password: "testpass",
			},
			expectValid: true,
		},
		{
			name: "valid SSH auth with key path",
			auth: ports.AuthenticationConfiguration{
				Type:       ports.AuthenticationTypeSSH,
				SSHKeyPath: "/path/to/key",
			},
			expectValid: true,
		},
		{
			name: "valid SSH auth with key content",
			auth: ports.AuthenticationConfiguration{
				Type:   ports.AuthenticationTypeSSH,
				SSHKey: "ssh-key-content",
			},
			expectValid: true,
		},
		{
			name: "invalid auth type",
			auth: ports.AuthenticationConfiguration{
				Type:  "invalid",
				Token: "test-token",
			},
			expectValid: false,
		},
		{
			name: "token auth missing token",
			auth: ports.AuthenticationConfiguration{
				Type:  ports.AuthenticationTypeToken,
				Token: "",
			},
			expectValid: false,
		},
		{
			name: "basic auth missing username",
			auth: ports.AuthenticationConfiguration{
				Type:     ports.AuthenticationTypeBasic,
				Username: "",
				Password: "testpass",
			},
			expectValid: false,
		},
		{
			name: "basic auth missing password",
			auth: ports.AuthenticationConfiguration{
				Type:     ports.AuthenticationTypeBasic,
				Username: "testuser",
				Password: "",
			},
			expectValid: false,
		},
		{
			name: "SSH auth missing key",
			auth: ports.AuthenticationConfiguration{
				Type:       ports.AuthenticationTypeSSH,
				SSHKeyPath: "",
				SSHKey:     "",
			},
			expectValid: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := ValidateAuthentication(test.auth)

			assert.Equal(t, test.expectValid, result.Valid)

			if !test.expectValid {
				assert.NotEmpty(t, result.Results)
			}
		})
	}
}

func TestValidateEnvironmentOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		options     ports.EnvironmentOptions
		expectValid bool
	}{
		{
			name: "valid options",
			options: ports.EnvironmentOptions{
				MaxConcurrency: 5,
				Timeout:        time.Minute * 30,
				RetryAttempts:  3,
				RetryDelay:     time.Second * 5,
			},
			expectValid: true,
		},
		{
			name: "zero values (should be valid)",
			options: ports.EnvironmentOptions{
				MaxConcurrency: 0,
				Timeout:        0,
				RetryAttempts:  0,
				RetryDelay:     0,
			},
			expectValid: true,
		},
		{
			name: "negative max concurrency",
			options: ports.EnvironmentOptions{
				MaxConcurrency: -1,
				Timeout:        time.Minute * 30,
				RetryAttempts:  3,
				RetryDelay:     time.Second * 5,
			},
			expectValid: false,
		},
		{
			name: "negative timeout",
			options: ports.EnvironmentOptions{
				MaxConcurrency: 5,
				Timeout:        -time.Minute,
				RetryAttempts:  3,
				RetryDelay:     time.Second * 5,
			},
			expectValid: false,
		},
		{
			name: "negative retry attempts",
			options: ports.EnvironmentOptions{
				MaxConcurrency: 5,
				Timeout:        time.Minute * 30,
				RetryAttempts:  -1,
				RetryDelay:     time.Second * 5,
			},
			expectValid: false,
		},
		{
			name: "negative retry delay",
			options: ports.EnvironmentOptions{
				MaxConcurrency: 5,
				Timeout:        time.Minute * 30,
				RetryAttempts:  3,
				RetryDelay:     -time.Second,
			},
			expectValid: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := ValidateEnvironmentOptions(test.options)

			assert.Equal(t, test.expectValid, result.Valid)

			if !test.expectValid {
				assert.NotEmpty(t, result.Results)
			}
		})
	}
}

func TestValidateRepositoryName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		repoName     string
		providerType string
		expectValid  bool
	}{
		// GitHub tests
		{"valid GitHub name", "my-repo", "github", true},
		{"GitHub with dots", "my.repo", "github", true},
		{"GitHub with underscores", "my_repo", "github", true},
		{"GitHub empty name", "", "github", false},
		{"GitHub name too long", strings.Repeat("a", 101), "github", false},
		{"GitHub invalid characters", "my repo", "github", false},
		{"GitHub starts with dot", ".myrepo", "github", false},
		{"GitHub ends with dot", "myrepo.", "github", false},
		{"GitHub starts with hyphen", "-myrepo", "github", false},
		{"GitHub ends with hyphen", "myrepo-", "github", false},

		// GitLab tests
		{"valid GitLab name", "my-repo", "gitlab", true},
		{"GitLab with dots and underscores", "my.repo_test", "gitlab", true},
		{"GitLab empty name", "", "gitlab", false},
		{"GitLab name too long", strings.Repeat("a", 101), "gitlab", false},
		{"GitLab invalid characters", "my repo", "gitlab", false},

		// Gitea tests
		{"valid Gitea name", "my-repo", "gitea", true},
		{"Gitea with dots", "my.repo", "gitea", true},
		{"Gitea empty name", "", "gitea", false},
		{"Gitea name too long", strings.Repeat("a", 101), "gitea", false},
		{"Gitea invalid characters", "my repo", "gitea", false},

		// Generic tests
		{"valid generic name", "my-repo", "other", true},
		{"generic empty name", "", "other", false},
		{"generic name too long", strings.Repeat("a", 101), "other", false},
		{"generic invalid characters", "my repo", "other", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := ValidateRepositoryName(test.repoName, test.providerType)

			assert.Equal(t, test.expectValid, result.Valid)

			if !test.expectValid {
				assert.NotEmpty(t, result.Code)
				assert.NotEmpty(t, result.Message)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		url         string
		expectValid bool
	}{
		{"valid HTTP URL", "http://example.com", true},
		{"valid HTTPS URL", "https://example.com", true},
		{"valid Git URL", "git://example.com/repo.git", true},
		{"valid URL with path", "https://example.com/path", true},
		{"valid URL with query", "https://example.com/path?query=value", true},
		{"empty URL", "", false},
		{"whitespace only URL", "   ", false},
		{"URL without scheme", "example.com", false},
		{"URL without host", "https://", false},
		{"invalid URL format", "not a url", false},
		{"malformed URL", "http://[invalid", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := ValidateURL(test.url)

			assert.Equal(t, test.expectValid, result.Valid)

			if !test.expectValid {
				assert.NotEmpty(t, result.Code)
				assert.NotEmpty(t, result.Message)
			}
		})
	}
}

func TestIsValidDomain(t *testing.T) {
	t.Parallel()

	validDomains := []string{"example.com", "sub.example.com", "test-github.example.com", "gitlab.example.org"}
	invalidDomains := []string{"", "invalid..domain", ".example.com", "example.com.", strings.Repeat("a", 256) + ".com"}

	for _, domain := range validDomains {
		assert.True(t, isValidDomain(domain), "Domain %s should be valid", domain)
	}

	for _, domain := range invalidDomains {
		assert.False(t, isValidDomain(domain), "Domain %s should be invalid", domain)
	}
}

func TestIsValidOwner(t *testing.T) {
	t.Parallel()

	validOwners := []string{"testuser", "test-user", "test_user", "user123"}
	invalidOwners := []string{"", "test user", "test@user", strings.Repeat("a", 108)}

	for _, owner := range validOwners {
		assert.True(t, isValidOwner(owner), "Owner %s should be valid", owner)
	}

	for _, owner := range invalidOwners {
		assert.False(t, isValidOwner(owner), "Owner %s should be invalid", owner)
	}
}

func TestProviderSpecificRepositoryNameValidation(t *testing.T) {
	t.Parallel()

	t.Run("validateGitHubRepositoryName", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name        string
			repoName    string
			expectValid bool
			expectCode  string
		}{
			{"valid name", "my-repo", true, ""},
			{"valid with dots", "my.repo", true, ""},
			{"valid with underscores", "my_repo", true, ""},
			{"too long", strings.Repeat("a", 101), false, "NAME_TOO_LONG"},
			{"invalid characters", "my repo", false, "INVALID_CHARACTERS"},
			{"starts with dot", ".myrepo", false, "INVALID_START_END"},
			{"ends with dot", "myrepo.", false, "INVALID_START_END"},
			{"starts with hyphen", "-myrepo", false, "INVALID_START_END"},
			{"ends with hyphen", "myrepo-", false, "INVALID_START_END"},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				result := validateGitHubRepositoryName(test.repoName)

				assert.Equal(t, test.expectValid, result.Valid)

				if !test.expectValid {
					assert.Equal(t, test.expectCode, result.Code)
				}
			})
		}
	})

	t.Run("validateGitLabRepositoryName", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name        string
			repoName    string
			expectValid bool
			expectCode  string
		}{
			{"valid name", "my-repo", true, ""},
			{"valid with dots and underscores", "my.repo_test", true, ""},
			{"too long", strings.Repeat("a", 101), false, "NAME_TOO_LONG"},
			{"invalid characters", "my repo", false, "INVALID_CHARACTERS"},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				result := validateGitLabRepositoryName(test.repoName)

				assert.Equal(t, test.expectValid, result.Valid)

				if !test.expectValid {
					assert.Equal(t, test.expectCode, result.Code)
				}
			})
		}
	})

	t.Run("validateGiteaRepositoryName", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name        string
			repoName    string
			expectValid bool
			expectCode  string
		}{
			{"valid name", "my-repo", true, ""},
			{"valid with dots", "my.repo", true, ""},
			{"too long", strings.Repeat("a", 101), false, "NAME_TOO_LONG"},
			{"invalid characters", "my repo", false, "INVALID_CHARACTERS"},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				result := validateGiteaRepositoryName(test.repoName)

				assert.Equal(t, test.expectValid, result.Valid)

				if !test.expectValid {
					assert.Equal(t, test.expectCode, result.Code)
				}
			})
		}
	})

	t.Run("validateGenericRepositoryName", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name        string
			repoName    string
			expectValid bool
			expectCode  string
		}{
			{"valid name", "my-repo", true, ""},
			{"valid with dots", "my.repo", true, ""},
			{"too long", strings.Repeat("a", 101), false, "NAME_TOO_LONG"},
			{"invalid characters", "my repo", false, "INVALID_CHARACTERS"},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				result := validateGenericRepositoryName(test.repoName)

				assert.Equal(t, test.expectValid, result.Valid)

				if !test.expectValid {
					assert.Equal(t, test.expectCode, result.Code)
				}
			})
		}
	})
}

func TestResultStructure(t *testing.T) {
	t.Parallel()

	t.Run("Result fields", func(t *testing.T) {
		t.Parallel()

		result := Result{
			Valid:      false,
			Field:      "test_field",
			Code:       "TEST_ERROR",
			Message:    "Test error message",
			Value:      "test_value",
			Suggestion: "Try this instead",
		}

		assert.False(t, result.Valid)
		assert.Equal(t, "test_field", result.Field)
		assert.Equal(t, "TEST_ERROR", result.Code)
		assert.Equal(t, "Test error message", result.Message)
		assert.Equal(t, "test_value", result.Value)
		assert.Equal(t, "Try this instead", result.Suggestion)
	})

	t.Run("Results aggregation", func(t *testing.T) {
		t.Parallel()

		results := Results{
			Valid: false,
			Results: []Result{
				{Valid: false, Code: "ERROR1", Message: "First error"},
				{Valid: false, Code: "ERROR2", Message: "Second error"},
			},
		}

		assert.False(t, results.Valid)
		require.Len(t, results.Results, 2)
		assert.Equal(t, "ERROR1", results.Results[0].Code)
		assert.Equal(t, "ERROR2", results.Results[1].Code)
	})
}
