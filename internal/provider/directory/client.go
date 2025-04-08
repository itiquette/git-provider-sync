// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package directory

import (
	"context"

	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
)

type Client struct{}

func (Client) CreateProject(_ context.Context, _ model.CreateProjectOption) (string, error) {
	return "", nil
}

func (Client) Name() string {
	return config.DIRECTORY
}

func (Client) SetDefaultBranch(_ context.Context, _ string, _ string, _ string) error {
	return nil
}

func (Client) IsValidProjectName(_ context.Context, _ string) bool {
	return true
}

func (Client) ProjectExists(_ context.Context, _, _ string) (bool, string, error) {
	return false, "", nil
}

func (Client) GetProjectInfos(_ context.Context, _ model.ProviderOption, _ bool) ([]model.ProjectInfo, error) {
	return nil, nil
}

func (Client) Protect(_ context.Context, _, _, _ string) error {
	return nil
}

func (Client) Unprotect(_ context.Context, _, _ string) error {
	return nil
}
