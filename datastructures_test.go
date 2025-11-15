package watchexec

import (
	"reflect"
	"testing"
)

func TestLruSet(t *testing.T) {
	t.Run("put and toSlice", func(t *testing.T) {
		s := newLruSet(3)
		s.put("a")
		s.put("b")
		s.put("c")

		expected := []string{"c", "b", "a"}
		if got := s.toSlice(); !reflect.DeepEqual(got, expected) {
			t.Errorf("toSlice() = %v, want %v", got, expected)
		}
	})

	t.Run("put refreshes existing element", func(t *testing.T) {
		s := newLruSet(3)
		s.put("a")
		s.put("b")
		s.put("c")
		s.put("b")

		expected := []string{"b", "c", "a"}
		if got := s.toSlice(); !reflect.DeepEqual(got, expected) {
			t.Errorf("toSlice() = %v, want %v", got, expected)
		}
	})

	t.Run("put evicts oldest element", func(t *testing.T) {
		s := newLruSet(3)
		s.put("a")
		s.put("b")
		s.put("c")
		s.put("d")

		expected := []string{"d", "c", "b"}
		if got := s.toSlice(); !reflect.DeepEqual(got, expected) {
			t.Errorf("toSlice() = %v, want %v", got, expected)
		}
	})

	t.Run("put with zero capacity", func(t *testing.T) {
		s := newLruSet(0)
		s.put("a")

		if got := s.toSlice(); len(got) != 0 {
			t.Errorf("toSlice() = %v, want empty slice", got)
		}
	})

	t.Run("put with one capacity", func(t *testing.T) {
		s := newLruSet(1)
		s.put("a")
		s.put("b")

		expected := []string{"b"}
		if got := s.toSlice(); !reflect.DeepEqual(got, expected) {
			t.Errorf("toSlice() = %v, want %v", got, expected)
		}
	})
}
