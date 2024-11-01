// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

type ProjectOption struct {
	Description string `koanf:"description"`
}

func (p ProjectOption) String() string {
	return "ProjectOption: Type: " + p.Description
}

func NewProjectOption() *ProjectOption {
	return &ProjectOption{
		Description: "",
	}
}
