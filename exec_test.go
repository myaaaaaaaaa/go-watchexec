package watchexec

import (
	"bytes"
	"strings"
	"testing"
)

func TestExecOutput(t *testing.T) {
	assertContains := func(cmd, want string) {
		var buf bytes.Buffer
		ExecOutput(&buf, strings.Split(cmd, " "))
		got := buf.String()

		if !strings.Contains(got, want) {
			t.Error(got)
		}
	}

	assertContains("go version", "go version go1")
	assertContains("go pher", "error:")
	assertContains("gopher", "error:")
}
