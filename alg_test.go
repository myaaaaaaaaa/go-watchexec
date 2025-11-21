package watchexec

import (
	"bytes"
	"iter"
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

		chunks := slices.Collect(repeatChunks(slices.Values([]byte(s)), chunkSize, numChunks))
		got := bytes.Join(chunks, []byte(" "))
		assertEquals(t, string(got), want)

		if s != "" {
			assertEquals(t, len(chunks), numChunks)
			if numChunks > 0 {
				assertEquals(t, len(got), (chunkSize+1)*numChunks-1)
			} else {
				assertEquals(t, len(got), 0)
			}
		} else {
			assertEquals(t, len(chunks), 0)
			assertEquals(t, len(got), 0)
		}
	}

	assert("hello", 1, 1, "h")
	assert("hello", 2, 1, "he")
	assert("hello", 1, 2, "h e")
	assert("hello", 2, 2, "he ll")
	assert("hello", 3, 4, "hel loh ell ohe")
	assert("hello", 6, 3, "helloh ellohe llohel")

	assert("", 1, 1, "")
	assert("", 2, 1, "")
	assert("", 3, 1, "")
	assert("", 1, 2, "")
	assert("", 1, 3, "")
	assert("", 2, 2, "")
}

func TestRepeatIter(t *testing.T) {
	t.Run("non-empty", func(t *testing.T) {
		src := []int{1, 2, 3}
		seq := repeatIter(slices.Values(src))
		pull, stop := iter.Pull(seq)
		defer stop()

		var got []int
		for i := 0; i < 5; i++ {
			v, ok := pull()
			if !ok {
				t.Fatal("iterator finished unexpectedly")
			}
			got = append(got, v)
		}

		want := []int{1, 2, 3, 1, 2}
		if !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("empty", func(t *testing.T) {
		src := []int{}
		seq := repeatIter(slices.Values(src))
		pull, stop := iter.Pull(seq)
		defer stop()

		v, ok := pull()
		if ok {
			t.Fatalf("got %v, want nothing", v)
		}
	})

	t.Run("can be stopped", func(t *testing.T) {
		src := []int{1, 2, 3}
		seq := repeatIter(slices.Values(src))

		var got []int
		for v := range seq {
			got = append(got, v)
			if len(got) >= 5 {
				break
			}
		}

		want := []int{1, 2, 3, 1, 2}
		if !slices.Equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}
