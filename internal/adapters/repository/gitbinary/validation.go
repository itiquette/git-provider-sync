// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package gitbinary

import (
	"os/exec"
	"strings"
)

// ValidateGitBinary validates that git binary is available and working.
// This restores the critical git binary validation from main branch.
func ValidateGitBinary() (string, error) {
	paths := []string{"git", "/usr/bin/git", "/usr/local/bin/git", "/opt/homebrew/bin/git"}

	for _, path := range paths {
		// #nosec G204 - Validating known git binary paths
		if output, err := exec.Command(path, "--version").Output(); err == nil && strings.HasPrefix(string(output), "git version") {
			return path, nil
		}
	}

	return "", ErrGitBinaryNotFound
}

// SetupSSHCommandEnv sets up SSH environment for git commands.
// This restores the critical SSH environment setup from main branch.
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
