// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

// Package cli provides CLI-specific adapters for cross-cutting concerns.
package cli

import (
	"context"

	"itiquette/git-provider-sync/internal/domain/entities"
)

// CLIConfigKey is used as a key for storing and retrieving CLIConfig from a context.
type CLIConfigKey struct{}

// CLIConfigFromContext retrieves the CLIConfig from the given context.
// This is an infrastructure concern for CLI applications and is placed in adapters layer.
// Returns the CLIConfig and a boolean indicating if it was found.
func CLIConfigFromContext(ctx context.Context) (entities.CLIConfig, bool) {
	config, ok := ctx.Value(CLIConfigKey{}).(entities.CLIConfig)

	return config, ok
}

// WithCLIConfig returns a new context with the given CLIConfig added.
// This is an infrastructure concern for CLI applications and is placed in adapters layer.
func WithCLIConfig(ctx context.Context, config entities.CLIConfig) context.Context {
	return context.WithValue(ctx, CLIConfigKey{}, config)
}
