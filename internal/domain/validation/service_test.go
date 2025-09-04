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

// Mock file system validator for testing.
type mockFileSystemValidator struct {
	results map[string]FileSystemResult
}

func (m *mockFileSystemValidator) ValidateFileSystem(_ context.Context, validation FileSystemValidation) FileSystemResult {
	if result, exists := m.results[validation.Path]; exists {
		result.Validation = validation

		return result
	}

	// Default successful result
	return FileSystemResult{
		Validation: validation,
		Success:    true,
		Exists:     true,
		Readable:   true,
		Writable:   validation.Writable,
		Details:    make(map[string]any),
	}
}

func TestNewDefaultValidationService(t *testing.T) {
	t.Parallel()

	connectivityValidator := &mockConnectivityValidator{}
	fileSystemValidator := &mockFileSystemValidator{}

	service := NewDefaultValidationService(connectivityValidator, fileSystemValidator)

	require.NotNil(t, service)
	assert.True(t, service.config.EnableConnectivityTests)
	assert.True(t, service.config.EnableFileSystemTests)
	assert.Equal(t, time.Second*30, service.config.ConnectivityTimeout)
	assert.False(t, service.config.SkipOptionalTests)
	assert.Equal(t, 5, service.config.MaxConcurrentTests)
}

func TestNewQuickValidationService(t *testing.T) {
	t.Parallel()

	service := NewQuickValidationService()

	require.NotNil(t, service)
	assert.False(t, service.config.EnableConnectivityTests)
	assert.False(t, service.config.EnableFileSystemTests)
	assert.Equal(t, time.Second*5, service.config.ConnectivityTimeout)
	assert.True(t, service.config.SkipOptionalTests)
	assert.Equal(t, 1, service.config.MaxConcurrentTests)
	assert.Nil(t, service.connectivityValidator)
	assert.Nil(t, service.fileSystemValidator)
}

func TestNewFullValidationService(t *testing.T) {
	t.Parallel()

	connectivityValidator := &mockConnectivityValidator{}
	fileSystemValidator := &mockFileSystemValidator{}

	service := NewFullValidationService(connectivityValidator, fileSystemValidator)

	require.NotNil(t, service)
	assert.True(t, service.config.EnableConnectivityTests)
	assert.True(t, service.config.EnableFileSystemTests)
	assert.Equal(t, time.Second*60, service.config.ConnectivityTimeout)
	assert.False(t, service.config.SkipOptionalTests)
	assert.Equal(t, 10, service.config.MaxConcurrentTests)
}

