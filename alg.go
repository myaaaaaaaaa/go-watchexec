package watchexec

import (
	"iter"
	"slices"
)

func lruPut[T comparable](s []T, value T, cap int) []T {
	s = slices.DeleteFunc(s, func(e T) bool {
		return e == value
	})
	s = slices.Insert(s, 0, value)
	s = (s)[:min(len(s), cap)]
	return s
}

func repeatIter[T any](src iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for {
			isEmpty := true
			for elem := range src {
				isEmpty = false
				if !yield(elem) {
					return
				}
			}
			if isEmpty {
				return
			}
		}
	}
}
func repeatChunks[T any](src iter.Seq[T], chunkSize, numChunks int) iter.Seq[[]T] {
	return func(yield func([]T) bool) {
		var rt []T

		for elem := range repeatIter(src) {
			rt = append(rt, elem)
			if len(rt) >= chunkSize {
				if !yield(slices.Clone(rt)) {
					return
				}
				numChunks--
				if numChunks <= 0 {
					return
				}
				rt = rt[:0]
			}
		}
	}
}
