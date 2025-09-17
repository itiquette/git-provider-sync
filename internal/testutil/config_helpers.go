// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package testutil

import (
	"fmt"
	"testing"
)

// ConfigBuilder helps build test configurations with a fluent API.
type ConfigBuilder struct {
	t       testing.TB
	content string
}

// NewConfigBuilder creates a new configuration builder.
func NewConfigBuilder(tb testing.TB) *ConfigBuilder {
	tb.Helper()

	return &ConfigBuilder{
		t:       tb,
		content: "gitprovidersync:\n",
	}
}

// AddEnvironment adds an environment to the configuration.
func (cb *ConfigBuilder) AddEnvironment(name string) *EnvironmentBuilder {
	return &EnvironmentBuilder{
		parent: cb,
		name:   name,
	}
}

// Build returns the final configuration as a string.
func (cb *ConfigBuilder) Build() string {
	return cb.content
}

// EnvironmentBuilder builds an environment configuration.
type EnvironmentBuilder struct {
	parent *ConfigBuilder
	name   string
}

// WithSource adds a source configuration.
func (eb *EnvironmentBuilder) WithSource(provider, owner string) *EnvironmentBuilder {
	content := fmt.Sprintf(`  %s:
    source:
      provider_type: %s
      owner: %s
      owner_type: user
`, eb.name, provider, owner)
	eb.parent.content += content

	return eb
}

// WithMirror adds a mirror configuration.
func (eb *EnvironmentBuilder) WithMirror(name, provider, owner string) *EnvironmentBuilder {
	content := fmt.Sprintf(`      mirrors:
        %s:
          provider_type: %s
          owner: %s
          owner_type: user
`, name, provider, owner)
	eb.parent.content += content

	return eb
}

// WithToken adds authentication with a token.
func (eb *EnvironmentBuilder) WithToken(token string) *EnvironmentBuilder {
	content := fmt.Sprintf(`      auth:
        token: "%s"
`, token)
	eb.parent.content += content

	return eb
}

// Build returns to the parent ConfigBuilder.
func (eb *EnvironmentBuilder) Build() *ConfigBuilder {
	return eb.parent
}

// TestConfigFixture provides pre-built configuration fixtures.
type TestConfigFixture struct {
	t    testing.TB
	Path string
	FS   *AferoTestFS
}

// NewTestConfigFixture creates a new configuration fixture.
func NewTestConfigFixture(tb testing.TB) *TestConfigFixture {
	tb.Helper()

	fs := NewMemFS(tb)

	return &TestConfigFixture{
		t:  tb,
		FS: fs,
	}
}

// CreateMinimalConfig creates a minimal valid configuration.
func (tcf *TestConfigFixture) CreateMinimalConfig() string {
	tcf.t.Helper()

	config := `gitprovidersync:
  test:
    source:
      provider_type: github
      owner: test-owner
      owner_type: user
`
	tcf.Path = tcf.FS.WriteConfig(config)

	return tcf.Path
}

// CreateGitHubConfig creates a GitHub-specific configuration.
func (tcf *TestConfigFixture) CreateGitHubConfig(owner, token string) string {
	tcf.t.Helper()

	config := fmt.Sprintf(`gitprovidersync:
  production:
    source:
      provider_type: github
      domain: github.com
      owner: %s
      owner_type: user
      auth:
        token: "%s"
`, owner, token)
	tcf.Path = tcf.FS.WriteConfig(config)

	return tcf.Path
}

// CreateGitLabConfig creates a GitLab-specific configuration.
func (tcf *TestConfigFixture) CreateGitLabConfig(owner, token string) string {
	tcf.t.Helper()

	config := fmt.Sprintf(`gitprovidersync:
  production:
    source:
      provider_type: gitlab
      domain: gitlab.com
      owner: %s
      owner_type: group
      auth:
        token: "%s"
`, owner, token)
	tcf.Path = tcf.FS.WriteConfig(config)

	return tcf.Path
}

