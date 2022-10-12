// Copyright 2022 Daniel Erat.
// All rights reserved.

package main

import (
	"archive/tar"
	"bufio"
	"compress/bzip2"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	logFreq    = 5 * time.Second
	mb         = 1024 * 1024
	timeLayout = "2006-01-02 15:04:05.999-07" // time format in PostgreSQL dumps
	emptyCol   = `\N`                         // empty column value in PostgreSQL dumps
)

// readArchive opens a .tar.bz2 file at path p and reads the named file within it.
// fn is invoked with each line from the file.
func readArchive(p, name string, fn func(*lineParser)) error {
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()

	tr := tar.NewReader(bzip2.NewReader(f))
	var size int64
	for {
		head, err := tr.Next()
		if err != nil {
			f.Close()
			if err == io.EOF {
				return fmt.Errorf("file %q not found in archive", name)
			}
			return err
		}
		if head.Name == name {
			size = head.Size
			break
		}
	}

	log.Printf("Processing %v (%0.1f MB)", name, float64(size)/mb)
	logTime := time.Now()

	r := &countReader{r: tr}
	sc := bufio.NewScanner(r)
	var nrows int
	for sc.Scan() {
		p := newLineParser(sc.Text())
		fn(p)
		if p.err != nil {
			return fmt.Errorf("bad row %q: %v", sc.Text(), p.err)
		}

		nrows++
		if now := time.Now(); now.Sub(logTime) > logFreq {
			log.Printf("Read %4.1f%% (%d rows, %0.1f MB)",
				float64(r.nbytes)/float64(size)*100,
				nrows, float64(r.nbytes)/mb)
			logTime = now
		}
	}
	return nil
}

// countReader wraps an io.Reader and counts the number of bytes that have been read.
type countReader struct {
	r      io.Reader
	nbytes int
}

func (cr *countReader) Read(p []byte) (int, error) {
	n, err := cr.r.Read(p)
	cr.nbytes += n
	return n, err
}

// lineParser extracts tab-separated values from a single line.
type lineParser struct {
	cols []string
	err  error
}

func newLineParser(ln string) *lineParser {
	return &lineParser{cols: strings.Split(ln, "\t")}
}

func (p *lineParser) getString(i int) string {
	if p.err != nil {
		return ""
	}
	if i >= len(p.cols) {
		p.err = fmt.Errorf("column %d requested but only have %d", i, len(p.cols))
		return ""
	}
	return p.cols[i]
}

func (p *lineParser) getInt(i int) int32 {
	s := p.getString(i)
	if p.err != nil {
		return 0
	}
	var v int64
	v, p.err = strconv.ParseInt(s, 10, 32)
	return int32(v)
}

func (p *lineParser) getTime(i int) time.Time {
	s := p.getString(i)
	if p.err != nil {
		return time.Time{}
	}
	var t time.Time
	t, p.err = time.Parse(timeLayout, s)
	return t
}
