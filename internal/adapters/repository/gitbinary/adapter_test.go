// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package gitbinary

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"itiquette/git-provider-sync/internal/adapters/logging"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

func TestNew_ValidConfig_CreatesGitBinaryAdapter(t *testing.T) {
	t.Parallel()

	config := ports.GitConfig{
		UserName:  "test",
		UserEmail: "test@example.com",
	}

	adapter := New(config)
	assert.NotNil(t, adapter)
	assert.Equal(t, config, adapter.config)
	assert.False(t, adapter.initialized)
	assert.Empty(t, adapter.tempDir)
	assert.Nil(t, adapter.mirrorSvc)
}

func TestInitialize(t *testing.T) {
	t.Parallel()

	config := ports.GitConfig{
		UserName:  "test",
		UserEmail: "test@example.com",
	}

	adapter := New(config)
	zerologInstance := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	logger := logging.NewZerologAdapter(&zerologInstance)
	ctx := context.Background()

	// First initialization should succeed
	err := adapter.Initialize(ctx, logger)
	require.NoError(t, err)
	assert.True(t, adapter.initialized)
	assert.NotEmpty(t, adapter.tempDir)
	assert.NotNil(t, adapter.mirrorSvc)

	// Second initialization should not fail (already initialized)
	err = adapter.Initialize(ctx, logger)
	require.NoError(t, err)
	assert.True(t, adapter.initialized)
}

func TestGetName(t *testing.T) {
	t.Parallel()

	adapter := New(ports.GitConfig{})
	assert.Equal(t, "git-binary", adapter.GetName())
}

