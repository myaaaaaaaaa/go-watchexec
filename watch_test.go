package watchexec

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

type errorFS struct {
	fs.FS
}

func (fsys errorFS) Open(name string) (fs.File, error) {
	if strings.Contains(name, "error") {
		return nil, errors.New("error file: " + name)
	}
	return fsys.FS.Open(name)
}

func TestWalk(t *testing.T) {
	mapfs := errorFS{fstest.MapFS{
		"a/f":            &fstest.MapFile{},
		"b/d/d/d/.f.txt": &fstest.MapFile{},
		"c/d/.d/d/f.txt": &fstest.MapFile{},
		".d":             &fstest.MapFile{},
		".e/d/d/f":       &fstest.MapFile{},
		".f/d/f.txt":     &fstest.MapFile{},

		"y/1/f.txt":       &fstest.MapFile{},
		"y/2/error/f.txt": &fstest.MapFile{},
		"y/3/f.txt":       &fstest.MapFile{},

		"z/1/f.txt": &fstest.MapFile{},
		"z/2/f.txt": &fstest.MapFile{},
		"z/3/f.txt": &fstest.MapFile{},
	}}

	assert := func(arg string, want ...string) {
		t.Helper()

		got := slices.Sorted(maps.Keys(walk(mapfs, arg)))

		if fmt.Sprint(got) != fmt.Sprint(want) {
			t.Error("got", got, "    want", want)
		}
	}

	assert("a", "a/f")
	assert("b", "b/d/d/d/.f.txt")
	assert("c", "")
	assert(".d", ".d")
	assert(".e", "")
	assert(".f", "")

	assert("y",
		"y/1/f.txt",
		"y/3/f.txt",
	)

	assert("z",
		"z/1/f.txt",
		"z/2/f.txt",
		"z/3/f.txt",
	)
}

func TestScan(t *testing.T) {
	mapfs := fstest.MapFS{
		"a.txt": &fstest.MapFile{ModTime: time.UnixMilli(1)},
		"b.txt": &fstest.MapFile{ModTime: time.UnixMilli(2)},
		"c.txt": &fstest.MapFile{ModTime: time.UnixMilli(3)},
	}

	var w watcher
	w.reindex(mapfs)

	want := "[a.txt b.txt c.txt]"
	got := slices.Collect(w.scan(mapfs))
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Error("got", got, "    want", want)
	}

	got = slices.Collect(w.scan(mapfs))
	if len(got) != 0 {
		t.Error(got)
	}

	mapfs["a.txt"].ModTime = time.UnixMilli(4)
	want = "[a.txt]"
	got = slices.Collect(w.scan(mapfs))
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Error("got", got, "    want", want)
	}

	got = slices.Collect(w.scan(mapfs))
	if len(got) != 0 {
		t.Error(got)
	}
}
