// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test the important With/Without operations that modify state.
func TestImmutableAppConfiguration_WithEnvironment(t *testing.T) {
	t.Parallel()

	originalEnv := ImmutableEnvironmentConfiguration{name: "original"}
	newEnv := ImmutableEnvironmentConfiguration{name: "new"}

	config := ImmutableAppConfiguration{
		environments: map[string]ImmutableEnvironmentConfiguration{
			"dev": originalEnv,
		},
	}

	// Add new environment
	updated := config.WithEnvironment("prod", newEnv)

	// Original should be unchanged
	assert.Len(t, config.environments, 1)
	assert.Contains(t, config.environments, "dev")
	assert.NotContains(t, config.environments, "prod")

	// Updated should have both
	assert.Len(t, updated.environments, 2)
	assert.Contains(t, updated.environments, "dev")
	assert.Contains(t, updated.environments, "prod")
	assert.Equal(t, newEnv, updated.environments["prod"])

	// Overwrite existing environment
	replacementEnv := ImmutableEnvironmentConfiguration{name: "replacement"}
	replaced := config.WithEnvironment("dev", replacementEnv)

	assert.Len(t, replaced.environments, 1)
	assert.Equal(t, replacementEnv, replaced.environments["dev"])
	// Original remains unchanged
	assert.Equal(t, originalEnv, config.environments["dev"])
}

func TestImmutableAppConfiguration_WithoutEnvironment(t *testing.T) {
	t.Parallel()

	env1 := ImmutableEnvironmentConfiguration{name: "env1"}
	env2 := ImmutableEnvironmentConfiguration{name: "env2"}

	config := ImmutableAppConfiguration{
		environments: map[string]ImmutableEnvironmentConfiguration{
			"dev":  env1,
			"prod": env2,
		},
	}

	// Remove existing environment
	updated := config.WithoutEnvironment("dev")

	// Original unchanged
	assert.Len(t, config.environments, 2)
	assert.Contains(t, config.environments, "dev")

	// Updated has one removed
	assert.Len(t, updated.environments, 1)
	assert.NotContains(t, updated.environments, "dev")
	assert.Contains(t, updated.environments, "prod")

	// Remove non-existent environment (should be safe)
	unchanged := config.WithoutEnvironment("nonexistent")
	assert.Len(t, unchanged.environments, 2)
}

func TestImmutableAppConfiguration_GetEnvironment(t *testing.T) {
	t.Parallel()

	env1 := ImmutableEnvironmentConfiguration{name: "env1"}
	config := ImmutableAppConfiguration{
		environments: map[string]ImmutableEnvironmentConfiguration{
			"dev": env1,
		},
	}

	// Get existing
	result, exists := config.GetEnvironment("dev")
	assert.True(t, exists)
	assert.Equal(t, env1, result)

	// Get non-existing
	result, exists = config.GetEnvironment("nonexistent")
	assert.False(t, exists)
	assert.Equal(t, ImmutableEnvironmentConfiguration{}, result)
}

func TestImmutableEnvironmentConfiguration_WithMirror(t *testing.T) {
	t.Parallel()

	mirror1 := ImmutableMirrorConfiguration{name: "mirror1"}
	mirror2 := ImmutableMirrorConfiguration{name: "mirror2"}

	env := ImmutableEnvironmentConfiguration{
		mirrors: map[string]ImmutableMirrorConfiguration{
			"backup1": mirror1,
		},
	}

	// Add new mirror
	updated := env.WithMirror("backup2", mirror2)

	// Original unchanged
	assert.Len(t, env.mirrors, 1)
	assert.NotContains(t, env.mirrors, "backup2")

	// Updated has both
	assert.Len(t, updated.mirrors, 2)
	assert.Contains(t, updated.mirrors, "backup1")
	assert.Contains(t, updated.mirrors, "backup2")
}

func TestImmutableEnvironmentConfiguration_WithoutMirror(t *testing.T) {
	t.Parallel()

	mirror1 := ImmutableMirrorConfiguration{name: "mirror1"}
	mirror2 := ImmutableMirrorConfiguration{name: "mirror2"}

	env := ImmutableEnvironmentConfiguration{
		mirrors: map[string]ImmutableMirrorConfiguration{
			"backup1": mirror1,
			"backup2": mirror2,
		},
	}

	updated := env.WithoutMirror("backup1")

	// Original unchanged
	assert.Len(t, env.mirrors, 2)

	// Updated has one removed
	assert.Len(t, updated.mirrors, 1)
	assert.NotContains(t, updated.mirrors, "backup1")
	assert.Contains(t, updated.mirrors, "backup2")
}

// Test complex operations with multiple changes.
func TestImmutableConfiguration_ChainedOperations(t *testing.T) {
	t.Parallel()

	config := ImmutableAppConfiguration{}

	env1 := ImmutableEnvironmentConfiguration{name: "env1"}
	env2 := ImmutableEnvironmentConfiguration{name: "env2"}
	settings := ImmutableGlobalSettings{logLevel: LogLevelInfo}

	// Chain multiple operations
	result := config.
		WithEnvironment("dev", env1).
		WithEnvironment("prod", env2).
		WithGlobalSettings(settings)

	assert.Len(t, result.environments, 2)
	assert.Equal(t, settings, result.globalSettings)

	// Remove and re-add
	modified := result.
		WithoutEnvironment("dev").
		WithEnvironment("staging", env1)

	assert.Len(t, modified.environments, 2)
	assert.Contains(t, modified.environments, "prod")
	assert.Contains(t, modified.environments, "staging")
	assert.NotContains(t, modified.environments, "dev")
}

