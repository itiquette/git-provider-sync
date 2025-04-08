// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package gitlib

import (
	"testing"

	"github.com/go-git/go-git/v5"
	gogitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_buildCloneOptions(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		mirror bool
		auth   transport.AuthMethod
		want   *git.CloneOptions
	}{
		{
			name:   "basic clone without auth",
			url:    "https://github.com/test/repo.git",
			mirror: false,
			auth:   nil,
			want: &git.CloneOptions{
				URL:    "https://github.com/test/repo.git",
				Mirror: false,
				Auth:   nil,
			},
		},
		{
			name:   "mirror clone with auth",
			url:    "https://github.com/test/repo.git",
			mirror: true,
			auth: &http.BasicAuth{
				Username: "user",
				Password: "pass",
			},
			want: &git.CloneOptions{
				URL:    "https://github.com/test/repo.git",
				Mirror: true,
				Auth: &http.BasicAuth{
					Username: "user",
					Password: "pass",
				},
			},
		},
		{
			name:   "empty url",
			url:    "",
			mirror: false,
			auth:   nil,
			want: &git.CloneOptions{
				URL:    "",
				Mirror: false,
				Auth:   nil,
			},
		},
	}

	s := &Service{}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			got := s.buildCloneOptions(tabletest.url, tabletest.mirror, tabletest.auth)
			require.Equal(t, tabletest.want.URL, got.URL)
			require.Equal(t, tabletest.want.Mirror, got.Mirror)
			require.Equal(t, tabletest.want.Auth, got.Auth)
		})
	}
}

func TestService_buildPullOptions(t *testing.T) {
	tests := []struct {
		name   string
		remote string
		url    string
		auth   transport.AuthMethod
		want   *git.PullOptions
	}{
		{
			name:   "pull without auth",
			remote: "origin",
			url:    "https://github.com/test/repo.git",
			auth:   nil,
			want: &git.PullOptions{
				RemoteName: "origin",
				RemoteURL:  "https://github.com/test/repo.git",
				Auth:       nil,
			},
		},
		{
			name:   "pull with auth",
			remote: "upstream",
			url:    "https://github.com/test/repo.git",
			auth: &http.BasicAuth{
				Username: "user",
				Password: "pass",
			},
			want: &git.PullOptions{
				RemoteName: "upstream",
				RemoteURL:  "https://github.com/test/repo.git",
				Auth: &http.BasicAuth{
					Username: "user",
					Password: "pass",
				},
			},
		},
		{
			name:   "empty remote and url",
			remote: "",
			url:    "",
			auth:   nil,
			want: &git.PullOptions{
				RemoteName: "",
				RemoteURL:  "",
				Auth:       nil,
			},
		},
	}

	s := &Service{}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			got := s.buildPullOptions(tabletest.remote, tabletest.url, tabletest.auth)
			require.Equal(t, tabletest.want.RemoteName, got.RemoteName)
			require.Equal(t, tabletest.want.RemoteURL, got.RemoteURL)
			require.Equal(t, tabletest.want.Auth, got.Auth)
		})
	}
}

