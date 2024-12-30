// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package archive

import "errors"

var (
	ErrArchiveCompression = errors.New("failed to compress archive")
	ErrArchiveCreation    = errors.New("failed to create archive file")
	ErrDirectoryCreation  = errors.New("failed to create target directory")
	ErrNoFilesToArchive   = errors.New("no files found to archive")
	ErrRepoInitialization = errors.New("failed to initialize repository")
	ErrPushRepository     = errors.New("failed to push to repository")
)
