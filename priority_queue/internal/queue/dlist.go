package queue

// DList is a simple, GC-friendly doubly linked list specialized for *QueueItem.
type DList struct {
	Head, Tail *QueueItem
	Size       int
}

func (q *DList) PushBack(item *QueueItem) {
	if q.Tail == nil {
		q.Head = item
		q.Tail = item
	} else {
		item.Prev = q.Tail
		q.Tail.Next = item
		q.Tail = item
	}
	q.Size++
}

func (q *DList) Front() *QueueItem { return q.Head }

func (q *DList) PopFront() *QueueItem {
	if q.Head == nil {
		return nil
	}
	value := q.Head
	q.Head = q.Head.Next
	if q.Head != nil {
		q.Head.Prev = nil
	} else {
		q.Tail = nil
	}
	value.Next = nil
	value.Prev = nil
	q.Size--
	return value
}

func (q *DList) Remove(item *QueueItem) {
	if item == nil {
		return
	}
	if item.Prev != nil {
		item.Prev.Next = item.Next
	} else {
		q.Head = item.Next
	}
	if item.Next != nil {
		item.Next.Prev = item.Prev
	} else {
		q.Tail = item.Prev
	}
	item.Next = nil
	item.Prev = nil
	q.Size--
}
