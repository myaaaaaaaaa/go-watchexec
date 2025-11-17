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

func (w *Watcher) pollAndUpdate(fsys fs.FS, filesToPoll []string) string {
	modifiedFile := ""
	for _, file := range filesToPoll {
		modifiedTime := int64(0)

		stat, err := fs.Stat(fsys, file)
		if err == nil {
			modifiedTime = stat.ModTime().UnixMilli()
		}

		if w.lastModified < modifiedTime {
			modifiedFile = file
			w.lastModified = modifiedTime
		}
	}
	return modifiedFile
}

func (w *Watcher) RunFor(t time.Duration, fsys fs.FS, f func(string)) {
	for file := range w.iterations(fsys, int(t/w.WaitBetweenPolls)) {
		if file != "" {
			f(file)
			w.lastModified = time.Now().UnixMilli()
		}
	}
}

func (w *Watcher) iterations(fsys fs.FS, cycles int) iter.Seq[string] {
	filesAtOnce := max(1, w.FilesAtOnce)

	return func(yield func(string) bool) {
		var likelyEditing []string

		for filesToPoll := range repeatChunks(walk(fsys), filesAtOnce, cycles) {
			time.Sleep(w.WaitBetweenPolls)

			modifiedFile := w.pollAndUpdate(fsys, filesToPoll)
			if modifiedFile == "" {
				modifiedFile = w.pollAndUpdate(fsys, likelyEditing)
			}
			if modifiedFile != "" {
				likelyEditing = lruPut(likelyEditing, modifiedFile, filesAtOnce)
			}

			if !yield(modifiedFile) {
				return
			}
		}
	}
}
