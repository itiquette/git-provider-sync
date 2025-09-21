// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package cmd

import (
	"context"
	"errors"
	"strings"
	"syscall"
)

// Standard exit codes following Unix/Linux conventions.
const (
	// ExitSuccess indicates successful completion.
	ExitSuccess = 0

	// ExitGeneralError indicates a general error occurred.
	ExitGeneralError = 1

	// ExitMisuse indicates incorrect command usage or invalid arguments.
	ExitMisuse = 2

	// ExitNoPermission indicates permission denied (API or file permissions).
	ExitNoPermission = 77

	// ExitConfigError indicates a configuration file error.
	ExitConfigError = 78

	// ExitCantExecute indicates a command exists but cannot be executed.
	ExitCantExecute = 126

	// ExitNotFound indicates a command was not found.
	ExitNotFound = 127

	// Signal-based exit codes (128 + signal number).

	// ExitSIGHUP indicates termination by SIGHUP (terminal hangup).
	ExitSIGHUP = 128 + 1 // 129

	// ExitSIGINT indicates termination by SIGINT (Ctrl+C).
	ExitSIGINT = 128 + 2 // 130

	// ExitSIGQUIT indicates termination by SIGQUIT (Ctrl+\).
	ExitSIGQUIT = 128 + 3 // 131

	// ExitSIGTERM indicates termination by SIGTERM (termination request).
	ExitSIGTERM = 128 + 15 // 143
)

// SignalToExitCode converts a signal to its corresponding exit code.
// Only handles the signals we explicitly care about.
//
//nolint:exhaustive // We only handle specific signals, not all possible signals
func SignalToExitCode(sig syscall.Signal) int {
	switch sig {
	case syscall.SIGHUP:
		return ExitSIGHUP
	case syscall.SIGINT:
		return ExitSIGINT
	case syscall.SIGQUIT:
		return ExitSIGQUIT
	case syscall.SIGTERM:
		return ExitSIGTERM
	default:
		return 0
	}
}

// DetermineExitCode analyzes an error and returns the appropriate exit code.
func DetermineExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}

	// Check for context cancellation (might be from signal)
	if errors.Is(err, context.Canceled) {
		// Signal exit codes are handled separately
		return ExitSuccess
	}

	errStr := strings.ToLower(err.Error())

	return mapErrorToExitCode(errStr)
}

// mapErrorToExitCode maps error string to appropriate exit code.
func mapErrorToExitCode(errStr string) int {
	// Check each error category
	if isConfigError(errStr) {
		return ExitConfigError
	}

	if isUsageError(errStr) {
		return ExitMisuse
	}

	if isCommandNotFoundError(errStr) {
		return ExitNotFound
	}

	if isCannotExecuteError(errStr) {
		return ExitCantExecute
	}

	if isPermissionError(errStr) {
		return ExitNoPermission
	}

	return ExitGeneralError
}

// isConfigError checks if the error is a configuration error.
func isConfigError(errStr string) bool {
	return strings.Contains(errStr, "configuration") ||
		strings.Contains(errStr, "config file") ||
		strings.Contains(errStr, "yaml")
}

// isUsageError checks if the error is a command usage error.
func isUsageError(errStr string) bool {
	return strings.Contains(errStr, "invalid argument") ||
		strings.Contains(errStr, "unknown flag") ||
		strings.Contains(errStr, "required flag")
}

// isCommandNotFoundError checks if the error indicates a command was not found.
func isCommandNotFoundError(errStr string) bool {
	return strings.Contains(errStr, "executable file not found") ||
		strings.Contains(errStr, "command not found") ||
		strings.Contains(errStr, "not found in path")
}

// isCannotExecuteError checks if the error indicates a command cannot be executed.
func isCannotExecuteError(errStr string) bool {
	return strings.Contains(errStr, "permission denied") &&
		(strings.Contains(errStr, "git") || strings.Contains(errStr, "exec"))
}

// isPermissionError checks if the error is a permission error.
func isPermissionError(errStr string) bool {
	// API permission errors
	if strings.Contains(errStr, "401") ||
		strings.Contains(errStr, "403") ||
		strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "forbidden") {
		return true
	}
	// General permission errors
	return strings.Contains(errStr, "permission denied")
}
