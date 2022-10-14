// Copyright 2022 Daniel Erat.
// All rights reserved.

// Package main implements the mbstats executable for generating MusicBrainz from read-mbdump data.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/derat/mbstats"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: mbstats [flag]... <INPUT_DIR>")
		fmt.Fprintln(flag.CommandLine.Output(), "Generate MusicBrainz stats using JSON data written by read-mbdump.")
		fmt.Fprintln(flag.CommandLine.Output())
		flag.PrintDefaults()
	}
	year := flag.Int("year", time.Now().Year()-1, "Year to display stats from (for applicable actions)")
	minYear := flag.Int("min-year", 2000, "Minimum year to display stats from (for applicable actions)")
	maxYear := flag.Int("max-year", time.Now().Year()-1, "Maximum year to display stats from (for applicable actions)")
	editor := flag.String("editor", "", "Print edit type counts for the named editor")
	editorHist := flag.String("editor-histogram", "", "Print editor edit-count histogram for specified edit type")
	editorList := flag.String("editor-list", "", "Print editor names and edits for specified edit type")
	histMin := flag.Int("histogram-min", 1, "Minimum value for histograms")
	histMax := flag.Int("histogram-max", 100, "Maximum value for histograms")
	histBuckets := flag.Int("histogram-buckets", 10, "Buckets to use for histograms")
	yearlyAge := flag.String("yearly-age", "", "Print yearly average account age in years of editors with specified edit type")
	yearlyEditors := flag.String("yearly-editors", "", "Print yearly editors for specified edit type")
	yearlyEdits := flag.String("yearly-edits", "", "Print yearly edits of specified type")
	flag.Parse()

	os.Exit(func() int {
		if flag.NArg() != 1 {
			flag.Usage()
			return 2
		}
		jsonDir := flag.Arg(0)

		switch {
		case *editor != "":
			stats, _, ret := doSingleYearEditsCmd(jsonDir, *year, "")
			if ret != 0 {
				return ret
			}
			for _, es := range stats {
				if es.Name == *editor {
					for et, cnt := range es.Edits {
						fmt.Printf("%-37s  %5d\n", mbstats.EditTypeName(et), cnt)
					}
					break
				}
			}
			return 0

		case *editorHist != "":
			stats, et, ret := doSingleYearEditsCmd(jsonDir, *year, *editorHist)
			if ret != 0 {
				return ret
			}
			printEditorHistogram(os.Stdout, stats, et, *histMin, *histMax, *histBuckets)
			return 0

		case *editorList != "":
			stats, et, ret := doSingleYearEditsCmd(jsonDir, *year, *editorList)
			if ret != 0 {
				return ret
			}
			for _, es := range stats {
				if cnt := es.Edits[et]; cnt > 0 {
					fmt.Printf("%5d  %v\n", cnt, es.Name)
				}
			}
			return 0

		case *yearlyAge != "":
			yearStats, et, ret := doYearlyEditsCmd(jsonDir, *minYear, *maxYear, *yearlyAge)
			if ret != 0 {
				return ret
			}
			for _, ys := range yearStats {
				end := time.Date(ys.year+1, 1, 1, 0, 0, 0, 0, time.UTC)
				var sum float64
				var cnt int
				for _, es := range ys.stats {
					if es.Edits[et] > 0 && !es.Created.IsZero() {
						sum += end.Sub(es.Created).Seconds() / (86400 * 365)
						cnt++
					}
				}
				var avg float64
				if cnt > 0 {
					avg = sum / float64(cnt)
				}
				fmt.Printf("%4d  %0.1f\n", ys.year, avg)
			}
			return 0

		case *yearlyEditors != "":
			yearStats, et, ret := doYearlyEditsCmd(jsonDir, *minYear, *maxYear, *yearlyEditors)
			if ret != 0 {
				return ret
			}
			for _, ys := range yearStats {
				fmt.Printf("%4d  %5d\n", ys.year, countEditors(ys.stats, et))
			}
			return 0

		case *yearlyEdits != "":
			yearStats, et, ret := doYearlyEditsCmd(jsonDir, *minYear, *maxYear, *yearlyEdits)
			if ret != 0 {
				return ret
			}
			for _, ys := range yearStats {
				fmt.Printf("%4d  %6d\n", ys.year, countEditTypes(ys.stats)[et])
			}
			return 0

		default:
			fmt.Fprintln(os.Stderr, "No action specified (e.g. -editor-histogram ARTIST_CREATE)")
			return 2
		}
	}())
}

// doSingleYearEditsCmd contains common code for commands that read a single year's editor stats.
// If editName is non-empty, it will be parsed and the corresponding edit type will be returned.
// If the returned int is non-zero, a failure occurred and it should be used as the exit code.
func doSingleYearEditsCmd(jsonDir string, year int, editName string) (
	[]mbstats.EditorStats, mbstats.EditType, int) {
	var et mbstats.EditType
	if editName != "" {
		var err error
		if et, err = mbstats.NamedEditType(editName); err != nil {
			fmt.Fprintf(os.Stderr, "Failed looking up %q: %v\n", editName, err)
			return nil, 0, 2
		}
	}
	stats, err := readEditorStats(filepath.Join(jsonDir, fmt.Sprintf("editors-%d.json", year)))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed reading editor stats:", err)
		return nil, 0, 1
	}
	return stats, et, 0
}

// doYearlyEditsCmd contains common code for commands that read multiple years' editor stats.
// If editName is non-empty, it will be parsed and the corresponding edit type will be returned.
// If the returned int is non-zero, a failure occurred and it should be used as the exit code.
func doYearlyEditsCmd(jsonDir string, minYear, maxYear int, editName string) (
	[]yearEditorStats, mbstats.EditType, int) {
	var et mbstats.EditType
	if editName != "" {
		var err error
		if et, err = mbstats.NamedEditType(editName); err != nil {
			fmt.Fprintf(os.Stderr, "Failed looking up %q: %v\n", editName, err)
			return nil, 0, 2
		}
	}
	stats, err := readAllEditorStats(jsonDir, minYear, maxYear)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed reading editor stats:", err)
		return nil, 0, 1
	}
	return stats, et, 0
}
