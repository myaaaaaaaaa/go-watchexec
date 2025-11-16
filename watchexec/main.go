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

func watch(f func(string)) {
	var w watchexec.Watcher
	w.FilesAtOnce = 4
	w.WaitBetweenPolls = time.Millisecond * 50

	for {
		w.RunFor(time.Second*5, os.DirFS("."), f)
	}
}

func main() {
	args := os.Args[1:]

	if !isTerminal(os.Stdout) {
		watch(func(s string) {
			fmt.Println(s)
		})
	}
	if len(args) > 0 {
		watch(func(string) {
			watchexec.ExecOutput(os.Stderr, args)
		})
	}
	if isTerminal(os.Stdin) {
		fmt.Println("stdin is terminal")
	}
}
