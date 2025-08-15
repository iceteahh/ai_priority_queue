package queue

import (
	"icetea/priority_queue/internal/ads"
	"math"
	"time"
)

func (q *VideoProcessingQueue) Dequeue() *ads.Ad {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()
	selected := -1
	bestScore := math.Inf(-1)

	for _, p := range q.priorities {
		queue := q.queueMap[p]
		if queue == nil || queue.Size == 0 {
			continue
		}
		if selected == -1 {
			selected = p
			if !q.enableAntiStarvation {
				break
			}
		}
		head := queue.Front()
		waited := now.Sub(head.EnqueueAt)
		if waited >= time.Duration(head.Ad.MaxWaitTime)*time.Second {
			score := float64(p) + waited.Seconds()/float64(q.maximumWaitTime)*2
			if score > bestScore {
				bestScore = score
				selected = p
			}
		}
	}

	if selected == -1 {
		return nil
	}

	item := q.queueMap[selected].PopFront()
	q.removeFromFamilyIndex(item)
	q.removeFromTimeIndex(item)
	return item.Ad
}
