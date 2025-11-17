package watchexec

import (
	"iter"
	"slices"
)

type set[T comparable] = map[T]struct{}

func lruPut[T comparable](s []T, value T, cap int) []T {
	s = slices.DeleteFunc(s, func(e T) bool {
		return e == value
	})
	s = slices.Insert(s, 0, value)
	s = (s)[:min(len(s), cap)]
	return s
}
func repeatChunks[T any](src iter.Seq[T], chunkSize, numChunks int) iter.Seq[[]T] {
	return func(yield func([]T) bool) {
		var rt []T

		put := func(elem T) bool {
			rt = append(rt, elem)

			if len(rt) >= chunkSize {
				if !yield(slices.Clone(rt)) {
					return false
				}
				numChunks--
				if numChunks <= 0 {
					return false
				}
				rt = rt[:0]
			}
			return true
		}

		for {
			isEmpty := true
			for elem := range src {
				isEmpty = false
				if !put(elem) {
					return
				}
			}
			if isEmpty {
				var zero T
				if !put(zero) {
					return
				}
			}
		}
	}
}
