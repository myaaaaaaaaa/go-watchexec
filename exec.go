package watchexec

import (
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// https://en.wikipedia.org/wiki/ANSI_escape_code#3-bit_and_4-bit
func wrapColor(v string, colors ...int) string {
	for _, color := range colors {
		v = "\x1b[" + strconv.Itoa(color) + "m" + v
	}
	v += "\x1b[0m"
	return v
}

func ExecOutput(out io.Writer, args []string) {
	if false {
		// capture mode - loses colors
		output, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		fmt.Println(string(output), err)
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = out
	cmd.Stderr = out

	header := strings.Join([]string{
		"",
		time.Now().Format("3:04:05 PM"),
		strings.Join(args, " "),
		"",
	}, "  ◄►  ")

	// Home cursor; Clear formatting; Clear screen x2
	const CLEAR = "\x1b[H\x1b[0m\x1b[2J\x1b[3J"
	fmt.Fprintln(out, CLEAR+wrapColor("\x1b[2K"+header, 1))
	err := cmd.Run()

	if err != nil {
		fmt.Fprint(out, "\n\x1b[999H"+wrapColor("\x1b[2Kerror: "+err.Error(), 97, 41, 1)+"\x1b[H")
	}
}
