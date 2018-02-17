package wikipath

import "container/list"

type Direction bool

const FORWARD Direction = true
const REVERSE Direction = false

type PathQueue struct {
	q *list.List
}

func NewPathQueue() *PathQueue {
	return &PathQueue{
		q: list.New(),
	}
}

func (pq *PathQueue) Enqueue(path *IndexPath) {
	pq.q.PushBack(path)
}

func (pq *PathQueue) Dequeue() *IndexPath {
	item := pq.q.Front()
	if item == nil {
		return nil
	}
	pq.q.Remove(item)
	return item.Value.(*IndexPath)
}
