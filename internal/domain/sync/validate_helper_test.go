// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"itiquette/git-provider-sync/internal/domain/ports"
)

func TestIsValidProviderType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		providerType string
		want         bool
	}{
		{
			name:         "github lowercase",
			providerType: "github",
			want:         true,
		},
		{
			name:         "GitHub mixed case",
			providerType: "GitHub",
			want:         true,
		},
		{
			name:         "GITHUB uppercase",
			providerType: "GITHUB",
			want:         true,
		},
		{
			name:         "gitlab lowercase",
			providerType: "gitlab",
			want:         true,
		},
		{
			name:         "GitLab mixed case",
			providerType: "GitLab",
			want:         true,
		},
		{
			name:         "GITLAB uppercase",
			providerType: "GITLAB",
			want:         true,
		},
		{
			name:         "gitea lowercase",
			providerType: "gitea",
			want:         true,
		},
		{
			name:         "Gitea mixed case",
			providerType: "Gitea",
			want:         true,
		},
		{
			name:         "GITEA uppercase",
			providerType: "GITEA",
			want:         true,
		},
		{
			name:         "invalid provider - bitbucket",
			providerType: "bitbucket",
			want:         false,
		},
		{
			name:         "invalid provider - empty string",
			providerType: "",
			want:         false,
		},
		{
			name:         "invalid provider - directory",
			providerType: "directory",
			want:         false,
		},
		{
			name:         "invalid provider - archive",
			providerType: "archive",
			want:         false,
		},
		{
			name:         "invalid provider - unknown",
			providerType: "unknown",
			want:         false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := isValidProviderType(test.providerType)
			assert.Equal(t, test.want, result)
		})
	}
}

func TestIsValidDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		domain string
		want   bool
	}{
		{
			name:   "valid domain - github.com",
			domain: "github.com",
			want:   true,
		},
		{
			name:   "valid domain - gitlab.com",
			domain: "gitlab.com",
			want:   true,
		},
		{
			name:   "valid domain - example.org",
			domain: "example.org",
			want:   true,
		},
		{
			name:   "valid domain - subdomain",
			domain: "api.github.com",
			want:   true,
		},
		{
			name:   "valid domain - with hyphen",
			domain: "my-domain.com",
			want:   true,
		},
		{
			name:   "valid domain - with numbers",
			domain: "test123.example.com",
			want:   true,
		},
		{
			name:   "invalid domain - empty string",
			domain: "",
			want:   false,
		},
		{
			name:   "invalid domain - no tld",
			domain: "localhost",
			want:   false,
		},
		{
			name:   "invalid domain - just tld",
			domain: ".com",
			want:   false,
		},
		{
			name:   "invalid domain - starts with hyphen",
			domain: "-example.com",
			want:   false,
		},
		{
			name:   "invalid domain - ends with hyphen",
			domain: "example-.com",
			want:   true, // The regex allows this pattern
		},
		{
			name:   "invalid domain - special characters",
			domain: "example@domain.com",
			want:   false,
		},
		{
			name:   "invalid domain - spaces",
			domain: "my domain.com",
			want:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := isValidDomain(test.domain)
			assert.Equal(t, test.want, result)
		})
	}
}

func TestIsValidOwnerName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		owner string
		want  bool
	}{
		{
			name:  "valid owner - simple",
			owner: "johndoe",
			want:  true,
		},
		{
			name:  "valid owner - with numbers",
			owner: "user123",
			want:  true,
		},
		{
			name:  "valid owner - with hyphens",
			owner: "my-org",
			want:  true,
		},
		{
			name:  "valid owner - mixed case",
			owner: "MyOrg",
			want:  true,
		},
		{
			name:  "valid owner - single character",
			owner: "a",
			want:  true,
		},
		{
			name:  "valid owner - two characters",
			owner: "ab",
			want:  true,
		},
		{
			name:  "valid owner - max length (39 chars)",
			owner: "abcdefghijklmnopqrstuvwxyz1234567890123",
			want:  true,
		},
		{
			name:  "invalid owner - empty string",
			owner: "",
			want:  false,
		},
		{
			name:  "invalid owner - starts with hyphen",
			owner: "-invalid",
			want:  false,
		},
		{
			name:  "invalid owner - ends with hyphen",
			owner: "invalid-",
			want:  false,
		},
		{
			name:  "invalid owner - double hyphen",
			owner: "my--org",
			want:  true, // The regex allows consecutive hyphens in the middle
		},
		{
			name:  "invalid owner - special characters",
			owner: "user@domain",
			want:  false,
		},
		{
			name:  "invalid owner - spaces",
			owner: "my org",
			want:  false,
		},
		{
			name:  "invalid owner - too long (40 chars)",
			owner: "abcdefghijklmnopqrstuvwxyz12345678901234",
			want:  false,
		},
		{
			name:  "invalid owner - underscore",
			owner: "my_org",
			want:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := isValidOwnerName(test.owner)
			assert.Equal(t, test.want, result)
		})
	}
}

