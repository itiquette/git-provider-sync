// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package cmd

import (
	"bytes"
	"context"
	"testing"

	"itiquette/git-provider-sync/internal/model"

	"github.com/stretchr/testify/require"
)

func TestExecutePrintCommandNoArgNoConfPanics(t *testing.T) {
	require := require.New(t)

	ctx := context.Background()
	cliOption := model.CLIOption{}
	ctx = model.WithCLIOption(ctx, cliOption)
	cmd := newPrintCommand()
	cmd.PersistentFlags().String("config-file", "testdasadfasdfta/testconfig.yaml", "path to a git provider sync configuration file.")
	cmd.PersistentFlags().Bool("config-file-only", false, "read configuration from file only (ignore ENV, dotenv, XDG_CONFIG_HOME)")
	cmd.PersistentFlags().Bool("verbosity-with-caller", false, "")
	cmd.PersistentFlags().String("output-format", "co", "")
	cmd.Root().SetContext(ctx)

	cmdOutput := bytes.NewBufferString("")
	cmd.SetOut(cmdOutput)

	require.Empty(cmd.Commands())

	require.Panics(func() { _ = cmd.Execute() })
}

func TestExecutePrintCommandFileConfArgSuccess(t *testing.T) {
	bak := configPrintWriter
	configPrintWriter = new(bytes.Buffer)

	defer func() { configPrintWriter = bak }()

	require := require.New(t)

	cmd := newPrintCommand()
	cmd.PersistentFlags().String("config-file", "testdata/testconfig.yaml", "path to a git provider sync configuration file.")
	cmd.PersistentFlags().Bool("config-file-only", false, "read configuration from file only (ignore ENV, dotenv, XDG_CONFIG_HOME)")
	cmd.PersistentFlags().Bool("verbosity-with-caller", false, "")
	cmd.PersistentFlags().String("output-format", "co", "")

	require.Empty(cmd.Commands())

	// configFilePath := "testdata/testconfig.yaml"
	// inputFlag := model.InputOption{
	// 	ConfigFilePath: configFilePath,
	// }
	//
	ctx := context.Background()
	// ctx = context.WithValue(ctx, model.InputOptionKey{}, inputFlag)

	cmd.Root().SetContext(ctx)
	_ = cmd.Execute()
	buffer, _ := configPrintWriter.(*bytes.Buffer)
	require.Contains(buffer.String(), "Configuration Name")
}
