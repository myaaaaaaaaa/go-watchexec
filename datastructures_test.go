package watchexec

import (
	"bytes"
	"slices"
	"testing"
)

func TestLruPut(t *testing.T) {
	t.Run("put", func(t *testing.T) {
		var s []string
		s = lruPut(s, "a", 3)
		s = lruPut(s, "b", 3)
		s = lruPut(s, "c", 3)
		assertEquals(t, s, "[c b a]")
	})

	t.Run("put refreshes existing element", func(t *testing.T) {
		var s []string
		s = lruPut(s, "a", 3)
		s = lruPut(s, "b", 3)
		s = lruPut(s, "c", 3)
		s = lruPut(s, "b", 3)
		assertEquals(t, s, "[b c a]")
	})

	t.Run("put evicts oldest element", func(t *testing.T) {
		var s []string
		s = lruPut(s, "a", 3)
		s = lruPut(s, "b", 3)
		s = lruPut(s, "c", 3)
		s = lruPut(s, "d", 3)
		assertEquals(t, s, "[d c b]")
	})

	t.Run("put with zero capacity", func(t *testing.T) {
		var s []string
		s = lruPut(s, "a", 0)
		assertEquals(t, s, "[]")
	})

	t.Run("put with one capacity", func(t *testing.T) {
		var s []string
		s = lruPut(s, "a", 1)
		s = lruPut(s, "b", 1)
		assertEquals(t, s, "[b]")
	})
}

func TestRepeatChunks(t *testing.T) {
	assert := func(s string, chunkSize, numChunks int, want string) {
		t.Helper()

		got := bytes.Join(slices.Collect(repeatChunks([]byte(s), chunkSize, numChunks)), []byte(" "))
		assertEquals(t, string(got), want)
	}

	assert("hello", 1, 1, "h")
	assert("hello", 2, 1, "he")
	assert("hello", 1, 2, "h e")
	assert("hello", 2, 2, "he ll")
	assert("hello", 3, 4, "hel lo hel lo")

	assert("", 1, 1, "")
	assert("", 2, 1, "")
	assert("", 3, 1, "")
	assert("", 1, 2, " ")
	assert("", 1, 3, "  ")
}
