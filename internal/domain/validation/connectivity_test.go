// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package validation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// Mock connectivity validator for testing.
type mockConnectivityValidator struct {
	results map[string]ConnectivityResult
}

func (m *mockConnectivityValidator) ValidateConnectivity(_ context.Context, validation ConnectivityValidation) ConnectivityResult {
	if result, exists := m.results[validation.Target]; exists {
		result.Validation = validation

		return result
	}

	// Default successful result
	return ConnectivityResult{
		Validation: validation,
		Success:    true,
		Duration:   time.Millisecond * 100,
		Details:    make(map[string]any),
	}
}

func TestPlanConnectivityValidations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		config             ports.AppConfiguration
		expectedValidation int
	}{
		{
			name: "single environment with source and mirror",
			config: ports.AppConfiguration{
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
					},
				},
			},
			expectedValidation: 6, // 3 for source (HTTP + API + Git) + 3 for mirror (HTTP + API + Git)
		},
		{
			name: "disabled environment",
			config: ports.AppConfiguration{
				Environments: map[string]ports.EnvironmentConfiguration{
					"disabled": {
						Name:    "disabled",
						Enabled: false,
						Source: ports.SourceConfiguration{
							ProviderType: "github",
							Domain:       "test-github.example.com",
							Owner:        "testorg",
						},
					},
				},
			},
			expectedValidation: 0,
		},
		{
			name: "local providers",
			config: ports.AppConfiguration{
				Environments: map[string]ports.EnvironmentConfiguration{
					"local": {
						Name:    "local",
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
							"directory": {
								Name:         "directory",
								ProviderType: "directory",
								Path:         "/path/to/backup",
								Enabled:      true,
							},
							"archive": {
								Name:         "archive",
								ProviderType: "archive",
								Path:         "/path/to/backup.tar.gz",
								Enabled:      true,
							},
						},
					},
				},
			},
			expectedValidation: 3, // Only source validations (HTTP + API + Git), no mirror validations for local providers
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			validations := PlanConnectivityValidations(test.config)

			assert.Len(t, validations, test.expectedValidation)

			// Verify all validations have required fields
			for _, validation := range validations {
				assert.NotEmpty(t, validation.Type)
				assert.NotEmpty(t, validation.Target)
				assert.NotEmpty(t, validation.Description)
				assert.NotZero(t, validation.Timeout)
			}
		})
	}
}

func TestPlanProviderConnectivity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		providerType        string
		domain              string
		owner               string
		auth                ports.AuthenticationConfiguration
		expectedValidations int
		expectedTypes       []ConnectivityType
	}{
		{
			name:         "GitHub with token auth",
			providerType: "github",
			domain:       "test-github.example.com",
			owner:        "testorg",
			auth: ports.AuthenticationConfiguration{
				Type:  ports.AuthenticationTypeToken,
				Token: "test-token",
			},
			expectedValidations: 3, // HTTP, API, Git
			expectedTypes:       []ConnectivityType{ConnectivityTypeHTTP, ConnectivityTypeProvider, ConnectivityTypeGit},
		},
		{
			name:         "GitHub with SSH auth",
			providerType: "github",
			domain:       "test-github.example.com",
			owner:        "testorg",
			auth: ports.AuthenticationConfiguration{
				Type:       ports.AuthenticationTypeSSH,
				SSHKeyPath: "/path/to/key",
			},
			expectedValidations: 4, // HTTP, API, Git, SSH
			expectedTypes:       []ConnectivityType{ConnectivityTypeHTTP, ConnectivityTypeProvider, ConnectivityTypeGit, ConnectivityTypeSSH},
		},
		{
			name:         "GitLab enterprise",
			providerType: "gitlab",
			domain:       "gitlab.company.com",
			owner:        "testorg",
			auth: ports.AuthenticationConfiguration{
				Type:  ports.AuthenticationTypeToken,
				Token: "test-token",
			},
			expectedValidations: 3, // HTTP, API, Git
			expectedTypes:       []ConnectivityType{ConnectivityTypeHTTP, ConnectivityTypeProvider, ConnectivityTypeGit},
		},
		{
			name:         "Gitea instance",
			providerType: "gitea",
			domain:       "gitea.company.com",
			owner:        "testorg",
			auth: ports.AuthenticationConfiguration{
				Type:  ports.AuthenticationTypeToken,
				Token: "test-token",
			},
			expectedValidations: 3, // HTTP, API, Git
			expectedTypes:       []ConnectivityType{ConnectivityTypeHTTP, ConnectivityTypeProvider, ConnectivityTypeGit},
		},
		{
			name:         "no auth",
			providerType: "github",
			domain:       "test-github.example.com",
			owner:        "testorg",
			auth: ports.AuthenticationConfiguration{
				Type: ports.AuthenticationTypeNone,
			},
			expectedValidations: 2, // HTTP, API (no Git without auth)
			expectedTypes:       []ConnectivityType{ConnectivityTypeHTTP, ConnectivityTypeProvider},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			validations := PlanProviderConnectivity(test.providerType, test.domain, test.owner, test.auth)

			assert.Len(t, validations, test.expectedValidations)

			// Check that all expected types are present
			actualTypes := make([]ConnectivityType, len(validations))
			for i, validation := range validations {
				actualTypes[i] = validation.Type
			}

			for _, expectedType := range test.expectedTypes {
				assert.Contains(t, actualTypes, expectedType)
			}

			// Verify all validations have required fields
			for _, validation := range validations {
				assert.NotEmpty(t, validation.Target)
				assert.NotEmpty(t, validation.Description)
				assert.NotZero(t, validation.Timeout)
			}
		})
	}
}

