// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package cli

// Verbosity levels.
const (
	VerbosityError = "error" // Only errors
	VerbosityWarn  = "warn"  // Warnings and errors
	VerbosityInfo  = "info"  // Normal information (default)
	VerbosityDebug = "debug" // Detailed debugging
	VerbosityTrace = "trace" // Very detailed tracing
)

// Output formats.
const (
	FormatDefault = "default" // Default human-friendly format with colors and icons
	FormatPlain   = "plain"   // Simple text format for logs and pipelines
	// FormatJSON is already defined in the package.
)

// Color modes.
const (
	ColorModeNever = "never"
	ColorModeAuto  = "auto"
)
