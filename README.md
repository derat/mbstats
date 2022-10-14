# mbstats

This repository contains command-line tools for generating statistics about the
[MusicBrainz] online music database.

[MusicBrainz]: https://musicbrainz.org/

## Background

MusicBrainz provides some
[high-level statistics](https://musicbrainz.org/statistics/), but
[I was curious about how and why people add and edit data](https://community.metabrainz.org/t/any-recent-musicbrainz-user-surveys/604509)
and couldn't find any existing information.

MusicBrainz releases periodic [database dumps], including edit-related
information licensed under the [Attribution-NonCommercial-ShareAlike 3.0]
license. The dumps are massive, though: in the `20221008-002009` dump,
`mbdump-edit.tar.bz2` is 8.7 GB and the uncompressed PostgreSQL dump files
within it are far larger! Even if I had enough free disk space to extract the
data, it'd still be hard and slow to work with.

[database dumps]: https://musicbrainz.org/doc/MusicBrainz_Database
[Attribution-NonCommercial-ShareAlike 3.0]: http://creativecommons.org/licenses/by-nc-sa/3.0/

So, I wrote [read-mbdump], a program that processes `mbdump-edit.tar.bz2` and
`mbdump-editor.tar.bz2` and writes much-smaller files containing JSON objects. A
separate [mbstats] program consumes the JSON data and generates stats.

[read-mbdump]: ./cmd/read-mbdump
[mbstats]: ./cmd/mbstats

## Usage

After making sure you have [Go] installed, run `go install ./cmd/...` from the
root of this repository.

[Go]: https://go.dev/

### read-mbdump

The [read-mbdump] program accepts a directory containing `mbdump-edit.tar.bz2`
and `mbdump-editor.tar.bz2` dump files and a second directory into which JSON
data will be written.

When I ran it on a laptop with an Intel Core i5-8250U CPU 1.60GHz processor, it
took around about 8.5 minutes and 3.1 GB of RAM to process the `20221008-002009`
dump with 419.5 MB of `mbdump/editor_sanitised` data and 9582.4 MB of
`mbdump/edit` data. The `editor-*.json` output files containing yearly editor
stats were each 4 MB or less.

### mbstats

The [mbstats] program accepts a directory containing the JSON files from
`read-mbdump`.

```
Usage: mbstats [flag]... <INPUT_DIR>
Generate MusicBrainz stats using JSON data written by read-mbdump.

  -editor string
    	Print edit type counts for the named editor
  -editor-histogram string
    	Print editor edit-count histogram for specified edit type
  -editor-list string
    	Print editor names and edits for specified edit type
  -histogram-buckets int
    	Buckets to use for histograms (default 10)
  -histogram-max int
    	Maximum value for histograms (default 100)
  -histogram-min int
    	Minimum value for histograms (default 1)
  -max-year int
    	Maximum year to display stats from (for applicable actions) (default 2021)
  -min-year int
    	Minimum year to display stats from (for applicable actions) (default 2000)
  -year int
    	Year to display stats from (for applicable actions) (default 2021)
  -yearly-age string
    	Print yearly average account age in years of editors with specified edit type
  -yearly-editors string
    	Print yearly editors for specified edit type
  -yearly-edits string
    	Print yearly edits of specified type
```
