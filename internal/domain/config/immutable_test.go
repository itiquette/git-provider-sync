// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const modifiedValue = "modified"

func TestImmutableAppConfiguration_Environments(t *testing.T) {
	t.Parallel()

	// Create test environments
	env1 := ImmutableEnvironmentConfiguration{}
	env2 := ImmutableEnvironmentConfiguration{}

	environments := map[string]ImmutableEnvironmentConfiguration{
		"dev":  env1,
		"prod": env2,
	}

	config := ImmutableAppConfiguration{
		environments: environments,
	}

	result := config.Environments()

	// Verify we get a copy, not the original
	assert.Equal(t, environments, result)
	// Maps are always copied in Go, so we can verify by modifying

	// Verify modifying the result doesn't affect the original
	result["test"] = ImmutableEnvironmentConfiguration{}
	assert.NotEqual(t, len(config.environments), len(result))
	assert.Len(t, config.environments, 2)
	assert.Len(t, result, 3)
}

func TestImmutableAppConfiguration_GlobalSettings(t *testing.T) {
	t.Parallel()

	globalSettings := ImmutableGlobalSettings{}
	config := ImmutableAppConfiguration{
		globalSettings: globalSettings,
	}

	result := config.GlobalSettings()
	assert.Equal(t, globalSettings, result)
}

func TestImmutableAppConfiguration_Metadata(t *testing.T) {
	t.Parallel()

	metadata := ImmutableConfigurationMetadata{}
	config := ImmutableAppConfiguration{
		metadata: metadata,
	}

	result := config.Metadata()
	assert.Equal(t, metadata, result)
}

func TestImmutableAppConfiguration_GetEnvironment(t *testing.T) {
	t.Parallel()

	env1 := ImmutableEnvironmentConfiguration{}
	env2 := ImmutableEnvironmentConfiguration{}

	environments := map[string]ImmutableEnvironmentConfiguration{
		"dev":  env1,
		"prod": env2,
	}

	config := ImmutableAppConfiguration{
		environments: environments,
	}

	// Test existing environment
	result, exists := config.GetEnvironment("dev")
	assert.True(t, exists)
	assert.Equal(t, env1, result)

	// Test non-existing environment
	result, exists = config.GetEnvironment("nonexistent")
	assert.False(t, exists)
	assert.Equal(t, ImmutableEnvironmentConfiguration{}, result)
}

func TestImmutableAppConfiguration_WithEnvironment(t *testing.T) {
	t.Parallel()

	originalEnv := ImmutableEnvironmentConfiguration{}
	newEnv := ImmutableEnvironmentConfiguration{}

	config := ImmutableAppConfiguration{
		environments: map[string]ImmutableEnvironmentConfiguration{
			"dev": originalEnv,
		},
	}

	// Add new environment
	newConfig := config.WithEnvironment("prod", newEnv)

	// Original config unchanged
	assert.Len(t, config.environments, 1)
	_, exists := config.GetEnvironment("prod")
	assert.False(t, exists)

	// New config has both environments
	assert.Len(t, newConfig.environments, 2)

	devEnv, exists := newConfig.GetEnvironment("dev")
	assert.True(t, exists)
	assert.Equal(t, originalEnv, devEnv)

	prodEnv, exists := newConfig.GetEnvironment("prod")
	assert.True(t, exists)
	assert.Equal(t, newEnv, prodEnv)

	// Test overwriting existing environment
	updatedEnv := ImmutableEnvironmentConfiguration{}
	updatedConfig := config.WithEnvironment("dev", updatedEnv)

	result, exists := updatedConfig.GetEnvironment("dev")
	assert.True(t, exists)
	assert.Equal(t, updatedEnv, result)
	// Since both are empty structs, they are equal - this tests the replacement works
}

func TestImmutableAppConfiguration_WithoutEnvironment(t *testing.T) {
	t.Parallel()

	env1 := ImmutableEnvironmentConfiguration{}
	env2 := ImmutableEnvironmentConfiguration{}

	config := ImmutableAppConfiguration{
		environments: map[string]ImmutableEnvironmentConfiguration{
			"dev":  env1,
			"prod": env2,
		},
	}

	// Remove existing environment
	newConfig := config.WithoutEnvironment("dev")

	// Original config unchanged
	assert.Len(t, config.environments, 2)

	// New config has one less environment
	assert.Len(t, newConfig.environments, 1)

	_, exists := newConfig.GetEnvironment("dev")
	assert.False(t, exists)

	prodEnv, exists := newConfig.GetEnvironment("prod")
	assert.True(t, exists)
	assert.Equal(t, env2, prodEnv)

	// Test removing non-existing environment (should be no-op)
	sameConfig := config.WithoutEnvironment("nonexistent")
	assert.Equal(t, config.environments, sameConfig.environments)
}

func TestImmutableAppConfiguration_WithGlobalSettings(t *testing.T) {
	t.Parallel()

	originalSettings := ImmutableGlobalSettings{}
	newSettings := ImmutableGlobalSettings{}

	config := ImmutableAppConfiguration{
		globalSettings: originalSettings,
	}

	newConfig := config.WithGlobalSettings(newSettings)

	// Original unchanged
	assert.Equal(t, originalSettings, config.globalSettings)

	// New config has new settings
	assert.Equal(t, newSettings, newConfig.globalSettings)
	// Since both are empty structs, they are equal - this tests the setter works
}

