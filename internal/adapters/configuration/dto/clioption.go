// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package dto

import (
	"fmt"
)

// CLIOptionKey is used as a key for storing and retrieving CLIOption from a context.
type CLIOptionKey struct{}

// CLIOption represents the set of command-line options available in the application.
type CLIOption struct {
	AlphaNumHyphName    bool   // Whether to clean up repository names
	ActiveFromLimit     string // Time limit for considering repositories as active
	ConfigFileOnly      bool   // Whether to use only the configuration file
	ConfigFilePath      string // Path to the configuration file
	DryRun              bool   // Whether to perform a dry run without making changes
	ForcePush           bool   // Whether to force push changes
	IgnoreInvalidName   bool   // Whether to ignore invalid repository names
	IncludeForks        bool   // Whether to include forked repositories
	IncludeArchived     bool   // Whether to include archived repositories
	UseGitBinary        bool   // Whether to use git binary instead of go-git
	OutputFormat        string // Output format for log
	Quiet               bool   // Whether to suppress non-essential output
	VerbosityWithCaller bool   // Whether to add caller information to log output
}

// String provides a string representation of CLIOption.
func (c CLIOption) String() string {
	return fmt.Sprintf("CLIOption{ForcePush: %v, IgnoreInvalidName: %v, ASCIIName: %v, "+
		"ActiveFromLimit: %s, DryRun: %v, ConfigFilePath: %s, ConfigFileOnly: %v, "+
		"Quiet: %v, OutputFormat: %v}",
		c.ForcePush, c.IgnoreInvalidName, c.AlphaNumHyphName, c.ActiveFromLimit,
		c.DryRun, c.ConfigFilePath, c.ConfigFileOnly, c.Quiet, c.OutputFormat)
}
