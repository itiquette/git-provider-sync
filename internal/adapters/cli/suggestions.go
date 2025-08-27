// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"fmt"
	"strings"
)

// GetErrorSuggestion returns helpful suggestions for common errors.
// Be functional: pure function with no side effects.
// Don't overengineer: only handle the most common cases.
func GetErrorSuggestion(err error) string {
	if err == nil {
		return ""
	}

	errMsg := err.Error()

	// Check error type and return appropriate suggestion
	switch {
	case isConfigError(errMsg):
		return getConfigSuggestion()
	case isAuthError(errMsg):
		return getAuthSuggestion()
	case isNetworkError(errMsg):
		return getNetworkSuggestion()
	case isPermissionError(errMsg):
		return getPermissionSuggestion()
	default:
		return ""
	}
}

func isConfigError(msg string) bool {
	return strings.Contains(msg, "configuration file") || strings.Contains(msg, "config")
}

func isAuthError(msg string) bool {
	return strings.Contains(msg, "authentication") || strings.Contains(msg, "401") ||
		strings.Contains(msg, "403") || strings.Contains(msg, "token")
}

func isNetworkError(msg string) bool {
	return strings.Contains(msg, "connection") || strings.Contains(msg, "network") ||
		strings.Contains(msg, "timeout")
}

func isPermissionError(msg string) bool {
	return strings.Contains(msg, "permission denied") || strings.Contains(msg, "access denied")
}

func getConfigSuggestion() string {
	return "  1. Check file exists: ls gitprovidersync.yaml\n" +
		"  2. Validate syntax: gitprovidersync status\n" +
		"  3. See example: gitprovidersync print --example"
}

func getAuthSuggestion() string {
	return "  1. Check token is set: echo $GITHUB_TOKEN (or relevant provider)\n" +
		"  2. Verify token permissions (repo access needed)\n" +
		"  3. Test connection: gitprovidersync status --connectivity-check"
}

func getNetworkSuggestion() string {
	return "  1. Check network connectivity\n" +
		"  2. Verify provider URL is correct\n" +
		"  3. Check if behind proxy/firewall"
}

func getPermissionSuggestion() string {
	return "  1. Check file/directory permissions\n" +
		"  2. Ensure you have write access to target directory\n" +
		"  3. Verify git credentials are configured"
}

// FormatErrorWithSuggestion formats an error with helpful suggestions.
// Be idiomatic: follow Go error handling patterns.
// Be functional: pure function with no side effects.
func FormatErrorWithSuggestion(err error, context string, symbols Symbols) string {
	if err == nil {
		return ""
	}

	var output strings.Builder

	// 1. Start with brief context of what failed
	fmt.Fprintf(&output, "\n%s %s\n", symbols.Cross, context)

	// 2. MOST IMPORTANT: Actionable help (moved to prominent position)
	suggestion := GetErrorSuggestion(err)
	if suggestion != "" {
		fmt.Fprintf(&output, "\n%s What to do:\n%s\n", symbols.Arrow, suggestion)
	}

	// 3. Technical details last (for those who need it)
	fmt.Fprintf(&output, "\nDetails: %s\n", err.Error())

	return output.String()
}