func TestValidateConfiguration(t *testing.T) {
	t.Parallel()

	t.Run("valid configuration", func(t *testing.T) {
		t.Parallel()

		connectivityValidator := &mockConnectivityValidator{}
		fileSystemValidator := &mockFileSystemValidator{}
		service := NewDefaultValidationService(connectivityValidator, fileSystemValidator)

		config := ports.AppConfiguration{
			GlobalSettings: ports.GlobalSettings{
				LogLevel:  ports.LogLevelInfo,
				LogFormat: ports.LogFormatJSON,
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
				},
			},
		}

		result := service.ValidateConfiguration(context.Background(), config)

		assert.True(t, result.OverallSuccess)
		assert.True(t, result.ConfigurationResults.Valid)
		assert.Zero(t, result.TotalErrors)
		assert.NotZero(t, result.Duration)
		assert.NotNil(t, result.Summary)
	})

	t.Run("invalid configuration", func(t *testing.T) {
		t.Parallel()

		connectivityValidator := &mockConnectivityValidator{}
		fileSystemValidator := &mockFileSystemValidator{}
		service := NewDefaultValidationService(connectivityValidator, fileSystemValidator)

		config := ports.AppConfiguration{
			GlobalSettings: ports.GlobalSettings{
				LogLevel:     "invalid",
				LogFormat:    "invalid",
				MaxCacheSize: -100,
			},
			Environments: map[string]ports.EnvironmentConfiguration{
				"invalid": {
					Name:    "",
					Enabled: true,
					Source: ports.SourceConfiguration{
						ProviderType: "invalid",
						Owner:        "",
						Authentication: ports.AuthenticationConfiguration{
							Type: "invalid",
						},
					},
				},
			},
		}

		result := service.ValidateConfiguration(context.Background(), config)

		assert.False(t, result.OverallSuccess)
		assert.False(t, result.ConfigurationResults.Valid)
		assert.Positive(t, result.TotalErrors)
		assert.NotZero(t, result.Duration)

		// Check summary
		assert.False(t, result.Summary.ConfigurationValid)
		assert.NotEmpty(t, result.Summary.CriticalIssues)
	})

	t.Run("connectivity tests disabled", func(t *testing.T) {
		t.Parallel()

		service := NewQuickValidationService()

		config := ports.AppConfiguration{
			GlobalSettings: ports.GlobalSettings{
				LogLevel:  ports.LogLevelInfo,
				LogFormat: ports.LogFormatJSON,
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
				},
			},
		}

		result := service.ValidateConfiguration(context.Background(), config)

		assert.True(t, result.OverallSuccess)
		assert.Empty(t, result.ConnectivityResults)
		assert.Empty(t, result.FileSystemResults)
	})

	t.Run("connectivity test failures", func(t *testing.T) {
		t.Parallel()

		connectivityValidator := &mockConnectivityValidator{
			results: map[string]ConnectivityResult{
				"https://test-github.example.com": {
					Success: false,
					Error:   errors.New("connection failed"),
				},
			},
		}
		fileSystemValidator := &mockFileSystemValidator{}
		service := NewDefaultValidationService(connectivityValidator, fileSystemValidator)

		config := ports.AppConfiguration{
			GlobalSettings: ports.GlobalSettings{
				LogLevel:  ports.LogLevelInfo,
				LogFormat: ports.LogFormatJSON,
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
				},
			},
		}

		result := service.ValidateConfiguration(context.Background(), config)

		assert.False(t, result.OverallSuccess)
		assert.Positive(t, result.TotalErrors)
		assert.NotEmpty(t, result.ConnectivityResults)

		// Check that some connectivity tests failed
		hasFailures := false

		for _, connResult := range result.ConnectivityResults {
			if !connResult.Success {
				hasFailures = true

				break
			}
		}

		assert.True(t, hasFailures)
	})

	t.Run("file system test failures", func(t *testing.T) {
		t.Parallel()

		connectivityValidator := &mockConnectivityValidator{}
		fileSystemValidator := &mockFileSystemValidator{
			results: map[string]FileSystemResult{
				"/tmp/git-sync": {
					Success: false,
					Error:   errors.New("permission denied"),
				},
			},
		}
		service := NewDefaultValidationService(connectivityValidator, fileSystemValidator)

		config := ports.AppConfiguration{
			GlobalSettings: ports.GlobalSettings{
				LogLevel:       ports.LogLevelInfo,
				LogFormat:      ports.LogFormatJSON,
				TempDirectory:  "/tmp/git-sync",
				CacheDirectory: "/cache/git-sync",
			},
			Environments: map[string]ports.EnvironmentConfiguration{},
		}

		result := service.ValidateConfiguration(context.Background(), config)

		assert.False(t, result.OverallSuccess)
		assert.Positive(t, result.TotalErrors)
		assert.NotEmpty(t, result.FileSystemResults)

		// Check that some file system tests failed
		hasFailures := false

		for _, fsResult := range result.FileSystemResults {
			if !fsResult.Success {
				hasFailures = true

				break
			}
		}

		assert.True(t, hasFailures)
	})
}

func TestService_ValidateEnvironment(t *testing.T) {
	t.Parallel()

	service := NewQuickValidationService()

	env := ports.EnvironmentConfiguration{
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
	}

	result := service.ValidateEnvironment(context.Background(), env)

	assert.True(t, result.Valid)
	assert.Empty(t, result.Results)
}

func TestService_ValidateRepositoryName(t *testing.T) {
	t.Parallel()

	service := NewQuickValidationService()

	tests := []struct {
		name         string
		repoName     string
		providerType string
		expectValid  bool
	}{
		{"valid GitHub repo", "my-repo", "github", true},
		{"invalid GitHub repo", "", "github", false},
		{"valid GitLab repo", "my-repo", "gitlab", true},
		{"invalid GitLab repo", "my repo", "gitlab", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := service.ValidateRepositoryName(test.repoName, test.providerType)

			assert.Equal(t, test.expectValid, result.Valid)
		})
	}
}

func TestService_ValidateURL(t *testing.T) {
	t.Parallel()

	service := NewQuickValidationService()

	tests := []struct {
		name        string
		url         string
		expectValid bool
	}{
		{"valid URL", "https://test-github.example.com", true},
		{"invalid URL", "not a url", false},
		{"empty URL", "", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := service.ValidateURL(test.url)

			assert.Equal(t, test.expectValid, result.Valid)
		})
	}
}

