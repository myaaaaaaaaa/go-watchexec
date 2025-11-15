package watchexec

import (
	"testing"
)

func TestLruSet(t *testing.T) {
	t.Run("put and slice", func(t *testing.T) {
		s := newLruSet(3)
		s.put("a")
		s.put("b")
		s.put("c")

		assertEquals(t, s.slice(), "[c b a]")
	})

	t.Run("put refreshes existing element", func(t *testing.T) {
		s := newLruSet(3)
		s.put("a")
		s.put("b")
		s.put("c")
		s.put("b")

		assertEquals(t, s.slice(), "[b c a]")
	})

	t.Run("put evicts oldest element", func(t *testing.T) {
		s := newLruSet(3)
		s.put("a")
		s.put("b")
		s.put("c")
		s.put("d")

		assertEquals(t, s.slice(), "[d c b]")
	})

	t.Run("put with zero capacity", func(t *testing.T) {
		s := newLruSet(0)
		s.put("a")

		assertEquals(t, s.slice(), "[]")
	})

	t.Run("put with one capacity", func(t *testing.T) {
		s := newLruSet(1)
		s.put("a")
		s.put("b")

		assertEquals(t, s.slice(), "[b]")
	})
}
