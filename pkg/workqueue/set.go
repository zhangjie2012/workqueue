package workqueue

type t interface{}
type empty struct{}
type Set map[t]empty

func NewSet() *Set {
	s := make(Set)
	return &s
}

func (s Set) Has(item interface{}) bool {
	_, exists := s[item]
	return exists
}

func (s Set) Insert(item interface{}) {
	s[item] = empty{}
}

func (s Set) Delete(item interface{}) {
	delete(s, item)
}
