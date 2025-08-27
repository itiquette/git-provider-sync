// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"context"
	"fmt"
	"io"
	"os"

	"itiquette/git-provider-sync/internal/domain"
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

// CLIOptions retrieves the CLIOption from the given context.
// If the CLIOption is not found or cannot be type-asserted, it calls HandleError.
func CLIOptions(ctx context.Context) CLIOption {
	cliOptions, ok := ctx.Value(CLIOptionKey{}).(CLIOption)
	if !ok {
		err := domain.ErrCLIOptionRetrievalFailed
		getErrorHandler(ctx).HandleError(ctx, err)
		// If HandleError doesn't terminate the program, return a zero-value CLIOption
		return CLIOption{}
	}

	return cliOptions
}

// getErrorHandler retrieves error handler from context or returns default.
func getErrorHandler(ctx context.Context) ErrorHandler {
	if handler, ok := ctx.Value(ErrorHandlerKey{}).(ErrorHandler); ok {
		return handler
	}

	return &StderrErrorHandler{}
}

// WithErrorHandler returns a new context with the given ErrorHandler added.
func WithErrorHandler(ctx context.Context, handler ErrorHandler) context.Context {
	return context.WithValue(ctx, ErrorHandlerKey{}, handler)
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
	getErrorHandler(ctx).HandleError(ctx, err)
}
