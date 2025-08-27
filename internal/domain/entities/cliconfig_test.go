// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCLIConfigBuilder(t *testing.T) {
	t.Parallel()

	builder := NewCLIConfigBuilder()
	config := builder.Build()

	// Test default values
	assert.False(t, config.AlphaNumHyphName())
	assert.Empty(t, config.ActiveFromLimit())
	assert.False(t, config.ConfigFileOnly())
	assert.Empty(t, config.ConfigFilePath())
	assert.False(t, config.DryRun())
	assert.False(t, config.ForcePush())
	assert.False(t, config.IgnoreInvalidName())
	assert.Equal(t, "console", config.OutputFormat())
	assert.False(t, config.Quiet())
	assert.False(t, config.VerbosityWithCaller())
}

func TestCLIConfigBuilder_WithMethods(t *testing.T) {
	t.Parallel()

	builder := NewCLIConfigBuilder().
		WithAlphaNumHyphName(true).
		WithActiveFromLimit("2023-01-01").
		WithConfigFileOnly(true).
		WithConfigFilePath("/path/to/config").
		WithDryRun(true).
		WithForcePush(true).
		WithIgnoreInvalidName(true).
		WithOutputFormat("json").
		WithQuiet(true).
		WithVerbosityWithCaller(true)

	config := builder.Build()

	assert.True(t, config.AlphaNumHyphName())
	assert.Equal(t, "2023-01-01", config.ActiveFromLimit())
	assert.True(t, config.ConfigFileOnly())
	assert.Equal(t, "/path/to/config", config.ConfigFilePath())
	assert.True(t, config.DryRun())
	assert.True(t, config.ForcePush())
	assert.True(t, config.IgnoreInvalidName())
	assert.Equal(t, "json", config.OutputFormat())
	assert.True(t, config.Quiet())
	assert.True(t, config.VerbosityWithCaller())
}

func TestCLIConfigBuilder_FluentInterface(t *testing.T) {
	t.Parallel()

	// Test that builder methods return builder instances for chaining
	builder := NewCLIConfigBuilder()
	result := builder.
		WithAlphaNumHyphName(true).
		WithActiveFromLimit("2023-01-01").
		WithConfigFileOnly(true)

	// Should be able to continue chaining
	config := result.
		WithDryRun(true).
		WithForcePush(true).
		Build()

	assert.True(t, config.AlphaNumHyphName())
	assert.Equal(t, "2023-01-01", config.ActiveFromLimit())
	assert.True(t, config.ConfigFileOnly())
	assert.True(t, config.DryRun())
	assert.True(t, config.ForcePush())
}

func TestCLIConfig_Accessors(t *testing.T) {
	t.Parallel()

	config := NewCLIConfigBuilder().
		WithAlphaNumHyphName(true).
		WithActiveFromLimit("test-limit").
		WithConfigFileOnly(true).
		WithConfigFilePath("/test/path").
		WithDryRun(true).
		WithForcePush(true).
		WithIgnoreInvalidName(true).
		WithOutputFormat("xml").
		WithQuiet(true).
		WithVerbosityWithCaller(true).
		Build()

	// Test all accessor methods
	assert.True(t, config.AlphaNumHyphName())
	assert.Equal(t, "test-limit", config.ActiveFromLimit())
	assert.True(t, config.ConfigFileOnly())
	assert.Equal(t, "/test/path", config.ConfigFilePath())
	assert.True(t, config.DryRun())
	assert.True(t, config.ForcePush())
	assert.True(t, config.IgnoreInvalidName())
	assert.Equal(t, "xml", config.OutputFormat())
	assert.True(t, config.Quiet())
	assert.True(t, config.VerbosityWithCaller())
}

func TestCLIConfig_String(t *testing.T) {
	t.Parallel()

	config := NewCLIConfigBuilder().
		WithAlphaNumHyphName(true).
		WithActiveFromLimit("2023-01-01").
		WithDryRun(true).
		WithOutputFormat("json").
		Build()

	str := config.String()
	assert.NotEmpty(t, str)
	// Check the actual format used by the String() method
	assert.Contains(t, str, "AlphaNumHyphName: true")
	assert.Contains(t, str, "ActiveFromLimit: 2023-01-01")
	assert.Contains(t, str, "DryRun: true")
	assert.Contains(t, str, "OutputFormat: json")
}

func TestCLIConfig_FromLegacyCLIOption(t *testing.T) {
	t.Parallel()

	// The function expects a map[string]interface{} and only handles 3 fields
	legacyMap := map[string]interface{}{
		"DryRun":         true,
		"ForcePush":      true,
		"ConfigFilePath": "/legacy/path",
	}

	config := FromLegacyCLIOption(legacyMap)

	// Only these 3 fields are handled by the current implementation
	assert.True(t, config.DryRun())
	assert.True(t, config.ForcePush())
	assert.Equal(t, "/legacy/path", config.ConfigFilePath())

	// All other fields should be defaults
	assert.False(t, config.AlphaNumHyphName())
	assert.Empty(t, config.ActiveFromLimit())
	assert.False(t, config.ConfigFileOnly())
	assert.False(t, config.IgnoreInvalidName())
	assert.Equal(t, "console", config.OutputFormat())
	assert.False(t, config.Quiet())
	assert.False(t, config.VerbosityWithCaller())
}

