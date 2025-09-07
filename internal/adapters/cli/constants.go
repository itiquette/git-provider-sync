// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package cli

// Verbosity levels.
const (
	VerbosityQuiet   = "quiet"
	VerbosityBrief   = "brief"
	VerbosityVerbose = "verbose"
	VerbosityDebug   = "debug"
	VerbosityTrace   = "trace"
)

// Output formats.
const (
	FormatConsole = "console"
	FormatPlain   = "plain"
	// FormatJSON is already defined in the package.
)

// Color modes.
const (
	ColorModeNever = "never"
	ColorModeAuto  = "auto"
)
