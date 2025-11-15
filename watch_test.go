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
		".d.txt":         &fstest.MapFile{},
		".e.txt/d/d/f":   &fstest.MapFile{},
		".f.txt/d/f.txt": &fstest.MapFile{},

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
	assert(".d.txt", ".d.txt")
	assert(".e.txt", "")
	assert(".f.txt", "")

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
