// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package dto

import (
	"fmt"
	"strings"
)

// SyncRunMetainfoKey is used as a key for context values
// allows SyncRunMetainfo to be stored and retrieved from a context.Context.
type SyncRunMetainfoKey struct{}

// SyncRunMetainfo holds metadata about a synchronization run
// captures essential information about the synchronization process,
// Including source and target identifiers, total items processed,
// And any failures encountered during the process.
type SyncRunMetainfo struct {
	// CtxID is a unique identifier for the synchronization context
	CtxID int

	// Source represents the identifier or location of the data source
	Source string

	// Target represents the identifier or location of the data target
	Target string

	// Total is the total number of items processed during the synchronization
	Total int

	// Fail is a map that stores any failures encountered during synchronization
	// Key is typically an identifier for the failure type or location,
	// And the value is a slice of strings providing details about the failures
	Fail *map[string][]string
}

// String provides a string representation of SyncRunMetainfo
// formats all the fields of SyncRunMetainfo into a human-readable string,
// Including a detailed representation of any failures.
func (s SyncRunMetainfo) String() string {
	var failInfo string

	if len(*s.Fail) > 0 {
		var failures []string
		for key, values := range *s.Fail {
			failures = append(failures, fmt.Sprintf("%s: %s", key, strings.Join(values, ", ")))
		}

		failInfo = fmt.Sprintf("Failures: {%s}", strings.Join(failures, "; "))
	} else {
		failInfo = "No failures"
	}

	return fmt.Sprintf("SyncRunMetainfo{CtxID: %d, Source: %s, Target: %s, Total: %d, %s}",
		s.CtxID, s.Source, s.Target, s.Total, failInfo)
}

// AddFailure adds a failure entry to the SyncRunMetainfo.
func (s *SyncRunMetainfo) AddFailure(key, value string) {
	if s.Fail == nil {
		s.Fail = &map[string][]string{}
	}

	(*s.Fail)[key] = append((*s.Fail)[key], value)
}

// GetFailuresByType returns failures of a specific type.
func (s *SyncRunMetainfo) GetFailuresByType(failureType string) []string {
	if s.Fail == nil {
		return nil
	}

	return (*s.Fail)[failureType]
}

// HasFailures returns true if there are any failures recorded.
func (s *SyncRunMetainfo) HasFailures() bool {
	if s.Fail == nil {
		return false
	}

	return len(*s.Fail) > 0
}

// GetTotalFailures returns the total number of failures across all types.
func (s *SyncRunMetainfo) GetTotalFailures() int {
	if s.Fail == nil {
		return 0
	}

	total := 0
	for _, failures := range *s.Fail {
		total += len(failures)
	}

	return total
}
