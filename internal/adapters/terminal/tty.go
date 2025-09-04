// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package terminal

import (
	"os"

	"golang.org/x/term"
)

// IsOutput checks if stdout is connected to a terminal
// Returns true if output is to a TTY, false if piped or redirected
// is the key heuristic for human vs machine output.
func IsOutput() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// IsError checks if stderr is connected to a terminal
// Useful for progress indicators and status messages.
func IsError() bool {
	return term.IsTerminal(int(os.Stderr.Fd()))
}

// IsInput checks if stdin is connected to a terminal
// Useful for interactive prompts.
func IsInput() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}
