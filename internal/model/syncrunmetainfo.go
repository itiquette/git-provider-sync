// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package model

import (
	"fmt"
	"strings"
)

// SyncRunMetainfoKey is used as a key for context values.
// It allows SyncRunMetainfo to be stored and retrieved from a context.Context.
type SyncRunMetainfoKey struct{}

// SyncRunMetainfo holds metadata about a synchronization run.
// It captures essential information about the synchronization process,
// including source and target identifiers, total items processed,
// and any failures encountered during the process.
type SyncRunMetainfo struct {
	// CtxID is a unique identifier for the synchronization context.
	CtxID int

	// Source represents the identifier or location of the data source.
	Source string

	// Target represents the identifier or location of the data target.
	Target string

	// Total is the total number of items processed during the synchronization.
	Total int

	// Fail is a map that stores any failures encountered during synchronization.
	// The key is typically an identifier for the failure type or location,
	// and the value is a slice of strings providing details about the failures.
	Fail map[string][]string
}

// String provides a string representation of SyncRunMetainfo.
// It formats all the fields of SyncRunMetainfo into a human-readable string,
// including a detailed representation of any failures.
//
// Returns:
//   - A string representation of the SyncRunMetainfo instance.
func (s SyncRunMetainfo) String() string {
	var failInfo string

	if len(s.Fail) > 0 {
		var failures []string
		for key, values := range s.Fail {
			failures = append(failures, fmt.Sprintf("%s: %s", key, strings.Join(values, ", ")))
		}

		failInfo = fmt.Sprintf("Failures: {%s}", strings.Join(failures, "; "))
	} else {
		failInfo = "No failures"
	}

	return fmt.Sprintf("SyncRunMetainfo{CtxID: %d, Source: %s, Target: %s, Total: %d, %s}",
		s.CtxID, s.Source, s.Target, s.Total, failInfo)
}

// NewSyncRunMetainfo creates a new SyncRunMetainfo instance.
// It initializes a SyncRunMetainfo struct with the provided values
// and an empty Fail map.
//
// Parameters:
//   - ctxID: An integer representing the unique identifier for the synchronization context.
//   - source: A string representing the source of the data being synchronized.
//   - target: A string representing the target of the synchronization.
//   - total: An integer representing the total number of items to be synchronized.
//
// Returns:
//   - A pointer to a new SyncRunMetainfo instance.
func NewSyncRunMetainfo(ctxID int, source, target string, total int) *SyncRunMetainfo {
	return &SyncRunMetainfo{
		CtxID:  ctxID,
		Source: source,
		Target: target,
		Total:  total,
		Fail:   make(map[string][]string, 200),
	}
}

// AddFailure adds a failure entry to the SyncRunMetainfo.
// This method is used to record any failures that occur during the synchronization process.
//
// Parameters:
//   - key: A string representing the type or location of the failure.
//   - value: A string providing details about the failure.
//
// Note: This method modifies the Fail map of the SyncRunMetainfo instance.
// If an entry for the given key already exists, the new value is appended to the existing slice.
func (s *SyncRunMetainfo) AddFailure(key, value string) {
	s.Fail[key] = append(s.Fail[key], value)
}

// Example usage:
//
//	metainfo := NewSyncRunMetainfo(1, "database_a", "database_b", 1000)
//	metainfo.AddFailure("data_integrity", "Checksum mismatch for record 42")
//	metainfo.AddFailure("network", "Connection timeout at 50% completion")
//	fmt.Println(metainfo)
