// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package main is the entry point for the git-provider-sync application.
package main

import (
	"fmt"
	"os"

	"itiquette/git-provider-sync/cmd"
	"itiquette/git-provider-sync/internal/adapters/cli"
)

var (
	version = "dev"
	commit  = "none"    //nolint:gochecknoglobals // Build variables set by linker
	date    = "unknown" //nolint:gochecknoglobals // Build variables set by linker
)

func main() {
	// Set up panic recovery to provide user-friendly error messages
	versionString := fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)
	panicHandler := cli.NewPanicHandler(os.Stderr, versionString)

	defer panicHandler.HandlePanic()

	cmd.RunApplication(version, commit, date)
}
