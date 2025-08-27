// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Tests for App Configuration Options

func TestWithEnvironment(t *testing.T) {
	t.Parallel()

	config := ImmutableAppConfiguration{
		environments: make(map[string]ImmutableEnvironmentConfiguration),
	}

	option := WithEnvironment("test-env",
		WithEnvironmentEnabled(true),
		WithDryRun(true),
	)

	updated := option(config)

	env, exists := updated.GetEnvironment("test-env")
	assert.True(t, exists)
	assert.True(t, env.Enabled())
	assert.True(t, env.Options().DryRun())
}

func TestWithGitHubEnvironment(t *testing.T) {
	t.Parallel()

	config := ImmutableAppConfiguration{
		environments: make(map[string]ImmutableEnvironmentConfiguration),
	}

	option := WithGitHubEnvironment("github-env", "testowner", "test-token")
	updated := option(config)

	env, exists := updated.GetEnvironment("github-env")
	assert.True(t, exists)

	source := env.Source()
	assert.Equal(t, "github", source.ProviderType())
	assert.Equal(t, "github.com", source.Domain())
	assert.Equal(t, "testowner", source.Owner())
	assert.Equal(t, "test-token", source.Authentication().Token())
}

func TestWithGitLabEnvironment(t *testing.T) {
	t.Parallel()

	config := ImmutableAppConfiguration{
		environments: make(map[string]ImmutableEnvironmentConfiguration),
	}

	option := WithGitLabEnvironment("gitlab-env", "gitlab.example.com", "testowner", "test-token")
	updated := option(config)

	env, exists := updated.GetEnvironment("gitlab-env")
	assert.True(t, exists)

	source := env.Source()
	assert.Equal(t, "gitlab", source.ProviderType())
	assert.Equal(t, "gitlab.example.com", source.Domain())
	assert.Equal(t, "testowner", source.Owner())
	assert.Equal(t, "test-token", source.Authentication().Token())
}

func TestWithGiteaEnvironment(t *testing.T) {
	t.Parallel()

	config := ImmutableAppConfiguration{
		environments: make(map[string]ImmutableEnvironmentConfiguration),
	}

	option := WithGiteaEnvironment("gitea-env", "gitea.example.com", "testowner", "test-token")
	updated := option(config)

	env, exists := updated.GetEnvironment("gitea-env")
	assert.True(t, exists)

	source := env.Source()
	assert.Equal(t, "gitea", source.ProviderType())
	assert.Equal(t, "gitea.example.com", source.Domain())
	assert.Equal(t, "testowner", source.Owner())
	assert.Equal(t, "test-token", source.Authentication().Token())
}

func TestWithLogLevel(t *testing.T) {
	t.Parallel()

	config := ImmutableAppConfiguration{
		globalSettings: ImmutableGlobalSettings{logLevel: LogLevelInfo},
	}

	option := WithLogLevel(LogLevelDebug)
	updated := option(config)

	assert.Equal(t, LogLevelDebug, updated.GlobalSettings().LogLevel())
}

func TestWithLogFormat(t *testing.T) {
	t.Parallel()

	config := ImmutableAppConfiguration{
		globalSettings: ImmutableGlobalSettings{logFormat: LogFormatJSON},
	}

	option := WithLogFormat(LogFormatText)
	updated := option(config)

	assert.Equal(t, LogFormatText, updated.GlobalSettings().LogFormat())
}

func TestWithTempDirectory(t *testing.T) {
	t.Parallel()

	config := ImmutableAppConfiguration{
		globalSettings: ImmutableGlobalSettings{tempDirectory: "/tmp"},
	}

	option := WithTempDirectory("/var/tmp")
	updated := option(config)

	assert.Equal(t, "/var/tmp", updated.GlobalSettings().TempDirectory())
}

func TestWithMetrics(t *testing.T) {
	t.Parallel()

	config := ImmutableAppConfiguration{
		globalSettings: ImmutableGlobalSettings{metricsEnabled: false, metricsPort: 8080},
	}

	option := WithMetrics(true, 9090)
	updated := option(config)

	settings := updated.GlobalSettings()
	assert.True(t, settings.MetricsEnabled())
	assert.Equal(t, 9090, settings.MetricsPort())
}

