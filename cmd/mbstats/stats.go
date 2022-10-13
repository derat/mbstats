// Copyright 2022 Daniel Erat.
// All rights reserved.

package main

import (
	"fmt"
	"io"
	"sort"

	"github.com/derat/mbstats"
	gostats "github.com/montanaflynn/stats"
)

// countEditors returns the total number of editors with at least one edit of type et.
func countEditors(stats []mbstats.EditorStats, et mbstats.EditType) int {
	var cnt int
	for _, es := range stats {
		if es.Edits[et] > 0 {
			cnt++
		}
	}
	return cnt
}

// countEditTypes returns a map from edit type to total number of edits.
func countEditTypes(stats []mbstats.EditorStats) map[mbstats.EditType]int {
	counts := make(map[mbstats.EditType]int)
	for _, es := range stats {
		for et, cnt := range es.Edits {
			counts[et] += int(cnt)
		}
	}
	return counts
}

// printEditTypeCounts prints edit types by descending number of editors.
func printEditTypeCounts(w io.Writer, stats []mbstats.EditorStats) {
	counts := countEditTypes(stats)
	type typeCount struct {
		et      mbstats.EditType
		total   int
		editors int
	}
	types := make([]typeCount, 0, len(counts))
	for et, cnt := range counts {
		types = append(types, typeCount{et, cnt, countEditors(stats, et)})
	}
	sort.Slice(types, func(i, j int) bool { return types[i].editors > types[j].editors })

	for _, t := range types {
		fmt.Fprintf(w, "%5d editors  %v (%d edits)\n", t.editors, mbstats.EditTypeName(t.et), t.total)
	}
}

// printEditorHistogram prints a histogram of per-editor edit counts.
func printEditorHistogram(w io.Writer, stats []mbstats.EditorStats, et mbstats.EditType) {
	hist := newHistogram(1, 100, 10)
	for _, es := range stats {
		if v := int64(es.Edits[et]); v > 0 {
			hist.add(v)
		}
	}
	hist.write(w, 0, 60)
}

func printEditTypeCorrelations(w io.Writer, stats []mbstats.EditorStats) error {
	typeCounts := countEditTypes(stats)
	types := make([]mbstats.EditType, 0, len(typeCounts))
	for et := range typeCounts {
		types = append(types, et)
	}
	sort.Slice(types, func(i, j int) bool { return types[i] < types[j] })

	edits := make(map[mbstats.EditType]gostats.Float64Data, len(types))
	for _, et := range types {
		vals := make(gostats.Float64Data, len(stats))
		for i, es := range stats {
			vals[i] = float64(es.Edits[et])
		}
		edits[et] = vals
	}

	for i := 0; i < len(types); i++ {
		for j := 0; j < i; j++ {
			et1, et2 := types[i], types[j]
			name1, name2 := mbstats.EditTypeName(et1), mbstats.EditTypeName(et2)
			if coeff, err := gostats.Pearson(edits[et1], edits[et2]); err != nil {
				return err
			} else if coeff > 0.5 || coeff < -0.5 {
				fmt.Fprintf(w, "(%v, %v) = %0.3f\n", name1, name2, coeff)
			}
		}
	}
	return nil
}
