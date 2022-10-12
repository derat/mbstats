// Copyright 2022 Daniel Erat.
// All rights reserved.

// Package mbstats contains MusicBrainz-related code shared between executables.
package mbstats

import (
	"fmt"
	"time"
)

//go:generate sh ./gen_types.sh

// I'm being careful with the sizes of these types since read-mbdump
// ends up holding a lot of them in memory at once.
type EditorID int32
type EditType int16

// EditorStats contains information about a single editor and counts of their
// edits within a given time period.
type EditorStats struct {
	ID      EditorID           `json:"id"`
	Name    string             `json:"name"`
	Created time.Time          `json:"created"`
	Active  time.Time          `json:"active"`
	Edits   map[EditType]int32 `json:"edits"`
}

// EditTypeName returns a human-readable string describing et.
func EditTypeName(et EditType) string {
	if v, ok := editTypeNames[et]; ok {
		return v
	}
	return fmt.Sprintf("UNKNOWN_%d", et)
}
