// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package mancmd

import (
	"fmt"

	"github.com/spf13/cobra"

	mcobra "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"
)

func NewManCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "man",
		Short:                 "Generates manpages",
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
		Hidden:                true,
		Args:                  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			manPage, err := mcobra.NewManPage(1, cmd.Root())
			if err != nil {
				return fmt.Errorf("failed new man command: %w", err)
			}

			manPage.
				WithSection("Author", "Git Provider Sync was written by Josef Andersson <https://github.com/itiquette/git-provider-sync>").
				WithSection("Copyright", "Copyright (C) 2024 Josef Andersson.\n"+
					"Released under the EUPL-1.2 license.")

			_, err = fmt.Println(manPage.Build(roff.NewDocument()))
			if err != nil {
				return fmt.Errorf("failed build man command: %w", err)
			}

			return nil
		},
	}

	return cmd
}
