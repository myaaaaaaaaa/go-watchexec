package watchexec

import (
	"io/fs"
	"iter"
	"time"
)

func walk(fsys fs.FS) iter.Seq[string] {
	return func(yield func(string) bool) {
		err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return fs.SkipDir
			}

			if !yield(p) {
				return fs.SkipAll
			}

			if d.IsDir() {
				if d.Name()[0] == '.' && p != "." {
					return fs.SkipDir
				}
			}

			return nil
		})

		if err != nil {
			// This shouldn't happen
			panic(err)
		}
	}
}

type Watcher struct {
	FilesAtOnce      int
	WaitBetweenPolls time.Duration

	lastModified int64
}

func (w *Watcher) updateModified(fsys fs.FS, files []string) string {
	s := ""
	for _, file := range files {
		modifiedTime := int64(0)

		stat, err := fs.Stat(fsys, file)
		if err == nil {
			modifiedTime = stat.ModTime().UnixMilli()
		}

		if w.lastModified < modifiedTime {
			s = file
			w.lastModified = modifiedTime
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
	filesAtOnce := max(1, w.FilesAtOnce)

	return func(yield func(string) bool) {
		var likelyEditing []string

		for chunk := range repeatChunks(walk(fsys), filesAtOnce, cycles) {
			time.Sleep(w.WaitBetweenPolls)

			s := w.updateModified(fsys, chunk)
			if s == "" {
				s = w.updateModified(fsys, likelyEditing)
			}
			if s != "" {
				likelyEditing = lruPut(likelyEditing, s, filesAtOnce)
			}

			if !yield(s) {
				return
			}
		}
	}
}
