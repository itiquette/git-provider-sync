// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"context"
	"os"
	"strings"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// ErrorHandler provides pure functional error handling without side effects.
type ErrorHandler struct {
	logger ports.Logger
}

// NewErrorHandler creates a new error handler.
func NewErrorHandler(logger ports.Logger) ErrorHandler {
	return ErrorHandler{logger: logger}
}

// HandleFatalError handles fatal errors and returns appropriate exit information.
// This is a pure function that returns what should be done rather than doing it.
func (h ErrorHandler) HandleFatalError(ctx context.Context, err error) (bool, int, string) {
	if err == nil {
		return false, 0, ""
	}

	h.logger.Error(ctx, "A fatal error occurred", map[string]interface{}{
		"error": err.Error(),
	})

	userFriendlyMessage := h.createUserFriendlyMessage(err)
	if userFriendlyMessage != "" {
		h.logger.Info(ctx, userFriendlyMessage, nil)
	}

	return true, 1, err.Error()
}

// CreateUserFriendlyMessage creates user-friendly messages for specific error types.
// This is a pure function with no side effects.
func (h ErrorHandler) createUserFriendlyMessage(err error) string {
	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "non-fast-forward update"):
		return "A fast-forward update to target failed. The target may have diverged from the original. Consider using the --force-push option or resolve it manually."
	case strings.Contains(errMsg, "flag accessed but not defined"):
		return "Reading a flag value failed. " + errMsg
	default:
		return ""
	}
}

// ExitIfError is a helper that calls os.Exit if the error is fatal.
// This is only used at the application boundary (main function).
func ExitIfError(ctx context.Context, logger ports.Logger, err error) {
	if err == nil {
		return
	}

	handler := NewErrorHandler(logger)
	shouldExit, exitCode, _ := handler.HandleFatalError(ctx, err)

	if shouldExit {
		os.Exit(exitCode)
	}
}
