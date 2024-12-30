// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package archive

import (
	"fmt"
	"time"
)

// FormatArchiveTimestamp formats time for archive filenames.
func FormatArchiveTimestamp(t time.Time) string {
	return fmt.Sprintf("_%d%02d%02d_%02d%02d%02d_%d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), t.UnixMilli())
}
