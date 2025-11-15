package watchexec

import (
	"io/fs"
	"iter"
	"maps"
	"slices"
	"time"
)

func walk(fsys fs.FS, root string) set[string] {
	rt := set[string]{}

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
	files        set[string]
}

func (w *watcher) reindex(fsys fs.FS) {
	f := walk(fsys, ".")
	maps.Insert(f, maps.All(w.files))
	w.files = f
}
func (w *watcher) statUpdate(fsys fs.FS, files []string) string {
	s := ""
	for _, file := range files {
		modified := statTime(fsys, file)
		if w.lastModified < modified {
			s = file
			w.lastModified = modified
		}
	}
	return s
}
func (w *watcher) scan(fsys fs.FS) iter.Seq[string] {
	chunkSize := max(1, w.filesPerCycle)
	chunks := slices.Chunk(slices.Sorted(maps.Keys(w.files)), chunkSize)

	return func(yield func(string) bool) {
		for chunk := range chunks {
			time.Sleep(w.wait)
			if !yield(w.statUpdate(fsys, chunk)) {
				return
			}
		}
	}
}
func (w *watcher) ScanCycles(fsys fs.FS, cycles int) iter.Seq[string] {
	return func(yield func(string) bool) {
		var likelyEditing []string
		for {
			for s := range w.scan(fsys) {
				if s == "" {
					s = w.statUpdate(fsys, likelyEditing)
				}
				if s != "" {
					cap := max(1, w.filesPerCycle)
					likelyEditing = lruPut(likelyEditing, s, cap)
				}

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
