package watchexec

import (
	"bytes"
	"iter"
	"slices"
	"strings"
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

func head[T any](seq iter.Seq[T], n int) []T {
	return slices.Collect(func(yield func(T) bool) {
		next, stop := iter.Pull(seq)
		defer stop()

		for range n {
			v, _ := next()
			yield(v)
		}
	})
}

func TestRepeatIter(t *testing.T) {
	t.Run("non-empty", func(t *testing.T) {
		const STR = "qwertyuiop"
		for iterSize := range 10 {
			for collectSize := range 10 {
				inputStr := STR[:iterSize+1]
				seq := repeatIter(slices.Values([]byte(inputStr)))
				got := head(seq, collectSize)
				assertEquals(t, string(got), strings.Repeat(inputStr, collectSize)[:collectSize])
			}
		}
	})

	t.Run("empty", func(t *testing.T) {
		for iterSize := range 10 {
			for collectSize := range 10 {
				seq := repeatIter(slices.Values(bytes.Repeat([]byte{0}, iterSize)))
				got := head(seq, collectSize)
				assertEquals(t, string(got), strings.Repeat("\x00", collectSize))
			}
		}
	})

	t.Run("live update", func(t *testing.T) {
		var n int
		seq := repeatIter(func(yield func(int) bool) {
			for range n {
				if !yield(n) {
					return
				}
			}
			n++
		})

		got := head(seq, 12)
		assertEquals(t, got, "[0 1 2 2 3 3 3 4 4 4 4 5]")
	})
}

func TestRepeatChunks(t *testing.T) {
	assert := func(s string, chunkSize, numChunks int, want string) {
		t.Helper()

		chunks := slices.Collect(repeatChunks(slices.Values([]byte(s)), chunkSize, numChunks))
		assertEquals(t, len(chunks), numChunks)
		got := bytes.Join(chunks, []byte(" "))
		assertEquals(t, len(got), (chunkSize+1)*numChunks-1)
		assertEquals(t, string(got), want)
	}

	assert("hello", 1, 1, "h")
	assert("hello", 2, 1, "he")
	assert("hello", 1, 2, "h e")
	assert("hello", 2, 2, "he ll")
	assert("hello", 3, 4, "hel loh ell ohe")
	assert("hello", 6, 3, "helloh ellohe llohel")

	assert("", 1, 1, "\x00")
	assert("", 2, 1, "\x00\x00")
	assert("", 3, 1, "\x00\x00\x00")
	assert("", 1, 2, "\x00 \x00")
	assert("", 1, 3, "\x00 \x00 \x00")
	assert("", 2, 2, "\x00\x00 \x00\x00")

	for chunkSize := range 5 {
		for numChunks := range 5 {
			chunkSize := chunkSize + 1
			numChunks := numChunks + 1

			want := " " + strings.Repeat("\x00", chunkSize)
			want = strings.Repeat(want, numChunks)
			want = want[1:]
			assert("", chunkSize, numChunks, want)
		}
	}
}
