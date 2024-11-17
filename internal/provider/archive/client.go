// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package archive

import (
	"context"

	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
)

type Client struct{}

func (Client) CreateProject(_ context.Context, _ config.ProviderConfig, _ model.CreateProjectOption) (string, error) {
	return "", nil
}

func (Client) Name() string {
	return config.ARCHIVE
}

func (Client) SetDefaultBranch(_ context.Context, _ string, _ string, _ string) error {
	return nil
}

func (Client) IsValidProjectName(_ context.Context, _ string) bool {
	return true
}

func (Client) ProjectInfos(_ context.Context, _ config.ProviderConfig, _ bool) ([]model.ProjectInfo, error) {
	return nil, nil
}

func (Client) ProtectProject(_ context.Context, _, _, _ string) error {
	return nil
}

func (Client) UnprotectProject(_ context.Context, _, _ string) error {
	return nil
}
