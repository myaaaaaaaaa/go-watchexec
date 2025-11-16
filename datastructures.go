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
func repeatChunks[T any](scanList []T, chunkSize, numChunks int) iter.Seq[[]T] {
	chunks := slices.Collect(slices.Chunk(scanList, chunkSize))
	if len(chunks) == 0 {
		chunks = append(chunks, nil)
	}

	return func(yield func([]T) bool) {
		for {
			for _, chunk := range chunks {
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
