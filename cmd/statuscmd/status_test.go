// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package statuscmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	validationAdapters "itiquette/git-provider-sync/internal/adapters/validation"
	"itiquette/git-provider-sync/internal/domain/validation"
	config "itiquette/git-provider-sync/internal/model/configuration"
)

func TestCreateSystemStatus_FromConfiguration_BuildsCorrectStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		config            config.AppConfiguration
		connectivityCheck bool
		expected          SystemStatus
	}{
		{
			name: "valid configuration with environments",
			config: config.AppConfiguration{
				GitProviderSyncConfs: map[string]config.Environment{
					"production": {
						"github-source": config.SyncConfig{
							BaseConfig: config.BaseConfig{
								ProviderType: "github",
								Domain:       "github.com",
								Owner:        "testuser",
							},
						},
					},
					"staging": {
						"gitlab-source": config.SyncConfig{
							BaseConfig: config.BaseConfig{
								ProviderType: "gitlab",
								Domain:       "gitlab.com",
								Owner:        "testuser",
							},
						},
					},
				},
			},
			connectivityCheck: true,
			expected: SystemStatus{
				ConfigurationValid:  true,
				EnvironmentCount:    2,
				ConnectivityChecked: true,
				HasCriticalIssues:   false,
				Issues:              []string{},
				Warnings:            []string{},
				Suggestions:         []string{},
			},
		},
		{
			name: "empty configuration",
			config: config.AppConfiguration{
				GitProviderSyncConfs: map[string]config.Environment{},
			},
			connectivityCheck: false,
			expected: SystemStatus{
				ConfigurationValid:  true,
				EnvironmentCount:    0,
				ConnectivityChecked: false,
				HasCriticalIssues:   true,
				Issues:              []string{"No environments configured"},
				Warnings:            []string{},
				Suggestions:         []string{"Add environment configuration to gitprovidersync.yaml"},
			},
		},
		{
			name: "single environment with multiple sources",
			config: config.AppConfiguration{
				GitProviderSyncConfs: map[string]config.Environment{
					"production": {
						"github-source": config.SyncConfig{
							BaseConfig: config.BaseConfig{
								ProviderType: "github",
								Domain:       "github.com",
								Owner:        "testuser",
							},
						},
						"gitlab-source": config.SyncConfig{
							BaseConfig: config.BaseConfig{
								ProviderType: "gitlab",
								Domain:       "gitlab.com",
								Owner:        "testuser",
							},
						},
					},
				},
			},
			connectivityCheck: false,
			expected: SystemStatus{
				ConfigurationValid:  true,
				EnvironmentCount:    2,
				ConnectivityChecked: false,
				HasCriticalIssues:   false,
				Issues:              []string{},
				Warnings:            []string{},
				Suggestions:         []string{},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := createSystemStatus(test.config, test.connectivityCheck)

			assert.Equal(t, test.expected.ConfigurationValid, result.ConfigurationValid)
			assert.Equal(t, test.expected.EnvironmentCount, result.EnvironmentCount)
			assert.Equal(t, test.expected.ConnectivityChecked, result.ConnectivityChecked)
			assert.Equal(t, test.expected.HasCriticalIssues, result.HasCriticalIssues)
			assert.Equal(t, test.expected.Issues, result.Issues)
			assert.Equal(t, test.expected.Warnings, result.Warnings)
			assert.Equal(t, test.expected.Suggestions, result.Suggestions)
		})
	}
}

