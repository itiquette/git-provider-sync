// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package terminal

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ConfirmOperation asks for confirmation when running interactively
// Returns true if confirmed (y/yes), false if cancelled or non-interactive
// Follows Go idiom: simple, explicit, no overengineering.
func ConfirmOperation(operation string) bool {
	// Skip if not interactive (pipes, scripts, CI)
	if !IsInput() || !IsError() {
		return false
	}

	fmt.Fprintf(os.Stderr, "⚠️  %s\n", operation)
	fmt.Fprintf(os.Stderr, "Continue? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))

	return response == "y" || response == "yes"
}
