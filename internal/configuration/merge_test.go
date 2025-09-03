// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package configuration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"itiquette/git-provider-sync/internal/configuration"
	config "itiquette/git-provider-sync/internal/model/configuration"
)

func TestMergeConfigs_Priority(t *testing.T) {
	t.Parallel()

	// Create base config
	base := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"prod": {
				"github": config.SyncConfig{
					BaseConfig: config.BaseConfig{
						Domain:       "github.com",
						Owner:        "base-owner",
						ProviderType: "github",
						Auth: config.AuthConfig{
							Token: "base-token",
						},
					},
				},
			},
		},
	}

	// Create override config
	override := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"prod": {
				"github": config.SyncConfig{
					BaseConfig: config.BaseConfig{
						Domain: "github.enterprise.com",
						Owner:  "override-owner",
						Auth: config.AuthConfig{
							Token: "override-token",
						},
					},
				},
			},
		},
	}

	// Test priority merging
	sources := []configuration.ConfigSource{
		{Name: "base", Config: base, Priority: 1},
		{Name: "override", Config: override, Priority: 2},
	}

	result := configuration.MergeConfigs(sources)

	// Verify override wins
	assert.Equal(t, "github.enterprise.com", result.GitProviderSyncConfs["prod"]["github"].Domain)
	assert.Equal(t, "override-owner", result.GitProviderSyncConfs["prod"]["github"].Owner)
	assert.Equal(t, "override-token", result.GitProviderSyncConfs["prod"]["github"].Auth.Token)
	assert.Equal(t, "github", result.GitProviderSyncConfs["prod"]["github"].ProviderType) // Not overridden
}

func TestMergeConfigs_MultipleEnvironments(t *testing.T) {
	t.Parallel()

	base := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"dev": {
				"source1": config.SyncConfig{
					BaseConfig: config.BaseConfig{Domain: "dev.github.com"},
				},
			},
			"prod": {
				"source1": config.SyncConfig{
					BaseConfig: config.BaseConfig{Domain: "prod.github.com"},
				},
			},
		},
	}

	override := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"prod": {
				"source1": config.SyncConfig{
					BaseConfig: config.BaseConfig{Domain: "prod-override.github.com"},
				},
				"source2": config.SyncConfig{
					BaseConfig: config.BaseConfig{Domain: "new.github.com"},
				},
			},
			"staging": {
				"source1": config.SyncConfig{
					BaseConfig: config.BaseConfig{Domain: "staging.github.com"},
				},
			},
		},
	}

	sources := []configuration.ConfigSource{
		{Name: "base", Config: base, Priority: 1},
		{Name: "override", Config: override, Priority: 2},
	}

	result := configuration.MergeConfigs(sources)

	// Check all environments are present
	assert.Len(t, result.GitProviderSyncConfs, 3)
	assert.Contains(t, result.GitProviderSyncConfs, "dev")
	assert.Contains(t, result.GitProviderSyncConfs, "prod")
	assert.Contains(t, result.GitProviderSyncConfs, "staging")

	// Check specific values
	assert.Equal(t, "dev.github.com", result.GitProviderSyncConfs["dev"]["source1"].Domain)
	assert.Equal(t, "prod-override.github.com", result.GitProviderSyncConfs["prod"]["source1"].Domain)
	assert.Equal(t, "new.github.com", result.GitProviderSyncConfs["prod"]["source2"].Domain)
	assert.Equal(t, "staging.github.com", result.GitProviderSyncConfs["staging"]["source1"].Domain)
}

func TestMergeConfigs_Mirrors(t *testing.T) {
	t.Parallel()

	base := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"prod": {
				"github": config.SyncConfig{
					BaseConfig: config.BaseConfig{Domain: "github.com"},
					Mirrors: map[string]config.MirrorConfig{
						"gitlab": {
							BaseConfig: config.BaseConfig{
								Domain: "gitlab.com",
								Owner:  "base-mirror",
							},
							Path: "/base/path",
						},
					},
				},
			},
		},
	}

	override := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"prod": {
				"github": config.SyncConfig{
					Mirrors: map[string]config.MirrorConfig{
						"gitlab": {
							BaseConfig: config.BaseConfig{
								Owner: "override-mirror",
							},
							Path: "/override/path",
						},
						"gitea": {
							BaseConfig: config.BaseConfig{
								Domain: "gitea.com",
							},
						},
					},
				},
			},
		},
	}

	sources := []configuration.ConfigSource{
		{Name: "base", Config: base, Priority: 1},
		{Name: "override", Config: override, Priority: 2},
	}

	result := configuration.MergeConfigs(sources)

	mirrors := result.GitProviderSyncConfs["prod"]["github"].Mirrors
	assert.Len(t, mirrors, 2)

	// Check gitlab mirror was merged
	assert.Equal(t, "gitlab.com", mirrors["gitlab"].Domain)
	assert.Equal(t, "override-mirror", mirrors["gitlab"].Owner)
	assert.Equal(t, "/override/path", mirrors["gitlab"].Path)

	// Check gitea mirror was added
	assert.Equal(t, "gitea.com", mirrors["gitea"].Domain)
}

