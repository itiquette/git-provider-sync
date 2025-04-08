// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package archive

import (
	"context"
	"fmt"
	"itiquette/git-provider-sync/internal/model"
	"os"
	"path/filepath"
	"strings"
)

type StorageHandler struct{}

func NewStorageHandler() StorageHandler {
	return StorageHandler{}
}

func (h StorageHandler) GetStoragePath(_ context.Context, opt model.PushOption) (string, error) {
	sourceDir := strings.TrimSuffix(opt.Target, ".tar.gz")
	if err := os.MkdirAll(filepath.Dir(sourceDir), os.ModePerm); err != nil {
		return "", fmt.Errorf("%w: %s: %w", ErrDirectoryCreation, filepath.Dir(sourceDir), err)
	}

	return sourceDir, nil
}
