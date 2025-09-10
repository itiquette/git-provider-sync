//go:build integration

// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

// Package testutil provides general test utilities for integration testing.
// These helpers focus on simplicity and reusability without overengineering.
package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// IsolateGitEnvironment provides complete Git isolation for tests.
// This is a general helper that works for any Git-related test.
//
// What it does:
//   - Isolates HOME directory (prevents reading ~/.gitconfig)
//   - Isolates Git configuration paths
//   - Sets consistent Git identity for tests
//   - Disables GPG signing and credential helpers
//   - Prevents SSH host key checking
//
// Usage:
//
//	func TestSomething(t *testing.T) {
//	    IsolateGitEnvironment(t)  // Must be before t.Parallel()
//	    // ... your test code
//	}
//
// Note: Cannot be used with t.Parallel() due to t.Setenv limitations.
// This is a Go testing framework limitation, not a design choice.
func IsolateGitEnvironment(t *testing.T) {
	t.Helper()

	// Create isolated directories
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	gitConfigDir := filepath.Join(tmpDir, "gitconfig")

	// Isolate HOME to prevent reading host configs
	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir) // Windows support

	// Isolate system temp directories
	t.Setenv("TMPDIR", tmpDir)
	t.Setenv("TMP", tmpDir)
	t.Setenv("TEMP", tmpDir)

	// Isolate Git configuration
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(gitConfigDir, "global.gitconfig"))
	t.Setenv("GIT_CONFIG_SYSTEM", filepath.Join(gitConfigDir, "system.gitconfig"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1") // Ignore system config

	// Set consistent test identity
	t.Setenv("GIT_AUTHOR_NAME", "Test Author")
	t.Setenv("GIT_AUTHOR_EMAIL", "test.author@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "Test Committer")
	t.Setenv("GIT_COMMITTER_EMAIL", "test.committer@example.com")

	// Disable features that could interfere with tests
	t.Setenv("GIT_NO_GPG_SIGN", "1")
	t.Setenv("GIT_TERMINAL_PROMPT", "0")
	t.Setenv("GIT_ASKPASS", "echo")

	// Isolate SSH configuration
	t.Setenv("GIT_SSH_COMMAND", "ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null")

	// Disable Git hooks
	t.Setenv("GIT_HOOKS_PATH", filepath.Join(tmpDir, "hooks"))
}

// CreateTempRepoPath creates a path for a test repository.
// This is a simple helper that ensures consistent path creation.
func CreateTempRepoPath(t *testing.T, repoName string) string {
	t.Helper()
	return filepath.Join(t.TempDir(), repoName)
}

// AssertDirExists verifies a directory exists.
// General helper for filesystem assertions.
func AssertDirExists(t *testing.T, path string, msgAndArgs ...interface{}) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Errorf("Expected directory to exist at %s: %v", path, err)
		return
	}
	if !info.IsDir() {
		t.Errorf("Expected %s to be a directory, but it's a file", path)
	}
}

// AssertFileExists verifies a file exists.
// General helper for filesystem assertions.
func AssertFileExists(t *testing.T, path string, msgAndArgs ...interface{}) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Errorf("Expected file to exist at %s: %v", path, err)
		return
	}
	if info.IsDir() {
		t.Errorf("Expected %s to be a file, but it's a directory", path)
	}
}