func TestCLIConfig_ToLegacyCLIOption(t *testing.T) {
	t.Parallel()

	config := NewCLIConfigBuilder().
		WithAlphaNumHyphName(true).
		WithActiveFromLimit("test-limit").
		WithConfigFileOnly(true).
		WithConfigFilePath("/test/config").
		WithDryRun(true).
		WithForcePush(true).
		WithIgnoreInvalidName(true).
		WithOutputFormat("json").
		WithQuiet(true).
		WithVerbosityWithCaller(true).
		Build()

	legacyOption := config.ToLegacyCLIOption()

	// Test the map values
	alphaNumHyphName, exists := legacyOption["AlphaNumHyphName"].(bool)
	require.True(t, exists)
	require.True(t, alphaNumHyphName)

	activeFromLimit, exists := legacyOption["ActiveFromLimit"].(string)
	require.True(t, exists)
	require.Equal(t, "test-limit", activeFromLimit)

	configFileOnly, exists := legacyOption["ConfigFileOnly"].(bool)
	require.True(t, exists)
	require.True(t, configFileOnly)

	configFilePath, exists := legacyOption["ConfigFilePath"].(string)
	require.True(t, exists)
	require.Equal(t, "/test/config", configFilePath)

	dryRun, exists := legacyOption["DryRun"].(bool)
	require.True(t, exists)
	require.True(t, dryRun)

	forcePush, exists := legacyOption["ForcePush"].(bool)
	require.True(t, exists)
	require.True(t, forcePush)

	ignoreInvalidName, exists := legacyOption["IgnoreInvalidName"].(bool)
	require.True(t, exists)
	require.True(t, ignoreInvalidName)

	outputFormat, exists := legacyOption["OutputFormat"].(string)
	require.True(t, exists)
	require.Equal(t, "json", outputFormat)

	quiet, exists := legacyOption["Quiet"].(bool)
	require.True(t, exists)
	require.True(t, quiet)

	verbosityWithCaller, exists := legacyOption["VerbosityWithCaller"].(bool)
	require.True(t, exists)
	require.True(t, verbosityWithCaller)
}

func TestCLIConfig_RoundTrip(t *testing.T) {
	t.Parallel()

	// Test round-trip conversion: CLIConfig -> map -> CLIConfig
	// Only test the fields that are actually handled by FromLegacyCLIOption
	original := NewCLIConfigBuilder().
		WithDryRun(true).
		WithForcePush(true).
		WithConfigFilePath("/test/path").
		Build()

	// Convert to map (like ToLegacyCLIOption does)
	legacyMap := map[string]interface{}{
		"DryRun":         original.DryRun(),
		"ForcePush":      original.ForcePush(),
		"ConfigFilePath": original.ConfigFilePath(),
	}

	converted := FromLegacyCLIOption(legacyMap)

	// Only these fields should be preserved in round-trip
	assert.Equal(t, original.DryRun(), converted.DryRun())
	assert.Equal(t, original.ForcePush(), converted.ForcePush())
	assert.Equal(t, original.ConfigFilePath(), converted.ConfigFilePath())
}

func TestCLIConfigBuilder_IndependentInstances(t *testing.T) {
	t.Parallel()

	// Test that builder instances are independent
	builder1 := NewCLIConfigBuilder().WithDryRun(true)
	builder2 := NewCLIConfigBuilder().WithDryRun(false)

	config1 := builder1.Build()
	config2 := builder2.Build()

	assert.True(t, config1.DryRun())
	assert.False(t, config2.DryRun())
}

func TestCLIConfig_DefaultValues(t *testing.T) {
	t.Parallel()

	config := NewCLIConfigBuilder().Build()

	// Verify all default values are correctly set
	assert.False(t, config.AlphaNumHyphName())
	assert.Empty(t, config.ActiveFromLimit())
	assert.False(t, config.ConfigFileOnly())
	assert.Empty(t, config.ConfigFilePath())
	assert.False(t, config.DryRun())
	assert.False(t, config.ForcePush())
	assert.False(t, config.IgnoreInvalidName())
	assert.Equal(t, "console", config.OutputFormat())
	assert.False(t, config.Quiet())
	assert.False(t, config.VerbosityWithCaller())
}

func TestCLIConfig_PartialConfiguration(t *testing.T) {
	t.Parallel()

	// Test that only some options can be set while others remain default
	config := NewCLIConfigBuilder().
		WithDryRun(true).
		WithOutputFormat("json").
		Build()

	// Set options should be as configured
	assert.True(t, config.DryRun())
	assert.Equal(t, "json", config.OutputFormat())

	// Unset options should remain default
	assert.False(t, config.AlphaNumHyphName())
	assert.Empty(t, config.ActiveFromLimit())
	assert.False(t, config.ConfigFileOnly())
	assert.Empty(t, config.ConfigFilePath())
	assert.False(t, config.ForcePush())
	assert.False(t, config.IgnoreInvalidName())
	assert.False(t, config.Quiet())
	assert.False(t, config.VerbosityWithCaller())
}

func TestCLIConfig_FromLegacyCLIOption_WithNonMap(t *testing.T) {
	t.Parallel()

	// Test that FromLegacyCLIOption handles non-map input gracefully
	config := FromLegacyCLIOption("not a map")

	// Should return defaults
	assert.False(t, config.AlphaNumHyphName())
	assert.Empty(t, config.ActiveFromLimit())
	assert.False(t, config.ConfigFileOnly())
	assert.Empty(t, config.ConfigFilePath())
	assert.False(t, config.DryRun())
	assert.False(t, config.ForcePush())
	assert.False(t, config.IgnoreInvalidName())
	assert.Equal(t, "console", config.OutputFormat())
	assert.False(t, config.Quiet())
	assert.False(t, config.VerbosityWithCaller())
}
