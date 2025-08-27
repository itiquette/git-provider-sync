// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
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
	colorMode           string
	dryRun              bool
	forcePush           bool
	ignoreInvalidName   bool
	outputFormat        string
	quiet               bool
	verbosityWithCaller bool
}

// CLIConfigOption is a functional option for configuring CLIConfig.
type CLIConfigOption func(*CLIConfig)

// NewCLIConfig creates a new CLI configuration with functional options.
func NewCLIConfig(options ...CLIConfigOption) CLIConfig {
	config := CLIConfig{
		alphaNumHyphName:    false,
		activeFromLimit:     "",
		colorMode:           "auto",
		configFileOnly:      false,
		configFilePath:      "",
		dryRun:              false,
		forcePush:           false,
		ignoreInvalidName:   false,
		outputFormat:        "console",
		quiet:               false,
		verbosityWithCaller: false,
	}

	for _, option := range options {
		option(&config)
	}

	return config
}

// WithAlphaNumHyphName sets the alpha-numeric and hyphen name cleaning option.
func WithAlphaNumHyphName(enabled bool) CLIConfigOption {
	return func(c *CLIConfig) {
		c.alphaNumHyphName = enabled
	}
}

// WithActiveFromLimit sets the time limit for considering repositories as active.
func WithActiveFromLimit(limit string) CLIConfigOption {
	return func(c *CLIConfig) {
		c.activeFromLimit = limit
	}
}

// WithConfigFileOnly sets whether to use only the configuration file.
func WithConfigFileOnly(enabled bool) CLIConfigOption {
	return func(c *CLIConfig) {
		c.configFileOnly = enabled
	}
}

// WithConfigFilePath sets the path to the configuration file.
func WithConfigFilePath(path string) CLIConfigOption {
	return func(c *CLIConfig) {
		c.configFilePath = path
	}
}

// WithColorMode sets when to use colors: auto, always, never.
func WithColorMode(mode string) CLIConfigOption {
	return func(c *CLIConfig) {
		c.colorMode = mode
	}
}

// WithDryRun sets whether to perform a dry run without making changes.
func WithDryRun(enabled bool) CLIConfigOption {
	return func(c *CLIConfig) {
		c.dryRun = enabled
	}
}

// WithForcePush sets whether to force push changes.
func WithForcePush(enabled bool) CLIConfigOption {
	return func(c *CLIConfig) {
		c.forcePush = enabled
	}
}

// WithIgnoreInvalidName sets whether to ignore invalid repository names.
func WithIgnoreInvalidName(enabled bool) CLIConfigOption {
	return func(c *CLIConfig) {
		c.ignoreInvalidName = enabled
	}
}

// WithOutputFormat sets the output format for logs.
func WithOutputFormat(format string) CLIConfigOption {
	return func(c *CLIConfig) {
		c.outputFormat = format
	}
}

// WithQuiet sets whether to suppress non-essential output.
func WithQuiet(enabled bool) CLIConfigOption {
	return func(c *CLIConfig) {
		c.quiet = enabled
	}
}

// WithVerbosityWithCaller sets whether to add caller information to log output.
func WithVerbosityWithCaller(enabled bool) CLIConfigOption {
	return func(c *CLIConfig) {
		c.verbosityWithCaller = enabled
	}
}

// Backward compatibility functions for existing code

// CLIConfigBuilder provides backward compatibility for existing builder usage.
type CLIConfigBuilder struct {
	options []CLIConfigOption
}

// NewCLIConfigBuilder creates a new CLI configuration builder for backward compatibility.
func NewCLIConfigBuilder() CLIConfigBuilder {
	return CLIConfigBuilder{
		options: []CLIConfigOption{},
	}
}

// WithAlphaNumHyphName sets the alpha-numeric and hyphen name cleaning option.
func (b CLIConfigBuilder) WithAlphaNumHyphName(enabled bool) CLIConfigBuilder {
	b.options = append(b.options, WithAlphaNumHyphName(enabled))

	return b
}

// WithActiveFromLimit sets the time limit for considering repositories as active.
func (b CLIConfigBuilder) WithActiveFromLimit(limit string) CLIConfigBuilder {
	b.options = append(b.options, WithActiveFromLimit(limit))

	return b
}

// WithConfigFileOnly sets whether to use only the configuration file.
func (b CLIConfigBuilder) WithConfigFileOnly(enabled bool) CLIConfigBuilder {
	b.options = append(b.options, WithConfigFileOnly(enabled))

	return b
}

// WithConfigFilePath sets the path to the configuration file.
func (b CLIConfigBuilder) WithConfigFilePath(path string) CLIConfigBuilder {
	b.options = append(b.options, WithConfigFilePath(path))

	return b
}

// WithColorMode sets when to use colors: auto, always, never.
func (b CLIConfigBuilder) WithColorMode(mode string) CLIConfigBuilder {
	b.options = append(b.options, WithColorMode(mode))

	return b
}

// WithDryRun sets whether to perform a dry run without making changes.
func (b CLIConfigBuilder) WithDryRun(enabled bool) CLIConfigBuilder {
	b.options = append(b.options, WithDryRun(enabled))

	return b
}

// WithForcePush sets whether to force push changes.
func (b CLIConfigBuilder) WithForcePush(enabled bool) CLIConfigBuilder {
	b.options = append(b.options, WithForcePush(enabled))

	return b
}

// WithIgnoreInvalidName sets whether to ignore invalid repository names.
func (b CLIConfigBuilder) WithIgnoreInvalidName(enabled bool) CLIConfigBuilder {
	b.options = append(b.options, WithIgnoreInvalidName(enabled))

	return b
}

// WithOutputFormat sets the output format for logs.
func (b CLIConfigBuilder) WithOutputFormat(format string) CLIConfigBuilder {
	b.options = append(b.options, WithOutputFormat(format))

	return b
}

// WithQuiet sets whether to suppress non-essential output.
func (b CLIConfigBuilder) WithQuiet(enabled bool) CLIConfigBuilder {
	b.options = append(b.options, WithQuiet(enabled))

	return b
}

// WithVerbosityWithCaller sets whether to add caller information to log output.
func (b CLIConfigBuilder) WithVerbosityWithCaller(enabled bool) CLIConfigBuilder {
	b.options = append(b.options, WithVerbosityWithCaller(enabled))

	return b
}

// Build creates an immutable CLIConfig instance using functional options.
func (b CLIConfigBuilder) Build() CLIConfig {
	return NewCLIConfig(b.options...)
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

// ColorMode returns when to use colors: auto, always, never.
func (c CLIConfig) ColorMode() string {
	return c.colorMode
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
		"ActiveFromLimit: %s, ColorMode: %s, DryRun: %v, ConfigFilePath: %s, ConfigFileOnly: %v, "+
		"Quiet: %v, OutputFormat: %v, VerbosityWithCaller: %v}",
		c.forcePush, c.ignoreInvalidName, c.alphaNumHyphName, c.activeFromLimit,
		c.colorMode, c.dryRun, c.configFilePath, c.configFileOnly, c.quiet, c.outputFormat,
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