func TestConnectivityURLBuilders(t *testing.T) {
	t.Parallel()

	t.Run("buildProviderURL", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name         string
			providerType string
			domain       string
			expected     string
		}{
			{"GitHub default", "github", "test-github.example.com", "https://test-github.example.com"},
			{"GitHub custom", "github", "github.company.com", "https://github.company.com"},
			{"GitLab default", "gitlab", "test-gitlab.example.com", "https://test-gitlab.example.com"},
			{"GitLab custom", "gitlab", "gitlab.company.com", "https://gitlab.company.com"},
			{"Unknown provider", "unknown", "", ""},
			{"Custom domain", "unknown", "custom.com", "https://custom.com"},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				result := buildProviderURL(test.providerType, test.domain)
				assert.Equal(t, test.expected, result)
			})
		}
	})

	t.Run("buildProviderAPIURL", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name         string
			providerType string
			domain       string
			owner        string
			expected     string
		}{
			{"GitHub default", "github", "test-github.example.com", "testorg", "https://test-github.example.com/api/v3/users/testorg"},
			{"GitHub enterprise", "github", "github.company.com", "testorg", "https://github.company.com/api/v3/users/testorg"},
			{"GitLab default", "gitlab", "test-gitlab.example.com", "testorg", "https://test-gitlab.example.com/api/v4/users?username=testorg"},
			{"GitLab self-hosted", "gitlab", "gitlab.company.com", "testorg", "https://gitlab.company.com/api/v4/users?username=testorg"},
			{"Gitea instance", "gitea", "gitea.company.com", "testorg", "https://gitea.company.com/api/v1/users/testorg"},
			{"Gitea no domain", "gitea", "", "testorg", ""},
			{"Unknown provider", "unknown", "example.com", "testorg", ""},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				result := buildProviderAPIURL(test.providerType, test.domain, test.owner)
				assert.Equal(t, test.expected, result)
			})
		}
	})

	t.Run("buildGitURL", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name         string
			providerType string
			domain       string
			owner        string
			repo         string
			expected     string
		}{
			{"GitHub default", "github", "test-github.example.com", "testorg", "testrepo", "https://test-github.example.com/testorg/testrepo.git"},
			{"GitHub enterprise", "github", "github.company.com", "testorg", "testrepo", "https://github.company.com/testorg/testrepo.git"},
			{"GitLab default", "gitlab", "test-gitlab.example.com", "testorg", "testrepo", "https://test-gitlab.example.com/testorg/testrepo.git"},
			{"GitLab self-hosted", "gitlab", "gitlab.company.com", "testorg", "testrepo", "https://gitlab.company.com/testorg/testrepo.git"},
			{"Gitea instance", "gitea", "gitea.company.com", "testorg", "testrepo", "https://gitea.company.com/testorg/testrepo.git"},
			{"Gitea no domain", "gitea", "", "testorg", "testrepo", ""},
			{"Unknown provider", "unknown", "example.com", "testorg", "testrepo", ""},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				result := buildGitURL(test.providerType, test.domain, test.owner, test.repo)
				assert.Equal(t, test.expected, result)
			})
		}
	})

	t.Run("buildSSHHost", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name         string
			providerType string
			domain       string
			expected     string
		}{
			{"GitHub default", "github", "test-github.example.com", "git@test-github.example.com"},
			{"GitHub enterprise", "github", "github.company.com", "git@github.company.com"},
			{"GitLab default", "gitlab", "test-gitlab.example.com", "git@test-gitlab.example.com"},
			{"GitLab self-hosted", "gitlab", "gitlab.company.com", "git@gitlab.company.com"},
			{"Gitea instance", "gitea", "gitea.company.com", "git@gitea.company.com"},
			{"Gitea no domain", "gitea", "", ""},
			{"Unknown provider", "unknown", "example.com", ""},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				result := buildSSHHost(test.providerType, test.domain)
				assert.Equal(t, test.expected, result)
			})
		}
	})
}