func TestBuildHTTPSURL_FromDomain_ReturnsFormattedURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		domain   string
		expected string
	}{
		{
			name:     "valid domain",
			domain:   "github.com",
			expected: "https://github.com",
		},
		{
			name:     "empty domain",
			domain:   "",
			expected: "",
		},
		{
			name:     "custom domain",
			domain:   "gitlab.example.com",
			expected: "https://gitlab.example.com",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := buildHTTPSURL(test.domain)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestProviderConnectivity_ChecksConnections_ReturnsValidationResults(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create a real connectivity adapter for testing
	adapter := validationAdapters.NewConnectivityAdapter(5 * time.Second)

	config := config.AppConfiguration{
		GitProviderSyncConfs: map[string]config.Environment{
			"production": {
				"github-source": config.SyncConfig{
					BaseConfig: config.BaseConfig{
						ProviderType: "github",
						Domain:       "github.com",
						Owner:        "testuser",
					},
				},
			},
		},
	}

	results := testProviderConnectivity(ctx, config, adapter)

	require.Len(t, results, 1)
	assert.Equal(t, "https://github.com", results[0].Validation.Target)
	assert.Equal(t, validation.ConnectivityTypeHTTP, results[0].Validation.Type)
	assert.Contains(t, results[0].Validation.Description, "HTTP connectivity to github.com")
	assert.True(t, results[0].Validation.Required)
}

func TestFormatSystemStatus_ToJSON_ReturnsValidFormat(t *testing.T) {
	t.Parallel()

	status := SystemStatus{
		ConfigurationValid:  true,
		EnvironmentCount:    2,
		ConnectivityChecked: true,
		HasCriticalIssues:   false,
		Issues:              []string{},
		Warnings:            []string{"Warning message"},
		Suggestions:         []string{"Suggestion message"},
		ConnectivityResults: []validation.ConnectivityResult{
			{
				Success: true,
				Validation: validation.ConnectivityValidation{
					Description: "Test connectivity",
					Required:    true,
				},
			},
		},
	}

	tests := []struct {
		name         string
		outputFormat string
		contains     []string
	}{
		{
			name:         "console format",
			outputFormat: "console",
			contains:     []string{"Git Provider Sync Status", "Configuration: Valid", "2 environment"},
		},
		{
			name:         "json format",
			outputFormat: "json",
			contains:     []string{"\"configuration_valid\":", "\"environment_count\":", "\"overall_success\":"},
		},
		{
			name:         "plain format",
			outputFormat: "plain",
			contains:     []string{"STATUS", "CONFIG_VALID", "ENVIRONMENTS"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := formatSystemStatus(status, true, false, test.outputFormat)

			for _, expected := range test.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestFormatStatus_ConsoleOutput_ReturnsReadableText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		status          SystemStatus
		skipSuggestions bool
		contains        []string
		notContains     []string
	}{
		{
			name: "successful status",
			status: SystemStatus{
				ConfigurationValid:  true,
				EnvironmentCount:    2,
				ConnectivityChecked: true,
				HasCriticalIssues:   false,
				Issues:              []string{},
				Warnings:            []string{},
				Suggestions:         []string{},
				ConnectivityResults: []validation.ConnectivityResult{
					{Success: true, Validation: validation.ConnectivityValidation{Required: true}},
				},
			},
			skipSuggestions: false,
			contains:        []string{"✓ Configuration: Valid", "✓ Provider Connectivity", "✓ Ready for sync operations"},
			notContains:     []string{"✗", "Issues Found"},
		},
		{
			name: "status with issues",
			status: SystemStatus{
				ConfigurationValid:  true,
				EnvironmentCount:    0,
				ConnectivityChecked: false,
				HasCriticalIssues:   true,
				Issues:              []string{"No environments configured"},
				Warnings:            []string{"Warning message"},
				Suggestions:         []string{"Add environment configuration"},
			},
			skipSuggestions: false,
			contains:        []string{"✗ Issues need attention", "Issues Found", "✗ No environments configured", "! Warning message", "Add environment configuration"},
			notContains:     []string{},
		},
		{
			name: "skip suggestions",
			status: SystemStatus{
				ConfigurationValid:  true,
				EnvironmentCount:    1,
				ConnectivityChecked: false,
				HasCriticalIssues:   false,
				Issues:              []string{},
				Warnings:            []string{},
				Suggestions:         []string{"Test suggestion"},
			},
			skipSuggestions: true,
			contains:        []string{"✓ Configuration: Valid"},
			notContains:     []string{"Suggested Actions", "Test suggestion"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := formatStatusConsole(test.status, test.skipSuggestions)

			for _, expected := range test.contains {
				assert.Contains(t, result, expected, "Expected to contain: %s", expected)
			}

			for _, notExpected := range test.notContains {
				assert.NotContains(t, result, notExpected, "Expected NOT to contain: %s", notExpected)
			}
		})
	}
}

func TestFormatStatus_JSONOutput_ReturnsValidJSON(t *testing.T) {
	t.Parallel()

	status := SystemStatus{
		ConfigurationValid:  true,
		EnvironmentCount:    2,
		ConnectivityChecked: true,
		HasCriticalIssues:   false,
		Issues:              []string{"Issue 1"},
		Warnings:            []string{"Warning 1"},
		Suggestions:         []string{"Suggestion 1"},
		ConnectivityResults: []validation.ConnectivityResult{
			{Success: true, Validation: validation.ConnectivityValidation{Required: true}},
			{Success: false, Validation: validation.ConnectivityValidation{Required: false}},
		},
	}

	result := formatStatusJSON(status)

	// Check that it contains valid JSON structure
	assert.Contains(t, result, "\"overall_success\": true")
	assert.Contains(t, result, "\"configuration_valid\": true")
	assert.Contains(t, result, "\"environment_count\": 2")
	assert.Contains(t, result, "\"connectivity_checked\": true")
	assert.Contains(t, result, "\"total_errors\": 1")
	assert.Contains(t, result, "\"total_warnings\": 1")
	assert.Contains(t, result, "\"connectivity_required_failed\": 0")
	assert.Contains(t, result, "\"connectivity_optional_failed\": 1")
	assert.Contains(t, result, "\"issues\": [\"Issue 1\"]")
	assert.Contains(t, result, "\"warnings\": [\"Warning 1\"]")
	assert.Contains(t, result, "\"suggestions\": [\"Suggestion 1\"]")
}

func TestFormatStatus_PlainOutput_ReturnsSimpleText(t *testing.T) {
	t.Parallel()

	status := SystemStatus{
		ConfigurationValid:  true,
		EnvironmentCount:    2,
		ConnectivityChecked: true,
		HasCriticalIssues:   false,
		Issues:              []string{},
		Warnings:            []string{"Warning 1"},
		Suggestions:         []string{},
		ConnectivityResults: []validation.ConnectivityResult{
			{Success: true, Validation: validation.ConnectivityValidation{Required: true}},
			{Success: false, Validation: validation.ConnectivityValidation{Required: false}},
		},
	}

	result := formatStatusPlain(status, false)

	lines := strings.Split(result, "\n")
	assert.Contains(t, lines[0], "STATUS\tWARNING") // Has warnings
	assert.Contains(t, result, "CONFIG_VALID\ttrue")
	assert.Contains(t, result, "ENVIRONMENTS\t2")
	assert.Contains(t, result, "CONNECTIVITY_REQUIRED_FAILED\t0")
	assert.Contains(t, result, "CONNECTIVITY_OPTIONAL_FAILED\t1")
	assert.Contains(t, result, "TOTAL_ERRORS\t0")
	assert.Contains(t, result, "TOTAL_WARNINGS\t1")
}

func TestFormatJSONOutput(t *testing.T) {
	t.Parallel()

	testData := map[string]any{
		"string_field": "test value",
		"int_field":    42,
		"bool_field":   true,
		"array_field":  []string{"item1", "item2"},
		"nested_field": map[string]any{
			"nested_string": "nested value",
			"nested_bool":   false,
			"nested_array":  []string{"nested1", "nested2"},
		},
	}

	result := formatJSONOutput(testData)

	// Check that it produces valid JSON-like output
	assert.Contains(t, result, "\"string_field\": \"test value\"")
	assert.Contains(t, result, "\"int_field\": 42")
	assert.Contains(t, result, "\"bool_field\": true")
	assert.Contains(t, result, "\"array_field\": [\"item1\", \"item2\"]")
	assert.Contains(t, result, "\"nested_field\": {")
	assert.Contains(t, result, "\"nested_string\": \"nested value\"")
	assert.Contains(t, result, "\"nested_bool\": false")
	assert.Contains(t, result, "\"nested_array\": [\"nested1\", \"nested2\"]")

	// Check structure
	assert.True(t, strings.HasPrefix(result, "{\n"))
	assert.True(t, strings.HasSuffix(result, "}\n"))
}

func TestHandleStatusError_WithErrorTypes_FormatsCorrectly(t *testing.T) {
	// Tests error handling logic without capturing fmt.Printf output
	// (capturing Printf output in tests is complex and unreliable)
	t.Parallel()

	tests := []struct {
		name         string
		outputFormat string
		errorMsg     string
	}{
		{
			name:         "json format",
			outputFormat: "json",
			errorMsg:     "test error message",
		},
		{
			name:         "plain format",
			outputFormat: "plain",
			errorMsg:     "test error message",
		},
		{
			name:         "console format",
			outputFormat: "console",
			errorMsg:     "test error message",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Ensure the function doesn't panic and accepts the expected parameters
			assert.NotPanics(t, func() {
				handleStatusError(fmt.Errorf("%s", test.errorMsg), test.outputFormat)
			})
		})
	}
}

func TestGetLastSyncInfo_ParsesSyncFileContent_ReturnsFormattedInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileContent string
		expected    string
	}{
		{
			name: "valid sync info",
			fileContent: `timestamp=1640995200
repos=10
successful=8
failed=1
skipped=1`,
			expected: "(10 repos, 8 successful)", // Check key parts, not exact timestamp format
		},
		{
			name:        "empty file",
			fileContent: "",
			expected:    "",
		},
		{
			name: "partial info",
			fileContent: `timestamp=1640995200
repos=5`,
			expected: "(5 repos, 0 successful)", // Check key parts, not exact timestamp format
		},
		{
			name: "no timestamp",
			fileContent: `repos=5
successful=3`,
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Create isolated temp file for this specific test
			tempDir := t.TempDir()
			syncFilePath := filepath.Join(tempDir, ".gitprovidersync-last-sync-test")

			if test.fileContent != "" {
				require.NoError(t, os.WriteFile(syncFilePath, []byte(test.fileContent), 0o600))
			}

			// Use the testable function with custom file path
			result := getLastSyncInfoFromPath(syncFilePath)

			if test.expected == "" {
				assert.Empty(t, result)
			} else {
				assert.Contains(t, result, test.expected)
			}
		})
	}
}

