// Copyright 2022 Daniel Erat.
// All rights reserved.

// Package main implements the read-mbdump executable for summarizing MusicBrainz database dumps.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/derat/mbstats"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: read-mbdump [flag]... <DUMP_DIR> <OUT_DIR>")
		fmt.Fprintln(flag.CommandLine.Output(), "Process MusicBrainz database dumps and write JSON data for gen-mb-stats.")
		fmt.Fprintln(flag.CommandLine.Output())
		flag.PrintDefaults()
	}
	flag.Parse()

	os.Exit(func() int {
		if flag.NArg() != 2 {
			flag.Usage()
			return 2
		}
		dumpDir := flag.Arg(0)
		outDir := flag.Arg(1)

		editors, err := readEditorArchive(filepath.Join(dumpDir, "mbdump-editor.tar.bz2"))
		if err != nil {
			log.Print("Failed reading editors: ", err)
			return 1
		}
		stats, err := readEditArchive(filepath.Join(dumpDir, "mbdump-edit.tar.bz2"))
		if err != nil {
			log.Print("Failed reading edits: ", err)
			return 1
		}
		if err := writeEditorStats(outDir, stats, editors); err != nil {
			log.Print("Failed writing stats: ", err)
			return 1
		}
		return 0
	}())
}

// The MusicBrainz database schema lives here:
// https://github.com/metabrainz/musicbrainz-server/blob/master/admin/sql/CreateTables.sql
//
//  CREATE TABLE editor
//  (
//  	id                  SERIAL,
//  	name                VARCHAR(64) NOT NULL,
//  	privs               INTEGER DEFAULT 0,
//  	email               VARCHAR(64) DEFAULT NULL,
//  	website             VARCHAR(255) DEFAULT NULL,
//  	bio                 TEXT DEFAULT NULL,
//  	member_since        TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
//  	email_confirm_date  TIMESTAMP WITH TIME ZONE,
//  	last_login_date     TIMESTAMP WITH TIME ZONE DEFAULT now(),
//  	last_updated        TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
//  	birth_date          DATE,
//  	gender              INTEGER, -- references gender.id
//  	area                INTEGER, -- references area.id
//  	password            VARCHAR(128) NOT NULL,
//  	ha1                 CHAR(32) NOT NULL,
//  	deleted             BOOLEAN NOT NULL DEFAULT FALSE
//  );
//
//  CREATE TABLE edit
//  (
//      id                  SERIAL,
//      editor              INTEGER NOT NULL, -- references editor.id
//      type                SMALLINT NOT NULL,
//      status              SMALLINT NOT NULL,
//      autoedit            SMALLINT NOT NULL DEFAULT 0,
//      open_time            TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
//      close_time           TIMESTAMP WITH TIME ZONE,
//      expire_time          TIMESTAMP WITH TIME ZONE NOT NULL,
//      language            INTEGER, -- references language.id
//      quality             SMALLINT NOT NULL DEFAULT 1
//  );

type editStats map[mbstats.EditType]int32
type editorStatsMap map[mbstats.EditorID]editStats

// editorInfo contains a subset of information from the editor table.
type editorInfo struct {
	name    string
	created time.Time // member_since
	active  time.Time // last_login_date
}

// readEditorArchive reads an mbdump-editor.tar.bz2 file at the specified path.
func readEditorArchive(p string) (map[mbstats.EditorID]editorInfo, error) {
	editors := make(map[mbstats.EditorID]editorInfo)
	err := readArchive(p, "mbdump/editor_sanitised", func(p *lineParser) {
		id := mbstats.EditorID(p.getInt(0))
		ed := editorInfo{
			name:   p.getString(1),
			active: p.getTime(8),
		}
		// Some accounts are missing a 'member_since' value.
		// No idea why -- maybe it wasn't recorded initially?
		if p.getString(6) != emptyCol {
			ed.created = p.getTime(6)
		}
		editors[id] = ed
	})
	return editors, err
}

// readEditArchive reads an mbdump-edit.tar.bz2 file at the specified path.
// The returned map contains per-editor edit type counts keyed by year.
func readEditArchive(p string) (map[int]editorStatsMap, error) {
	stats := make(map[int]editorStatsMap)
	err := readArchive(p, "mbdump/edit", func(p *lineParser) {
		// Skip non-applied edits.
		// https://github.com/metabrainz/musicbrainz-server/blob/master/root/types/edit.js:
		//
		//  declare type EditStatusT =
		//    | 1 // OPEN
		//    | 2 // APPLIED
		//    | 3 // FAILEDVOTE
		//    | 4 // FAILEDDEP
		//    | 5 // ERROR
		//    | 6 // FAILEDPREREQ
		//    | 7 // NOVOTES
		//    | 9; // DELETED
		if p.getInt(3) != 2 {
			return
		}

		year := p.getTime(5).Year()
		editors := stats[year]
		if editors == nil {
			editors = make(editorStatsMap)
			stats[year] = editors
		}
		ed := mbstats.EditorID(p.getInt(1))
		counts := editors[ed]
		if counts == nil {
			counts = make(editStats)
			editors[ed] = counts
		}
		counts[mbstats.EditType(p.getInt(2))]++
	})
	return stats, err
}

// writeEditorStats writes per-year files (e.g. "editors-2020.json") into dir
// containing JSON-marshaled mbstats.EditorStats objects.
func writeEditorStats(dir string, stats map[int]editorStatsMap,
	editors map[mbstats.EditorID]editorInfo) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	for year, em := range stats {
		p := filepath.Join(dir, fmt.Sprintf("editors-%d.json", year))
		log.Print("Writing ", p)
		f, err := os.Create(p)
		if err != nil {
			return err
		}
		enc := json.NewEncoder(f)

		for id, edits := range em {
			es := mbstats.EditorStats{
				ID:    id,
				Edits: edits,
			}
			if ed, ok := editors[id]; ok {
				es.Name = ed.name
				es.Created = ed.created
				es.Active = ed.active
			}
			if err := enc.Encode(es); err != nil {
				f.Close()
				return err
			}
		}

		if err := f.Close(); err != nil {
			return err
		}
	}
	return nil
}