func TestPlanFileSystemValidations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		config              ports.AppConfiguration
		expectedValidations int
	}{
		{
			name: "global directories configured",
			config: ports.AppConfiguration{
				GlobalSettings: ports.GlobalSettings{
					TempDirectory:  "/tmp/git-sync",
					CacheDirectory: "/cache/git-sync",
				},
				Environments: map[string]ports.EnvironmentConfiguration{},
			},
			expectedValidations: 2,
		},
		{
			name: "local mirror paths",
			config: ports.AppConfiguration{
				Environments: map[string]ports.EnvironmentConfiguration{
					"local": {
						Name:    "local",
						Enabled: true,
						Source: ports.SourceConfiguration{
							ProviderType: "github",
							Domain:       "test-github.example.com",
							Owner:        "testorg",
						},
						Mirrors: map[string]ports.MirrorConfiguration{
							"directory": {
								Name:         "directory",
								ProviderType: "directory",
								Path:         "/path/to/backup",
								Enabled:      true,
							},
							"archive": {
								Name:         "archive",
								ProviderType: "archive",
								Path:         "/path/to/backup.tar.gz",
								Enabled:      true,
							},
						},
					},
				},
			},
			expectedValidations: 2,
		},
		{
			name: "mixed configuration",
			config: ports.AppConfiguration{
				GlobalSettings: ports.GlobalSettings{
					TempDirectory: "/tmp/git-sync",
				},
				Environments: map[string]ports.EnvironmentConfiguration{
					"mixed": {
						Name:    "mixed",
						Enabled: true,
						Source: ports.SourceConfiguration{
							ProviderType: "github",
							Domain:       "test-github.example.com",
							Owner:        "testorg",
						},
						Mirrors: map[string]ports.MirrorConfiguration{
							"directory": {
								Name:         "directory",
								ProviderType: "directory",
								Path:         "/path/to/backup",
								Enabled:      true,
							},
							"remote": {
								Name:         "remote",
								ProviderType: "gitlab",
								Domain:       "test-gitlab.example.com",
								Owner:        "backuporg",
								Enabled:      true,
							},
						},
					},
				},
			},
			expectedValidations: 2, // 1 for temp dir + 1 for directory mirror
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			validations := PlanFileSystemValidations(test.config)

			assert.Len(t, validations, test.expectedValidations)

			// Verify all validations have required fields
			for _, validation := range validations {
				assert.NotEmpty(t, validation.Type)
				assert.NotEmpty(t, validation.Path)
				assert.NotEmpty(t, validation.Description)
			}
		})
	}
}