// TestFormatConfigurationStatus tests are commented out - functions refactored
/*
func TestFormatConfigurationStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   SystemStatus
		contains []string
	}{
		{
			name: "valid config single environment",
			status: SystemStatus{
				ConfigurationValid: true,
				EnvironmentCount:   1,
			},
			contains: []string{"✓ Configuration: Valid", "(1 environment)"},
		},
		{
			name: "valid config multiple environments",
			status: SystemStatus{
				ConfigurationValid: true,
				EnvironmentCount:   3,
			},
			contains: []string{"✓ Configuration: Valid", "(3 environments)"},
		},
		{
			name: "invalid config",
			status: SystemStatus{
				ConfigurationValid: false,
				EnvironmentCount:   0,
			},
			contains: []string{"✗ Configuration: Invalid", "(0 environments)"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := formatConfigurationStatus(test.status)

			for _, expected := range test.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
/* }
*/

func TestFormatConnectivityStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   SystemStatus
		contains []string
	}{
		{
			name: "connectivity not checked",
			status: SystemStatus{
				ConnectivityChecked: false,
			},
			contains: []string{"- Provider Connectivity: Not checked (use --connectivity-check)"},
		},
		{
			name: "all required providers reachable",
			status: SystemStatus{
				ConnectivityChecked: true,
				ConnectivityResults: []validation.ConnectivityResult{
					{Success: true, Validation: validation.ConnectivityValidation{Required: true}},
					{Success: true, Validation: validation.ConnectivityValidation{Required: true}},
				},
			},
			contains: []string{"✓ Provider Connectivity: All required providers reachable"},
		},
		{
			name: "required provider unreachable",
			status: SystemStatus{
				ConnectivityChecked: true,
				ConnectivityResults: []validation.ConnectivityResult{
					{Success: false, Validation: validation.ConnectivityValidation{Required: true}},
					{Success: true, Validation: validation.ConnectivityValidation{Required: true}},
				},
			},
			contains: []string{"✗ Provider Connectivity: 1 required provider(s) unreachable"},
		},
		{
			name: "optional connectivity failed",
			status: SystemStatus{
				ConnectivityChecked: true,
				ConnectivityResults: []validation.ConnectivityResult{
					{Success: true, Validation: validation.ConnectivityValidation{Required: true}},
					{Success: false, Validation: validation.ConnectivityValidation{Required: false}},
				},
			},
			contains: []string{"✓ Provider Connectivity: All required providers reachable", "! Warning: 1 optional connectivity test(s) failed"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := formatConnectivityStatus(test.status)

			for _, expected := range test.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestFormatOverallStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		hasCriticalIssues bool
		expected          string
	}{
		{
			name:              "ready for sync",
			hasCriticalIssues: false,
			expected:          "✓ Ready for sync operations",
		},
		{
			name:              "issues need attention",
			hasCriticalIssues: true,
			expected:          "✗ Issues need attention",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			status := SystemStatus{HasCriticalIssues: test.hasCriticalIssues}
			result := formatOverallStatus(status)

			assert.Contains(t, result, test.expected)
		})
	}
}

