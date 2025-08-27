// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package synccmd provides sync command implementation with hexagonal architecture.
package synccmd

import (
	"github.com/rs/zerolog"
)

type syncInputOption struct {
	activeFromLimit   string
	alphaNumHyphName  bool
	dryRun            bool
	forcePush         bool
	ignoreInvalidName bool
}

func (sio syncInputOption) DebugLog(logger *zerolog.Logger) *zerolog.Event {
	return logger.Debug(). //nolint:zerologlint
				Bool("alphaNumHyphName", sio.alphaNumHyphName).
				Bool("dryRun", sio.dryRun).
				Bool("forcePush", sio.forcePush).
				Bool("ignoreInvalidName", sio.ignoreInvalidName).
				Str("activeFromLimit", sio.activeFromLimit)
}
