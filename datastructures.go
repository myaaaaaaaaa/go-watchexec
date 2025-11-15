package watchexec

import "container/list"

type lruSet struct {
	capacity int
	list     *list.List
	cache    map[string]*list.Element
}

func newLruSet(capacity int) *lruSet {
	return &lruSet{
		capacity: capacity,
		list:     list.New(),
		cache:    make(map[string]*list.Element),
	}
}

func (s *lruSet) put(value string) {
	if s.capacity == 0 {
		return
	}

	if elem, ok := s.cache[value]; ok {
		s.list.MoveToFront(elem)
		return
	}

	if s.list.Len() == s.capacity {
		elem := s.list.Back()
		if elem != nil {
			s.list.Remove(elem)
			delete(s.cache, elem.Value.(string))
		}
	}

	elem := s.list.PushFront(value)
	s.cache[value] = elem
}

func (s *lruSet) toSlice() []string {
	var result []string
	for elem := s.list.Front(); elem != nil; elem = elem.Next() {
		result = append(result, elem.Value.(string))
	}
	return result
}
