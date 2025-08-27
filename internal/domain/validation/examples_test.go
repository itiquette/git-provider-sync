// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package validation

import (
	"fmt"
	"testing"

	"itiquette/git-provider-sync/internal/domain/ports"

	"github.com/stretchr/testify/require"
)

// TestPureValidationFunctions demonstrates pure validation functions.
func TestPureValidationFunctions(t *testing.T) {
	t.Parallel()

	t.Run("ValidateAuthentication", func(t *testing.T) {
		t.Parallel()

		// Test auth validation directly with existing function
		auth := ports.AuthenticationConfiguration{
			Type:  ports.AuthenticationTypeToken,
			Token: "github_pat_123",
		}
		results := ValidateAuthentication(auth)
		require.True(t, results.Valid)
	})
}

// TestConfigurationValidation demonstrates configuration validation.
func TestConfigurationValidation(t *testing.T) {
	t.Parallel()

	t.Run("ValidConfiguration", func(t *testing.T) {
		t.Parallel()

		config := ports.AppConfiguration{
			GlobalSettings: ports.GlobalSettings{
				LogLevel:     ports.LogLevelInfo,
				LogFormat:    ports.LogFormatJSON,
				MaxCacheSize: 100 * 1024 * 1024,
				CacheTTL:     3600,
			},
			Environments: map[string]ports.EnvironmentConfiguration{
				"dev": {
					Name: "dev",
					Source: ports.SourceConfiguration{
						ProviderType: "github",
						Domain:       "test-github.example.com",
						Owner:        "myorg",
						Authentication: ports.AuthenticationConfiguration{
							Type:  ports.AuthenticationTypeToken,
							Token: "github_pat_123",
						},
					},
					Mirrors: map[string]ports.MirrorConfiguration{
						"backup": {
							Name:         "backup",
							ProviderType: "directory",
							Path:         "/tmp/backup",
							Enabled:      true,
						},
					},
					Enabled: true,
				},
			},
		}

		results := ValidateAppConfiguration(config)
		require.True(t, results.Valid, "Configuration should be valid")
		require.Empty(t, results.Results, "Should have no validation errors")
	})

	t.Run("InvalidConfiguration", func(t *testing.T) {
		t.Parallel()

		config := ports.AppConfiguration{
			GlobalSettings: ports.GlobalSettings{
				LogLevel:     "invalid_level", // Invalid log level
				LogFormat:    ports.LogFormatJSON,
				MaxCacheSize: -100, // Negative cache size
			},
			Environments: map[string]ports.EnvironmentConfiguration{
				"dev": {
					Name: "", // Empty name
					Source: ports.SourceConfiguration{
						ProviderType: "invalid_provider", // Invalid provider
						Owner:        "",                 // Empty owner
					},
					Enabled: true,
				},
			},
		}

		results := ValidateAppConfiguration(config)
		require.False(t, results.Valid, "Configuration should be invalid")
		require.NotEmpty(t, results.Results, "Should have validation errors")

		// Check specific errors
		errorCodes := make(map[string]bool)
		for _, result := range results.Results {
			errorCodes[result.Code] = true
		}

		require.True(t, errorCodes["INVALID_LOG_LEVEL"])
		require.True(t, errorCodes["NEGATIVE_CACHE_SIZE"])
		require.True(t, errorCodes["EMPTY_ENVIRONMENT_NAME"])
		require.True(t, errorCodes["INVALID_PROVIDER_TYPE"])
		require.True(t, errorCodes["INVALID_OWNER"])
	})
}

// TestConnectivityValidationPlanning demonstrates connectivity validation planning.
func TestConnectivityValidationPlanning(t *testing.T) {
	t.Parallel()

	config := ports.AppConfiguration{
		Environments: map[string]ports.EnvironmentConfiguration{
			"prod": {
				Name: "prod",
				Source: ports.SourceConfiguration{
					ProviderType: "github",
					Domain:       "test-github.example.com",
					Owner:        "myorg",
					Authentication: ports.AuthenticationConfiguration{
						Type:  ports.AuthenticationTypeToken,
						Token: "github_pat_123",
					},
				},
				Mirrors: map[string]ports.MirrorConfiguration{
					"gitlab": {
						Name:         "gitlab",
						ProviderType: "gitlab",
						Domain:       "test-gitlab.example.com",
						Owner:        "myorg",
						Enabled:      true,
						Authentication: ports.AuthenticationConfiguration{
							Type:  ports.AuthenticationTypeToken,
							Token: "gitlab_token_456",
						},
					},
				},
				Enabled: true,
			},
		},
	}

	validations := PlanConnectivityValidations(config)
	require.NotEmpty(t, validations, "Should plan connectivity validations")

	// Should have validations for GitHub and GitLab
	hasGitHub := false
	hasGitLab := false

	for _, val := range validations {
		if val.Description != "" {
			if val.Description == "HTTP connectivity to github provider (source for prod)" {
				hasGitHub = true
			}

			if val.Description == "HTTP connectivity to gitlab provider (mirror gitlab for prod)" {
				hasGitLab = true
			}
		}
	}

	require.True(t, hasGitHub, "Should have GitHub connectivity validation")
	require.True(t, hasGitLab, "Should have GitLab connectivity validation")
}

// ExampleValidateAppConfiguration demonstrates the complete validation workflow.
func ExampleValidateAppConfiguration() {
	// Create a simple configuration
	config := ports.AppConfiguration{
		GlobalSettings: ports.GlobalSettings{
			LogLevel:  ports.LogLevelInfo,
			LogFormat: ports.LogFormatJSON,
		},
		Environments: map[string]ports.EnvironmentConfiguration{
			"dev": {
				Name: "dev",
				Source: ports.SourceConfiguration{
					ProviderType: "github",
					Owner:        "myorg",
					Authentication: ports.AuthenticationConfiguration{
						Type:  ports.AuthenticationTypeToken,
						Token: "github_pat_123",
					},
				},
				Enabled: true,
			},
		},
	}

	// Validate configuration using pure functions
	results := ValidateAppConfiguration(config)
	if results.Valid {
		fmt.Println("Configuration is valid!")
	} else {
		fmt.Println("Configuration has errors:")

		for _, result := range results.Results {
			fmt.Println("- " + result.Field + ": " + result.Message)
		}
	}

	// Validate individual repository name
	// Example of using other validation functions
	fmt.Println("Validation complete!")

	// Plan connectivity validations (without executing them)
	connectivityPlans := PlanConnectivityValidations(config)
	fmt.Printf("Planned %d connectivity validations\n", len(connectivityPlans))

	// Output: Configuration is valid!
	// Validation complete!
	// Planned 3 connectivity validations
}
