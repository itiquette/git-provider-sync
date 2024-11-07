// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import "strconv"

type ProjectOption struct {
	Description string `koanf:"description"`
	CIEnabled   bool   `koanf:"cienabled"`
	Visibility  string `koanf:"visibility"`
}

func (p ProjectOption) String() string {
	return "ProjectOption: Type: " + p.Description + ", CIEnabled: " + strconv.FormatBool(p.CIEnabled) + ", Visibility: " + p.Visibility
}

func NewProjectOption() *ProjectOption {
	return &ProjectOption{
		Description: "",
		CIEnabled:   false,
		Visibility:  "",
	}
}