func TestConnectivity(t *testing.T) {
	t.Parallel()

	t.Run("connectivity tests enabled", func(t *testing.T) {
		t.Parallel()

		connectivityValidator := &mockConnectivityValidator{}
		service := NewDefaultValidationService(connectivityValidator, nil)

		auth := ports.AuthenticationConfiguration{
			Type:  ports.AuthenticationTypeToken,
			Token: "test-token",
		}

		results := service.TestConnectivity(context.Background(), "github", "test-github.example.com", "testorg", auth)

		assert.NotEmpty(t, results)

		for _, result := range results {
			assert.True(t, result.Success) // Mock returns success by default
		}
	})

	t.Run("connectivity tests disabled", func(t *testing.T) {
		t.Parallel()

		service := NewQuickValidationService()

		auth := ports.AuthenticationConfiguration{
			Type:  ports.AuthenticationTypeToken,
			Token: "test-token",
		}

		results := service.TestConnectivity(context.Background(), "github", "test-github.example.com", "testorg", auth)

		assert.Empty(t, results)
	})
}

func TestHasCriticalErrors(t *testing.T) {
	t.Parallel()

	t.Run("no errors", func(t *testing.T) {
		t.Parallel()

		result := ComprehensiveResult{
			OverallSuccess: true,
			TotalErrors:    0,
		}

		assert.False(t, HasCriticalErrors(result))
	})

	t.Run("overall success false", func(t *testing.T) {
		t.Parallel()

		result := ComprehensiveResult{
			OverallSuccess: false,
			TotalErrors:    0,
		}

		assert.True(t, HasCriticalErrors(result))
	})

	t.Run("has errors", func(t *testing.T) {
		t.Parallel()

		result := ComprehensiveResult{
			OverallSuccess: true,
			TotalErrors:    1,
		}

		assert.True(t, HasCriticalErrors(result))
	})
}

func TestGetValidationErrors(t *testing.T) {
	t.Parallel()

	result := ComprehensiveResult{
		ConfigurationResults: Results{
			Results: []Result{
				{Valid: false, Field: "log_level", Message: "Invalid log level"},
				{Valid: true, Field: "domain", Message: "Valid domain"},
			},
		},
		ConnectivityResults: []ConnectivityResult{
			{
				Success: false,
				Validation: ConnectivityValidation{
					Description: "HTTP connectivity",
					Required:    true,
				},
			},
			{
				Success: false,
				Validation: ConnectivityValidation{
					Description: "Git connectivity",
					Required:    false,
				},
			},
		},
		FileSystemResults: []FileSystemResult{
			{
				Success: false,
				Validation: FileSystemValidation{
					Description: "Temp directory access",
					Required:    true,
				},
			},
		},
		RepositoryResults: []RepositoryResult{
			{
				Valid:          false,
				RepositoryName: "test-repo",
				Results: []Result{
					{Valid: false, Message: "Invalid repository name"},
				},
			},
		},
	}

	errors := GetValidationErrors(result)

	assert.Contains(t, errors, "log_level: Invalid log level")
	assert.Contains(t, errors, "Connectivity: HTTP connectivity failed")
	assert.Contains(t, errors, "File system: Temp directory access failed")
	assert.Contains(t, errors, "Repository test-repo: Invalid repository name")
	assert.NotContains(t, errors, "Git connectivity") // Optional, not included
}

func TestGetValidationWarnings(t *testing.T) {
	t.Parallel()

	result := ComprehensiveResult{
		ConnectivityResults: []ConnectivityResult{
			{
				Success: false,
				Validation: ConnectivityValidation{
					Description: "HTTP connectivity",
					Required:    true,
				},
			},
			{
				Success: false,
				Validation: ConnectivityValidation{
					Description: "Git connectivity",
					Required:    false,
				},
			},
		},
		FileSystemResults: []FileSystemResult{
			{
				Success: false,
				Validation: FileSystemValidation{
					Description: "Cache directory access",
					Required:    false,
				},
			},
		},
	}

	warnings := GetValidationWarnings(result)

	assert.Contains(t, warnings, "Connectivity: Git connectivity failed (optional)")
	assert.Contains(t, warnings, "File system: Cache directory access failed (optional)")
	assert.NotContains(t, warnings, "HTTP connectivity") // Required, not a warning
}