// Tests for Environment Options

func TestWithSource(t *testing.T) {
	t.Parallel()

	env := ImmutableEnvironmentConfiguration{}

	option := WithSource(
		WithSourceProvider("github"),
		WithSourceDomain("github.com"),
	)

	updated := option(env)
	source := updated.Source()

	assert.Equal(t, "github", source.ProviderType())
	assert.Equal(t, "github.com", source.Domain())
}

func TestWithMirror(t *testing.T) {
	t.Parallel()

	env := ImmutableEnvironmentConfiguration{
		mirrors: make(map[string]ImmutableMirrorConfiguration),
	}

	option := WithMirror("test-mirror",
		WithMirrorProvider("gitlab"),
		WithMirrorDomain("gitlab.com"),
	)

	updated := option(env)
	mirrors := updated.Mirrors()

	mirror, exists := mirrors["test-mirror"]
	assert.True(t, exists)
	assert.Equal(t, "gitlab", mirror.ProviderType())
	assert.Equal(t, "gitlab.com", mirror.Domain())
}

func TestWithGitHubMirror(t *testing.T) {
	t.Parallel()

	env := ImmutableEnvironmentConfiguration{
		mirrors: make(map[string]ImmutableMirrorConfiguration),
	}

	option := WithGitHubMirror("github-mirror", "testowner", "test-token")
	updated := option(env)

	mirrors := updated.Mirrors()
	mirror, exists := mirrors["github-mirror"]
	assert.True(t, exists)
	assert.Equal(t, "github", mirror.ProviderType())
	assert.Equal(t, "github.com", mirror.Domain())
	assert.Equal(t, "testowner", mirror.Owner())
	assert.Equal(t, "test-token", mirror.Authentication().Token())
}

func TestWithGitLabMirror(t *testing.T) {
	t.Parallel()

	env := ImmutableEnvironmentConfiguration{
		mirrors: make(map[string]ImmutableMirrorConfiguration),
	}

	option := WithGitLabMirror("gitlab-mirror", "gitlab.example.com", "testowner", "test-token")
	updated := option(env)

	mirrors := updated.Mirrors()
	mirror, exists := mirrors["gitlab-mirror"]
	assert.True(t, exists)
	assert.Equal(t, "gitlab", mirror.ProviderType())
	assert.Equal(t, "gitlab.example.com", mirror.Domain())
	assert.Equal(t, "testowner", mirror.Owner())
	assert.Equal(t, "test-token", mirror.Authentication().Token())
}

func TestWithDirectoryMirror(t *testing.T) {
	t.Parallel()

	env := ImmutableEnvironmentConfiguration{
		mirrors: make(map[string]ImmutableMirrorConfiguration),
	}

	option := WithDirectoryMirror("dir-mirror", "/tmp/mirrors")
	updated := option(env)

	mirrors := updated.Mirrors()
	mirror, exists := mirrors["dir-mirror"]
	assert.True(t, exists)
	assert.Equal(t, "directory", mirror.ProviderType())
	assert.Equal(t, "/tmp/mirrors", mirror.Path())
}

func TestWithArchiveMirror(t *testing.T) {
	t.Parallel()

	env := ImmutableEnvironmentConfiguration{
		mirrors: make(map[string]ImmutableMirrorConfiguration),
	}

	option := WithArchiveMirror("archive-mirror", "/tmp/archives")
	updated := option(env)

	mirrors := updated.Mirrors()
	mirror, exists := mirrors["archive-mirror"]
	assert.True(t, exists)
	assert.Equal(t, "archive", mirror.ProviderType())
	assert.Equal(t, "/tmp/archives", mirror.Path())
}

func TestWithEnvironmentEnabled(t *testing.T) {
	t.Parallel()

	env := ImmutableEnvironmentConfiguration{enabled: false}

	option := WithEnvironmentEnabled(true)
	updated := option(env)

	assert.True(t, updated.Enabled())
}

func TestWithDryRun(t *testing.T) {
	t.Parallel()

	env := ImmutableEnvironmentConfiguration{
		options: ImmutableEnvironmentOptions{dryRun: false},
	}

	option := WithDryRun(true)
	updated := option(env)

	assert.True(t, updated.Options().DryRun())
}

