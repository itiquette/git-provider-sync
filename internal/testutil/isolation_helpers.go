// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package testutil

import (
	"path/filepath"
	"testing"
)

// IsolateTestEnvironment provides complete test isolation from the host system.
// This includes filesystem, environment variables, and network settings.
// Note: Cannot be used with t.Parallel() due to t.Setenv limitations.
func IsolateTestEnvironment(t *testing.T) {
	t.Helper()

	// Create isolated directories
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	configDir := filepath.Join(tmpDir, "config")

	// Isolate HOME directories
	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir) // Windows
	t.Setenv("XDG_CONFIG_HOME", configDir)
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(tmpDir, "cache"))

	// Isolate temp directories
	t.Setenv("TMPDIR", tmpDir)
	t.Setenv("TMP", tmpDir)
	t.Setenv("TEMP", tmpDir)

	// Isolate network settings (prevent accidental external calls)
	t.Setenv("HTTP_PROXY", "")
	t.Setenv("HTTPS_PROXY", "")
	t.Setenv("NO_PROXY", "*")
	t.Setenv("http_proxy", "")
	t.Setenv("https_proxy", "")
	t.Setenv("no_proxy", "*")

	// Isolate SSH settings
	t.Setenv("SSH_AUTH_SOCK", "")
	t.Setenv("SSH_AGENT_PID", "")
	t.Setenv("SSH_CONFIG", filepath.Join(configDir, "ssh", "config"))

	// Isolate Git settings
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(configDir, "git", "config"))
	t.Setenv("GIT_CONFIG_SYSTEM", filepath.Join(configDir, "git", "system"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

	// Set test identity for Git
	t.Setenv("GIT_AUTHOR_NAME", "Test Author")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "Test Committer")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.com")

	// Disable Git features that might interfere
	t.Setenv("GIT_NO_GPG_SIGN", "1")
	t.Setenv("GIT_TERMINAL_PROMPT", "0")
	t.Setenv("GIT_ASKPASS", "")

	// Isolate Docker settings if present
	t.Setenv("DOCKER_CONFIG", filepath.Join(configDir, "docker"))
	t.Setenv("DOCKER_HOST", "")

	// Isolate Kubernetes settings if present
	t.Setenv("KUBECONFIG", filepath.Join(configDir, "kube", "config"))
}

// IsolateGitEnvironment provides Git-specific test isolation.
// This is a lighter version focused only on Git settings.
// Note: Cannot be used with t.Parallel() due to t.Setenv limitations.
func IsolateGitEnvironment(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	gitConfigDir := filepath.Join(tmpDir, "gitconfig")

	// Isolate Git configuration
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(gitConfigDir, "global"))
	t.Setenv("GIT_CONFIG_SYSTEM", filepath.Join(gitConfigDir, "system"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

	// Set consistent test identity
	t.Setenv("GIT_AUTHOR_NAME", "Test Author")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "Test Committer")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.com")

	// Disable features that could interfere
	t.Setenv("GIT_NO_GPG_SIGN", "1")
	t.Setenv("GIT_TERMINAL_PROMPT", "0")
	t.Setenv("GIT_ASKPASS", "")
	t.Setenv("GIT_SSH_COMMAND", "")
}

// IsolateNetworkEnvironment prevents accidental network calls.
func IsolateNetworkEnvironment(t *testing.T) {
	t.Helper()

	// Set proxy to prevent external network calls
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:0")
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:0")
	t.Setenv("NO_PROXY", "")
	t.Setenv("http_proxy", "http://127.0.0.1:0")
	t.Setenv("https_proxy", "http://127.0.0.1:0")
	t.Setenv("no_proxy", "")
}
