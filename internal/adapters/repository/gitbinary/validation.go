// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package gitbinary

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// ValidateGitBinary validates that git binary is available and working
//
//	critical git binary validation .
func ValidateGitBinary(ctx context.Context) (string, error) {
	paths := []string{"git", "/usr/bin/git", "/usr/local/bin/git", "/opt/homebrew/bin/git"}

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for _, path := range paths {
		// #nosec G204 - Validating known git binary paths
		if output, err := exec.CommandContext(timeoutCtx, path, "--version").Output(); err == nil && strings.HasPrefix(string(output), "git version") {
			return path, nil
		}
	}

	return "", ErrGitBinaryNotFound
}

// SetupSSHCommandEnv sets up SSH environment for git commands
//
//	critical SSH environment setup .
func SetupSSHCommandEnv(sshcommand, rewriteurlfrom, rewriteurlto string) []string {
	if sshcommand == "" {
		return []string{}
	}

	env := []string{
		"GIT_SSH_COMMAND=" + sshcommand,
	}

	// Add URL rewriting if specified
	if rewriteurlfrom != "" && rewriteurlto != "" {
		env = append(env,
			"GIT_CONFIG_COUNT=1",
			"GIT_CONFIG_KEY_0=url."+rewriteurlto+".insteadOf",
			"GIT_CONFIG_VALUE_0="+rewriteurlfrom,
		)
	}

	return env
}