func TestWithParallelExecution(t *testing.T) {
	t.Parallel()

	env := ImmutableEnvironmentConfiguration{
		options: ImmutableEnvironmentOptions{parallel: false, maxConcurrency: 1},
	}

	option := WithParallelExecution(true, 5)
	updated := option(env)

	options := updated.Options()
	assert.True(t, options.Parallel())
	assert.Equal(t, 5, options.MaxConcurrency())
}

func TestWithTimeout(t *testing.T) {
	t.Parallel()

	env := ImmutableEnvironmentConfiguration{
		options: ImmutableEnvironmentOptions{timeout: time.Minute},
	}

	option := WithTimeout(time.Hour)
	updated := option(env)

	assert.Equal(t, time.Hour, updated.Options().Timeout())
}

func TestWithRetryPolicy(t *testing.T) {
	t.Parallel()

	env := ImmutableEnvironmentConfiguration{
		options: ImmutableEnvironmentOptions{retryAttempts: 1, retryDelay: time.Second},
	}

	option := WithRetryPolicy(3, time.Second*5)
	updated := option(env)

	options := updated.Options()
	assert.Equal(t, 3, options.RetryAttempts())
	assert.Equal(t, time.Second*5, options.RetryDelay())
}

// Tests for Source Options

func TestWithSourceProvider(t *testing.T) {
	t.Parallel()

	source := ImmutableSourceConfiguration{providerType: "github"}

	option := WithSourceProvider("gitlab")
	updated := option(source)

	assert.Equal(t, "gitlab", updated.ProviderType())
}

func TestWithSourceDomain(t *testing.T) {
	t.Parallel()

	source := ImmutableSourceConfiguration{domain: "github.com"}

	option := WithSourceDomain("gitlab.com")
	updated := option(source)

	assert.Equal(t, "gitlab.com", updated.Domain())
}

func TestWithSourceOwner(t *testing.T) {
	t.Parallel()

	source := ImmutableSourceConfiguration{owner: "oldowner"}

	option := WithSourceOwner("newowner")
	updated := option(source)

	assert.Equal(t, "newowner", updated.Owner())
}

func TestWithSourceAuth(t *testing.T) {
	t.Parallel()

	source := ImmutableSourceConfiguration{
		authentication: ImmutableAuthenticationConfiguration{authType: AuthenticationTypeNone},
	}

	option := WithSourceAuth(WithAuthToken("test-token"))
	updated := option(source)

	auth := updated.Authentication()
	assert.Equal(t, AuthenticationTypeToken, auth.Type())
	assert.Equal(t, "test-token", auth.Token())
}

func TestWithIncludePatterns(t *testing.T) {
	t.Parallel()

	source := ImmutableSourceConfiguration{
		repository: ImmutableRepositoryConfiguration{includePatterns: []string{"old"}},
	}

	patterns := []string{"new1", "new2"}
	option := WithIncludePatterns(patterns)
	updated := option(source)

	result := updated.Repository().IncludePatterns()
	assert.Equal(t, patterns, result)
}

func TestWithExcludePatterns(t *testing.T) {
	t.Parallel()

	source := ImmutableSourceConfiguration{
		repository: ImmutableRepositoryConfiguration{excludePatterns: []string{"old"}},
	}

	patterns := []string{"new1", "new2"}
	option := WithExcludePatterns(patterns)
	updated := option(source)

	result := updated.Repository().ExcludePatterns()
	assert.Equal(t, patterns, result)
}

func TestWithIncludeForks(t *testing.T) {
	t.Parallel()

	source := ImmutableSourceConfiguration{
		repository: ImmutableRepositoryConfiguration{includeForks: false},
	}

	option := WithIncludeForks(true)
	updated := option(source)

	assert.True(t, updated.Repository().IncludeForks())
}

func TestWithIncludeArchived(t *testing.T) {
	t.Parallel()

	source := ImmutableSourceConfiguration{
		repository: ImmutableRepositoryConfiguration{includeArchived: false},
	}

	option := WithIncludeArchived(true)
	updated := option(source)

	assert.True(t, updated.Repository().IncludeArchived())
}

// Tests for Mirror Options