func TestBuildSummary(t *testing.T) {
	t.Parallel()

	service := NewDefaultValidationService(&mockConnectivityValidator{}, &mockFileSystemValidator{})

	result := ComprehensiveResult{
		ConfigurationResults: Results{
			Valid: false,
			Results: []Result{
				{Valid: false, Field: "log_level", Message: "Invalid log level", Suggestion: "Use info, debug, or error"},
			},
		},
		ConnectivityResults: []ConnectivityResult{
			{
				Success: false,
				Validation: ConnectivityValidation{
					Description: "HTTP connectivity",
					Required:    true,
				},
			},
			{
				Success: false,
				Validation: ConnectivityValidation{
					Description: "Git connectivity",
					Required:    false,
				},
			},
		},
		FileSystemResults: []FileSystemResult{
			{
				Success: false,
				Validation: FileSystemValidation{
					Description: "Temp directory access",
					Required:    true,
				},
			},
			{
				Success: false,
				Validation: FileSystemValidation{
					Description: "Cache directory access",
					Required:    false,
				},
			},
		},
		RepositoryResults: []RepositoryResult{
			{
				Valid:          false,
				RepositoryName: "test-repo",
				Results: []Result{
					{Valid: false, Message: "Invalid repository name", Suggestion: "Use alphanumeric characters only"},
				},
			},
		},
	}

	summary := service.buildSummary(result)

	assert.False(t, summary.ConfigurationValid)
	assert.False(t, summary.ConnectivityValid)
	assert.False(t, summary.FileSystemValid)
	assert.False(t, summary.RepositoryNamesValid)

	assert.NotEmpty(t, summary.CriticalIssues)
	assert.NotEmpty(t, summary.Warnings)
	assert.NotEmpty(t, summary.Suggestions)

	// Check that critical issues include required failures
	assert.Contains(t, summary.CriticalIssues, "Configuration error in log_level: Invalid log level")
	assert.Contains(t, summary.CriticalIssues, "Critical connectivity issues found")
	assert.Contains(t, summary.CriticalIssues, "File system error: Temp directory access")

	// Check that warnings include optional failures
	assert.Contains(t, summary.Warnings, "Some optional connectivity tests failed")
	assert.Contains(t, summary.Warnings, "File system warning: Cache directory access")

	// Check that suggestions are included
	assert.Contains(t, summary.Suggestions, "Use info, debug, or error")
	assert.Contains(t, summary.Suggestions, "Use alphanumeric characters only")
}

func TestResultStructures(t *testing.T) {
	t.Parallel()

	t.Run("ComprehensiveResult", func(t *testing.T) {
		t.Parallel()

		result := ComprehensiveResult{
			ConfigurationResults: Results{Valid: true},
			ConnectivityResults:  []ConnectivityResult{},
			FileSystemResults:    []FileSystemResult{},
			RepositoryResults:    []RepositoryResult{},
			OverallSuccess:       true,
			TotalErrors:          0,
			TotalWarnings:        0,
			Duration:             time.Second,
			Summary: Summary{
				ConfigurationValid: true,
			},
		}

		assert.True(t, result.OverallSuccess)
		assert.Zero(t, result.TotalErrors)
		assert.Equal(t, time.Second, result.Duration)
		assert.True(t, result.Summary.ConfigurationValid)
	})

	t.Run("Summary", func(t *testing.T) {
		t.Parallel()

		summary := Summary{
			ConfigurationValid:   true,
			ConnectivityValid:    true,
			FileSystemValid:      true,
			RepositoryNamesValid: true,
			CriticalIssues:       []string{"Critical issue"},
			Warnings:             []string{"Warning"},
			Suggestions:          []string{"Suggestion"},
		}

		assert.True(t, summary.ConfigurationValid)
		assert.Contains(t, summary.CriticalIssues, "Critical issue")
		assert.Contains(t, summary.Warnings, "Warning")
		assert.Contains(t, summary.Suggestions, "Suggestion")
	})

	t.Run("RepositoryResult", func(t *testing.T) {
		t.Parallel()

		result := RepositoryResult{
			RepositoryName: "test-repo",
			ProviderType:   "github",
			Results:        []Result{{Valid: true}},
			Valid:          true,
		}

		assert.Equal(t, "test-repo", result.RepositoryName)
		assert.Equal(t, "github", result.ProviderType)
		assert.True(t, result.Valid)
		assert.Len(t, result.Results, 1)
	})

	t.Run("Config", func(t *testing.T) {
		t.Parallel()

		config := Config{
			EnableConnectivityTests: true,
			EnableFileSystemTests:   true,
			ConnectivityTimeout:     time.Second * 30,
			SkipOptionalTests:       false,
			MaxConcurrentTests:      5,
		}

		assert.True(t, config.EnableConnectivityTests)
		assert.True(t, config.EnableFileSystemTests)
		assert.Equal(t, time.Second*30, config.ConnectivityTimeout)
		assert.False(t, config.SkipOptionalTests)
		assert.Equal(t, 5, config.MaxConcurrentTests)
	})
}
