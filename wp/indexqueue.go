package wikipath

import "container/list"

// PathQueue represents a queue of `IndexPath`s.
type PathQueue struct {
	q *list.List
}

// NewPathQueue creates a `PathQueue`.
func NewPathQueue() *PathQueue {
	return &PathQueue{
		q: list.New(),
	}
}

// Enqueue pushes an item to the `PathQueue`.
func (pq *PathQueue) Enqueue(path *IndexPath) {
	pq.q.PushBack(path)
}

// Dequeue pops an item from the `PathQueue`.
func (pq *PathQueue) Dequeue() *IndexPath {
	item := pq.q.Front()
	if item == nil {
		return nil
	}
	pq.q.Remove(item)
	return item.Value.(*IndexPath)
}