func TestImmutableAppConfiguration_WithMetadata(t *testing.T) {
	t.Parallel()

	originalMetadata := ImmutableConfigurationMetadata{}
	newMetadata := ImmutableConfigurationMetadata{}

	config := ImmutableAppConfiguration{
		metadata: originalMetadata,
	}

	newConfig := config.WithMetadata(newMetadata)

	// Original unchanged
	assert.Equal(t, originalMetadata, config.metadata)

	// New config has new metadata
	assert.Equal(t, newMetadata, newConfig.metadata)
	// Since both are empty structs, they are equal - this tests the setter works
}

func TestImmutableEnvironmentConfiguration_Name(t *testing.T) {
	t.Parallel()

	envName := "test-environment"
	config := ImmutableEnvironmentConfiguration{
		name: envName,
	}

	assert.Equal(t, envName, config.Name())
}

func TestImmutableEnvironmentConfiguration_Source(t *testing.T) {
	t.Parallel()

	source := ImmutableSourceConfiguration{}
	config := ImmutableEnvironmentConfiguration{
		source: source,
	}

	assert.Equal(t, source, config.Source())
}

func TestConfigurationFormat_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, ConfigurationFormatYAML, ConfigurationFormat("yaml"))
	assert.Equal(t, ConfigurationFormatJSON, ConfigurationFormat("json"))
	assert.Equal(t, ConfigurationFormatENV, ConfigurationFormat("env"))
}

func TestImmutableAppConfiguration_EmptyConfig(t *testing.T) {
	t.Parallel()

	config := ImmutableAppConfiguration{}

	environments := config.Environments()
	assert.NotNil(t, environments)
	assert.Empty(t, environments)

	_, exists := config.GetEnvironment("any")
	assert.False(t, exists)
}

func TestImmutableAppConfiguration_Immutability(t *testing.T) {
	t.Parallel()

	// Test that WithEnvironment creates new instances
	config1 := ImmutableAppConfiguration{
		environments: make(map[string]ImmutableEnvironmentConfiguration),
	}

	env := ImmutableEnvironmentConfiguration{}
	config2 := config1.WithEnvironment("test", env)
	config3 := config2.WithEnvironment("another", env)

	// All should be different instances
	assert.NotSame(t, &config1, &config2)
	assert.NotSame(t, &config2, &config3)
	assert.NotSame(t, &config1, &config3)

	// State should be preserved correctly
	assert.Empty(t, config1.environments)
	assert.Len(t, config2.environments, 1)
	assert.Len(t, config3.environments, 2)
}

// Tests for ImmutableEnvironmentConfiguration methods

func TestImmutableEnvironmentConfiguration_Mirrors(t *testing.T) {
	t.Parallel()

	mirror1 := ImmutableMirrorConfiguration{name: "mirror1"}
	mirror2 := ImmutableMirrorConfiguration{name: "mirror2"}

	mirrors := map[string]ImmutableMirrorConfiguration{
		"mirror1": mirror1,
		"mirror2": mirror2,
	}

	config := ImmutableEnvironmentConfiguration{
		mirrors: mirrors,
	}

	result := config.Mirrors()

	// Verify we get a copy
	assert.Equal(t, mirrors, result)
	// Verify modifying the result doesn't affect the original
	result["test"] = ImmutableMirrorConfiguration{}
	assert.NotEqual(t, len(config.mirrors), len(result))
	assert.Len(t, config.mirrors, 2)
	assert.Len(t, result, 3)
}

func TestImmutableEnvironmentConfiguration_Options(t *testing.T) {
	t.Parallel()

	options := ImmutableEnvironmentOptions{dryRun: true, parallel: false}
	config := ImmutableEnvironmentConfiguration{options: options}

	result := config.Options()
	assert.Equal(t, options, result)
}

func TestImmutableEnvironmentConfiguration_Enabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{"enabled true", true, true},
		{"enabled false", false, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			config := ImmutableEnvironmentConfiguration{enabled: test.enabled}
			assert.Equal(t, test.expected, config.Enabled())
		})
	}
}

func TestImmutableEnvironmentConfiguration_WithName(t *testing.T) {
	t.Parallel()

	original := ImmutableEnvironmentConfiguration{name: "original"}
	updated := original.WithName("updated")

	// Original unchanged
	assert.Equal(t, "original", original.name)
	// New has updated name
	assert.Equal(t, "updated", updated.name)
}

func TestImmutableEnvironmentConfiguration_WithSource(t *testing.T) {
	t.Parallel()

	originalSource := ImmutableSourceConfiguration{domain: "original.com"}
	newSource := ImmutableSourceConfiguration{domain: "new.com"}

	config := ImmutableEnvironmentConfiguration{source: originalSource}
	updated := config.WithSource(newSource)

	// Original unchanged
	assert.Equal(t, "original.com", config.source.domain)
	// New has updated source
	assert.Equal(t, "new.com", updated.source.domain)
}