func TestWithMirrorProvider(t *testing.T) {
	t.Parallel()

	mirror := ImmutableMirrorConfiguration{providerType: "github"}

	option := WithMirrorProvider("gitlab")
	updated := option(mirror)

	assert.Equal(t, "gitlab", updated.ProviderType())
}

func TestWithMirrorDomain(t *testing.T) {
	t.Parallel()

	mirror := ImmutableMirrorConfiguration{domain: "github.com"}

	option := WithMirrorDomain("gitlab.com")
	updated := option(mirror)

	assert.Equal(t, "gitlab.com", updated.Domain())
}

func TestWithMirrorOwner(t *testing.T) {
	t.Parallel()

	mirror := ImmutableMirrorConfiguration{owner: "oldowner"}

	option := WithMirrorOwner("newowner")
	updated := option(mirror)

	assert.Equal(t, "newowner", updated.Owner())
}

func TestWithMirrorPath(t *testing.T) {
	t.Parallel()

	mirror := ImmutableMirrorConfiguration{path: "/old/path"}

	option := WithMirrorPath("/new/path")
	updated := option(mirror)

	assert.Equal(t, "/new/path", updated.Path())
}

func TestWithMirrorAuth(t *testing.T) {
	t.Parallel()

	mirror := ImmutableMirrorConfiguration{
		authentication: ImmutableAuthenticationConfiguration{authType: AuthenticationTypeNone},
	}

	option := WithMirrorAuth(WithAuthToken("test-token"))
	updated := option(mirror)

	auth := updated.Authentication()
	assert.Equal(t, AuthenticationTypeToken, auth.Type())
	assert.Equal(t, "test-token", auth.Token())
}

func TestWithMirrorEnabled(t *testing.T) {
	t.Parallel()

	mirror := ImmutableMirrorConfiguration{enabled: false}

	option := WithMirrorEnabled(true)
	updated := option(mirror)

	assert.True(t, updated.Enabled())
}

func TestWithCreateIfNotExists(t *testing.T) {
	t.Parallel()

	mirror := ImmutableMirrorConfiguration{
		options: ImmutableMirrorOptionsConfiguration{createIfNotExists: false},
	}

	option := WithCreateIfNotExists(true)
	updated := option(mirror)

	assert.True(t, updated.Options().createIfNotExists)
}

func TestWithUpdateDescription(t *testing.T) {
	t.Parallel()

	mirror := ImmutableMirrorConfiguration{
		options: ImmutableMirrorOptionsConfiguration{updateDescription: false},
	}

	option := WithUpdateDescription(true)
	updated := option(mirror)

	assert.True(t, updated.Options().updateDescription)
}

func TestWithSyncVisibility(t *testing.T) {
	t.Parallel()

	mirror := ImmutableMirrorConfiguration{
		options: ImmutableMirrorOptionsConfiguration{syncVisibility: false},
	}

	option := WithSyncVisibility(true)
	updated := option(mirror)

	assert.True(t, updated.Options().syncVisibility)
}

func TestWithSyncTopics(t *testing.T) {
	t.Parallel()

	mirror := ImmutableMirrorConfiguration{
		options: ImmutableMirrorOptionsConfiguration{syncTopics: false},
	}

	option := WithSyncTopics(true)
	updated := option(mirror)

	assert.True(t, updated.Options().syncTopics)
}

func TestWithBranchProtection(t *testing.T) {
	t.Parallel()

	mirror := ImmutableMirrorConfiguration{
		options: ImmutableMirrorOptionsConfiguration{syncBranchProtection: false},
	}

	option := WithBranchProtection(true)
	updated := option(mirror)

	assert.True(t, updated.Options().syncBranchProtection)
}

func TestWithGitLFS(t *testing.T) {
	t.Parallel()

	mirror := ImmutableMirrorConfiguration{
		options: ImmutableMirrorOptionsConfiguration{enableLFS: false},
	}

	option := WithGitLFS(true)
	updated := option(mirror)

	assert.True(t, updated.Options().enableLFS)
}

// Tests for Auth Options

func TestWithAuthToken(t *testing.T) {
	t.Parallel()

	auth := ImmutableAuthenticationConfiguration{authType: AuthenticationTypeNone}

	option := WithAuthToken("test-token")
	updated := option(auth)

	assert.Equal(t, AuthenticationTypeToken, updated.Type())
	assert.Equal(t, "test-token", updated.Token())
}

