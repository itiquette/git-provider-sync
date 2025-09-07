// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

// Package printcmd implements CLI commands for printing
// configurations and connectivity status
package printcmd

import (
	"context"
	"fmt"
	"io"
	"time"

	gps "itiquette/git-provider-sync/internal/adapters/configuration/dto"
	"itiquette/git-provider-sync/internal/domain/validation"
)

// TestAndDisplayConnectivityMocked is a mocked version for testing that avoids real network calls.
func testAndDisplayConnectivityMocked(_ context.Context, config gps.AppConfiguration, outputFormat string, writer io.Writer) error {
	// Create mock connectivity results instead of real network calls
	var allResults []validation.ConnectivityResult

	for envName, env := range config.GitProviderSyncConfs {
		for configName, syncConfig := range env {
			if syncConfig.ProviderType == "" || syncConfig.Domain == "" {
				continue
			}

			// Create mock result that simulates successful connectivity
			result := validation.ConnectivityResult{
				Validation: validation.ConnectivityValidation{
					Type:        validation.ConnectivityTypeHTTP,
					Target:      buildHTTPSURL(syncConfig.Domain),
					Timeout:     10 * time.Second,
					Description: fmt.Sprintf("HTTP connectivity to %s (env: %s, config: %s)", syncConfig.Domain, envName, configName),
					Required:    true,
				},
				Success:  true,                  // Mock successful connection
				Duration: 50 * time.Millisecond, // Mock fast response
				Error:    nil,
			}
			allResults = append(allResults, result)
		}
	}

	// Display results using the same function as the real version
	if err := displayConnectivityResults(allResults, outputFormat, writer); err != nil {
		return fmt.Errorf("failed to display connectivity results: %w", err)
	}

	return nil
}