func TestValidateAllConnectivity(t *testing.T) {
	t.Parallel()

	t.Run("all successful validations", func(t *testing.T) {
		t.Parallel()

		validator := &mockConnectivityValidator{
			results: map[string]ConnectivityResult{},
		}

		validations := []ConnectivityValidation{
			{Type: ConnectivityTypeHTTP, Target: "https://test-github.example.com", Required: true},
			{Type: ConnectivityTypeProvider, Target: "https://api.test-github.example.com", Required: true},
		}

		results := ValidateAllConnectivity(context.Background(), validator, validations)

		require.Len(t, results, 2)

		for _, result := range results {
			assert.True(t, result.Success)
		}
	})

	t.Run("required validation failure stops execution", func(t *testing.T) {
		t.Parallel()

		validator := &mockConnectivityValidator{
			results: map[string]ConnectivityResult{
				"https://test-github.example.com": {
					Success: false,
					Error:   errors.New("connection failed"),
				},
			},
		}

		validations := []ConnectivityValidation{
			{Type: ConnectivityTypeHTTP, Target: "https://test-github.example.com", Required: true},
			{Type: ConnectivityTypeProvider, Target: "https://api.test-github.example.com", Required: true},
			{Type: ConnectivityTypeGit, Target: "https://test-github.example.com/test/repo.git", Required: false},
		}

		results := ValidateAllConnectivity(context.Background(), validator, validations)

		require.Len(t, results, 3)
		assert.False(t, results[0].Success)
		assert.False(t, results[1].Success) // Skipped
		assert.False(t, results[2].Success) // Skipped
		assert.Equal(t, ErrValidationSkipped, results[1].Error)
		assert.Equal(t, ErrValidationSkipped, results[2].Error)
	})

	t.Run("optional validation failure continues execution", func(t *testing.T) {
		t.Parallel()

		validator := &mockConnectivityValidator{
			results: map[string]ConnectivityResult{
				"https://test-github.example.com/test/repo.git": {
					Success: false,
					Error:   errors.New("repository not found"),
				},
			},
		}

		validations := []ConnectivityValidation{
			{Type: ConnectivityTypeHTTP, Target: "https://test-github.example.com", Required: true},
			{Type: ConnectivityTypeGit, Target: "https://test-github.example.com/test/repo.git", Required: false},
			{Type: ConnectivityTypeProvider, Target: "https://api.test-github.example.com", Required: true},
		}

		results := ValidateAllConnectivity(context.Background(), validator, validations)

		require.Len(t, results, 3)
		assert.True(t, results[0].Success)  // HTTP - success
		assert.False(t, results[1].Success) // Git - failed but optional
		assert.True(t, results[2].Success)  // API - success (execution continued)
	})
}

func TestCountConnectivityFailures(t *testing.T) {
	t.Parallel()

	results := []ConnectivityResult{
		{Success: true, Validation: ConnectivityValidation{Required: true}},
		{Success: false, Validation: ConnectivityValidation{Required: true}},
		{Success: false, Validation: ConnectivityValidation{Required: false}},
		{Success: false, Validation: ConnectivityValidation{Required: true}},
		{Success: true, Validation: ConnectivityValidation{Required: false}},
	}

	total, required, optional := CountConnectivityFailures(results)

	assert.Equal(t, 3, total)    // 3 failures total
	assert.Equal(t, 2, required) // 2 required failures
	assert.Equal(t, 1, optional) // 1 optional failure
}

func TestFilterConnectivityResults(t *testing.T) {
	t.Parallel()

	results := []ConnectivityResult{
		{Success: true, Validation: ConnectivityValidation{Type: ConnectivityTypeHTTP}},
		{Success: false, Validation: ConnectivityValidation{Type: ConnectivityTypeHTTP}},
		{Success: true, Validation: ConnectivityValidation{Type: ConnectivityTypeProvider}},
		{Success: false, Validation: ConnectivityValidation{Type: ConnectivityTypeGit}},
	}

	t.Run("filter by type", func(t *testing.T) {
		t.Parallel()

		httpResults := FilterConnectivityResults(results, ConnectivityTypeHTTP, false)
		assert.Len(t, httpResults, 2)

		providerResults := FilterConnectivityResults(results, ConnectivityTypeProvider, false)
		assert.Len(t, providerResults, 1)
	})

	t.Run("filter by success", func(t *testing.T) {
		t.Parallel()

		successResults := FilterConnectivityResults(results, "", true)
		assert.Len(t, successResults, 2)

		for _, result := range successResults {
			assert.True(t, result.Success)
		}
	})

	t.Run("filter by type and success", func(t *testing.T) {
		t.Parallel()

		httpSuccessResults := FilterConnectivityResults(results, ConnectivityTypeHTTP, true)
		assert.Len(t, httpSuccessResults, 1)
		assert.True(t, httpSuccessResults[0].Success)
		assert.Equal(t, ConnectivityTypeHTTP, httpSuccessResults[0].Validation.Type)
	})

	t.Run("no filter", func(t *testing.T) {
		t.Parallel()

		allResults := FilterConnectivityResults(results, "", false)
		assert.Len(t, allResults, 4)
	})
}

