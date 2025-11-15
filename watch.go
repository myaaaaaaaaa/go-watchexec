package watchexec

import (
	"io/fs"
	"maps"
)

func walk(fsys fs.FS, root string) map[string]int64 {
	rt := map[string]int64{}

	err := fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return fs.SkipDir
		}
		if d.IsDir() && d.Name()[0] == '.' {
			return fs.SkipDir
		}

		if !d.IsDir() {
			rt[p] = 0
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
	lastModified int64
	files        map[string]int64
}

func (w *watcher) reindex(fsys fs.FS) {
	f := walk(fsys, ".")
	maps.Insert(f, maps.All(w.files))
	w.files = f
}

func statTime(fsys fs.FS, f string) int64 {
	stat, err := fs.Stat(fsys, f)
	if err != nil {
		return 0
	}
	return stat.ModTime().UnixMilli()
}
