// SPDX-FileCopyrightText: 2024 Josef Andersson
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
	ForcePush           bool   // Whether to force push changesj
	IgnoreInvalidName   bool   // Whether to ignore invalid repository names
	ASCIIName           bool   // Whether to clean up repository names
	ActiveFromLimit     string // Time limit for considering repositories as active
	DryRun              bool   // Whether to perform a dry run without making changes
	ConfigFilePath      string // Path to the configuration file
	ConfigFileOnly      bool   // Whether to use only the configuration file
	Quiet               bool   // Whether to suppress non-essential output
	VerbosityWithCaller bool   // Whether to add caller information to log output
	OutputFormat        string // Output format for log
}

// CLIOptions retrieves the CLIOption from the given context.
// If the CLIOption is not found or cannot be type-asserted, it calls HandleError.
//
// Parameters:
//   - ctx: The context containing the CLIOption.
//
// Returns:
//   - The CLIOption stored in the context.
//
// Note: This function may terminate the program if HandleError is configured to do so on errors.
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

// WithCLIOption returns a new context with the given CLIOption added.
//
// Parameters:
//   - ctx: The parent context.
//   - options: The CLIOption to add to the context.
//
// Returns:
//   - A new context containing the CLIOption.
func WithCLIOption(ctx context.Context, option CLIOption) context.Context {
	return context.WithValue(ctx, CLIOptionKey{}, option)
}

// String provides a string representation of CLIOption.
func (c CLIOption) String() string {
	return fmt.Sprintf("CLIOption{ForcePush: %v, IgnoreInvalidName: %v, ASCIIName: %v, "+
		"ActiveFromLimit: %s, DryRun: %v, ConfigFilePath: %s, ConfigFileOnly: %v, "+
		"Quiet: %v, OutputFormat: %v}",
		c.ForcePush, c.IgnoreInvalidName, c.ASCIIName, c.ActiveFromLimit,
		c.DryRun, c.ConfigFilePath, c.ConfigFileOnly, c.Quiet, c.OutputFormat)
}

// Example usage:
//
//	options := CLIOption{
//		ForcePush: true,
//		DryRun: true,
//		ConfigFilePath: "/path/to/config.yaml",
//	}
//	ctx := context.Background()
//	ctx = WithCLIOptions(ctx, options)
//
//	// Later in the code:
//	retrievedOptions := CLIOptions(ctx)
//	fmt.Println(retrievedOptions)
