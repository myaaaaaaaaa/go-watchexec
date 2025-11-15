package watchexec

import (
	"io/fs"
)

func walk(fsys fs.FS, root string) []string {
	var rt []string

	err := fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return fs.SkipDir
		}
		if d.IsDir() && d.Name()[0] == '.' {
			return fs.SkipDir
		}

		if !d.IsDir() {
			rt = append(rt, p)
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
}

func statTime(fsys fs.FS, f string) int64 {
	stat, err := fs.Stat(fsys, f)
	if err != nil {
		return 0
	}
	return stat.ModTime().UnixMilli()
}
