// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package directory

import (
	"context"
	"fmt"
	"itiquette/git-provider-sync/internal/log"
	"os"
	"path/filepath"
)

type StorageHandler struct{}

func NewStorageHandler() StorageHandler {
	return StorageHandler{}
}

func (h *StorageHandler) GetTargetPath(ctx context.Context, targetDir, name string) (string, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Directory:GetTargetPath")

	fullPath := filepath.Join(targetDir, name)
	logger.Debug().Str("path", fullPath).Msg("Targeting directory")

	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("%w: %s: %w", ErrDirCreate, targetDir, err)
	}

	return fullPath, nil
}

func (h *StorageHandler) DirectoryExists(dir string) bool {
	_, err := os.Stat(dir)

	return !os.IsNotExist(err)
}
