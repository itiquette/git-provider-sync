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

// package main

// func main() {
// 	gitHandler := git.NewHandler(NewGitLib())
// 	storageHandler := storage.NewHandler()

// 	dirService := directory.NewService(gitHandler, storageHandler)

// 	// Use for push
// 	err := dirService.Push(ctx, repo, opt, gitOpt)
// 	if err != nil {
// 		log.Error().Err(err).Msg("Failed to push to directory")
// 		return
// 	}

// 	// Use for pull
// 	err = dirService.Pull(ctx, sourceCfg, targetPath, repo)
// 	if err != nil {
// 		log.Error().Err(err).Msg("Failed to pull from directory")
// 		return
// 	}
// }