func TestImmutableEnvironmentConfiguration_WithMirror(t *testing.T) {
	t.Parallel()

	originalMirror := ImmutableMirrorConfiguration{name: "original"}
	newMirror := ImmutableMirrorConfiguration{name: "new"}

	config := ImmutableEnvironmentConfiguration{
		mirrors: map[string]ImmutableMirrorConfiguration{
			"mirror1": originalMirror,
		},
	}

	// Add new mirror
	updated := config.WithMirror("mirror2", newMirror)

	// Original unchanged
	assert.Len(t, config.mirrors, 1)
	_, exists := config.mirrors["mirror2"]
	assert.False(t, exists)

	// New has both mirrors
	assert.Len(t, updated.mirrors, 2)
	result, exists := updated.mirrors["mirror2"]
	assert.True(t, exists)
	assert.Equal(t, newMirror, result)

	// Test overwriting existing mirror
	updatedMirror := ImmutableMirrorConfiguration{name: "updated"}
	overwritten := config.WithMirror("mirror1", updatedMirror)
	result, exists = overwritten.mirrors["mirror1"]
	assert.True(t, exists)
	assert.Equal(t, updatedMirror, result)
}

func TestImmutableEnvironmentConfiguration_WithoutMirror(t *testing.T) {
	t.Parallel()

	mirror1 := ImmutableMirrorConfiguration{name: "mirror1"}
	mirror2 := ImmutableMirrorConfiguration{name: "mirror2"}

	config := ImmutableEnvironmentConfiguration{
		mirrors: map[string]ImmutableMirrorConfiguration{
			"mirror1": mirror1,
			"mirror2": mirror2,
		},
	}

	// Remove existing mirror
	updated := config.WithoutMirror("mirror1")

	// Original unchanged
	assert.Len(t, config.mirrors, 2)

	// New has one less mirror
	assert.Len(t, updated.mirrors, 1)
	_, exists := updated.mirrors["mirror1"]
	assert.False(t, exists)
	result, exists := updated.mirrors["mirror2"]
	assert.True(t, exists)
	assert.Equal(t, mirror2, result)

	// Test removing non-existing mirror
	same := config.WithoutMirror("nonexistent")
	assert.Equal(t, config.mirrors, same.mirrors)
}

func TestImmutableEnvironmentConfiguration_WithOptions(t *testing.T) {
	t.Parallel()

	originalOptions := ImmutableEnvironmentOptions{dryRun: false}
	newOptions := ImmutableEnvironmentOptions{dryRun: true}

	config := ImmutableEnvironmentConfiguration{options: originalOptions}
	updated := config.WithOptions(newOptions)

	// Original unchanged
	assert.False(t, config.options.dryRun)
	// New has updated options
	assert.True(t, updated.options.dryRun)
}

func TestImmutableEnvironmentConfiguration_WithEnabled(t *testing.T) {
	t.Parallel()

	config := ImmutableEnvironmentConfiguration{enabled: false}
	updated := config.WithEnabled(true)

	// Original unchanged
	assert.False(t, config.enabled)
	// New has updated enabled status
	assert.True(t, updated.enabled)
}

// Tests for ImmutableSourceConfiguration methods

func TestImmutableSourceConfiguration_AccessorMethods(t *testing.T) {
	t.Parallel()

	auth := ImmutableAuthenticationConfiguration{authType: AuthenticationTypeToken}
	repo := ImmutableRepositoryConfiguration{includeForks: true}
	filtering := ImmutableFilterConfiguration{}
	rateLimit := ImmutableRateLimitConfiguration{requestsPerHour: 100}

	config := ImmutableSourceConfiguration{
		providerType:   "github",
		domain:         "github.com",
		owner:          "testowner",
		authentication: auth,
		repository:     repo,
		filtering:      filtering,
		rateLimit:      rateLimit,
	}

	assert.Equal(t, "github", config.ProviderType())
	assert.Equal(t, "github.com", config.Domain())
	assert.Equal(t, "testowner", config.Owner())
	assert.Equal(t, auth, config.Authentication())
	assert.Equal(t, repo, config.Repository())
	assert.Equal(t, filtering, config.Filtering())
	assert.Equal(t, rateLimit, config.RateLimit())
}

func TestImmutableSourceConfiguration_WithProviderType(t *testing.T) {
	t.Parallel()

	config := ImmutableSourceConfiguration{providerType: "github"}
	updated := config.WithProviderType("gitlab")

	// Original unchanged
	assert.Equal(t, "github", config.providerType)
	// New has updated provider type
	assert.Equal(t, "gitlab", updated.providerType)
}

func TestImmutableSourceConfiguration_WithDomain(t *testing.T) {
	t.Parallel()

	config := ImmutableSourceConfiguration{domain: "github.com"}
	updated := config.WithDomain("gitlab.com")

	// Original unchanged
	assert.Equal(t, "github.com", config.domain)
	// New has updated domain
	assert.Equal(t, "gitlab.com", updated.domain)
}

func TestImmutableSourceConfiguration_WithOwner(t *testing.T) {
	t.Parallel()

	config := ImmutableSourceConfiguration{owner: "oldowner"}
	updated := config.WithOwner("newowner")

	// Original unchanged
	assert.Equal(t, "oldowner", config.owner)
	// New has updated owner
	assert.Equal(t, "newowner", updated.owner)
}

