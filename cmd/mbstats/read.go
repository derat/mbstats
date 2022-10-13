// Copyright 2022 Daniel Erat.
// All rights reserved.

package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/derat/mbstats"
)

// readEditorStats reads the specified editor-<year>.json file written by read-mbdump.
func readEditorStats(p string) ([]mbstats.EditorStats, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var stats []mbstats.EditorStats
	dec := json.NewDecoder(f)
	for {
		var es mbstats.EditorStats
		if err := dec.Decode(&es); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		stats = append(stats, es)
	}
	return stats, nil
}

type yearEditorStats struct {
	year  int
	stats []mbstats.EditorStats
}

// readAllEditorStats reads and returns all editor-<year>.json files within the
// specified range from dir. The returned slice is sorted by ascending year.
func readAllEditorStats(dir string, minYear, maxYear int) ([]yearEditorStats, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "editors-????.json"))
	if err != nil {
		return nil, err
	}
	all := make([]yearEditorStats, 0, len(paths))
	for _, p := range paths {
		ys := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(p), "editors-"), ".json")
		year, err := strconv.Atoi(ys)
		if err != nil {
			continue
		}
		if year < minYear || year > maxYear {
			continue
		}
		stats, err := readEditorStats(p)
		if err != nil {
			return nil, err
		}
		all = append(all, yearEditorStats{year, stats})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].year < all[j].year })
	return all, nil
}
