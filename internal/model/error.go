// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"context"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

// HandleError logs the error and exits the program if an error is present.
// It also provides user-friendly messages for specific error types.
//
// This function should be used at key points in the application where errors
// are critical and should result in program termination.
//
// Parameters:
//   - ctx: A context.Context that contains the zerolog.Logger.
//   - err: The error to handle. If nil, the function returns immediately.
//
// Behavior:
//   - If err is not nil:
//     1. Logs the error using zerolog.
//     2. Provides a user-friendly message if applicable.
//     3. Terminates the program with an exit code of 1.
//
// Usage:
//
//	err := someFunction()
//	HandleError(ctx, err)
func HandleError(ctx context.Context, err error) {
	if err == nil {
		return
	}

	logger := zerolog.Ctx(ctx)
	logger.Error().Err(err).Msg("An error occurred")

	provideUserFriendlyMessage(err, logger)
	os.Exit(1)
}

// provideUserFriendlyMessage logs additional user-friendly information for specific error types.
// This function is intended to be used internally by HandleError.
//
// It checks the error message for known patterns and provides more detailed,
// user-friendly explanations or suggestions for these common error scenarios.
//
// Parameters:
//   - err: The error to analyze.
//   - logger: A pointer to the zerolog.Logger to use for logging.
//
// Current supported error types:
//   - "non-fast-forward update": Suggests using --force-push or manual resolution.
//
// Note: This function can be extended to handle more error types by adding
// additional cases to the switch statement.
func provideUserFriendlyMessage(err error, logger *zerolog.Logger) {
	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "non-fast-forward update"):
		logger.Info().Msg("A fast-forward update to target failed. The target may have diverged from the original. Consider using the --force-push option or resolve it manually.")
	case strings.Contains(errMsg, "flag accessed but not defined"):
		logger.Warn().Msgf("Reading a flag value failed. %s", errMsg)
	}
}