func TestService_buildPushOptions(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		refSpec []string
		prune   bool
		auth    transport.AuthMethod
		want    git.PushOptions
	}{
		{
			name:    "push without refspecs",
			url:     "https://github.com/test/repo.git",
			refSpec: []string{},
			prune:   false,
			auth:    nil,
			want: git.PushOptions{
				RemoteURL: "https://github.com/test/repo.git",
				RefSpecs:  []gogitconfig.RefSpec{},
				Prune:     false,
				Auth:      nil,
			},
		},
		{
			name:    "push with single refspec",
			url:     "https://github.com/test/repo.git",
			refSpec: []string{"refs/heads/main:refs/heads/main"},
			prune:   true,
			auth: &http.BasicAuth{
				Username: "user",
				Password: "pass",
			},
			want: git.PushOptions{
				RemoteURL: "https://github.com/test/repo.git",
				RefSpecs:  []gogitconfig.RefSpec{"refs/heads/main:refs/heads/main"},
				Prune:     true,
				Auth: &http.BasicAuth{
					Username: "user",
					Password: "pass",
				},
			},
		},
		{
			name:    "push with multiple refspecs",
			url:     "https://github.com/test/repo.git",
			refSpec: []string{"refs/heads/*:refs/heads/*", "refs/tags/*:refs/tags/*"},
			prune:   false,
			auth:    nil,
			want: git.PushOptions{
				RemoteURL: "https://github.com/test/repo.git",
				RefSpecs:  []gogitconfig.RefSpec{"refs/heads/*:refs/heads/*", "refs/tags/*:refs/tags/*"},
				Prune:     false,
				Auth:      nil,
			},
		},
		{
			name:    "empty url and refspecs",
			url:     "",
			refSpec: []string{},
			prune:   false,
			auth:    nil,
			want: git.PushOptions{
				RemoteURL: "",
				RefSpecs:  []gogitconfig.RefSpec{},
				Prune:     false,
				Auth:      nil,
			},
		},
	}

	s := &Service{}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			got := s.buildPushOptions(tabletest.url, tabletest.refSpec, tabletest.prune, tabletest.auth)
			require.Equal(t, tabletest.want.RemoteURL, got.RemoteURL)
			require.Equal(t, tabletest.want.RefSpecs, got.RefSpecs)
			require.Equal(t, tabletest.want.Prune, got.Prune)
			require.Equal(t, tabletest.want.Auth, got.Auth)
		})
	}
}
func TestService_buildPushOptions_RefSpecs(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		refSpec []string
		prune   bool
		auth    transport.AuthMethod
		want    git.PushOptions
	}{
		{
			name:    "single branch push refspec",
			url:     "https://github.com/test/repo.git",
			refSpec: []string{"refs/heads/main:refs/heads/main"},
			prune:   false,
			auth:    nil,
			want: git.PushOptions{
				RemoteURL: "https://github.com/test/repo.git",
				RefSpecs:  []gogitconfig.RefSpec{"refs/heads/main:refs/heads/main"},
				Prune:     false,
				Auth:      nil,
			},
		},
		{
			name:    "wildcard all branches refspec",
			url:     "https://github.com/test/repo.git",
			refSpec: []string{"refs/heads/*:refs/heads/*"},
			prune:   false,
			auth:    nil,
			want: git.PushOptions{
				RemoteURL: "https://github.com/test/repo.git",
				RefSpecs:  []gogitconfig.RefSpec{"refs/heads/*:refs/heads/*"},
				Prune:     false,
				Auth:      nil,
			},
		},
		{
			name: "multiple specific branches refspec",
			url:  "https://github.com/test/repo.git",
			refSpec: []string{
				"refs/heads/main:refs/heads/main",
				"refs/heads/develop:refs/heads/develop",
				"refs/heads/feature/*:refs/heads/feature/*",
			},
			prune: false,
			auth:  nil,
			want: git.PushOptions{
				RemoteURL: "https://github.com/test/repo.git",
				RefSpecs: []gogitconfig.RefSpec{
					"refs/heads/main:refs/heads/main",
					"refs/heads/develop:refs/heads/develop",
					"refs/heads/feature/*:refs/heads/feature/*",
				},
				Prune: false,
				Auth:  nil,
			},
		},
		{
			name: "tags refspec",
			url:  "https://github.com/test/repo.git",
			refSpec: []string{
				"refs/tags/*:refs/tags/*",
			},
			prune: false,
			auth:  nil,
			want: git.PushOptions{
				RemoteURL: "https://github.com/test/repo.git",
				RefSpecs:  []gogitconfig.RefSpec{"refs/tags/*:refs/tags/*"},
				Prune:     false,
				Auth:      nil,
			},
		},
		{
			name: "combined branches and tags refspec",
			url:  "https://github.com/test/repo.git",
			refSpec: []string{
				"refs/heads/*:refs/heads/*",
				"refs/tags/*:refs/tags/*",
			},
			prune: true,
			auth: &http.BasicAuth{
				Username: "user",
				Password: "pass",
			},
			want: git.PushOptions{
				RemoteURL: "https://github.com/test/repo.git",
				RefSpecs: []gogitconfig.RefSpec{
					"refs/heads/*:refs/heads/*",
					"refs/tags/*:refs/tags/*",
				},
				Prune: true,
				Auth: &http.BasicAuth{
					Username: "user",
					Password: "pass",
				},
			},
		},
		{
			name: "branch rename refspec",
			url:  "https://github.com/test/repo.git",
			refSpec: []string{
				"refs/heads/source:refs/heads/destination",
			},
			prune: false,
			auth:  nil,
			want: git.PushOptions{
				RemoteURL: "https://github.com/test/repo.git",
				RefSpecs:  []gogitconfig.RefSpec{"refs/heads/source:refs/heads/destination"},
				Prune:     false,
				Auth:      nil,
			},
		},
		{
			name: "delete branch refspec",
			url:  "https://github.com/test/repo.git",
			refSpec: []string{
				":refs/heads/branch-to-delete",
			},
			prune: false,
			auth:  nil,
			want: git.PushOptions{
				RemoteURL: "https://github.com/test/repo.git",
				RefSpecs:  []gogitconfig.RefSpec{":refs/heads/branch-to-delete"},
				Prune:     false,
				Auth:      nil,
			},
		},
		{
			name:    "empty refspec slice",
			url:     "https://github.com/test/repo.git",
			refSpec: []string{},
			prune:   false,
			auth:    nil,
			want: git.PushOptions{
				RemoteURL: "https://github.com/test/repo.git",
				RefSpecs:  []gogitconfig.RefSpec{},
				Prune:     false,
				Auth:      nil,
			},
		},
		{
			name: "mixed valid and directory refspecs",
			url:  "https://github.com/test/repo.git",
			refSpec: []string{
				"refs/heads/main:refs/heads/main",
				"refs/heads/feature-*:refs/heads/feature-*",
				"refs/environments/*/main:refs/environments/*/main",
			},
			prune: false,
			auth:  nil,
			want: git.PushOptions{
				RemoteURL: "https://github.com/test/repo.git",
				RefSpecs: []gogitconfig.RefSpec{
					"refs/heads/main:refs/heads/main",
					"refs/heads/feature-*:refs/heads/feature-*",
					"refs/environments/*/main:refs/environments/*/main",
				},
				Prune: false,
				Auth:  nil,
			},
		},
		{
			name: "refspecs with plus force notation",
			url:  "https://github.com/test/repo.git",
			refSpec: []string{
				"+refs/heads/*:refs/heads/*",
				"+refs/heads/feature:refs/heads/feature",
			},
			prune: false,
			auth:  nil,
			want: git.PushOptions{
				RemoteURL: "https://github.com/test/repo.git",
				RefSpecs: []gogitconfig.RefSpec{
					"+refs/heads/*:refs/heads/*",
					"+refs/heads/feature:refs/heads/feature",
				},
				Prune: false,
				Auth:  nil,
			},
		},
	}

	s := &Service{}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			got := s.buildPushOptions(tabletest.url, tabletest.refSpec, tabletest.prune, tabletest.auth)
			assert.Equal(t, tabletest.want.RemoteURL, got.RemoteURL)
			assert.Equal(t, tabletest.want.RefSpecs, got.RefSpecs)
			assert.Equal(t, tabletest.want.Prune, got.Prune)
			assert.Equal(t, tabletest.want.Auth, got.Auth)

			// Additional checks for refspec validation
			for i, refSpec := range got.RefSpecs {
				// Check that the refspec was properly converted to the RefSpec type
				assert.Equal(t, gogitconfig.RefSpec(tabletest.refSpec[i]), refSpec)
			}
		})
	}
}
