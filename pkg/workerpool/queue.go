package workerpool

import (
	"sync"
)

type Queue[resType any] interface {
	Push(j Job[resType])
	Pop() (Job[resType], bool)
}

type node[resType any] struct {
	job  Job[resType]
	next *node[resType]
}

type queue[resType any] struct {
	start *node[resType]
	end   *node[resType]
	size  uint64
	mu    *sync.Mutex
}

func CreateQueue[resType any]() Queue[resType] {
	return &queue[resType]{
		mu: &sync.Mutex{},
	}
}

func (q *queue[resType]) Push(j Job[resType]) {
	q.mu.Lock()
	nn := &node[resType]{
		next: nil,
		job:  j,
	}
	if q.size == 0 {
		q.size++
		q.start = nn
		q.end = nn
		q.mu.Unlock()
		return
	}
	q.size++
	q.start.next = nn
	q.start = q.start.next
	q.mu.Unlock()
}

func (q *queue[resType]) Pop() (Job[resType], bool) {
	q.mu.Lock()
	if q.size == 0 {
		q.mu.Unlock()
		return nil, false
	}
	ret := q.end
	q.end = q.end.next
	q.size--
	q.mu.Unlock()
	return ret.job, true
}
