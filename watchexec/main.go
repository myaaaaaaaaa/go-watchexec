package main

import (
	"fmt"
	"io/fs"
	"os"
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

	if len(args) > 0 {
		fmt.Println("args:", args)
	}
	if isTerminal(os.Stdin) {
		fmt.Println("stdin is terminal")
	}
	if isTerminal(os.Stdout) {
		fmt.Println("stdout is terminal")
	}
}
