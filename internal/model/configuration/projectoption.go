// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import "strconv"

type ProjectOption struct {
	Description string `koanf:"description"`
	Disabled    bool   `koanf:"disabled"`
	Visibility  string `koanf:"visibility"`
}

func (p ProjectOption) String() string {
	return "ProjectOption: Type: " + p.Description + ", Disabled: " + strconv.FormatBool(p.Disabled) + ", Visibility: " + p.Visibility
}

func NewProjectOption() *ProjectOption {
	return &ProjectOption{
		Description: "",
		Disabled:    false,
		Visibility:  "",
	}
}
