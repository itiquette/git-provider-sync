// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import "strings"

type SSHClientOption struct {
	ProxyCommand      string `koanf:"proxycommand"`
	RewriteSSHURLFrom string `koanf:"rewritesshurlfrom"`
	RewriteSSHURLTo   string `koanf:"rewritesshurlto"`
}

func (p SSHClientOption) String() string {
	var parts []string

	parts = append(parts, "SSHClientOption{")

	if p.ProxyCommand != "" {
		parts = append(parts, "ProxyCommand: "+p.ProxyCommand)
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
