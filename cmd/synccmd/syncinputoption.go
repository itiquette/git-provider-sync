// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
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

type syncInputOption struct {
	activeFromLimit   string
	alphaNumHyphName  bool
	dryRun            bool
	forcePush         bool
	ignoreInvalidName bool
}

func addSyncInputOptions(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.Bool("alphanumhyph-name", false, "Mirror target name will only contain alpha numeric or hyphen (lessen incompatible characters)")
	flags.Bool("dry-run", false, "Simulate sync run without performing clone and push actions")
	flags.Bool("force-push", false, "Overwrite existing mirror target with force")
	flags.Bool("ignore-invalid-name", false, "Don't fail on invalid mirror target names, ignore them")
	flags.String("active-from-limit", "", "A negative time duration (e.g., '-1h') to consider repositories active from")
}

func (sio syncInputOption) DebugLog(logger *zerolog.Logger) *zerolog.Event {
	return logger.Debug(). //nolint:zerologlint
				Bool("alphaNumHyphName", sio.alphaNumHyphName).
				Bool("dryRun", sio.dryRun).
				Bool("forcePush", sio.forcePush).
				Bool("ignoreInvalidName", sio.ignoreInvalidName).
				Str("activeFromLimit", sio.activeFromLimit)
}

func getSyncInputOptions(_ context.Context, cmd *cobra.Command) (*syncInputOption, error) {
	flags := &syncInputOption{}

	var err error

	if flags.alphaNumHyphName, err = cmd.Flags().GetBool("alphanumhyph-name"); err != nil {
		return nil, fmt.Errorf("get alphanumhyph-name flag: %w", err)
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