// Test builder pattern.
func TestAppConfigurationBuilder(t *testing.T) {
	t.Parallel()

	builder := NewAppConfigurationBuilder()
	require.NotNil(t, builder)

	env := ImmutableEnvironmentConfiguration{name: "test"}
	settings := ImmutableGlobalSettings{logLevel: LogLevelDebug}

	config := builder.
		WithEnvironment("dev", env).
		WithGlobalSettings(settings).
		Build()

	assert.Len(t, config.environments, 1)
	assert.Equal(t, settings, config.globalSettings)

	// Test adding multiple environments
	builder2 := NewAppConfigurationBuilder()
	config2 := builder2.
		WithEnvironment("prod", env).
		WithEnvironment("staging", env).
		Build()

	assert.Len(t, config2.environments, 2)
	assert.Contains(t, config2.environments, "prod")
	assert.Contains(t, config2.environments, "staging")
}

// Test edge cases.
func TestImmutableConfiguration_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("empty configuration", func(t *testing.T) {
		t.Parallel()

		config := ImmutableAppConfiguration{}
		assert.Empty(t, config.Environments())
		assert.Equal(t, ImmutableGlobalSettings{}, config.GlobalSettings())
		assert.Equal(t, ImmutableConfigurationMetadata{}, config.Metadata())

		_, exists := config.GetEnvironment("any")
		assert.False(t, exists)
	})

	t.Run("nil maps initialization", func(t *testing.T) {
		t.Parallel()

		env := ImmutableEnvironmentConfiguration{}
		updated := env.WithMirror("test", ImmutableMirrorConfiguration{})
		assert.Len(t, updated.mirrors, 1)
	})

	t.Run("removing from empty", func(t *testing.T) {
		t.Parallel()

		config := ImmutableAppConfiguration{}
		updated := config.WithoutEnvironment("nonexistent")
		assert.Empty(t, updated.environments)
	})
}

// Test that functional updates create proper copies.
func TestImmutableConfiguration_DeepCopy(t *testing.T) {
	t.Parallel()

	source := ImmutableSourceConfiguration{
		providerType: "github",
		domain:       "github.com",
	}

	env := ImmutableEnvironmentConfiguration{
		name:   "dev",
		source: source,
	}

	config := ImmutableAppConfiguration{
		environments: map[string]ImmutableEnvironmentConfiguration{
			"dev": env,
		},
	}

	// Modify through WithEnvironment
	modifiedEnv := env.WithName("modified")
	updated := config.WithEnvironment("dev", modifiedEnv)

	// Check original is unchanged
	assert.Equal(t, "dev", config.environments["dev"].name)
	assert.Equal(t, "modified", updated.environments["dev"].name)
}

// Test authentication configuration updates.
func TestImmutableAuthenticationConfiguration_Updates(t *testing.T) {
	t.Parallel()

	auth := ImmutableAuthenticationConfiguration{
		authType: AuthenticationTypeToken,
		token:    "original-token",
	}

	// Update token
	updated := auth.WithToken("new-token")
	assert.Equal(t, "original-token", auth.token)
	assert.Equal(t, "new-token", updated.token)
	assert.Equal(t, AuthenticationTypeToken, updated.authType)

	// Change auth type
	basic := auth.
		WithType(AuthenticationTypeBasic).
		WithUsername("user").
		WithPassword("pass")

	assert.Equal(t, AuthenticationTypeBasic, basic.authType)
	assert.Equal(t, "user", basic.username)
	assert.Equal(t, "pass", basic.password)
}

// Test repository configuration with patterns.
func TestImmutableRepositoryConfiguration_Patterns(t *testing.T) {
	t.Parallel()

	repo := ImmutableRepositoryConfiguration{
		includePatterns: []string{"*.go"},
		excludePatterns: []string{"*_test.go"},
	}

	// Add patterns
	updated := repo.
		WithIncludePatterns([]string{"*.go", "*.md"}).
		WithExcludePatterns([]string{"*_test.go", "vendor/*"})

	// Original unchanged
	assert.Len(t, repo.includePatterns, 1)
	assert.Len(t, repo.excludePatterns, 1)

	// Updated has new patterns
	assert.Equal(t, []string{"*.go", "*.md"}, updated.includePatterns)
	assert.Equal(t, []string{"*_test.go", "vendor/*"}, updated.excludePatterns)
}

// Test environment options.
func TestImmutableEnvironmentOptions_Updates(t *testing.T) {
	t.Parallel()

	options := ImmutableEnvironmentOptions{
		dryRun:         false,
		maxConcurrency: 5,
	}

	updated := options.
		WithDryRun(true).
		WithMaxConcurrency(10).
		WithTimeout(30 * time.Second)

	// Original unchanged
	assert.False(t, options.dryRun)
	assert.Equal(t, 5, options.maxConcurrency)

	// Updated has new values
	assert.True(t, updated.dryRun)
	assert.Equal(t, 10, updated.maxConcurrency)
	assert.Equal(t, 30*time.Second, updated.timeout)
}
