// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package utilities

// SetupSSHCommandEnv sets up SSH command environment variables for git operations.
// This restores the critical SSH functionality from main branch gitbinary/service.go.
//
// Parameters:
//   - sshCommand: Custom SSH command to use (e.g., "ssh -i ~/.ssh/custom_key")
//   - rewriteURLFrom: URL pattern to rewrite from (e.g., "git@github.com:")
//   - rewriteURLTo: URL pattern to rewrite to (e.g., "https://github.com/")
//
// Returns environment variables array for git command execution.
func SetupSSHCommandEnv(sshCommand, rewriteURLFrom, rewriteURLTo string) []string {
	if sshCommand == "" {
		return []string{}
	}

	env := []string{
		"GIT_SSH_COMMAND=" + sshCommand,
	}

	// Add URL rewriting configuration if specified
	if rewriteURLFrom != "" && rewriteURLTo != "" {
		env = append(env,
			"GIT_CONFIG_COUNT=1",
			"GIT_CONFIG_KEY_0=url."+rewriteURLTo+".insteadOf",
			"GIT_CONFIG_VALUE_0="+rewriteURLFrom,
		)
	}

	return env
}
