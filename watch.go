package watchexec

import (
	"io/fs"
)

type globber struct {
	files *[]string
}

func (g globber) walkDirFunc(p string, d fs.DirEntry, err error) error {
	if err != nil {
		return fs.SkipDir
	}
	if d.IsDir() && d.Name()[0] == '.' {
		return fs.SkipDir
	}

	if !d.IsDir() {
		*g.files = append(*g.files, p)
	}
	return nil
}
