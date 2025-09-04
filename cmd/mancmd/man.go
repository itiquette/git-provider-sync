// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

// Package mancmd generates manual pages and help documentation
package mancmd

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

// NewManCommand creates and returns a new cli.Command for the 'man' subcommand
// command generates man pages in markdown format
// Command is hidden from normal help output as it's primarily used for build processes.
func NewManCommand() *cli.Command {
	cmd := &cli.Command{
		Name:   "man",
		Usage:  "Generate man page documentation",
		Hidden: true, // Hidden from regular help
		Action: runManGeneration,
	}

	return cmd
}

// RunManGeneration generates man page content
// outputs basic man page content
// In a full implementation, this would generate markdown that gets converted to man format.
func runManGeneration(_ context.Context, _ *cli.Command) error {
	// Basic man page content output
	// In practice, this would generate proper markdown documentation
	manContent := `# gitprovidersync(1)

## NAME

gitprovidersync - Utility for mirroring and storing Git repositories

## SYNOPSIS

**gitprovidersync** [OPTIONS] COMMAND [ARGS]...

## DESCRIPTION

A utility for mirroring Git repositories to various Git providers or storage.
Supports GitHub, Gitea, GitLab, uncompressed directories, and a compressed archive format (tar.gz).
Allows syncing to multiple mirror destinations.

## COMMANDS

**sync**
    Mirror repositories from a source Git provider to targets

**print**
    Print the current configuration

## OPTIONS

**--config-file** *FILE*
    Path to the configuration file (default: gitprovidersync.yaml)

**--format** *FORMAT*
    Output format (console,json,plain) (default: console)

**--quiet**
    Only output errors

**--verbosity** *LEVEL*
    Set output verbosity: quiet | brief | verbose | debug | trace (default: brief)

## AUTHOR

Git Provider Sync was initially created by Josef Andersson <https://github.com/itiquette/git-provider-sync>

## COPYRIGHT

Copyright (C) 2025 The Git Provider Sync Authors.
Released under the EUPL-1.2 license.
`

	_, err := fmt.Fprint(os.Stdout, manContent)
	if err != nil {
		return fmt.Errorf("stdout write error during man page output: %w", err)
	}

	return nil
}
