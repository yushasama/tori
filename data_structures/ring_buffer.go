package data_structures

type Ring[T any] struct {
	data  []T
	head  int
	count int
}

func New[T any](size int) *Ring[T] {
	if size&(size-1) != 0 {
		panic("ring size must be power of 2")
	}

	return &Ring[T]{
		data: make([]T, size),
	}
}

func (r *Ring[T]) Push(v T) {
	if r.count < len(r.data) {
		r.data[(r.head+r.count)&(len(r.data)-1)] = v
		r.count++
	} else {
		r.data[r.head] = v
		r.head = (r.head + 1) & (len(r.data) - 1)
	}
}

func (r *Ring[T]) PruneBefore(threshold T, beforeFn func(a, b T) bool) {
	for r.count > 0 && beforeFn(r.data[r.head], threshold) {
		r.head = (r.head + 1) & (len(r.data) - 1)
		r.count--
	}
}

func (r *Ring[T]) PopNewest() (T, bool) {
	var zero T
	if r.count == 0 {
		return zero, false
	}

	idx := (r.head + r.count - 1) & (len(r.data) - 1)

	v := r.data[idx]
	r.count--

	return v, true
}

func (r *Ring[T]) PopOldest() (T, bool) {
	var zero T
	if r.count == 0 {
		return zero, false
	}

	v := r.data[r.head]
	r.head = (r.head + 1) & (len(r.data) - 1)
	r.count--

	return v, true
}

func (r *Ring[T]) Len() int      { return r.count }
func (r *Ring[T]) Capacity() int { return len(r.data) }
func (r *Ring[T]) IsFull() bool  { return r.count == len(r.data) }
func (r *Ring[T]) IsEmpty() bool { return r.count == 0 }

func (r *Ring[T]) Oldest() (T, bool) {
	var zero T

	if r.count == 0 {
		return zero, false
	}

	return r.data[r.head], true
}

func (r *Ring[T]) Reset() {
	r.head = 0
	r.count = 0
}
