package watchexec

import (
	"cmp"
	"iter"
	"maps"
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
func repeatChunks[T cmp.Ordered](s set[T], chunkSize, numChunks int) iter.Seq[[]T] {
	scanList := slices.Sorted(maps.Keys(s))
	if len(scanList) == 0 {
		var zero T
		scanList = append(scanList, zero)
	}
	chunks := slices.Chunk(scanList, chunkSize)

	return func(yield func([]T) bool) {
		for {
			for chunk := range chunks {
				if !yield(chunk) {
					return
				}
				numChunks--
				if numChunks <= 0 {
					return
				}
			}
		}
	}
}