func TestWithAuthBasic(t *testing.T) {
	t.Parallel()

	auth := ImmutableAuthenticationConfiguration{authType: AuthenticationTypeNone}

	option := WithAuthBasic("testuser", "testpass")
	updated := option(auth)

	assert.Equal(t, AuthenticationTypeBasic, updated.Type())
	assert.Equal(t, "testuser", updated.Username())
	assert.Equal(t, "testpass", updated.Password())
}

func TestWithAuthSSH(t *testing.T) {
	t.Parallel()

	auth := ImmutableAuthenticationConfiguration{authType: AuthenticationTypeNone}

	option := WithAuthSSH("/path/to/key", "testuser")
	updated := option(auth)

	assert.Equal(t, AuthenticationTypeSSH, updated.Type())
	assert.Equal(t, "/path/to/key", updated.SSHKeyPath())
	assert.Equal(t, "testuser", updated.Username())
}

func TestWithAuthSSHKey(t *testing.T) {
	t.Parallel()

	auth := ImmutableAuthenticationConfiguration{authType: AuthenticationTypeNone}

	option := WithAuthSSHKey("ssh-key-content", "testuser")
	updated := option(auth)

	assert.Equal(t, AuthenticationTypeSSH, updated.Type())
	assert.Equal(t, "ssh-key-content", updated.SSHKey())
	assert.Equal(t, "testuser", updated.Username())
}

// Tests for Configuration Factory Functions

func TestNewAppConfiguration(t *testing.T) {
	t.Parallel()

	config := NewAppConfiguration(
		WithLogLevel(LogLevelDebug),
		WithLogFormat(LogFormatText),
		WithEnvironment("test",
			WithEnvironmentEnabled(true),
		),
	)

	assert.Equal(t, LogLevelDebug, config.GlobalSettings().LogLevel())
	assert.Equal(t, LogFormatText, config.GlobalSettings().LogFormat())

	env, exists := config.GetEnvironment("test")
	assert.True(t, exists)
	assert.True(t, env.Enabled())
}

func TestComposeAppConfigOptions(t *testing.T) {
	t.Parallel()

	options := []AppConfigOption{
		WithLogLevel(LogLevelDebug),
		WithLogFormat(LogFormatText),
	}

	composedOption := ComposeAppConfigOptions(options...)
	config := ImmutableAppConfiguration{
		globalSettings: ImmutableGlobalSettings{
			logLevel:  LogLevelInfo,
			logFormat: LogFormatJSON,
		},
	}

	updated := composedOption(config)

	assert.Equal(t, LogLevelDebug, updated.GlobalSettings().LogLevel())
	assert.Equal(t, LogFormatText, updated.GlobalSettings().LogFormat())
}

func TestProductionDefaults(t *testing.T) {
	t.Parallel()

	config := ImmutableAppConfiguration{}
	option := ProductionDefaults()
	updated := option(config)

	settings := updated.GlobalSettings()
	assert.Equal(t, LogLevelInfo, settings.LogLevel())
	assert.True(t, settings.MetricsEnabled())
	assert.Equal(t, "/tmp/git-provider-sync", settings.TempDirectory())
}

func TestDevelopmentDefaults(t *testing.T) {
	t.Parallel()

	config := ImmutableAppConfiguration{}
	option := DevelopmentDefaults()
	updated := option(config)

	settings := updated.GlobalSettings()
	assert.Equal(t, LogLevelDebug, settings.LogLevel())
	assert.False(t, settings.MetricsEnabled())
	assert.Equal(t, "/tmp/git-provider-sync-dev", settings.TempDirectory())
}

