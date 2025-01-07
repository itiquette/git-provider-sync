// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package directory

import "errors"

var (
	ErrDirCreate          = errors.New("failed to create directory")
	ErrDirGetPath         = errors.New("failed to get directory path")
	ErrRepoInitialization = errors.New("failed to initialize repository")
	ErrPushRepository     = errors.New("failed to push repository")
	ErrPullRepository     = errors.New("failed to pull repository")
)
