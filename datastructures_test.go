package watchexec

import (
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
