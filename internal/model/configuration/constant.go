// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

// Mirror target types.
const (
	GITHUB    string = "github"
	GITLAB    string = "gitlab"
	GITEA     string = "gitea"
	ARCHIVE   string = "archive"
	DIRECTORY string = "directory"
)

// Git branch.
const (
	ORIGIN      string = "origin"
	GPSUPSTREAM string = "gpsupstream"
)

// AuthCfg Protocols.
const (
	TLS string = "tls"
	SSH string = "ssh"

	// Auth protocol impl/methods.
	SSHAGENT string = "sshagent"
)

// URL Schemes.
const (
	HTTP  string = "http"
	HTTPS string = "https"
)

// Git Provider Owner types.
const (
	USER  string = "user"
	GROUP string = "group"
)
