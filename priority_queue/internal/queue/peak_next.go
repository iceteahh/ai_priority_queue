package queue

import (
	"icetea/priority_queue/internal/ads"
	"time"
)

// PeekNext returns the next n ads in the exact order Dequeue would pick, without mutation.
func (q *VideoProcessingQueue) PeekNext(n int) []*ads.Ad {
	q.mu.Lock()
	defer q.mu.Unlock()

	if n <= 0 {
		return nil
	}
	now := time.Now()
	result := make([]*ads.Ad, 0, n)

	type cursor struct {
		node *QueueItem
		size int
	}
	cursors := make(map[int]cursor, len(q.priorities))
	for _, p := range q.priorities {
		if dq := q.queueMap[p]; dq != nil && dq.Size > 0 {
			cursors[p] = cursor{node: dq.Head, size: dq.Size}
		}
	}

	for len(result) < n {
		selected := -1
		for _, p := range q.priorities {
			cur, ok := cursors[p]
			if !ok || cur.size == 0 || cur.node == nil {
				continue
			}
			if selected == -1 {
				selected = p
				if !q.enableAntiStarvation {
					break
				}
			}
			waited := now.Sub(cur.node.EnqueueAt)
			if waited >= time.Duration(cur.node.Ad.MaxWaitTime)*time.Second {
				selected = p
				break
			}
		}
		if selected == -1 {
			break
		}
		cur := cursors[selected]
		result = append(result, cur.node.Ad)
		cur.node = cur.node.Next
		cur.size--
		if cur.size <= 0 || cur.node == nil {
			delete(cursors, selected)
		} else {
			cursors[selected] = cur
		}
	}
	return result
}
