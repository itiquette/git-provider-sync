// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package model

import "strings"

// SSHClientOption represents SSH client configuration options.
type SSHClientOption struct {
	SSHCommand        string `koanf:"sshcommand"`
	RewriteSSHURLFrom string `koanf:"rewritesshurlfrom"`
	RewriteSSHURLTo   string `koanf:"rewritesshurlto"`
}

func (p SSHClientOption) String() string {
	var parts []string

	parts = append(parts, "SSHClientOption{")

	if p.SSHCommand != "" {
		parts = append(parts, "SSHCommand: "+p.SSHCommand)
	}

	if p.RewriteSSHURLFrom != "" {
		parts = append(parts, "RewriteSSHURLFrom: "+p.RewriteSSHURLFrom)
	}

	if p.RewriteSSHURLTo != "" {
		parts = append(parts, "RewriteSSHURLTo: "+p.RewriteSSHURLTo)
	}

	parts = append(parts, "}")

	return strings.Join(parts, " ")
}