func TestImmutableSourceConfiguration_WithAuthentication(t *testing.T) {
	t.Parallel()

	originalAuth := ImmutableAuthenticationConfiguration{authType: AuthenticationTypeToken}
	newAuth := ImmutableAuthenticationConfiguration{authType: AuthenticationTypeBasic}

	config := ImmutableSourceConfiguration{authentication: originalAuth}
	updated := config.WithAuthentication(newAuth)

	// Original unchanged
	assert.Equal(t, AuthenticationTypeToken, config.authentication.authType)
	// New has updated authentication
	assert.Equal(t, AuthenticationTypeBasic, updated.authentication.authType)
}

func TestImmutableSourceConfiguration_WithRepository(t *testing.T) {
	t.Parallel()

	originalRepo := ImmutableRepositoryConfiguration{includeForks: false}
	newRepo := ImmutableRepositoryConfiguration{includeForks: true}

	config := ImmutableSourceConfiguration{repository: originalRepo}
	updated := config.WithRepository(newRepo)

	// Original unchanged
	assert.False(t, config.repository.includeForks)
	// New has updated repository
	assert.True(t, updated.repository.includeForks)
}

// Tests for ImmutableAuthenticationConfiguration methods

func TestImmutableAuthenticationConfiguration_AccessorMethods(t *testing.T) {
	t.Parallel()

	config := ImmutableAuthenticationConfiguration{
		authType:   AuthenticationTypeToken,
		token:      "test-token",
		username:   "test-user",
		password:   "test-pass",
		sshKeyPath: "/path/to/key",
		sshKey:     "ssh-key-content",
		passphrase: "test-phrase",
	}

	assert.Equal(t, AuthenticationTypeToken, config.Type())
	assert.Equal(t, "test-token", config.Token())
	assert.Equal(t, "test-user", config.Username())
	assert.Equal(t, "test-pass", config.Password())
	assert.Equal(t, "/path/to/key", config.SSHKeyPath())
	assert.Equal(t, "ssh-key-content", config.SSHKey())
	assert.Equal(t, "test-phrase", config.Passphrase())
}

func TestImmutableAuthenticationConfiguration_WithToken(t *testing.T) {
	t.Parallel()

	config := ImmutableAuthenticationConfiguration{token: "old-token"}
	updated := config.WithToken("new-token")

	// Original unchanged
	assert.Equal(t, "old-token", config.token)
	// New has updated token
	assert.Equal(t, "new-token", updated.token)
}

func TestImmutableAuthenticationConfiguration_WithUsername(t *testing.T) {
	t.Parallel()

	config := ImmutableAuthenticationConfiguration{username: "olduser"}
	updated := config.WithUsername("newuser")

	// Original unchanged
	assert.Equal(t, "olduser", config.username)
	// New has updated username
	assert.Equal(t, "newuser", updated.username)
}

func TestImmutableAuthenticationConfiguration_WithType(t *testing.T) {
	t.Parallel()

	config := ImmutableAuthenticationConfiguration{authType: AuthenticationTypeToken}
	updated := config.WithType(AuthenticationTypeSSH)

	// Original unchanged
	assert.Equal(t, AuthenticationTypeToken, config.authType)
	// New has updated type
	assert.Equal(t, AuthenticationTypeSSH, updated.authType)
}

func TestImmutableAuthenticationConfiguration_WithPassword(t *testing.T) {
	t.Parallel()

	config := ImmutableAuthenticationConfiguration{password: "oldpass"}
	updated := config.WithPassword("newpass")

	// Original unchanged
	assert.Equal(t, "oldpass", config.password)
	// New has updated password
	assert.Equal(t, "newpass", updated.password)
}

func TestImmutableAuthenticationConfiguration_WithSSHKeyPath(t *testing.T) {
	t.Parallel()

	config := ImmutableAuthenticationConfiguration{sshKeyPath: "/old/path"}
	updated := config.WithSSHKeyPath("/new/path")

	// Original unchanged
	assert.Equal(t, "/old/path", config.sshKeyPath)
	// New has updated SSH key path
	assert.Equal(t, "/new/path", updated.sshKeyPath)
}

func TestImmutableAuthenticationConfiguration_WithSSHKey(t *testing.T) {
	t.Parallel()

	config := ImmutableAuthenticationConfiguration{sshKey: "old-key"}
	updated := config.WithSSHKey("new-key")

	// Original unchanged
	assert.Equal(t, "old-key", config.sshKey)
	// New has updated SSH key
	assert.Equal(t, "new-key", updated.sshKey)
}

// Tests for ImmutableGlobalSettings methods

func TestImmutableGlobalSettings_AccessorMethods(t *testing.T) {
	t.Parallel()

	config := ImmutableGlobalSettings{
		logLevel:       LogLevelDebug,
		logFormat:      LogFormatJSON,
		logFile:        "/var/log/app.log",
		tempDirectory:  "/tmp",
		metricsEnabled: true,
		metricsPort:    8080,
	}

	assert.Equal(t, LogLevelDebug, config.LogLevel())
	assert.Equal(t, LogFormatJSON, config.LogFormat())
	assert.Equal(t, "/var/log/app.log", config.LogFile())
	assert.Equal(t, "/tmp", config.TempDirectory())
	assert.True(t, config.MetricsEnabled())
	assert.Equal(t, 8080, config.MetricsPort())
}

