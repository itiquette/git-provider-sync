// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package configuration

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	model "itiquette/git-provider-sync/internal/model/configuration"
)

func TestPrintConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		config         *model.AppConfiguration
		expectedOutput []string // Strings that should be present in output
		notExpected    []string // Strings that should NOT be present
	}{
		{
			name: "Simple configuration",
			config: &model.AppConfiguration{
				GitProviderSyncConfs: map[string]model.Environment{
					"production": {
						"main": model.SyncConfig{
							BaseConfig: model.BaseConfig{
								ProviderType: "github",
								Domain:       "github.com",
								Owner:        "test-org",
								OwnerType:    "group",
								Auth: model.AuthConfig{
									Token: "github-token",
								},
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"Environment: production",
				"Configuration: main",
				"Provider Type: github",
				"Domain: github.com",
				"Owner: test-org (group)",
				"Token: github-token",
			},
			notExpected: []string{
				"Mirrors:",
				"Username:",
				"Password:",
			},
		},
		{
			name: "Configuration with mirrors",
			config: &model.AppConfiguration{
				GitProviderSyncConfs: map[string]model.Environment{
					"development": {
						"source": model.SyncConfig{
							BaseConfig: model.BaseConfig{
								ProviderType: "github",
								Domain:       "github.com",
								Owner:        "dev-org",
								Auth: model.AuthConfig{
									Token: "dev-token",
								},
							},
							Mirrors: map[string]model.MirrorConfig{
								"backup": {
									BaseConfig: model.BaseConfig{
										ProviderType: "gitlab",
										Domain:       "gitlab.company.com",
										Owner:        "backup-org",
										Auth: model.AuthConfig{
											Token: "backup-token",
										},
									},
									Settings: model.MirrorSettings{
										DescriptionPrefix: "[BACKUP]",
										Visibility:        "private",
									},
								},
								"archive": {
									BaseConfig: model.BaseConfig{
										ProviderType: "archive",
									},
									Path: "/backup/archives",
								},
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"Environment: development",
				"Configuration: source",
				"Provider Type: github",
				"Mirrors:",
				"Mirror: backup",
				"Provider Type: gitlab",
				"Domain: gitlab.company.com",
				"Owner: backup-org",
				"Description Prefix: [BACKUP]",
				"Visibility: private",
				"Mirror: archive",
				"Provider Type: archive",
				"Directory Path: /backup/archives",
			},
		},
		{
			name: "Configuration with all auth types",
			config: &model.AppConfiguration{
				GitProviderSyncConfs: map[string]model.Environment{
					"auth-test": {
						"basic-auth": model.SyncConfig{
							BaseConfig: model.BaseConfig{
								ProviderType: "gitlab",
								Domain:       "gitlab.com",
								Owner:        "test-user",
								Auth: model.AuthConfig{
									Token: "basic-auth-token",
								},
							},
						},
						"ssh-auth": model.SyncConfig{
							BaseConfig: model.BaseConfig{
								ProviderType: "github",
								Domain:       "github.com",
								Owner:        "ssh-org",
								Auth: model.AuthConfig{
									SSHCommand: "ssh -i ~/.ssh/custom_key",
								},
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"Environment: auth-test",
				"Configuration: basic-auth",
				"Token: basic-auth-token",
				"Configuration: ssh-auth",
				"SSH Command: ssh -i ~/.ssh/custom_key",
			},
		},
		{
			name: "Configuration with repository options",
			config: &model.AppConfiguration{
				GitProviderSyncConfs: map[string]model.Environment{
					"repo-test": {
						"filtered": model.SyncConfig{
							BaseConfig: model.BaseConfig{
								ProviderType: "github",
								Domain:       "github.com",
								Owner:        "filtered-org",
								Auth: model.AuthConfig{
									Token: "token",
								},
							},
							Repositories: model.RepositoriesOption{
								Include: []string{"repo1", "repo2", "repo3"},
								Exclude: []string{"private-repo", "test-repo"},
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"Environment: repo-test",
				"Configuration: filtered",
				"Repositories:",
				"Include: [repo1, repo2, repo3]",
				"Exclude: [private-repo, test-repo]",
			},
		},
		{
			name: "Configuration with advanced settings",
			config: &model.AppConfiguration{
				GitProviderSyncConfs: map[string]model.Environment{
					"advanced": {
						"complex": model.SyncConfig{
							BaseConfig: model.BaseConfig{
								ProviderType: "github",
								Domain:       "github.enterprise.com",
								Owner:        "enterprise-org",
								Auth: model.AuthConfig{
									Token:             "enterprise-token",
									HTTPScheme:        "https",
									ProxyURL:          "http://proxy.company.com:8080",
									CertDirPath:       "/etc/ssl/certs",
									SSHURLRewriteFrom: "ssh://github.enterprise.com",
									SSHURLRewriteTo:   "ssh://git.company.com",
								},
							},
							IncludeForks:    false,
							ActiveFromLimit: "30",
							Mirrors: map[string]model.MirrorConfig{
								"directory-backup": {
									BaseConfig: model.BaseConfig{
										ProviderType: "directory",
									},
									Path: "/enterprise/backups",
									Settings: model.MirrorSettings{
										IgnoreInvalidName: true,
									},
								},
							},
						},
					},
				},
			},
			expectedOutput: []string{
				"Environment: advanced",
				"Configuration: complex",
				"Provider Type: github",
				"Domain: github.enterprise.com",
				"Owner: enterprise-org",
				"Token: enterprise-token",
				"HTTP Scheme: https",
				"Proxy URL: http://proxy.company.com:8080",
				"Certificate Directory: /etc/ssl/certs",
				"SSH URL Rewrite From: ssh://github.enterprise.com",
				"SSH URL Rewrite To: ssh://git.company.com",
				"Include Forks: false",
				"Active From Limit: 30",
				"Mirror: directory-backup",
				"Provider Type: directory",
				"Directory Path: /enterprise/backups",
				"Ignore Invalid Name: true",
			},
		},
		{
			name: "Empty configuration",
			config: &model.AppConfiguration{
				GitProviderSyncConfs: map[string]model.Environment{},
			},
			expectedOutput: []string{},
			notExpected: []string{
				"Environment:",
				"Configuration:",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Capture output
			var buf bytes.Buffer

			PrintConfiguration(*test.config, &buf)
			output := buf.String()

			// Check expected strings are present
			for _, expected := range test.expectedOutput {
				assert.Contains(t, output, expected, "Expected output to contain: %s", expected)
			}

			// Check unwanted strings are not present
			for _, notExpected := range test.notExpected {
				assert.NotContains(t, output, notExpected, "Expected output to NOT contain: %s", notExpected)
			}
		})
	}
}

func TestPrintEnvironment(t *testing.T) {
	t.Parallel()

	envConf := model.Environment{
		"primary": model.SyncConfig{
			BaseConfig: model.BaseConfig{
				ProviderType: "github",
				Domain:       "github.com",
				Owner:        "primary-org",
				Auth: model.AuthConfig{
					Token: "primary-token",
				},
			},
		},
		"secondary": model.SyncConfig{
			BaseConfig: model.BaseConfig{
				ProviderType: "gitlab",
				Domain:       "gitlab.com",
				Owner:        "secondary-org",
				Auth: model.AuthConfig{
					Token: "secondary-token",
				},
			},
		},
	}

	var buf bytes.Buffer

	printEnvironment("test-env", envConf, &buf, 0)
	output := buf.String()

	assert.Contains(t, output, "Environment: test-env")
	assert.Contains(t, output, "Configuration: primary")
	assert.Contains(t, output, "Configuration: secondary")
	assert.Contains(t, output, "Provider Type: github")
	assert.Contains(t, output, "Provider Type: gitlab")
}

func TestPrintSyncConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		syncConfig     model.SyncConfig
		expectedOutput []string
	}{
		{
			name: "Basic sync config",
			syncConfig: model.SyncConfig{
				BaseConfig: model.BaseConfig{
					ProviderType: "github",
					Domain:       "github.com",
					Owner:        "test-owner",
					OwnerType:    "user",
				},
			},
			expectedOutput: []string{
				"Provider Type: github",
				"Domain: github.com",
				"Owner: test-owner (user)",
			},
		},
		{
			name: "Sync config with all optional fields",
			syncConfig: model.SyncConfig{
				BaseConfig: model.BaseConfig{
					ProviderType: "gitlab",
					Domain:       "gitlab.company.com",
					Owner:        "company-group",
					OwnerType:    "group",
				},
				IncludeForks:    false,
				ActiveFromLimit: "60",
			},
			expectedOutput: []string{
				"Provider Type: gitlab",
				"Domain: gitlab.company.com",
				"Owner: company-group (group)",
				"Include Forks: false",
				"Active From Limit: 60",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			printSyncConfig("test-config", test.syncConfig, &buf, 0)
			output := buf.String()

			assert.Contains(t, output, "Configuration: test-config")

			for _, expected := range test.expectedOutput {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestPrintAuthConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		authConfig     model.AuthConfig
		expectedOutput []string
		notExpected    []string
	}{
		{
			name: "Token authentication",
			authConfig: model.AuthConfig{
				Token: "github-token-123",
			},
			expectedOutput: []string{
				"Token: github-token-123",
			},
			notExpected: []string{
				"Username:",
				"Password:",
				"SSH Key:",
			},
		},
		{
			name: "HTTP authentication",
			authConfig: model.AuthConfig{
				HTTPScheme: "https",
				ProxyURL:   "http://proxy.example.com:8080",
			},
			expectedOutput: []string{
				"HTTP Scheme: https",
				"Proxy URL: http://proxy.example.com:8080",
			},
			notExpected: []string{
				"Token:",
				"SSH Command:",
			},
		},
		{
			name: "SSH authentication",
			authConfig: model.AuthConfig{
				SSHCommand:        "ssh -i ~/.ssh/custom_key -o StrictHostKeyChecking=no",
				SSHURLRewriteFrom: "ssh://original.com",
				SSHURLRewriteTo:   "ssh://rewritten.com",
			},
			expectedOutput: []string{
				"SSH Command: ssh -i ~/.ssh/custom_key -o StrictHostKeyChecking=no",
				"SSH URL Rewrite From: ssh://original.com",
				"SSH URL Rewrite To: ssh://rewritten.com",
			},
			notExpected: []string{
				"Token:",
				"HTTP Scheme:",
			},
		},
		{
			name: "Complete auth configuration",
			authConfig: model.AuthConfig{
				Token:             "complete-token",
				SSHCommand:        "ssh-command",
				HTTPScheme:        "https",
				ProxyURL:          "http://proxy.example.com:8080",
				CertDirPath:       "/etc/ssl/certs",
				SSHURLRewriteFrom: "ssh://original.com",
				SSHURLRewriteTo:   "ssh://rewritten.com",
				Protocol:          "tls",
				RequestTimeout:    30,
			},
			expectedOutput: []string{
				"Token: complete-token",
				"SSH Command: ssh-command",
				"HTTP Scheme: https",
				"Proxy URL: http://proxy.example.com:8080",
				"Certificate Directory: /etc/ssl/certs",
				"SSH URL Rewrite From: ssh://original.com",
				"SSH URL Rewrite To: ssh://rewritten.com",
				"Protocol: tls",
				"Request Timeout: 30",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			printAuthConfig(test.authConfig, &buf, 0)
			output := buf.String()

			for _, expected := range test.expectedOutput {
				assert.Contains(t, output, expected)
			}

			for _, notExpected := range test.notExpected {
				assert.NotContains(t, output, notExpected)
			}
		})
	}
}

func TestPrintMirrorConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		mirrorName     string
		mirrorConfig   model.MirrorConfig
		expectedOutput []string
	}{
		{
			name:       "Remote mirror configuration",
			mirrorName: "remote-backup",
			mirrorConfig: model.MirrorConfig{
				BaseConfig: model.BaseConfig{
					ProviderType: "gitlab",
					Domain:       "backup.gitlab.com",
					Owner:        "backup-org",
					Auth: model.AuthConfig{
						Token: "backup-token",
					},
				},
			},
			expectedOutput: []string{
				"Mirror: remote-backup",
				"Provider Type: gitlab",
				"Domain: backup.gitlab.com",
				"Owner: backup-org",
				"Token: backup-token",
			},
		},
		{
			name:       "Archive mirror configuration",
			mirrorName: "archive-backup",
			mirrorConfig: model.MirrorConfig{
				BaseConfig: model.BaseConfig{
					ProviderType: "archive",
				},
				Path: "/backup/archives/git-repos",
			},
			expectedOutput: []string{
				"Mirror: archive-backup",
				"Provider Type: archive",
				"Directory Path: /backup/archives/git-repos",
			},
		},
		{
			name:       "Directory mirror configuration",
			mirrorName: "directory-backup",
			mirrorConfig: model.MirrorConfig{
				BaseConfig: model.BaseConfig{
					ProviderType: "directory",
				},
				Path: "/local/backup/repositories",
				Settings: model.MirrorSettings{
					DescriptionPrefix: "[LOCAL-BACKUP]",
					Visibility:        "private",
				},
			},
			expectedOutput: []string{
				"Mirror: directory-backup",
				"Provider Type: directory",
				"Directory Path: /local/backup/repositories",
				"Description Prefix: [LOCAL-BACKUP]",
				"Visibility: private",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			printMirrorConfig(test.mirrorName, test.mirrorConfig, &buf, 0)
			output := buf.String()

			for _, expected := range test.expectedOutput {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestPrintMirrorSettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		settings       model.MirrorSettings
		expectedOutput []string
		notExpected    []string
	}{
		{
			name: "All mirror settings",
			settings: model.MirrorSettings{
				DescriptionPrefix: "[MIRROR]",
				Visibility:        "internal",
				IgnoreInvalidName: true,
			},
			expectedOutput: []string{
				"Description Prefix: [MIRROR]",
				"Visibility: internal",
				"Ignore Invalid Name: true",
			},
		},
		{
			name: "Partial mirror settings",
			settings: model.MirrorSettings{
				DescriptionPrefix: "[BACKUP]",
			},
			expectedOutput: []string{
				"Description Prefix: [BACKUP]",
			},
			notExpected: []string{
				"Visibility:",
				"Ignore Invalid Name:",
			},
		},
		{
			name:           "Empty mirror settings",
			settings:       model.MirrorSettings{},
			expectedOutput: []string{},
			notExpected: []string{
				"Description Prefix:",
				"Visibility:",
				"Ignore Invalid Name:",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			printMirrorSettings(test.settings, &buf, 0)
			output := buf.String()

			for _, expected := range test.expectedOutput {
				assert.Contains(t, output, expected)
			}

			for _, notExpected := range test.notExpected {
				assert.NotContains(t, output, notExpected)
			}
		})
	}
}

func TestPrintRepositoriesOption(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		repositories   model.RepositoriesOption
		expectedOutput []string
		notExpected    []string
	}{
		{
			name: "Include and exclude lists",
			repositories: model.RepositoriesOption{
				Include: []string{"repo1", "repo2", "important-repo"},
				Exclude: []string{"test-repo", "deprecated-repo"},
			},
			expectedOutput: []string{
				"Repositories:",
				"Include: [repo1, repo2, important-repo]",
				"Exclude: [test-repo, deprecated-repo]",
			},
		},
		{
			name: "Include only",
			repositories: model.RepositoriesOption{
				Include: []string{"only-this-repo"},
			},
			expectedOutput: []string{
				"Repositories:",
				"Include: [only-this-repo]",
			},
			notExpected: []string{
				"Exclude:",
			},
		},
		{
			name: "Exclude only",
			repositories: model.RepositoriesOption{
				Exclude: []string{"skip-this-repo"},
			},
			expectedOutput: []string{
				"Repositories:",
				"Exclude: [skip-this-repo]",
			},
			notExpected: []string{
				"Include:",
			},
		},
		{
			name:           "Empty repository option",
			repositories:   model.RepositoriesOption{},
			expectedOutput: []string{},
			notExpected: []string{
				"Repositories:",
				"Include:",
				"Exclude:",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			printRepositoriesOption(test.repositories, &buf, 0)
			output := buf.String()

			for _, expected := range test.expectedOutput {
				assert.Contains(t, output, expected)
			}

			for _, notExpected := range test.notExpected {
				assert.NotContains(t, output, notExpected)
			}
		})
	}
}

func TestIsEmptyAuthConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		auth     model.AuthConfig
		expected bool
	}{
		{
			name:     "Completely empty auth config",
			auth:     model.AuthConfig{},
			expected: true,
		},
		{
			name: "Auth config with token",
			auth: model.AuthConfig{
				Token: "some-token",
			},
			expected: false,
		},
		{
			name: "Auth config with username only",
			auth: model.AuthConfig{
				Token: "user-token",
			},
			expected: false,
		},
		{
			name: "Auth config with SSH key",
			auth: model.AuthConfig{
				SSHCommand: "ssh -i ~/.ssh/key",
			},
			expected: false,
		},
		{
			name: "Auth config with advanced settings only",
			auth: model.AuthConfig{
				HTTPScheme: "https",
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := isEmptyAuthConfig(test.auth)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestIsEmptyRepositoriesOption(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		repos    model.RepositoriesOption
		expected bool
	}{
		{
			name:     "Empty repository option",
			repos:    model.RepositoriesOption{},
			expected: true,
		},
		{
			name: "With include list",
			repos: model.RepositoriesOption{
				Include: []string{"repo1"},
			},
			expected: false,
		},
		{
			name: "With exclude list",
			repos: model.RepositoriesOption{
				Exclude: []string{"repo1"},
			},
			expected: false,
		},
		{
			name: "With both lists",
			repos: model.RepositoriesOption{
				Include: []string{"repo1"},
				Exclude: []string{"repo2"},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := isEmptyRepositoriesOption(test.repos)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestIsEmptyMirrorSettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings model.MirrorSettings
		expected bool
	}{
		{
			name:     "Empty mirror settings",
			settings: model.MirrorSettings{},
			expected: true,
		},
		{
			name: "With description prefix",
			settings: model.MirrorSettings{
				DescriptionPrefix: "[MIRROR]",
			},
			expected: false,
		},
		{
			name: "With visibility",
			settings: model.MirrorSettings{
				Visibility: "private",
			},
			expected: false,
		},
		{
			name: "With ignore invalid name",
			settings: model.MirrorSettings{
				IgnoreInvalidName: true,
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := isEmptyMirrorSettings(test.settings)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestPrintConfiguration_WithNilWriter(t *testing.T) {
	t.Parallel()

	// This test ensures the function handles nil writer gracefully
	config := &model.AppConfiguration{
		GitProviderSyncConfs: map[string]model.Environment{
			"test": {
				"main": model.SyncConfig{
					BaseConfig: model.BaseConfig{
						ProviderType: "github",
						Domain:       "github.com",
						Owner:        "test-owner",
					},
				},
			},
		},
	}

	// Should not panic with nil writer
	require.NotPanics(t, func() {
		PrintConfiguration(*config, nil)
	})
}

func TestPrintConfiguration_WithDiscardWriter(t *testing.T) {
	t.Parallel()

	// Test with discard writer to ensure all code paths are executed
	config := &model.AppConfiguration{
		GitProviderSyncConfs: map[string]model.Environment{
			"comprehensive": {
				"full": model.SyncConfig{
					BaseConfig: model.BaseConfig{
						ProviderType: "github",
						Domain:       "github.com",
						Owner:        "test-owner",
						OwnerType:    "group",
						Auth: model.AuthConfig{
							Token:             "test-token",
							SSHCommand:        "ssh-command",
							HTTPScheme:        "https",
							ProxyURL:          "http://proxy.com",
							CertDirPath:       "/certs",
							SSHURLRewriteFrom: "from",
							SSHURLRewriteTo:   "to",
						},
					},
					IncludeForks:    true,
					ActiveFromLimit: "30",
					Repositories: model.RepositoriesOption{
						Include: []string{"include1", "include2"},
						Exclude: []string{"exclude1", "exclude2"},
					},
					Mirrors: map[string]model.MirrorConfig{
						"mirror1": {
							BaseConfig: model.BaseConfig{
								ProviderType: "gitlab",
								Domain:       "gitlab.com",
								Owner:        "mirror-owner",
								Auth: model.AuthConfig{
									Token: "mirror-token",
								},
							},
							Settings: model.MirrorSettings{
								DescriptionPrefix: "[MIRROR]",
								Visibility:        "private",
								IgnoreInvalidName: true,
							},
						},
						"mirror2": {
							BaseConfig: model.BaseConfig{
								ProviderType: "directory",
							},
							Path: "/mirror/path",
						},
						"mirror3": {
							BaseConfig: model.BaseConfig{
								ProviderType: "archive",
							},
							Path: "/archive/path",
						},
					},
				},
			},
		},
	}

	// Should not panic and should execute all code paths
	require.NotPanics(t, func() {
		PrintConfiguration(*config, io.Discard)
	})
}
