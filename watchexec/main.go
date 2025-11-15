package main

import (
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/myaaaaaaaaa/go-watchexec"
)

func isTerminal(f fs.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return stat.Mode()&fs.ModeCharDevice != 0
}

func main() {
	args := os.Args[1:]

	if !isTerminal(os.Stdout) {
		var w watchexec.Watcher
		w.FilesPerCycle = 4
		w.Wait = time.Millisecond * 50

		for {
			for f := range w.ScanCycles(os.DirFS("."), 100) {
				if f != "" {
					fmt.Println(f)
				}
			}
		}
	}
	if len(args) > 0 {
		for {
			watchexec.ExecOutput(os.Stderr, args)
			time.Sleep(time.Second * 2)
		}
	}
	if isTerminal(os.Stdin) {
		fmt.Println("stdin is terminal")
	}
}