func TestImmutableGlobalSettings_WithLogLevel(t *testing.T) {
	t.Parallel()

	config := ImmutableGlobalSettings{logLevel: LogLevelInfo}
	updated := config.WithLogLevel(LogLevelDebug)

	// Original unchanged
	assert.Equal(t, LogLevelInfo, config.logLevel)
	// New has updated log level
	assert.Equal(t, LogLevelDebug, updated.logLevel)
}

func TestImmutableGlobalSettings_WithLogFormat(t *testing.T) {
	t.Parallel()

	config := ImmutableGlobalSettings{logFormat: LogFormatJSON}
	updated := config.WithLogFormat(LogFormatText)

	// Original unchanged
	assert.Equal(t, LogFormatJSON, config.logFormat)
	// New has updated log format
	assert.Equal(t, LogFormatText, updated.logFormat)
}

func TestImmutableGlobalSettings_WithTempDirectory(t *testing.T) {
	t.Parallel()

	config := ImmutableGlobalSettings{tempDirectory: "/tmp"}
	updated := config.WithTempDirectory("/var/tmp")

	// Original unchanged
	assert.Equal(t, "/tmp", config.tempDirectory)
	// New has updated temp directory
	assert.Equal(t, "/var/tmp", updated.tempDirectory)
}

func TestImmutableGlobalSettings_WithMetricsEnabled(t *testing.T) {
	t.Parallel()

	config := ImmutableGlobalSettings{metricsEnabled: false}
	updated := config.WithMetricsEnabled(true)

	// Original unchanged
	assert.False(t, config.metricsEnabled)
	// New has updated metrics enabled
	assert.True(t, updated.metricsEnabled)
}

func TestImmutableGlobalSettings_WithMetricsPort(t *testing.T) {
	t.Parallel()

	config := ImmutableGlobalSettings{metricsPort: 8080}
	updated := config.WithMetricsPort(9090)

	// Original unchanged
	assert.Equal(t, 8080, config.metricsPort)
	// New has updated metrics port
	assert.Equal(t, 9090, updated.metricsPort)
}

// Tests for ImmutableMirrorConfiguration methods

func TestImmutableMirrorConfiguration_AccessorMethods(t *testing.T) {
	t.Parallel()

	auth := ImmutableAuthenticationConfiguration{authType: AuthenticationTypeToken}
	options := ImmutableMirrorOptionsConfiguration{createIfNotExists: true}

	config := ImmutableMirrorConfiguration{
		name:           "test-mirror",
		providerType:   "gitlab",
		domain:         "gitlab.com",
		owner:          "testowner",
		path:           "/mirrors",
		authentication: auth,
		options:        options,
		enabled:        true,
	}

	assert.Equal(t, "test-mirror", config.Name())
	assert.Equal(t, "gitlab", config.ProviderType())
	assert.Equal(t, "gitlab.com", config.Domain())
	assert.Equal(t, "testowner", config.Owner())
	assert.Equal(t, "/mirrors", config.Path())
	assert.Equal(t, auth, config.Authentication())
	assert.Equal(t, options, config.Options())
	assert.True(t, config.Enabled())
}

func TestImmutableMirrorConfiguration_WithProviderType(t *testing.T) {
	t.Parallel()

	config := ImmutableMirrorConfiguration{providerType: "github"}
	updated := config.WithProviderType("gitlab")

	// Original unchanged
	assert.Equal(t, "github", config.providerType)
	// New has updated provider type
	assert.Equal(t, "gitlab", updated.providerType)
}

func TestImmutableMirrorConfiguration_WithDomain(t *testing.T) {
	t.Parallel()

	config := ImmutableMirrorConfiguration{domain: "github.com"}
	updated := config.WithDomain("gitlab.com")

	// Original unchanged
	assert.Equal(t, "github.com", config.domain)
	// New has updated domain
	assert.Equal(t, "gitlab.com", updated.domain)
}

func TestImmutableMirrorConfiguration_WithOwner(t *testing.T) {
	t.Parallel()

	config := ImmutableMirrorConfiguration{owner: "oldowner"}
	updated := config.WithOwner("newowner")

	// Original unchanged
	assert.Equal(t, "oldowner", config.owner)
	// New has updated owner
	assert.Equal(t, "newowner", updated.owner)
}

func TestImmutableMirrorConfiguration_WithPath(t *testing.T) {
	t.Parallel()

	config := ImmutableMirrorConfiguration{path: "/old/path"}
	updated := config.WithPath("/new/path")

	// Original unchanged
	assert.Equal(t, "/old/path", config.path)
	// New has updated path
	assert.Equal(t, "/new/path", updated.path)
}

