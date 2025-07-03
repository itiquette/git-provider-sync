// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package baseoption

import (
	"fmt"

	"github.com/spf13/cobra"

	"itiquette/git-provider-sync/internal/domain/entities"
)

// ExtractRootInputOptions creates CLI configuration from command flags.
// This follows hexagonal architecture by extracting options explicitly rather than hiding them in context.
func ExtractRootInputOptions(cmd *cobra.Command) (entities.CLIConfig, error) {
	configFilePath, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return entities.CLIConfig{}, fmt.Errorf("failed to get config-file flag: %w", err)
	}

	configFileOnly, err := cmd.Flags().GetBool("config-file-only")
	if err != nil {
		return entities.CLIConfig{}, fmt.Errorf("failed to get config-file-only flag: %w", err)
	}

	verbosityWithCaller, err := cmd.Flags().GetBool("verbosity-with-caller")
	if err != nil {
		return entities.CLIConfig{}, fmt.Errorf("failed to get verbosity-with-caller flag: %w", err)
	}

	outputFormat, err := cmd.Flags().GetString("output-format")
	if err != nil {
		return entities.CLIConfig{}, fmt.Errorf("failed to get output-format flag: %w", err)
	}

	// Build CLI configuration using functional builder
	cliConfig := entities.NewCLIConfigBuilder().
		WithConfigFilePath(configFilePath).
		WithConfigFileOnly(configFileOnly).
		WithVerbosityWithCaller(verbosityWithCaller).
		WithOutputFormat(outputFormat).
		Build()

	return cliConfig, nil
}
