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

type Watcher struct {
	FilesAtOnce      int
	LastModified     int64
	WaitBetweenPolls time.Duration

	files set[string]
}

func (w *Watcher) reindex(fsys fs.FS) {
	f := walk(fsys, ".")
	maps.Insert(f, maps.All(w.files))
	w.files = f
}
func (w *Watcher) statUpdate(fsys fs.FS, files []string) string {
	s := ""
	for _, file := range files {
		modified := statTime(fsys, file)
		if w.LastModified < modified {
			s = file
			w.LastModified = modified
		}
	}
	return s
}

func (w *Watcher) ScanCycles(fsys fs.FS, cycles int) iter.Seq[string] {
	w.reindex(fsys)

	filesAtOnce := max(1, w.FilesAtOnce)

	return func(yield func(string) bool) {
		var likelyEditing []string

		for chunk := range repeatChunks(slices.Sorted(maps.Keys(w.files)), filesAtOnce, cycles) {
			time.Sleep(w.WaitBetweenPolls)

			s := w.statUpdate(fsys, chunk)
			if s == "" {
				s = w.statUpdate(fsys, likelyEditing)
			}
			if s != "" {
				cap := filesAtOnce
				likelyEditing = lruPut(likelyEditing, s, cap)
			}

			if !yield(s) {
				return
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
