// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"context"
	"fmt"
	"io"
	"os"
)

// CLIOptionKey is used as a key for storing and retrieving CLIOption from a context.
type CLIOptionKey struct{}

// ErrorHandlerKey is used as a key for storing and retrieving error handler from a context.
type ErrorHandlerKey struct{}

// ErrorHandler defines an interface for handling CLI option errors.
type ErrorHandler interface {
	HandleError(ctx context.Context, err error)
}

// StderrErrorHandler is the default error handler that writes to stderr.
type StderrErrorHandler struct {
	Writer io.Writer
}

// HandleError implements ErrorHandler interface for stderr output.
func (h *StderrErrorHandler) HandleError(_ context.Context, err error) {
	writer := h.Writer
	if writer == nil {
		writer = os.Stderr
	}

	_, _ = fmt.Fprintf(writer, "CLI Option Error: %v\n", err)
}

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