func TestMergeConfigs_EmptyAndNil(t *testing.T) {
	t.Parallel()

	// Test empty sources
	result := configuration.MergeConfigs([]configuration.ConfigSource{})
	assert.NotNil(t, result)
	assert.NotNil(t, result.GitProviderSyncConfs)
	assert.Empty(t, result.GitProviderSyncConfs)

	// Test nil configs
	sources := []configuration.ConfigSource{
		{Name: "nil", Config: nil, Priority: 1},
		{Name: "empty", Config: &config.AppConfiguration{}, Priority: 2},
	}
	result = configuration.MergeConfigs(sources)
	assert.NotNil(t, result)
	assert.NotNil(t, result.GitProviderSyncConfs)
}

func TestMergeConfigs_AuthConfig(t *testing.T) {
	t.Parallel()

	base := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"prod": {
				"source": config.SyncConfig{
					BaseConfig: config.BaseConfig{
						Auth: config.AuthConfig{
							Token:          "base-token",
							HTTPScheme:     "https",
							RequestTimeout: 30,
							CertDirPath:    "/base/certs",
						},
					},
				},
			},
		},
	}

	override := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"prod": {
				"source": config.SyncConfig{
					BaseConfig: config.BaseConfig{
						Auth: config.AuthConfig{
							Token:          "override-token",
							RequestTimeout: 60,
							ProxyURL:       "http://proxy:8080",
							// HTTPScheme not set - should keep base
							// CertDirPath overridden with empty - should keep base
						},
					},
				},
			},
		},
	}

	sources := []configuration.ConfigSource{
		{Name: "base", Config: base, Priority: 1},
		{Name: "override", Config: override, Priority: 2},
	}

	result := configuration.MergeConfigs(sources)

	auth := result.GitProviderSyncConfs["prod"]["source"].Auth
	assert.Equal(t, "override-token", auth.Token)
	assert.Equal(t, "https", auth.HTTPScheme)           // Kept from base
	assert.Equal(t, 60, auth.RequestTimeout)            // Overridden
	assert.Equal(t, "/base/certs", auth.CertDirPath)    // Kept from base
	assert.Equal(t, "http://proxy:8080", auth.ProxyURL) // New from override
}

func TestMergeConfigs_Repositories(t *testing.T) {
	t.Parallel()

	base := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"prod": {
				"source": config.SyncConfig{
					Repositories: config.RepositoriesOption{
						Include: []string{"repo1", "repo2"},
						Exclude: []string{"excluded1"},
					},
				},
			},
		},
	}

	override := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"prod": {
				"source": config.SyncConfig{
					Repositories: config.RepositoriesOption{
						Include: []string{"repo3", "repo4"},
						// Exclude not set - should keep base
					},
				},
			},
		},
	}

	sources := []configuration.ConfigSource{
		{Name: "base", Config: base, Priority: 1},
		{Name: "override", Config: override, Priority: 2},
	}

	result := configuration.MergeConfigs(sources)

	repos := result.GitProviderSyncConfs["prod"]["source"].Repositories
	// Include completely replaced
	assert.Equal(t, []string{"repo3", "repo4"}, repos.Include)
	// Exclude kept from base
	assert.Equal(t, []string{"excluded1"}, repos.Exclude)
}

