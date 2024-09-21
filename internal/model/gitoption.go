// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import "fmt"

type GitOption struct {
	Type              string `koanf:"type"`
	SSHPrivateKeyPath string `koanf:"sshprivatekeypath"`
	SSHPrivateKeyPW   string `koanf:"sshprivatekeypw"`
	IncludeForks      bool   `koanf:"includeforks"`
}

func (p GitOption) String() string {
	return fmt.Sprintf("GitOption: Type: %s, SSHPrivateKeyPath: %s, SSHPrivateKeyPW: %s",
		p.Type, p.SSHPrivateKeyPath, "****")
}

const (
	SSHAGENT string = "sshagent"
	HTTPS    string = "https"
	HTTP     string = "http"
	SSHKEY   string = "sshkey"
)