func TestImmutableMirrorConfiguration_WithAuthentication(t *testing.T) {
	t.Parallel()

	originalAuth := ImmutableAuthenticationConfiguration{authType: AuthenticationTypeToken}
	newAuth := ImmutableAuthenticationConfiguration{authType: AuthenticationTypeSSH}

	config := ImmutableMirrorConfiguration{authentication: originalAuth}
	updated := config.WithAuthentication(newAuth)

	// Original unchanged
	assert.Equal(t, AuthenticationTypeToken, config.authentication.authType)
	// New has updated authentication
	assert.Equal(t, AuthenticationTypeSSH, updated.authentication.authType)
}

func TestImmutableMirrorConfiguration_WithOptions(t *testing.T) {
	t.Parallel()

	originalOptions := ImmutableMirrorOptionsConfiguration{createIfNotExists: false}
	newOptions := ImmutableMirrorOptionsConfiguration{createIfNotExists: true}

	config := ImmutableMirrorConfiguration{options: originalOptions}
	updated := config.WithOptions(newOptions)

	// Original unchanged
	assert.False(t, config.options.createIfNotExists)
	// New has updated options
	assert.True(t, updated.options.createIfNotExists)
}

func TestImmutableMirrorConfiguration_WithEnabled(t *testing.T) {
	t.Parallel()

	config := ImmutableMirrorConfiguration{enabled: false}
	updated := config.WithEnabled(true)

	// Original unchanged
	assert.False(t, config.enabled)
	// New has updated enabled status
	assert.True(t, updated.enabled)
}

// Tests for ImmutableMirrorOptionsConfiguration methods

func TestImmutableMirrorOptionsConfiguration_WithCreateIfNotExists(t *testing.T) {
	t.Parallel()

	config := ImmutableMirrorOptionsConfiguration{createIfNotExists: false}
	updated := config.WithCreateIfNotExists(true)

	// Original unchanged
	assert.False(t, config.createIfNotExists)
	// New has updated createIfNotExists
	assert.True(t, updated.createIfNotExists)
}

func TestImmutableMirrorOptionsConfiguration_WithUpdateDescription(t *testing.T) {
	t.Parallel()

	config := ImmutableMirrorOptionsConfiguration{updateDescription: false}
	updated := config.WithUpdateDescription(true)

	// Original unchanged
	assert.False(t, config.updateDescription)
	// New has updated updateDescription
	assert.True(t, updated.updateDescription)
}

func TestImmutableMirrorOptionsConfiguration_WithSyncVisibility(t *testing.T) {
	t.Parallel()

	config := ImmutableMirrorOptionsConfiguration{syncVisibility: false}
	updated := config.WithSyncVisibility(true)

	// Original unchanged
	assert.False(t, config.syncVisibility)
	// New has updated syncVisibility
	assert.True(t, updated.syncVisibility)
}

func TestImmutableMirrorOptionsConfiguration_WithSyncTopics(t *testing.T) {
	t.Parallel()

	config := ImmutableMirrorOptionsConfiguration{syncTopics: false}
	updated := config.WithSyncTopics(true)

	// Original unchanged
	assert.False(t, config.syncTopics)
	// New has updated syncTopics
	assert.True(t, updated.syncTopics)
}

func TestImmutableMirrorOptions_BranchProtection_ConfiguresCorrectly(t *testing.T) {
	t.Parallel()

	config := ImmutableMirrorOptionsConfiguration{syncBranchProtection: false}
	updated := config.WithSyncBranchProtection(true)

	// Original unchanged
	assert.False(t, config.syncBranchProtection)
	// New has updated syncBranchProtection
	assert.True(t, updated.syncBranchProtection)
}

func TestImmutableMirrorOptionsConfiguration_WithEnableLFS(t *testing.T) {
	t.Parallel()

	config := ImmutableMirrorOptionsConfiguration{enableLFS: false}
	updated := config.WithEnableLFS(true)

	// Original unchanged
	assert.False(t, config.enableLFS)
	// New has updated enableLFS
	assert.True(t, updated.enableLFS)
}

// Tests for ImmutableRepositoryConfiguration methods

func TestImmutableRepositoryConfiguration_AccessorMethods(t *testing.T) {
	t.Parallel()

	includePatterns := []string{"include1", "include2"}
	excludePatterns := []string{"exclude1", "exclude2"}

	config := ImmutableRepositoryConfiguration{
		includePatterns: includePatterns,
		excludePatterns: excludePatterns,
		includeForks:    true,
		includeArchived: false,
	}

	resultInclude := config.IncludePatterns()
	resultExclude := config.ExcludePatterns()

	// Verify we get copies
	assert.Equal(t, includePatterns, resultInclude)
	assert.Equal(t, excludePatterns, resultExclude)

	// Verify modifying results doesn't affect originals
	resultInclude[0] = modifiedValue
	resultExclude[0] = modifiedValue

	assert.Equal(t, "include1", config.includePatterns[0])
	assert.Equal(t, "exclude1", config.excludePatterns[0])

	assert.True(t, config.IncludeForks())
	assert.False(t, config.IncludeArchived())
}

func TestImmutableRepositoryConfiguration_WithIncludePatterns(t *testing.T) {
	t.Parallel()

	original := []string{"old1", "old2"}
	newPatterns := []string{"new1", "new2", "new3"}

	config := ImmutableRepositoryConfiguration{includePatterns: original}
	updated := config.WithIncludePatterns(newPatterns)

	// Original unchanged
	assert.Equal(t, original, config.includePatterns)
	// New has updated patterns
	assert.Equal(t, newPatterns, updated.includePatterns)

	// Verify independence
	newPatterns[0] = modifiedValue

	assert.Equal(t, "new1", updated.includePatterns[0])
}

