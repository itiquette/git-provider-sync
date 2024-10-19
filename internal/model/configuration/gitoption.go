// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import "fmt"

// GitOption represents configuration options for Git operations.
type GitOption struct {
	Type         string `koanf:"type"`
	IncludeForks bool   `koanf:"includeforks"`
	UseGitBinary bool   `koanf:"usegitbinary"`
}

// String returns a string representation of GitOption, masking sensitive information.
func (p GitOption) String() string {
	return fmt.Sprintf("GitOption: Type: %s, IncludeForks: %v, UseGitBinary: %v",
		p.Type, p.IncludeForks, p.UseGitBinary)
}

// NewGitOption creates a new GitOption with default values.
func NewGitOption() *GitOption {
	return &GitOption{
		Type:         "https",
		IncludeForks: false,
		UseGitBinary: false,
	}
}

const (
	SSHAGENT string = "sshagent"
	HTTPS    string = "https"
	HTTP     string = "http"
)
