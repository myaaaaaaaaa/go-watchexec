package watchexec

import (
	"io/fs"
	"iter"
	"maps"
	"slices"
	"time"
)

type fsIndex = map[string]struct{}

func walk(fsys fs.FS, root string) fsIndex {
	rt := fsIndex{}

	err := fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return fs.SkipDir
		}
		if d.IsDir() {
			name := d.Name()
			if len(name) > 1 && name[0] == '.' {
				return fs.SkipDir
			}
		}

		if !d.IsDir() {
			rt[p] = struct{}{}
		}
		return nil
	})

	if err != nil {
		// This shouldn't happen
		panic(err)
	}
	return rt
}

type watcher struct {
	filesPerCycle int
	wait          time.Duration

	lastModified int64
	files        fsIndex
}

func (w *watcher) reindex(fsys fs.FS) {
	f := walk(fsys, ".")
	maps.Insert(f, maps.All(w.files))
	w.files = f
}
func (w *watcher) scan(fsys fs.FS) iter.Seq[string] {
	chunkSize := max(1, w.filesPerCycle)
	chunks := slices.Chunk(slices.Sorted(maps.Keys(w.files)), chunkSize)

	return func(yield func(string) bool) {
		for chunk := range chunks {
			time.Sleep(w.wait)

			s := ""
			for _, file := range chunk {
				modified := statTime(fsys, file)
				if w.lastModified < modified {
					s = file
					w.lastModified = modified
				}
			}

			if !yield(s) {
				return
			}
		}
	}
}
func (w *watcher) ScanCycles(fsys fs.FS, cycles int) iter.Seq[string] {
	return func(yield func(string) bool) {
		for {
			for s := range w.scan(fsys) {
				if !yield(s) {
					return
				}
				cycles--
				if cycles <= 0 {
					return
				}
			}
		}
	}
}

func statTime(fsys fs.FS, f string) int64 {
	stat, err := fs.Stat(fsys, f)
	if err != nil {
		return 0
	}
	return stat.ModTime().UnixMilli()
}
