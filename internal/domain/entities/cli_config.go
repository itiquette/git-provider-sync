// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package entities

import (
	"fmt"
)

// CLIConfig represents command-line interface configuration as a pure domain entity.
// This replaces the legacy model.CLIOption with proper domain design.
type CLIConfig struct {
	alphaNumHyphName    bool
	activeFromLimit     string
	configFileOnly      bool
	configFilePath      string
	dryRun              bool
	forcePush           bool
	ignoreInvalidName   bool
	outputFormat        string
	quiet               bool
	verbosityWithCaller bool
}

// CLIConfigBuilder provides a functional builder for CLIConfig.
type CLIConfigBuilder struct {
	config CLIConfig
}

// NewCLIConfigBuilder creates a new CLI configuration builder with defaults.
func NewCLIConfigBuilder() CLIConfigBuilder {
	return CLIConfigBuilder{
		config: CLIConfig{
			alphaNumHyphName:    false,
			activeFromLimit:     "",
			configFileOnly:      false,
			configFilePath:      "",
			dryRun:              false,
			forcePush:           false,
			ignoreInvalidName:   false,
			outputFormat:        "text",
			quiet:               false,
			verbosityWithCaller: false,
		},
	}
}

// WithAlphaNumHyphName sets the alpha-numeric and hyphen name cleaning option.
func (b CLIConfigBuilder) WithAlphaNumHyphName(enabled bool) CLIConfigBuilder {
	b.config.alphaNumHyphName = enabled

	return b
}

// WithActiveFromLimit sets the time limit for considering repositories as active.
func (b CLIConfigBuilder) WithActiveFromLimit(limit string) CLIConfigBuilder {
	b.config.activeFromLimit = limit

	return b
}

// WithConfigFileOnly sets whether to use only the configuration file.
func (b CLIConfigBuilder) WithConfigFileOnly(enabled bool) CLIConfigBuilder {
	b.config.configFileOnly = enabled

	return b
}

// WithConfigFilePath sets the path to the configuration file.
func (b CLIConfigBuilder) WithConfigFilePath(path string) CLIConfigBuilder {
	b.config.configFilePath = path

	return b
}

// WithDryRun sets whether to perform a dry run without making changes.
func (b CLIConfigBuilder) WithDryRun(enabled bool) CLIConfigBuilder {
	b.config.dryRun = enabled

	return b
}

// WithForcePush sets whether to force push changes.
func (b CLIConfigBuilder) WithForcePush(enabled bool) CLIConfigBuilder {
	b.config.forcePush = enabled

	return b
}

// WithIgnoreInvalidName sets whether to ignore invalid repository names.
func (b CLIConfigBuilder) WithIgnoreInvalidName(enabled bool) CLIConfigBuilder {
	b.config.ignoreInvalidName = enabled

	return b
}

// WithOutputFormat sets the output format for logs.
func (b CLIConfigBuilder) WithOutputFormat(format string) CLIConfigBuilder {
	b.config.outputFormat = format

	return b
}

// WithQuiet sets whether to suppress non-essential output.
func (b CLIConfigBuilder) WithQuiet(enabled bool) CLIConfigBuilder {
	b.config.quiet = enabled

	return b
}

// WithVerbosityWithCaller sets whether to add caller information to log output.
func (b CLIConfigBuilder) WithVerbosityWithCaller(enabled bool) CLIConfigBuilder {
	b.config.verbosityWithCaller = enabled

	return b
}

// Build creates an immutable CLIConfig instance.
func (b CLIConfigBuilder) Build() CLIConfig {
	return b.config
}

// Getters for immutable access to CLIConfig fields

// AlphaNumHyphName returns whether repository names should be cleaned.
func (c CLIConfig) AlphaNumHyphName() bool {
	return c.alphaNumHyphName
}

// ActiveFromLimit returns the time limit for considering repositories as active.
func (c CLIConfig) ActiveFromLimit() string {
	return c.activeFromLimit
}

// ConfigFileOnly returns whether to use only the configuration file.
func (c CLIConfig) ConfigFileOnly() bool {
	return c.configFileOnly
}

// ConfigFilePath returns the path to the configuration file.
func (c CLIConfig) ConfigFilePath() string {
	return c.configFilePath
}

// DryRun returns whether to perform a dry run without making changes.
func (c CLIConfig) DryRun() bool {
	return c.dryRun
}

// ForcePush returns whether to force push changes.
func (c CLIConfig) ForcePush() bool {
	return c.forcePush
}

// IgnoreInvalidName returns whether to ignore invalid repository names.
func (c CLIConfig) IgnoreInvalidName() bool {
	return c.ignoreInvalidName
}

// OutputFormat returns the output format for logs.
func (c CLIConfig) OutputFormat() string {
	return c.outputFormat
}

// Quiet returns whether to suppress non-essential output.
func (c CLIConfig) Quiet() bool {
	return c.quiet
}

// VerbosityWithCaller returns whether to add caller information to log output.
func (c CLIConfig) VerbosityWithCaller() bool {
	return c.verbosityWithCaller
}

// String provides a string representation of CLIConfig.
func (c CLIConfig) String() string {
	return fmt.Sprintf("CLIConfig{ForcePush: %v, IgnoreInvalidName: %v, AlphaNumHyphName: %v, "+
		"ActiveFromLimit: %s, DryRun: %v, ConfigFilePath: %s, ConfigFileOnly: %v, "+
		"Quiet: %v, OutputFormat: %v, VerbosityWithCaller: %v}",
		c.forcePush, c.ignoreInvalidName, c.alphaNumHyphName, c.activeFromLimit,
		c.dryRun, c.configFilePath, c.configFileOnly, c.quiet, c.outputFormat,
		c.verbosityWithCaller)
}

// Context functions moved to internal/adapters/cli/config_context.go
// This maintains clean separation - domain entities should not handle infrastructure concerns like context manipulation.

// Legacy adapter for backward compatibility during migration
// This allows gradual migration from model.CLIOption to entities.CLIConfig

// FromLegacyCLIOption converts a legacy model.CLIOption to domain CLIConfig.
// This is a temporary migration helper that should be removed after full migration.
func FromLegacyCLIOption(legacy interface{}) CLIConfig {
	// Use reflection-like approach to extract fields safely
	// This handles the transition period gracefully
	builder := NewCLIConfigBuilder()

	// Type assertion with fallback to defaults
	if legacyMap, ok := legacy.(map[string]interface{}); ok {
		if val, exists := legacyMap["DryRun"].(bool); exists {
			builder = builder.WithDryRun(val)
		}

		if val, exists := legacyMap["ForcePush"].(bool); exists {
			builder = builder.WithForcePush(val)
		}

		if val, exists := legacyMap["ConfigFilePath"].(string); exists {
			builder = builder.WithConfigFilePath(val)
		}
	}

	return builder.Build()
}

// ToLegacyCLIOption converts domain CLIConfig to legacy format for transition.
// This is a temporary migration helper that should be removed after full migration.
func (c CLIConfig) ToLegacyCLIOption() map[string]interface{} {
	return map[string]interface{}{
		"AlphaNumHyphName":    c.alphaNumHyphName,
		"ActiveFromLimit":     c.activeFromLimit,
		"ConfigFileOnly":      c.configFileOnly,
		"ConfigFilePath":      c.configFilePath,
		"DryRun":              c.dryRun,
		"ForcePush":           c.forcePush,
		"IgnoreInvalidName":   c.ignoreInvalidName,
		"OutputFormat":        c.outputFormat,
		"Quiet":               c.quiet,
		"VerbosityWithCaller": c.verbosityWithCaller,
	}
}
