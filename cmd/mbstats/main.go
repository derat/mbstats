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
	editorHist := flag.String("editor-histogram", "", "Print editor edit-count histogram for specified edit type")
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
		case *editorHist != "":
			et, err := mbstats.NamedEditType(*editorHist)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed looking up %q: %v\n", *editorHist, err)
				return 2
			}
			stats, err := readEditorStats(filepath.Join(jsonDir, fmt.Sprintf("editors-%d.json", *year)))
			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed reading editor stats:", err)
				return 1
			}
			printEditorHistogram(os.Stdout, stats, et)

		case *yearlyEditors != "":
			et, err := mbstats.NamedEditType(*yearlyEditors)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed looking up %q: %v\n", *yearlyEditors, err)
				return 2
			}
			yearStats, err := readAllEditorStats(jsonDir, *minYear, *maxYear)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed reading editor stats:", err)
				return 1
			}
			for _, ys := range yearStats {
				fmt.Printf("%4d  %d\n", ys.year, countEditors(ys.stats, et))
			}
			return 0

		case *yearlyEdits != "":
			et, err := mbstats.NamedEditType(*yearlyEdits)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed looking up %q: %v\n", *yearlyEdits, err)
				return 2
			}
			yearStats, err := readAllEditorStats(jsonDir, *minYear, *maxYear)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed reading editor stats:", err)
				return 1
			}
			for _, ys := range yearStats {
				fmt.Printf("%4d  %d\n", ys.year, countEditTypes(ys.stats)[et])
			}
			return 0

		default:
			fmt.Fprintln(os.Stderr, "No action specified (e.g. -editor-histogram ARTIST_CREATE)")
			return 2
		}

		return 0
	}())
}