func TestProcessRepositoryLists(t *testing.T) {
	t.Parallel()

	cfg := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"prod": {
				"source": config.SyncConfig{
					Repositories: config.RepositoriesOption{
						Include: []string{"repo1,repo2,repo3"},
						Exclude: []string{"exclude1, exclude2 , exclude3"},
					},
				},
			},
		},
	}

	result := configuration.ProcessRepositoryLists(cfg)

	// Check original not modified
	assert.Equal(t, []string{"repo1,repo2,repo3"}, cfg.GitProviderSyncConfs["prod"]["source"].Repositories.Include)

	// Check result is processed
	repos := result.GitProviderSyncConfs["prod"]["source"].Repositories
	assert.Equal(t, []string{"repo1", "repo2", "repo3"}, repos.Include)
	assert.Equal(t, []string{"exclude1", "exclude2", "exclude3"}, repos.Exclude)
}

func TestExpandVariables(t *testing.T) {
	t.Parallel()

	cfg := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"prod": {
				"source": config.SyncConfig{
					BaseConfig: config.BaseConfig{
						Auth: config.AuthConfig{
							Token:       "${GITHUB_TOKEN}",
							TokenFile:   "$HOME/.tokens/github",
							ProxyURL:    "http://${PROXY_HOST}:${PROXY_PORT}",
							CertDirPath: "/etc/ssl/certs",
						},
					},
					Mirrors: map[string]config.MirrorConfig{
						"backup": {
							Path: "${BACKUP_DIR}/repos",
							BaseConfig: config.BaseConfig{
								Auth: config.AuthConfig{
									Token: "${GITLAB_TOKEN}",
								},
							},
						},
					},
				},
			},
		},
	}

	envVars := map[string]string{
		"GITHUB_TOKEN": "ghp_123456",
		"GITLAB_TOKEN": "glpat_789",
		"HOME":         "/home/user",
		"PROXY_HOST":   "proxy.example.com",
		"PROXY_PORT":   "8080",
		"BACKUP_DIR":   "/mnt/backup",
	}

	result := configuration.ExpandVariables(cfg, envVars)

	// Check expansions
	auth := result.GitProviderSyncConfs["prod"]["source"].Auth
	assert.Equal(t, "ghp_123456", auth.Token)
	assert.Equal(t, "/home/user/.tokens/github", auth.TokenFile)
	assert.Equal(t, "http://proxy.example.com:8080", auth.ProxyURL)
	assert.Equal(t, "/etc/ssl/certs", auth.CertDirPath) // No variable, unchanged

	// Check mirror expansions
	mirror := result.GitProviderSyncConfs["prod"]["source"].Mirrors["backup"]
	assert.Equal(t, "/mnt/backup/repos", mirror.Path)
	assert.Equal(t, "glpat_789", mirror.Auth.Token)
}

func TestApplyProviderTokens(t *testing.T) {
	t.Parallel()

	cfg := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"prod": {
				"github-source": config.SyncConfig{
					BaseConfig: config.BaseConfig{
						ProviderType: "github",
						Auth: config.AuthConfig{
							Token:     "old-github-token",
							TokenFile: "/path/to/token",
						},
					},
					Mirrors: map[string]config.MirrorConfig{
						"gitlab-mirror": {
							BaseConfig: config.BaseConfig{
								ProviderType: "gitlab",
								Auth: config.AuthConfig{
									Token: "old-gitlab-token",
								},
							},
						},
						"gitea-mirror": {
							BaseConfig: config.BaseConfig{
								ProviderType: "gitea",
								Auth: config.AuthConfig{
									TokenFile: "/path/to/gitea/token",
								},
							},
						},
					},
				},
			},
		},
	}

	providerTokens := map[string]string{
		"GPS_GITHUB_TOKEN": "new-github-token",
		"GPS_GITLAB_TOKEN": "new-gitlab-token",
		"GPS_GITEA_TOKEN":  "new-gitea-token",
	}

	result := configuration.ApplyProviderTokens(cfg, providerTokens)

	// Check tokens were applied
	source := result.GitProviderSyncConfs["prod"]["github-source"]
	assert.Equal(t, "new-github-token", source.Auth.Token)
	assert.Empty(t, source.Auth.TokenFile) // Cleared when env token set

	// Check mirror tokens
	gitlab := source.Mirrors["gitlab-mirror"]
	assert.Equal(t, "new-gitlab-token", gitlab.Auth.Token)

	gitea := source.Mirrors["gitea-mirror"]
	assert.Equal(t, "new-gitea-token", gitea.Auth.Token)
	assert.Empty(t, gitea.Auth.TokenFile) // Cleared
}

