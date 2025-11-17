package watchexec

import (
	"io/fs"
	"iter"
	"maps"
	"slices"
	"time"
)

func walk(fsys fs.FS) set[string] {
	rt := set[string]{}

	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return fs.SkipDir
		}

		rt[p] = struct{}{}
		if p == "." {
		} else if !d.IsDir() {
		} else if d.IsDir() {
			if d.Name()[0] == '.' {
				return fs.SkipDir
			}
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
	WaitBetweenPolls time.Duration

	lastModified int64
	files        set[string]
}

func (w *Watcher) reindex(fsys fs.FS) {
	f := walk(fsys)
	maps.Insert(f, maps.All(w.files))
	w.files = f
}
func (w *Watcher) statUpdate(fsys fs.FS, files []string) string {
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

func (w *Watcher) RunFor(t time.Duration, fsys fs.FS, f func(string)) {
	for file := range w.scanCycles(fsys, int(t/w.WaitBetweenPolls)) {
		if file != "" {
			f(file)
			w.lastModified = time.Now().UnixMilli()
		}
	}
}

func (w *Watcher) scanCycles(fsys fs.FS, cycles int) iter.Seq[string] {
	w.reindex(fsys)

	filesAtOnce := max(1, w.FilesAtOnce)

	return func(yield func(string) bool) {
		var likelyEditing []string

		files := slices.Values(slices.Sorted(maps.Keys(w.files)))
		for chunk := range repeatChunks(files, filesAtOnce, cycles) {
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
