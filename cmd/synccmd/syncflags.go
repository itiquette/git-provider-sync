// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

// sync_flags.go - Flag handling and parsing
package synccmd

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

type syncFlags struct {
	activeFromLimit   string
	asciiName         bool
	dryRun            bool
	forcePush         bool
	ignoreInvalidName bool
}

func addSyncFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.Bool("ascii-name", false, "Remove mirror incompatible characters from repository names")
	flags.Bool("dry-run", false, "Simulate sync run without performing clone and push actions")
	flags.Bool("force-push", false, "Overwrite existing mirror target with force")
	flags.Bool("ignore-invalid-name", false, "Don't fail on invalid mirror target names, ignore them")
	flags.String("active-from-limit", "", "A negative time duration (e.g., '-1h') to consider repositories active from")
}

func (syn syncFlags) DebugLog(logger *zerolog.Logger) *zerolog.Event {
	return logger.Debug(). //nolint:zerologlint
				Bool("asciiName", syn.asciiName).
				Bool("dryRun", syn.dryRun).
				Bool("forcePush", syn.forcePush).
				Bool("ignoreInvalidName", syn.ignoreInvalidName).
				Str("activeFromLimit", syn.activeFromLimit)
}

func getSyncFlags(_ context.Context, cmd *cobra.Command) (*syncFlags, error) {
	flags := &syncFlags{}

	var err error

	if flags.asciiName, err = cmd.Flags().GetBool("ascii-name"); err != nil {
		return nil, fmt.Errorf("get ascii-name flag: %w", err)
	}

	if flags.dryRun, err = cmd.Flags().GetBool("dry-run"); err != nil {
		return nil, fmt.Errorf("get dry-run flag: %w", err)
	}

	if flags.forcePush, err = cmd.Flags().GetBool("force-push"); err != nil {
		return nil, fmt.Errorf("get force-push flag: %w", err)
	}

	if flags.ignoreInvalidName, err = cmd.Flags().GetBool("ignore-invalid-name"); err != nil {
		return nil, fmt.Errorf("get ignore-invalid-name flag: %w", err)
	}

	if flags.activeFromLimit, err = cmd.Flags().GetString("active-from-limit"); err != nil {
		return nil, fmt.Errorf("get active-from-limit flag: %w", err)
	}

	return flags, nil
}
