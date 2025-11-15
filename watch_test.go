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

func assertEquals(t *testing.T, got any, want any) {
	t.Helper()
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Error("got", got, "    want", want)
	}
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
		assertEquals(t, got, want)
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
	mapfs := fstest.MapFS{}

	for i := range 7 {
		i := int64(i) + 1
		mapfs[fmt.Sprintf("%d.txt", i)] = &fstest.MapFile{ModTime: time.UnixMilli(i)}
	}

	var w watcher
	w.reindex(mapfs)

	{
		w := w
		w.filesPerCycle = 2

		got := slices.Collect(w.ScanCycles(mapfs, 2))
		assertEquals(t, got, "[2.txt 4.txt]")
	}

	want := "[1.txt 2.txt 3.txt 4.txt 5.txt 6.txt 7.txt]"
	got := slices.Collect(w.scan(mapfs))
	assertEquals(t, got, want)

	want = "[      ]"
	got = slices.Collect(w.scan(mapfs))
	assertEquals(t, got, want)

	mapfs["4.txt"].ModTime = time.UnixMilli(20)
	want = "[   4.txt   ]"
	got = slices.Collect(w.scan(mapfs))
	assertEquals(t, got, want)

	want = "[      ]"
	got = slices.Collect(w.scan(mapfs))
	assertEquals(t, got, want)
}