// CreateMirrorConfig creates a configuration with mirror setup.
func (tcf *TestConfigFixture) CreateMirrorConfig(sourceProvider, targetProvider string) string {
	tcf.t.Helper()

	config := fmt.Sprintf(`gitprovidersync:
  sync:
    source:
      provider_type: %s
      domain: %s.example.com
      owner: source-owner
      owner_type: user
      auth:
        token: "${SOURCE_TOKEN}"
      mirrors:
        backup:
          provider_type: %s
          domain: %s.example.com
          owner: backup-owner
          owner_type: user
          auth:
            token: "${TARGET_TOKEN}"
`, sourceProvider, sourceProvider, targetProvider, targetProvider)
	tcf.Path = tcf.FS.WriteConfig(config)

	return tcf.Path
}

// CreateMultiEnvironmentConfig creates a configuration with multiple environments.
func (tcf *TestConfigFixture) CreateMultiEnvironmentConfig() string {
	tcf.t.Helper()

	config := `gitprovidersync:
  development:
    source:
      provider_type: github
      owner: dev-owner
      owner_type: user
  staging:
    source:
      provider_type: gitlab
      owner: staging-owner
      owner_type: group
  production:
    source:
      provider_type: gitea
      owner: prod-owner
      owner_type: user
`
	tcf.Path = tcf.FS.WriteConfig(config)

	return tcf.Path
}

// CreateInvalidConfig creates an intentionally invalid configuration for testing error handling.
func (tcf *TestConfigFixture) CreateInvalidConfig(invalidType string) string {
	tcf.t.Helper()

	var config string

	switch invalidType {
	case "missing-owner":
		config = `gitprovidersync:
  test:
    source:
      provider_type: github
      owner_type: user
`
	case "invalid-provider":
		config = `gitprovidersync:
  test:
    source:
      provider_type: invalid-provider
      owner: test-owner
`
	case "malformed-yaml":
		config = `gitprovidersync:
  test
    source:
      invalid yaml structure
`
	default:
		config = `invalid: configuration`
	}

	tcf.Path = tcf.FS.WriteConfig(config)

	return tcf.Path
}

// StandardTestConfigs provides commonly used test configurations.
//
//nolint:gochecknoglobals // Test fixtures are acceptable as package-level variables
var StandardTestConfigs = struct {
	Minimal     string
	GitHub      string
	GitLab      string
	Gitea       string
	WithMirror  string
	WithSSH     string
	WithFilters string
	WithArchive string
	MultiEnv    string
	EmptyConfig string
}{
	Minimal: `gitprovidersync:
  test:
    source:
      provider_type: github
      owner: test-owner
      owner_type: user
`,
	GitHub: `gitprovidersync:
  github:
    source:
      provider_type: github
      domain: github.com
      owner: octocat
      owner_type: user
      auth:
        token: "${GITHUB_TOKEN}"
`,
	GitLab: `gitprovidersync:
  gitlab:
    source:
      provider_type: gitlab
      domain: gitlab.com
      owner: gitlab-org
      owner_type: group
      auth:
        token: "${GITLAB_TOKEN}"
`,
	Gitea: `gitprovidersync:
  gitea:
    source:
      provider_type: gitea
      domain: gitea.example.com
      owner: gitea-user
      owner_type: user
      auth:
        token: "${GITEA_TOKEN}"
`,
	WithMirror: `gitprovidersync:
  sync:
    source:
      provider_type: github
      owner: source-owner
      mirrors:
        backup:
          provider_type: gitlab
          owner: backup-owner
`,
	WithSSH: `gitprovidersync:
  ssh:
    source:
      provider_type: github
      owner: ssh-user
      auth:
        method: ssh
        ssh_key_path: ~/.ssh/id_rsa
`,
	WithFilters: `gitprovidersync:
  filtered:
    source:
      provider_type: github
      owner: filter-owner
      filters:
        include:
          - "^project-.*"
        exclude:
          - ".*-archived$"
        exclude_forks: true
`,
	WithArchive: `gitprovidersync:
  archive:
    source:
      provider_type: github
      owner: archive-owner
      mirrors:
        archive:
          provider_type: archive
          path: /backup/archives
`,
	MultiEnv: `gitprovidersync:
  dev:
    source:
      provider_type: github
      owner: dev-owner
  staging:
    source:
      provider_type: gitlab
      owner: staging-owner
  prod:
    source:
      provider_type: gitea
      owner: prod-owner
`,
	EmptyConfig: `gitprovidersync: {}
`,
}
