// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

// sync_flags.go - Flag handling and parsing
package synccmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

type syncFlags struct {
	forcePush         bool
	ignoreInvalidName bool
	cleanupName       bool
	activeFromLimit   string
	dryRun            bool
}

func addSyncFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.Bool("force-push", false, "Overwrite any existing target")
	flags.Bool("ignore-invalid-name", false, "Ignore repositories with invalid names")
	flags.Bool("cleanup-name", false, "Remove non-alphanumeric characters from repository names")
	flags.String("active-from-limit", "", "A negative time duration (e.g., '-1h') to consider repositories active from")
	flags.Bool("dry-run", false, "Simulate sync run without performing clone and push actions")
}

func getSyncFlags(_ context.Context, cmd *cobra.Command) (*syncFlags, error) {
	flags := &syncFlags{}

	var err error

	if flags.forcePush, err = cmd.Flags().GetBool("force-push"); err != nil {
		return nil, fmt.Errorf("get force-push flag: %w", err)
	}

	if flags.ignoreInvalidName, err = cmd.Flags().GetBool("ignore-invalid-name"); err != nil {
		return nil, fmt.Errorf("get ignore-invalid-name flag: %w", err)
	}

	if flags.cleanupName, err = cmd.Flags().GetBool("cleanup-name"); err != nil {
		return nil, fmt.Errorf("get cleanup-name flag: %w", err)
	}

	if flags.activeFromLimit, err = cmd.Flags().GetString("active-from-limit"); err != nil {
		return nil, fmt.Errorf("get active-from-limit flag: %w", err)
	}

	if flags.dryRun, err = cmd.Flags().GetBool("dry-run"); err != nil {
		return nil, fmt.Errorf("get dry-run flag: %w", err)
	}

	return flags, nil
}
