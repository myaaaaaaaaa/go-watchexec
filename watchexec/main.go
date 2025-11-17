package main

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
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
	w.FilesAtOnce = 6
	w.WaitBetweenPolls = time.Millisecond * 100

	for {
		w.RunFor(time.Minute, os.DirFS("."), f)
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
		if len(args) == 1 && strings.ContainsAny(args[0], "|&;") {
			args = []string{"sh", "-c", args[0]}
		}

		watch(func(string) {
			watchexec.ExecOutput(os.Stderr, args)
		})
	}
	if isTerminal(os.Stdin) {
		fmt.Println("stdin is terminal")
	}
}