func TestGitHubToGitLabMirror(t *testing.T) {
	t.Parallel()

	config := ImmutableAppConfiguration{
		environments: make(map[string]ImmutableEnvironmentConfiguration),
	}

	option := GitHubToGitLabMirror("test-env", "github-owner", "github-token",
		"gitlab.com", "gitlab-owner", "gitlab-token")
	updated := option(config)

	env, exists := updated.GetEnvironment("test-env")
	assert.True(t, exists)

	// Check source (GitHub)
	source := env.Source()
	assert.Equal(t, "github", source.ProviderType())
	assert.Equal(t, "github.com", source.Domain())
	assert.Equal(t, "github-owner", source.Owner())
	assert.Equal(t, "github-token", source.Authentication().Token())

	// Check mirror (GitLab)
	mirrors := env.Mirrors()
	mirror, exists := mirrors["gitlab-mirror"]
	assert.True(t, exists)
	assert.Equal(t, "gitlab", mirror.ProviderType())
	assert.Equal(t, "gitlab.com", mirror.Domain())
	assert.Equal(t, "gitlab-owner", mirror.Owner())
	assert.Equal(t, "gitlab-token", mirror.Authentication().Token())
}

func TestNewEnvironment(t *testing.T) {
	t.Parallel()

	env := NewEnvironment(
		WithEnvironmentEnabled(true),
		WithDryRun(true),
		WithSource(
			WithSourceProvider("github"),
			WithSourceDomain("github.com"),
		),
	)

	assert.True(t, env.Enabled())
	assert.True(t, env.Options().DryRun())

	source := env.Source()
	assert.Equal(t, "github", source.ProviderType())
	assert.Equal(t, "github.com", source.Domain())
}

func TestNewSource(t *testing.T) {
	t.Parallel()

	source := NewSource(
		WithSourceProvider("gitlab"),
		WithSourceDomain("gitlab.com"),
		WithSourceOwner("testowner"),
		WithSourceAuth(WithAuthToken("test-token")),
	)

	assert.Equal(t, "gitlab", source.ProviderType())
	assert.Equal(t, "gitlab.com", source.Domain())
	assert.Equal(t, "testowner", source.Owner())
	assert.Equal(t, "test-token", source.Authentication().Token())
}

func TestNewMirror(t *testing.T) {
	t.Parallel()

	mirror := NewMirror(
		WithMirrorProvider("github"),
		WithMirrorDomain("github.com"),
		WithMirrorOwner("testowner"),
		WithMirrorEnabled(true),
	)

	assert.Equal(t, "github", mirror.ProviderType())
	assert.Equal(t, "github.com", mirror.Domain())
	assert.Equal(t, "testowner", mirror.Owner())
	assert.True(t, mirror.Enabled())
}

func TestNewAuth(t *testing.T) {
	t.Parallel()

	auth := NewAuth(
		WithAuthToken("test-token"),
	)

	assert.Equal(t, AuthenticationTypeToken, auth.Type())
	assert.Equal(t, "test-token", auth.Token())
}

func TestDefaultGitHubConfig(t *testing.T) {
	t.Parallel()

	config := DefaultGitHubConfig("testowner", "test-token")

	env, exists := config.GetEnvironment("github")
	assert.True(t, exists)

	source := env.Source()
	assert.Equal(t, "github", source.ProviderType())
	assert.Equal(t, "github.com", source.Domain())
	assert.Equal(t, "testowner", source.Owner())
	assert.Equal(t, "test-token", source.Authentication().Token())
}

func TestDefaultGitLabConfig(t *testing.T) {
	t.Parallel()

	config := DefaultGitLabConfig("gitlab.example.com", "testowner", "test-token")

	env, exists := config.GetEnvironment("gitlab")
	assert.True(t, exists)

	source := env.Source()
	assert.Equal(t, "gitlab", source.ProviderType())
	assert.Equal(t, "gitlab.example.com", source.Domain())
	assert.Equal(t, "testowner", source.Owner())
	assert.Equal(t, "test-token", source.Authentication().Token())
}

func TestDevelopmentConfig(t *testing.T) {
	t.Parallel()

	config := DevelopmentConfig()

	settings := config.GlobalSettings()
	assert.Equal(t, LogLevelDebug, settings.LogLevel())
	assert.Equal(t, LogFormatConsole, settings.LogFormat())
	assert.False(t, settings.MetricsEnabled())
}

func TestProductionConfig(t *testing.T) {
	t.Parallel()

	config := ProductionConfig()

	settings := config.GlobalSettings()
	assert.Equal(t, LogLevelInfo, settings.LogLevel())
	assert.Equal(t, LogFormatJSON, settings.LogFormat())
	assert.True(t, settings.MetricsEnabled())
}
