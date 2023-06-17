package relativetime

type timer[T Time[T, D], D Duration] struct {
	f      func(T)
	when   T
	period D
	index  int
}

type queue[T Time[T, D], D Duration] []*timer[T, D]

func (q queue[T, D]) peek() *timer[T, D] {
	if len(q) == 0 {
		return nil
	}
	return q[0]
}

// Heap management

// If container/heap isn't good enough for the Go runtime, then it's not good
// enough for clock (see siftupTimer and siftdownTimer in runtime/time.go).

// insert adds the timer t and ensures the heap property is maintained.
// Inserting a timer that already exists in the queue will likely lead to
// undefined behavior.
func (q *queue[T, D]) insert(t *timer[T, D]) {
	t.index = len(*q)
	// Grow the queue and get it heapified again
	*q = append(*q, t)
	q.siftup(t)
}

// remove removes the timer t and ensures the heap property is maintained.
// Removing a timer that has never been inserted into the queue will likely
// lead to undefined behavior.
func (q *queue[T, D]) remove(t *timer[T, D]) {
	i := t.index
	n := len(*q) - 1

	if i != n {
		// Move the last timer into this one's old home
		(*q)[i] = (*q)[n]
		(*q)[i].index = i

		// Shrink the queue and get it heapified again
		(*q)[:n].fix((*q)[i])
	}

	(*q)[n] = nil
	t.index = -1
	*q = (*q)[:n]
}

// fix ensures the heap property is maintained after a change in timer t.
// Fixing a timer that is not in the queue will likely lead to undefined
// behavior.
func (q queue[T, D]) fix(t *timer[T, D]) {
	i0 := t.index
	if q.siftdown(t); t.index == i0 {
		q.siftup(t)
	}
}

// siftup maintains heap property by moving the timer t towards the top of
// the heap. Panics if it has an invalid index.
func (q queue[T, D]) siftup(t *timer[T, D]) {
	i := t.index
	for i > 0 {
		p := (i - 1) / 4 // parent

		// Swap needed in this direction?
		if !q[p].when.After(t.when) {
			break
		}

		// Move parent here
		q[i] = q[p]
		q[i].index = i

		// Check parent's old home
		i = p
	}
	if t != q[i] {
		// Place original timer in its new home
		q[i] = t
		q[i].index = i
	}
}

// siftdown maintains heap property by moving the timer t towards the bottom
// of the heap. Panics if it has an invalid index.
func (q queue[T, D]) siftdown(t *timer[T, D]) {
	i := t.index
	n := len(q)
	for {
		c := i*4 + 1 // left child
		c4 := c + 3  // right child
		if c >= n {
			// No children, can't go any lower from here
			break
		}
		if c4 >= n {
			c4 = n - 1
		}
		w := q[c].when

		// If there are additional children, make sure to pick the favorite
		for i := c + 1; i <= c4; i++ {
			if w.After(q[i].when) {
				w = q[i].when
				c = i
			}
		}

		// Swap needed in this direction?
		if !t.when.After(w) {
			break
		}

		// Move child here
		q[i] = q[c]
		q[i].index = i

		// Check child's old home
		i = c
	}
	if t != q[i] {
		// Place original timer in its new home
		q[i] = t
		q[i].index = i
	}
}
