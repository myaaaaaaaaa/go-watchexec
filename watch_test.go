package watchexec

import (
	"errors"
	"fmt"
	"io/fs"
	"iter"
	"maps"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"testing/synctest"
	"time"
)

func assertEquals(t *testing.T, got any, want any) {
	t.Helper()
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Error("got", got, "    want", want)
	}
}

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

func touch(mapfs fstest.MapFS, name string, i int) {
	mapfs[name] = &fstest.MapFile{ModTime: time.UnixMilli(int64(i))}
}

func TestScan(t *testing.T) {
	mapfs := fstest.MapFS{}

	for i := range 7 {
		i := i + 1
		touch(mapfs, fmt.Sprintf("%d.txt", i), i)
	}

	{
		var w Watcher
		w.FilesAtOnce = 2

		got := slices.Collect(w.ScanCycles(mapfs, 2))
		assertEquals(t, got, "[2.txt 4.txt]")
	}

	var w Watcher

	want := "[1.txt 2.txt 3.txt 4.txt 5.txt 6.txt 7.txt   ]"
	got := slices.Collect(w.ScanCycles(mapfs, 10))
	assertEquals(t, got, want)

	for range 3 {
		want = "[         ]"
		got = slices.Collect(w.ScanCycles(mapfs, 10))
		assertEquals(t, got, want)
	}

	mapfs["4.txt"].ModTime = time.UnixMilli(20)
	want = "[   4.txt      ]"
	got = slices.Collect(w.ScanCycles(mapfs, 10))
	assertEquals(t, got, want)

	for range 3 {
		want = "[         ]"
		got = slices.Collect(w.ScanCycles(mapfs, 10))
		assertEquals(t, got, want)
	}

	for n := range 4 {
		w.FilesAtOnce = n
		want = "[         ]"
		got = slices.Collect(w.ScanCycles(mapfs, 10))
		assertEquals(t, got, want)
	}
}

func pullInf[T any](it iter.Seq[T]) (func() T, func()) {
	next, done := iter.Pull(it)
	return func() T {
		rt, ok := next()
		if !ok {
			panic("iterator not infinite")
		}
		return rt
	}, done
}

func TestEditing(t *testing.T) {
	mapfs := fstest.MapFS{}
	touch(mapfs, "a.txt", 2)
	touch(mapfs, "b.txt", 4)
	touch(mapfs, "c.txt", 7)

	var w Watcher

	next, done := pullInf(w.ScanCycles(mapfs, 1000))
	defer done()

	assertEquals(t, next(), "a.txt")
	assertEquals(t, next(), "b.txt")
	assertEquals(t, next(), "c.txt")
	assertEquals(t, next(), "")
	assertEquals(t, next(), "")
	assertEquals(t, next(), "")

	touch(mapfs, "b.txt", 10)
	assertEquals(t, next(), "")
	assertEquals(t, next(), "b.txt")
	assertEquals(t, next(), "")
	assertEquals(t, next(), "")
	assertEquals(t, next(), "")
	assertEquals(t, next(), "")

	touch(mapfs, "b.txt", 11)
	assertEquals(t, next(), "b.txt")
	assertEquals(t, next(), "")

	for n := range 7 {
		for range n {
			assertEquals(t, next(), "")
		}
		touch(mapfs, "b.txt", 20+n)
		assertEquals(t, next(), "b.txt")
	}
}

type statRecordFS struct {
	fs.FS
	cb func(string)
}

func (fsys statRecordFS) Stat(name string) (fs.FileInfo, error) {
	var _ fs.StatFS = statRecordFS{}
	fsys.cb(name)
	return fs.Stat(fsys.FS, name)
}
func TestStat(t *testing.T) {
	var counter int
	fs := statRecordFS{
		FS: fstest.MapFS{
			"a.txt": &fstest.MapFile{},
			"b.txt": &fstest.MapFile{},
			"c.txt": &fstest.MapFile{},
			"d.txt": &fstest.MapFile{},
			"e.txt": &fstest.MapFile{},
		},
		cb: func(string) { counter++ },
	}

	w := Watcher{FilesAtOnce: 1}
	_ = slices.Collect(w.ScanCycles(fs, 2))
	assertEquals(t, counter, 3)
}

func TestWait(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		mapfs := fstest.MapFS{}
		touch(mapfs, "a.txt", 2)
		touch(mapfs, "b.txt", 4)
		touch(mapfs, "c.txt", 7)

		for n := range 8 {
			w := Watcher{
				FilesAtOnce:      n + 1,
				WaitBetweenPolls: time.Minute,
			}

			start := time.Now()
			_ = slices.Collect(w.ScanCycles(mapfs, 60))
			assertEquals(t, time.Since(start), time.Hour)
		}
	})

	synctest.Test(t, func(t *testing.T) {
		w := Watcher{WaitBetweenPolls: time.Minute}
		start := time.Now()
		_ = slices.Collect(w.ScanCycles(fstest.MapFS{}, 60))
		assertEquals(t, time.Since(start), time.Hour)
	})
}
