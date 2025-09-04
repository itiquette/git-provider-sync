// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"context"

	"itiquette/git-provider-sync/internal/domain/entities"
)

// ConfigKey is used as a key for storing and retrieving CLIConfig from a context.
type ConfigKey struct{}

// ConfigFromContext retrieves the CLIConfig from the given context
// is an infrastructure concern for CLI applications and is placed in adapters layer
// Returns the CLIConfig and a boolean indicating if it was found.
func ConfigFromContext(ctx context.Context) (entities.CLIConfig, bool) {
	config, ok := ctx.Value(ConfigKey{}).(entities.CLIConfig)

	return config, ok
}

// WithCLIConfig returns a new context with the given CLIConfig added
// is an infrastructure concern for CLI applications and is placed in adapters layer.
func WithCLIConfig(ctx context.Context, config entities.CLIConfig) context.Context {
	return context.WithValue(ctx, ConfigKey{}, config)
}