func TestSupportsURL(t *testing.T) {
	t.Parallel()

	adapter := New(ports.GitConfig{})

	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "https url",
			url:  "https://github.com/test/repo.git",
			want: true,
		},
		{
			name: "ssh url",
			url:  "git@github.com:test/repo.git",
			want: true,
		},
		{
			name: "ssh protocol url",
			url:  "ssh://git@github.com/test/repo.git",
			want: true,
		},
		{
			name: "file url",
			url:  "file:///path/to/repo.git",
			want: true,
		},
		{
			name: "local path",
			url:  "/path/to/repo.git",
			want: true,
		},
		{
			name: "empty url",
			url:  "",
			want: true, // git binary supports all URLs
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := adapter.SupportsURL(tt.url)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConvertAuthOptions(t *testing.T) {
	t.Parallel()

	adapter := New(ports.GitConfig{})

	tests := []struct {
		name     string
		authOpts ports.AuthOptions
		want     AuthConfig
	}{
		{
			name: "none auth",
			authOpts: ports.AuthOptions{
				Type: ports.AuthTypeNone,
			},
			want: AuthConfig{
				Protocol: "https",
			},
		},
		{
			name: "basic auth",
			authOpts: ports.AuthOptions{
				Type:     ports.AuthTypeBasic,
				Username: "user",
				Password: "pass",
			},
			want: AuthConfig{
				Protocol: "https",
			},
		},
		{
			name: "token auth",
			authOpts: ports.AuthOptions{
				Type:  ports.AuthTypeToken,
				Token: "token123",
			},
			want: AuthConfig{
				Protocol: "https",
				Token:    "token123",
			},
		},
		{
			name: "ssh auth",
			authOpts: ports.AuthOptions{
				Type: ports.AuthTypeSSH,
			},
			want: AuthConfig{
				Protocol: "ssh",
			},
		},
		{
			name: "ssh key auth",
			authOpts: ports.AuthOptions{
				Type:   ports.AuthTypeSSHKey,
				SSHKey: []byte("ssh-key-content"),
			},
			want: AuthConfig{
				Protocol:   "ssh",
				SSHCommand: "ssh -i /tmp/ssh_key",
			},
		},
		{
			name: "ssh agent auth",
			authOpts: ports.AuthOptions{
				Type:   ports.AuthTypeSSHAgent,
				SSHKey: []byte("ssh-key-content"),
			},
			want: AuthConfig{
				Protocol:   "ssh",
				SSHCommand: "ssh -i /tmp/ssh_key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := adapter.convertAuthOptions(tt.authOpts)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDetermineMirrorType(t *testing.T) {
	t.Parallel()

	adapter := New(ports.GitConfig{})

	tests := []struct {
		name    string
		options ports.CloneOptions
		want    string
	}{
		{
			name: "mirror clone",
			options: ports.CloneOptions{
				Mirror: true,
			},
			want: "mirror",
		},
		{
			name: "bare clone",
			options: ports.CloneOptions{
				Bare: true,
			},
			want: "bare",
		},
		{
			name: "shallow clone",
			options: ports.CloneOptions{
				Depth: 1,
			},
			want: "shallow",
		},
		{
			name: "full clone",
			options: ports.CloneOptions{
				Depth:  0,
				Mirror: false,
				Bare:   false,
			},
			want: "full",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := adapter.determineMirrorType(tt.options)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDetectSourceType(t *testing.T) {
	t.Parallel()

	adapter := New(ports.GitConfig{})

	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "ssh git url",
			url:  "git@github.com:test/repo.git",
			want: "ssh",
		},
		{
			name: "ssh protocol url",
			url:  "ssh://git@github.com/test/repo.git",
			want: "ssh",
		},
		{
			name: "https url",
			url:  "https://github.com/test/repo.git",
			want: "https",
		},
		{
			name: "http url",
			url:  "http://github.com/test/repo.git",
			want: "https",
		},
		{
			name: "file url",
			url:  "file:///path/to/repo.git",
			want: "https",
		},
		{
			name: "local path",
			url:  "/path/to/repo.git",
			want: "https",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := adapter.detectSourceType(tt.url)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCleanup(t *testing.T) {
	t.Parallel()

	adapter := New(ports.GitConfig{})
	ctx := context.Background()

	// Test cleanup with empty path
	err := adapter.Cleanup(ctx, "")
	require.NoError(t, err)

	// Test cleanup with non-existent path
	err = adapter.Cleanup(ctx, "/nonexistent/path")
	require.NoError(t, err) // Should not fail for non-existent path

	// Test cleanup with actual directory
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "test-cleanup")
	err = os.MkdirAll(testDir, 0750)
	require.NoError(t, err)

	// Create a test file
	testFile := filepath.Join(testDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0600)
	require.NoError(t, err)

	// Verify directory exists
	_, err = os.Stat(testDir)
	require.NoError(t, err)

	// Cleanup should remove the directory
	err = adapter.Cleanup(ctx, testDir)
	require.NoError(t, err)

	// Verify directory is gone
	_, err = os.Stat(testDir)
	require.ErrorIs(t, err, fs.ErrNotExist)
}

// Performance benchmarks.
func BenchmarkGitBinaryAdapter_Init(b *testing.B) {
	config := ports.GitConfig{
		UserName:  "bench-user",
		UserEmail: "bench@example.com",
	}

	adapter := New(config)
	zerologInstance := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	logger := logging.NewZerologAdapter(&zerologInstance)
	ctx := context.Background()

	err := adapter.Initialize(ctx, logger)
	require.NoError(b, err)

	tmpDir := b.TempDir()

	b.ResetTimer()

	for i := range b.N {
		repoPath := filepath.Join(tmpDir, "bench-repo-"+string(rune('0'+i%10)))

		repo, err := adapter.Init(ctx, repoPath, ports.InitOptions{})
		if err == nil {
			_ = repo.Close()
		}
	}
}

func BenchmarkGitBinaryAdapter_ConvertAuthOptions(b *testing.B) {
	adapter := New(ports.GitConfig{})

	authOpts := ports.AuthOptions{
		Type:     ports.AuthTypeToken,
		Token:    "test-token",
		Username: "test-user",
	}

	b.ResetTimer()

	for range b.N {
		_ = adapter.convertAuthOptions(authOpts)
	}
}

func BenchmarkGitBinaryAdapter_DetermineMirrorType(b *testing.B) {
	adapter := New(ports.GitConfig{})

	options := ports.CloneOptions{
		URL:          "https://github.com/test/repo.git",
		Path:         "/tmp/test",
		Depth:        1,
		SingleBranch: true,
	}

	b.ResetTimer()

	for range b.N {
		_ = adapter.determineMirrorType(options)
	}
}

func TestMirrorService_addBasicAuthToURL(t *testing.T) {
	t.Parallel()

	zerologInstance := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	logger := logging.NewZerologAdapter(&zerologInstance)
	service := &MirrorService{logger: logger}
	ctx := context.Background()

	tests := []struct {
		name     string
		url      string
		username string
		token    string
		expected string
	}{
		{
			name:     "https URL with auth",
			url:      "https://github.com/user/repo.git",
			username: "testuser",
			token:    "testtoken",
			expected: "https://testuser:testtoken@github.com/user/repo.git",
		},
		{
			name:     "http URL (unchanged)",
			url:      "http://github.com/user/repo.git",
			username: "testuser",
			token:    "testtoken",
			expected: "http://github.com/user/repo.git",
		},
		{
			name:     "git URL (unchanged)",
			url:      "git@github.com:user/repo.git",
			username: "testuser",
			token:    "testtoken",
			expected: "git@github.com:user/repo.git",
		},
		{
			name:     "empty auth",
			url:      "https://github.com/user/repo.git",
			username: "",
			token:    "",
			expected: "https://:@github.com/user/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := service.addBasicAuthToURL(ctx, tt.url, tt.username, tt.token)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMirrorService_removeBasicAuthFromURL(t *testing.T) {
	t.Parallel()

	zerologInstance := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	logger := logging.NewZerologAdapter(&zerologInstance)
	service := &MirrorService{logger: logger}
	ctx := context.Background()

	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "URL with auth",
			url:      "https://testuser:testtoken@github.com/user/repo.git",
			expected: "https://github.com/user/repo.git",
		},
		{
			name:     "URL without auth",
			url:      "https://github.com/user/repo.git",
			expected: "https://github.com/user/repo.git",
		},
		{
			name:     "URL with @ in path (no auth)",
			url:      "https://github.com/user@company/repo.git",
			expected: "https://company/repo.git",
		},
		{
			name:     "empty URL",
			url:      "",
			expected: "",
		},
		{
			name:     "malformed URL with @",
			url:      "https://user@",
			expected: "https://",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := service.removeBasicAuthFromURL(ctx, tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMirrorService_setupSSHCommandEnv(t *testing.T) {
	t.Parallel()

	zerologInstance := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	logger := logging.NewZerologAdapter(&zerologInstance)
	service := &MirrorService{logger: logger}

	tests := []struct {
		name       string
		authConfig AuthConfig
		expected   []string
	}{
		{
			name: "empty SSH command",
			authConfig: AuthConfig{
				SSHCommand: "",
			},
			expected: []string{},
		},
		{
			name: "SSH command only",
			authConfig: AuthConfig{
				SSHCommand: "ssh -i /path/to/key",
			},
			expected: []string{
				"GIT_SSH_COMMAND=ssh -i /path/to/key",
			},
		},
		{
			name: "SSH command with URL rewrite",
			authConfig: AuthConfig{
				SSHCommand:        "ssh -i /path/to/key",
				SSHURLRewriteFrom: "git@github.com:",
				SSHURLRewriteTo:   "ssh://git@github.com/",
			},
			expected: []string{
				"GIT_SSH_COMMAND=ssh -i /path/to/key",
				"GIT_CONFIG_COUNT=1",
				"GIT_CONFIG_KEY_0=url.ssh://git@github.com/.insteadOf",
				"GIT_CONFIG_VALUE_0=git@github.com:",
			},
		},
		{
			name: "SSH command with partial URL rewrite config",
			authConfig: AuthConfig{
				SSHCommand:        "ssh -i /path/to/key",
				SSHURLRewriteFrom: "git@github.com:",
				SSHURLRewriteTo:   "", // Missing rewrite target
			},
			expected: []string{
				"GIT_SSH_COMMAND=ssh -i /path/to/key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := service.setupSSHCommandEnv(tt.authConfig)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMirrorService_sanitizeURL(t *testing.T) {
	t.Parallel()

	zerologInstance := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	logger := logging.NewZerologAdapter(&zerologInstance)
	service := &MirrorService{logger: logger}
	ctx := context.Background()

	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "URL with auth to sanitize",
			url:      "https://user:token@github.com/repo.git",
			expected: "https://***:***@github.com/repo.git",
		},
		{
			name:     "clean URL unchanged",
			url:      "https://github.com/repo.git",
			expected: "https://github.com/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := service.sanitizeURL(ctx, tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMirrorService_getRepositoryName(t *testing.T) {
	t.Parallel()

	zerologInstance := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	logger := logging.NewZerologAdapter(&zerologInstance)
	service := &MirrorService{logger: logger}

	// is a placeholder implementation that returns "unknown"
	result := service.getRepositoryName(entities.Repository{})
	assert.Equal(t, "unknown", result)
}

func TestMirrorService_createRepositoryEntity(t *testing.T) {
	t.Parallel()

	zerologInstance := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	logger := logging.NewZerologAdapter(&zerologInstance)
	service := &MirrorService{logger: logger}
	ctx := context.Background()

	config := MirrorConfig{
		Name:       "test-repo",
		SourceURL:  "https://github.com/user/test-repo.git",
		MirrorType: "full",
	}

	// is a placeholder implementation that returns empty Repository
	result := service.createRepositoryEntity(ctx, "/tmp/test-repo", config)
	assert.NotNil(t, result)
}

func TestMirrorService_createTempDirectory(t *testing.T) {
	t.Parallel()

	zerologInstance := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	logger := logging.NewZerologAdapter(&zerologInstance)
	tempDir := t.TempDir()
	service := &MirrorService{
		logger:  logger,
		tempDir: tempDir,
	}
	ctx := context.Background()

	// Test successful temp directory creation
	result, err := service.createTempDirectory(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, tempDir)

	// Verify directory was created
	_, err = os.Stat(result)
	require.NoError(t, err)
}