func TestImmutableRepositoryConfiguration_WithExcludePatterns(t *testing.T) {
	t.Parallel()

	original := []string{"old1", "old2"}
	newPatterns := []string{"new1", "new2", "new3"}

	config := ImmutableRepositoryConfiguration{excludePatterns: original}
	updated := config.WithExcludePatterns(newPatterns)

	// Original unchanged
	assert.Equal(t, original, config.excludePatterns)
	// New has updated patterns
	assert.Equal(t, newPatterns, updated.excludePatterns)

	// Verify independence
	newPatterns[0] = modifiedValue

	assert.Equal(t, "new1", updated.excludePatterns[0])
}

func TestImmutableRepositoryConfiguration_WithIncludeForks(t *testing.T) {
	t.Parallel()

	config := ImmutableRepositoryConfiguration{includeForks: false}
	updated := config.WithIncludeForks(true)

	// Original unchanged
	assert.False(t, config.includeForks)
	// New has updated includeForks
	assert.True(t, updated.includeForks)
}

func TestImmutableRepositoryConfiguration_WithIncludeArchived(t *testing.T) {
	t.Parallel()

	config := ImmutableRepositoryConfiguration{includeArchived: false}
	updated := config.WithIncludeArchived(true)

	// Original unchanged
	assert.False(t, config.includeArchived)
	// New has updated includeArchived
	assert.True(t, updated.includeArchived)
}

// Tests for ImmutableEnvironmentOptions methods

func TestImmutableEnvironmentOptions_AccessorMethods(t *testing.T) {
	t.Parallel()

	config := ImmutableEnvironmentOptions{
		dryRun:         true,
		parallel:       false,
		maxConcurrency: 5,
		timeout:        time.Minute * 30,
		retryAttempts:  3,
		retryDelay:     time.Second * 5,
	}

	assert.True(t, config.DryRun())
	assert.False(t, config.Parallel())
	assert.Equal(t, 5, config.MaxConcurrency())
	assert.Equal(t, time.Minute*30, config.Timeout())
	assert.Equal(t, 3, config.RetryAttempts())
	assert.Equal(t, time.Second*5, config.RetryDelay())
}

func TestImmutableEnvironmentOptions_WithDryRun(t *testing.T) {
	t.Parallel()

	config := ImmutableEnvironmentOptions{dryRun: false}
	updated := config.WithDryRun(true)

	// Original unchanged
	assert.False(t, config.dryRun)
	// New has updated dryRun
	assert.True(t, updated.dryRun)
}

func TestImmutableEnvironmentOptions_WithParallel(t *testing.T) {
	t.Parallel()

	config := ImmutableEnvironmentOptions{parallel: false}
	updated := config.WithParallel(true)

	// Original unchanged
	assert.False(t, config.parallel)
	// New has updated parallel
	assert.True(t, updated.parallel)
}

func TestImmutableEnvironmentOptions_WithMaxConcurrency(t *testing.T) {
	t.Parallel()

	config := ImmutableEnvironmentOptions{maxConcurrency: 5}
	updated := config.WithMaxConcurrency(10)

	// Original unchanged
	assert.Equal(t, 5, config.maxConcurrency)
	// New has updated maxConcurrency
	assert.Equal(t, 10, updated.maxConcurrency)
}

func TestImmutableEnvironmentOptions_WithTimeout(t *testing.T) {
	t.Parallel()

	config := ImmutableEnvironmentOptions{timeout: time.Minute * 30}
	updated := config.WithTimeout(time.Hour)

	// Original unchanged
	assert.Equal(t, time.Minute*30, config.timeout)
	// New has updated timeout
	assert.Equal(t, time.Hour, updated.timeout)
}

func TestImmutableEnvironmentOptions_WithRetryAttempts(t *testing.T) {
	t.Parallel()

	config := ImmutableEnvironmentOptions{retryAttempts: 3}
	updated := config.WithRetryAttempts(5)

	// Original unchanged
	assert.Equal(t, 3, config.retryAttempts)
	// New has updated retryAttempts
	assert.Equal(t, 5, updated.retryAttempts)
}

func TestImmutableEnvironmentOptions_WithRetryDelay(t *testing.T) {
	t.Parallel()

	config := ImmutableEnvironmentOptions{retryDelay: time.Second * 5}
	updated := config.WithRetryDelay(time.Second * 10)

	// Original unchanged
	assert.Equal(t, time.Second*5, config.retryDelay)
	// New has updated retryDelay
	assert.Equal(t, time.Second*10, updated.retryDelay)
}

// Tests for AppConfigurationBuilder

func TestNewAppConfigurationBuilder(t *testing.T) {
	t.Parallel()

	builder := NewAppConfigurationBuilder()

	assert.NotNil(t, builder)
	assert.NotNil(t, builder.config.environments)
	assert.Empty(t, builder.config.environments)
	assert.Equal(t, LogLevelInfo, builder.config.globalSettings.logLevel)
	assert.Equal(t, LogFormatJSON, builder.config.globalSettings.logFormat)
	assert.False(t, builder.config.metadata.loadTime.IsZero())
}

