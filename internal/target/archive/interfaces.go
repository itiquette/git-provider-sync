// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package archive

import (
	"context"
	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/model"
)

type GitHandlerer interface {
	InitializeRepository(ctx context.Context, path string, repo interfaces.GitRepository) error
	Push(ctx context.Context, repo interfaces.GitRepository, opt model.PushOption) error
	ConfigureRepository(ctx context.Context, repo interfaces.GitRepository, path string) error
}

type StorageHandlerer interface {
	GetStoragePath(ctx context.Context, opt model.PushOption) (string, error)
	CreateTargetPath(name, targetDir string) string
}

type Handlerer interface {
	CreateArchive(ctx context.Context, sourceDir, targetPath, name string) error
}
