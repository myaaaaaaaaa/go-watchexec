package watchexec

import "slices"

type set[T comparable] = map[T]struct{}

func lruPut[T comparable](s []T, value T, cap int) []T {
	s = slices.DeleteFunc(s, func(e T) bool {
		return e == value
	})
	s = slices.Insert(s, 0, value)
	s = (s)[:min(len(s), cap)]
	return s
}