func TestMergeConfigs_BooleanFields(t *testing.T) {
	t.Parallel()

	base := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"prod": {
				"source": config.SyncConfig{
					BaseConfig: config.BaseConfig{
						UseGitBinary: false,
					},
					IncludeForks: true,
				},
			},
		},
	}

	override := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"prod": {
				"source": config.SyncConfig{
					BaseConfig: config.BaseConfig{
						UseGitBinary: true, // Override to true
					},
					IncludeForks: false, // Override to false
				},
			},
		},
	}

	sources := []configuration.ConfigSource{
		{Name: "base", Config: base, Priority: 1},
		{Name: "override", Config: override, Priority: 2},
	}

	result := configuration.MergeConfigs(sources)

	source := result.GitProviderSyncConfs["prod"]["source"]
	assert.True(t, source.UseGitBinary)  // Changed from false to true
	assert.False(t, source.IncludeForks) // Changed from true to false
}

func TestMergeConfigs_MirrorSettings(t *testing.T) {
	t.Parallel()

	base := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"prod": {
				"source": config.SyncConfig{
					Mirrors: map[string]config.MirrorConfig{
						"mirror1": {
							Settings: config.MirrorSettings{
								AlphaNumHyphName:  true,
								DescriptionPrefix: "Mirror: ",
								Disabled:          false,
								ForcePush:         false,
								IgnoreInvalidName: true,
								Visibility:        "private",
							},
						},
					},
				},
			},
		},
	}

	override := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"prod": {
				"source": config.SyncConfig{
					Mirrors: map[string]config.MirrorConfig{
						"mirror1": {
							Settings: config.MirrorSettings{
								AlphaNumHyphName:  false,                       // Override
								DescriptionPrefix: "Backup: ",                  // Override
								Disabled:          true,                        // Override
								ForcePush:         true,                        // Override
								IgnoreInvalidName: false,                       // Override to false (Go zero value)
								Visibility:        "public",                    // Override
								GitHubUploadURL:   "https://upload.github.com", // New
							},
						},
					},
				},
			},
		},
	}

	sources := []configuration.ConfigSource{
		{Name: "base", Config: base, Priority: 1},
		{Name: "override", Config: override, Priority: 2},
	}

	result := configuration.MergeConfigs(sources)

	settings := result.GitProviderSyncConfs["prod"]["source"].Mirrors["mirror1"].Settings
	assert.False(t, settings.AlphaNumHyphName)
	assert.Equal(t, "Backup: ", settings.DescriptionPrefix)
	assert.True(t, settings.Disabled)
	assert.True(t, settings.ForcePush)
	assert.False(t, settings.IgnoreInvalidName) // Overridden to false (Go zero value)
	assert.Equal(t, "public", settings.Visibility)
	assert.Equal(t, "https://upload.github.com", settings.GitHubUploadURL)
}

func TestMergeConfigs_Immutability(t *testing.T) {
	t.Parallel()

	// Create original configs
	base := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"prod": {
				"source": config.SyncConfig{
					BaseConfig: config.BaseConfig{
						Domain: "original.com",
					},
				},
			},
		},
	}

	override := &config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"prod": {
				"source": config.SyncConfig{
					BaseConfig: config.BaseConfig{
						Domain: "override.com",
					},
				},
			},
		},
	}

	sources := []configuration.ConfigSource{
		{Name: "base", Config: base, Priority: 1},
		{Name: "override", Config: override, Priority: 2},
	}

	// Perform merge
	result := configuration.MergeConfigs(sources)

	// Verify result
	assert.Equal(t, "override.com", result.GitProviderSyncConfs["prod"]["source"].Domain)

	// Verify originals unchanged
	assert.Equal(t, "original.com", base.GitProviderSyncConfs["prod"]["source"].Domain)
	assert.Equal(t, "override.com", override.GitProviderSyncConfs["prod"]["source"].Domain)

	// Modify result (need to get, modify, and put back due to Go's map semantics)
	modifiedSource := result.GitProviderSyncConfs["prod"]["source"]
	modifiedSource.Domain = "modified.com"
	result.GitProviderSyncConfs["prod"]["source"] = modifiedSource

	// Verify originals still unchanged
	assert.Equal(t, "original.com", base.GitProviderSyncConfs["prod"]["source"].Domain)
	assert.Equal(t, "override.com", override.GitProviderSyncConfs["prod"]["source"].Domain)
}