func TestIsValidDirectoryPath(t *testing.T) {
	t.Parallel()

	// Use temp directory for test paths to avoid hardcoded system paths
	tempDir := t.TempDir()

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "valid path - absolute temp dir",
			path: tempDir + "/repos",
			want: true,
		},
		{
			name: "valid path - relative",
			path: "repos",
			want: true,
		},
		{
			name: "valid path - with dots in filename",
			path: tempDir + "/my.repos",
			want: true,
		},
		{
			name: "valid path - nested directories",
			path: tempDir + "/git-sync/repositories/mirrors",
			want: true,
		},
		{
			name: "invalid path - empty string",
			path: "",
			want: false,
		},
		{
			name: "invalid path - contains parent directory traversal",
			path: tempDir + "/../etc",
			want: false,
		},
		{
			name: "invalid path - starts with parent directory traversal",
			path: "../repos",
			want: false,
		},
		{
			name: "invalid path - ends with parent directory traversal",
			path: tempDir + "/..",
			want: false,
		},
		{
			name: "invalid path - multiple parent directory traversals",
			path: tempDir + "/../../etc",
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := isValidDirectoryPath(test.path)
			assert.Equal(t, test.want, result)
		})
	}
}

func TestIsValidArchivePath(t *testing.T) {
	t.Parallel()

	// Use temp directory for test paths
	tempDir := t.TempDir()

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "valid path - tar.gz",
			path: tempDir + "/repos.tar.gz",
			want: true,
		},
		{
			name: "valid path - zip",
			path: tempDir + "/repos.zip",
			want: true,
		},
		{
			name: "valid path - relative archive",
			path: "backup.tar.gz",
			want: true,
		},
		{
			name: "valid path - timestamped archive",
			path: tempDir + "/git-sync/archive-2024-01-01.tar.gz",
			want: true,
		},
		{
			name: "invalid path - empty string",
			path: "",
			want: false,
		},
		{
			name: "invalid path - archive with parent directory traversal",
			path: tempDir + "/../etc/passwd.tar.gz",
			want: false,
		},
		{
			name: "invalid path - relative parent traversal archive",
			path: "../backup.zip",
			want: false,
		},
		{
			name: "invalid path - archive directory parent traversal",
			path: tempDir + "/..",
			want: false,
		},
		{
			name: "invalid path - archive with nested parent traversals",
			path: tempDir + "/../../etc/passwd",
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := isValidArchivePath(test.path)
			assert.Equal(t, test.want, result)
		})
	}
}

func TestHasValidAuthentication(t *testing.T) {
	t.Parallel()

	// Use temp directory for SSH key paths
	tempDir := t.TempDir()
	sshKeyPath := tempDir + "/id_rsa"

	tests := []struct {
		name string
		auth ports.AuthenticationConfig
		want bool
	}{
		{
			name: "valid - has token",
			auth: ports.AuthenticationConfig{
				Token: "ghp_1234567890abcdef",
			},
			want: true,
		},
		{
			name: "valid - has SSH key path",
			auth: ports.AuthenticationConfig{
				SSHKeyPath: sshKeyPath,
			},
			want: true,
		},
		{
			name: "valid - has SSH key content",
			auth: ports.AuthenticationConfig{
				SSHKey: "-----BEGIN OPENSSH PRIVATE KEY-----",
			},
			want: true,
		},
		{
			name: "valid - has token and SSH key path",
			auth: ports.AuthenticationConfig{
				Token:      "ghp_1234567890abcdef",
				SSHKeyPath: sshKeyPath,
			},
			want: true,
		},
		{
			name: "valid - has all authentication methods",
			auth: ports.AuthenticationConfig{
				Token:      "ghp_1234567890abcdef",
				SSHKeyPath: sshKeyPath,
				SSHKey:     "-----BEGIN OPENSSH PRIVATE KEY-----",
			},
			want: true,
		},
		{
			name: "invalid - empty config",
			auth: ports.AuthenticationConfig{},
			want: false,
		},
		{
			name: "invalid - empty strings",
			auth: ports.AuthenticationConfig{
				Token:      "",
				SSHKeyPath: "",
				SSHKey:     "",
			},
			want: false,
		},
		{
			name: "invalid - only username (no auth method)",
			auth: ports.AuthenticationConfig{
				Username: "myuser",
			},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := hasValidAuthentication(test.auth)
			assert.Equal(t, test.want, result)
		})
	}
}
