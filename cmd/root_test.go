// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package cmd

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExecuteRootCommandNoArgumentShowHelp(t *testing.T) {
	require := require.New(t)

	ctx := context.Background()
	cmd := newRootCommand(ctx, "tests")

	cmdOutput := bytes.NewBufferString("")
	cmd.SetOut(cmdOutput)

	require.Len(cmd.Commands(), 3)

	subCmdNames := make([]string, 0, 2)
	for _, v := range cmd.Commands() {
		subCmdNames = append(subCmdNames, v.Name())
	}

	require.Contains(subCmdNames, "print", "sync")

	_ = cmd.Execute()

	out, err := io.ReadAll(cmdOutput)
	if err != nil {
		t.Fatal(err)
	}

	require.Contains(string(out), "A utility")
}
