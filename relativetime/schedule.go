package relativetime

type timer[T Time[T, D], D Duration] struct {
	f      func(T)
	when   T
	period D
	index  int
}

type queue[T Time[T, D], D Duration] []*timer[T, D]

// Implement sort.Interface
func (q queue[T, D]) Len() int {
	return len(q)
}

func (q queue[T, D]) Less(i, j int) bool {
	return q[i].when.Before(q[j].when)
}

func (q queue[T, D]) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index, q[j].index = i, j
}

// Implement container.heap.Interface
func (q *queue[T, D]) Push(x any) {
	t := x.(*timer[T, D])
	t.index = len(*q)
	*q = append(*q, t)
}

func (q *queue[T, D]) Pop() any {
	n := len(*q) - 1
	t := (*q)[n]
	(*q)[n] = nil
	t.index = -1
	*q = (*q)[:n]
	return t
}

func (q queue[T, D]) peek() *timer[T, D] {
	if len(q) == 0 {
		return nil
	}
	return q[0]
}
