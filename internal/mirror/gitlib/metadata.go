// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package gitlib

import (
	"context"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	"itiquette/git-provider-sync/internal/provider/stringconvert"
)

type MetadataHandlerer interface {
	UpdateSyncMetadata(ctx context.Context, key, targetDir string)
}

type MetadataHandler struct {
}

func NewMetadataHandler() *MetadataHandler {
	return &MetadataHandler{}
}

func (h *MetadataHandler) UpdateSyncMetadata(ctx context.Context, key, targetDir string) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLib:updateSyncRunMetainfo")
	logger.Debug().Str("key", key).Str("targetDir", stringconvert.RemoveBasicAuthFromURL(ctx, targetDir, false)).Msg("GitLib:updateSyncRunMetainfo")

	if syncRunMeta, ok := ctx.Value(model.SyncRunMetainfoKey{}).(model.SyncRunMetainfo); ok {
		(*syncRunMeta.Fail)[key] = append((*syncRunMeta.Fail)[key], targetDir)
	}
}
