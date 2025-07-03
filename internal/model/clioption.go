// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"context"
	"errors"
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

// CLIOptions retrieves the CLIOption from the given context.
// If the CLIOption is not found or cannot be type-asserted, it calls HandleError.
func CLIOptions(ctx context.Context) CLIOption {
	cliOptions, ok := ctx.Value(CLIOptionKey{}).(CLIOption)
	if !ok {
		err := errors.New("failed to retrieve or type-assert CLIOption from context")
		HandleError(ctx, err)
		// If HandleError doesn't terminate the program, return a zero-value CLIOption
		return CLIOption{}
	}

	return cliOptions
}

// WithCLIOpt returns a new context with the given CLIOption added.
func WithCLIOpt(ctx context.Context, opt CLIOption) context.Context {
	return context.WithValue(ctx, CLIOptionKey{}, opt)
}

// String provides a string representation of CLIOption.
func (c CLIOption) String() string {
	return fmt.Sprintf("CLIOption{ForcePush: %v, IgnoreInvalidName: %v, ASCIIName: %v, "+
		"ActiveFromLimit: %s, DryRun: %v, ConfigFilePath: %s, ConfigFileOnly: %v, "+
		"Quiet: %v, OutputFormat: %v}",
		c.ForcePush, c.IgnoreInvalidName, c.AlphaNumHyphName, c.ActiveFromLimit,
		c.DryRun, c.ConfigFilePath, c.ConfigFileOnly, c.Quiet, c.OutputFormat)
}

// HandleError handles errors that occur during CLI option processing.
// This is a placeholder - in the actual implementation this would integrate with
// the application's error handling strategy.
func HandleError(ctx context.Context, err error) {
	// For now, we'll just log the error
	// In a real implementation, this might terminate the program or take other actions
	fmt.Printf("CLI Option Error: %v\n", err)
}