func TestConnectivityTypes(t *testing.T) {
	t.Parallel()

	// Test that constants are defined correctly
	assert.Equal(t, ConnectivityTypeHTTP, ConnectivityType("http"))
	assert.Equal(t, ConnectivityTypeGit, ConnectivityType("git"))
	assert.Equal(t, ConnectivityTypeSSH, ConnectivityType("ssh"))
	assert.Equal(t, ConnectivityTypeProvider, ConnectivityType("provider"))
}

func TestFileSystemTypes(t *testing.T) {
	t.Parallel()

	// Test that constants are defined correctly
	assert.Equal(t, FileSystemTypeFile, FileSystemType("file"))
	assert.Equal(t, FileSystemTypeDirectory, FileSystemType("directory"))
	assert.Equal(t, FileSystemTypeArchive, FileSystemType("archive"))

	t.Run("getFileSystemType", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			providerType string
			expected     FileSystemType
		}{
			{"directory", FileSystemTypeDirectory},
			{"archive", FileSystemTypeArchive},
			{"unknown", FileSystemTypeFile},
			{"", FileSystemTypeFile},
		}

		for _, test := range tests {
			t.Run(test.providerType, func(t *testing.T) {
				t.Parallel()

				result := getFileSystemType(test.providerType)
				assert.Equal(t, test.expected, result)
			})
		}
	})
}

func TestValidationError(t *testing.T) {
	t.Parallel()

	err := Error{
		Code:    "TEST_ERROR",
		Message: "This is a test error",
	}

	assert.Equal(t, "This is a test error", err.Error())
	assert.Equal(t, "TEST_ERROR", err.Code)
	assert.Equal(t, "This is a test error", err.Message)
}

func TestConnectivityStructures(t *testing.T) {
	t.Parallel()

	t.Run("ConnectivityValidation", func(t *testing.T) {
		t.Parallel()

		validation := ConnectivityValidation{
			Type:        ConnectivityTypeHTTP,
			Target:      "https://test-github.example.com",
			Timeout:     time.Second * 30,
			Description: "Test connectivity",
			Required:    true,
		}

		assert.Equal(t, ConnectivityTypeHTTP, validation.Type)
		assert.Equal(t, "https://test-github.example.com", validation.Target)
		assert.Equal(t, time.Second*30, validation.Timeout)
		assert.Equal(t, "Test connectivity", validation.Description)
		assert.True(t, validation.Required)
	})

	t.Run("ConnectivityResult", func(t *testing.T) {
		t.Parallel()

		validation := ConnectivityValidation{Type: ConnectivityTypeHTTP, Target: "https://test-github.example.com"}
		details := map[string]any{"status_code": 200}

		result := ConnectivityResult{
			Validation: validation,
			Success:    true,
			Error:      nil,
			Duration:   time.Millisecond * 100,
			Details:    details,
		}

		assert.Equal(t, validation, result.Validation)
		assert.True(t, result.Success)
		require.NoError(t, result.Error)
		assert.Equal(t, time.Millisecond*100, result.Duration)
		assert.Equal(t, details, result.Details)
	})

	t.Run("FileSystemValidation", func(t *testing.T) {
		t.Parallel()

		validation := FileSystemValidation{
			Type:        FileSystemTypeDirectory,
			Path:        "/path/to/directory",
			Required:    true,
			Writable:    true,
			Description: "Test directory access",
		}

		assert.Equal(t, FileSystemTypeDirectory, validation.Type)
		assert.Equal(t, "/path/to/directory", validation.Path)
		assert.True(t, validation.Required)
		assert.True(t, validation.Writable)
		assert.Equal(t, "Test directory access", validation.Description)
	})

	t.Run("FileSystemResult", func(t *testing.T) {
		t.Parallel()

		validation := FileSystemValidation{Type: FileSystemTypeDirectory, Path: "/path/to/directory"}
		details := map[string]any{"permissions": "755"}

		result := FileSystemResult{
			Validation: validation,
			Success:    true,
			Error:      nil,
			Exists:     true,
			Readable:   true,
			Writable:   true,
			Details:    details,
		}

		assert.Equal(t, validation, result.Validation)
		assert.True(t, result.Success)
		require.NoError(t, result.Error)
		assert.True(t, result.Exists)
		assert.True(t, result.Readable)
		assert.True(t, result.Writable)
		assert.Equal(t, details, result.Details)
	})
}