func TestAppConfigurationBuilder_WithEnvironment(t *testing.T) {
	t.Parallel()

	builder := NewAppConfigurationBuilder()
	env := ImmutableEnvironmentConfiguration{name: "test"}

	result := builder.WithEnvironment("test", env)

	// Should return the same builder instance
	assert.Same(t, builder, result)

	// Environment should be added
	config := builder.Build()
	_, exists := config.GetEnvironment("test")
	assert.True(t, exists)
}

func TestAppConfigurationBuilder_WithGlobalSettings(t *testing.T) {
	t.Parallel()

	builder := NewAppConfigurationBuilder()
	settings := ImmutableGlobalSettings{logLevel: LogLevelDebug}

	result := builder.WithGlobalSettings(settings)

	// Should return the same builder instance
	assert.Same(t, builder, result)

	// Global settings should be updated
	config := builder.Build()
	assert.Equal(t, LogLevelDebug, config.GlobalSettings().LogLevel())
}

func TestAppConfigurationBuilder_Build(t *testing.T) {
	t.Parallel()

	builder := NewAppConfigurationBuilder()
	env := ImmutableEnvironmentConfiguration{name: "test"}
	settings := ImmutableGlobalSettings{logLevel: LogLevelDebug}

	config := builder.WithEnvironment("test", env).WithGlobalSettings(settings).Build()

	// Should have the environment and settings
	_, exists := config.GetEnvironment("test")
	assert.True(t, exists)
	assert.Equal(t, LogLevelDebug, config.GlobalSettings().LogLevel())
}

// Tests for enum constants

func TestAuthenticationTypeConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, AuthenticationTypeNone, AuthenticationType("none"))
	assert.Equal(t, AuthenticationTypeToken, AuthenticationType("token"))
	assert.Equal(t, AuthenticationTypeBasic, AuthenticationType("basic"))
	assert.Equal(t, AuthenticationTypeSSH, AuthenticationType("ssh"))
	assert.Equal(t, AuthenticationTypeOAuth, AuthenticationType("oauth"))
}

func TestBackoffStrategyConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, BackoffStrategyLinear, BackoffStrategy("linear"))
	assert.Equal(t, BackoffStrategyExponential, BackoffStrategy("exponential"))
	assert.Equal(t, BackoffStrategyFixed, BackoffStrategy("fixed"))
}

func TestLogLevelConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, LogLevelTrace, LogLevel("trace"))
	assert.Equal(t, LogLevelDebug, LogLevel("debug"))
	assert.Equal(t, LogLevelInfo, LogLevel("info"))
	assert.Equal(t, LogLevelWarn, LogLevel("warn"))
	assert.Equal(t, LogLevelError, LogLevel("error"))
	assert.Equal(t, LogLevelFatal, LogLevel("fatal"))
}

func TestLogFormatConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, LogFormatJSON, LogFormat("json"))
	assert.Equal(t, LogFormatText, LogFormat("text"))
	assert.Equal(t, LogFormatConsole, LogFormat("console"))
}

func TestSourceTypeConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, SourceTypeFile, SourceType("file"))
	assert.Equal(t, SourceTypeEnvironment, SourceType("environment"))
	assert.Equal(t, SourceTypeEtcd, SourceType("etcd"))
	assert.Equal(t, SourceTypeConsul, SourceType("consul"))
	assert.Equal(t, SourceTypeVault, SourceType("vault"))
	assert.Equal(t, SourceTypeHTTP, SourceType("http"))
	assert.Equal(t, SourceTypeDefaults, SourceType("defaults"))
}

func TestConfigurationFormatConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, ConfigurationFormatYAML, ConfigurationFormat("yaml"))
	assert.Equal(t, ConfigurationFormatTOML, ConfigurationFormat("toml"))
	assert.Equal(t, ConfigurationFormatINI, ConfigurationFormat("ini"))
}

// Additional edge case tests

func TestImmutableAppConfiguration_GetEnvironment_Empty(t *testing.T) {
	t.Parallel()

	config := ImmutableAppConfiguration{}

	env, exists := config.GetEnvironment("nonexistent")
	assert.False(t, exists)
	assert.Equal(t, ImmutableEnvironmentConfiguration{}, env)
}

func TestImmutableEnvironmentConfiguration_EmptyMirrors(t *testing.T) {
	t.Parallel()

	config := ImmutableEnvironmentConfiguration{}

	mirrors := config.Mirrors()
	assert.NotNil(t, mirrors)
	assert.Empty(t, mirrors)
}

func TestImmutableRepositoryConfiguration_EmptySlices(t *testing.T) {
	t.Parallel()

	config := ImmutableRepositoryConfiguration{}

	includePatterns := config.IncludePatterns()
	excludePatterns := config.ExcludePatterns()

	// When underlying slices are nil, the copy returns nil
	assert.Nil(t, includePatterns)
	assert.Nil(t, excludePatterns)
	assert.Empty(t, includePatterns)
	assert.Empty(t, excludePatterns)
}