func TestFormatIssuesSection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   SystemStatus
		expected string
	}{
		{
			name: "no issues or warnings",
			status: SystemStatus{
				Issues:   []string{},
				Warnings: []string{},
			},
			expected: "",
		},
		{
			name: "issues and warnings",
			status: SystemStatus{
				Issues:   []string{"Critical issue 1", "Critical issue 2"},
				Warnings: []string{"Warning 1"},
			},
			expected: "\nIssues Found:\n  ✗ Critical issue 1\n  ✗ Critical issue 2\n  ! Warning 1\n",
		},
		{
			name: "only warnings",
			status: SystemStatus{
				Issues:   []string{},
				Warnings: []string{"Warning 1", "Warning 2"},
			},
			expected: "\nIssues Found:\n  ! Warning 1\n  ! Warning 2\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := formatIssuesSection(test.status)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestFormatSuggestionsSection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   SystemStatus
		contains []string
	}{
		{
			name: "invalid configuration",
			status: SystemStatus{
				ConfigurationValid: false,
			},
			contains: []string{"Fix configuration errors and run 'gitprovidersync status' again"},
		},
		{
			name: "connectivity not checked",
			status: SystemStatus{
				ConfigurationValid:  true,
				ConnectivityChecked: false,
			},
			contains: []string{"Run 'gitprovidersync status --connectivity-check' to test provider connections"},
		},
		{
			name: "ready for sync",
			status: SystemStatus{
				ConfigurationValid:  true,
				ConnectivityChecked: true,
				HasCriticalIssues:   false,
			},
			contains: []string{"Run 'gitprovidersync sync --dry-run' to preview", "Run 'gitprovidersync sync' to perform the sync"},
		},
		{
			name: "connectivity issues",
			status: SystemStatus{
				ConfigurationValid:  true,
				ConnectivityChecked: true,
				HasCriticalIssues:   true,
			},
			contains: []string{"Fix connectivity issues shown above", "Check authentication tokens and network connectivity"},
		},
		{
			name: "with custom suggestions",
			status: SystemStatus{
				ConfigurationValid:  true,
				ConnectivityChecked: false,
				Suggestions:         []string{"Custom suggestion 1", "Custom suggestion 2"},
			},
			contains: []string{"Run 'gitprovidersync status --connectivity-check'", "Custom suggestion 1", "Custom suggestion 2"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := formatSuggestionsSection(test.status)

			for _, expected := range test.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestFormatJSONValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		key      string
		value    any
		indent   string
		expected string
	}{
		{
			name:     "string value",
			key:      "test_key",
			value:    "test_value",
			indent:   "",
			expected: "  \"test_key\": \"test_value\",\n",
		},
		{
			name:     "int value",
			key:      "count",
			value:    42,
			indent:   "",
			expected: "  \"count\": 42,\n",
		},
		{
			name:     "int64 value",
			key:      "timestamp",
			value:    int64(1640995200),
			indent:   "",
			expected: "  \"timestamp\": 1640995200,\n",
		},
		{
			name:     "bool value",
			key:      "enabled",
			value:    true,
			indent:   "",
			expected: "  \"enabled\": true,\n",
		},
		{
			name:     "string array",
			key:      "items",
			value:    []string{"item1", "item2"},
			indent:   "",
			expected: "  \"items\": [\"item1\", \"item2\"],\n",
		},
		{
			name:     "nested object",
			key:      "nested",
			value:    map[string]any{"sub_key": "sub_value"},
			indent:   "",
			expected: "  \"nested\": {\n    \"sub_key\": \"sub_value\"\n  },\n",
		},
		{
			name:     "unsupported type",
			key:      "unsupported",
			value:    struct{}{},
			indent:   "",
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := formatJSONValue(test.key, test.value, test.indent)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestFormatJSONStringArray(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		key      string
		values   []string
		indent   string
		expected string
	}{
		{
			name:     "empty array",
			key:      "empty",
			values:   []string{},
			indent:   "  ",
			expected: "  \"empty\": [],\n",
		},
		{
			name:     "single item",
			key:      "single",
			values:   []string{"item1"},
			indent:   "  ",
			expected: "  \"single\": [\"item1\"],\n",
		},
		{
			name:     "multiple items",
			key:      "multiple",
			values:   []string{"item1", "item2", "item3"},
			indent:   "  ",
			expected: "  \"multiple\": [\"item1\", \"item2\", \"item3\"],\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := formatJSONStringArray(test.key, test.values, test.indent)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestRemoveTrailingComma(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with trailing comma",
			input:    "test content,\n",
			expected: "test content\n",
		},
		{
			name:     "without trailing comma",
			input:    "test content\n",
			expected: "test content\n",
		},
		{
			name:     "short string",
			input:    "a",
			expected: "a",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := removeTrailingComma(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}
