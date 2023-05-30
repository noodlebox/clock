package steppedtime

import (
	"container/heap"
)

type timer struct {
	f      func(Time)
	when   Time
	period Duration
	index  int
}

type queue []*timer

// Implement sort.Interface
func (q queue) Len() int {
	return len(q)
}

func (q queue) Less(i, j int) bool {
	return q[i].when.Before(q[j].when)
}

func (q queue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index, q[j].index = i, j
}

// Implement container.heap.Interface
func (q *queue) Push(x any) {
	t := x.(*timer)
	t.index = len(*q)
	*q = append(*q, t)
}

func (q *queue) Pop() any {
	n := len(*q) - 1
	t := (*q)[n]
	(*q)[n] = nil
	t.index = -1
	*q = (*q)[:n]
	return t
}

func (q queue) peek() *timer {
	if len(q) == 0 {
		return nil
	}
	return q[0]
}

// Check schedule for pending events that should trigger now.
func (c *Clock) checkSchedule() {
	for t := c.queue.peek(); t != nil && !t.when.After(c.now); t = c.queue.peek() {
		if t.period.Seconds() <= 0 {
			c.unschedule(t)
		} else {
			t.when = c.now.Add(t.period)
			c.reschedule(t)
		}
		t.f(c.now)
	}
}

func (c *Clock) schedule(t *timer) {
	heap.Push(&c.queue, t)
}

func (c *Clock) unschedule(t *timer) {
	if t.index == -1 {
		return
	}
	heap.Remove(&c.queue, t.index)
}

func (c *Clock) reschedule(t *timer) {
	if t.index == -1 {
		c.schedule(t)
		return
	}
	heap.Fix(&c.queue, t.index)
}
