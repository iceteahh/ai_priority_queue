package queue

import (
	"time"

	"github.com/google/btree"
)

func (q *VideoProcessingQueue) ReprioritizeByAgeOlderThan(age time.Duration, newPriority int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.timeIndex == nil {
		return
	}

	targetPriority := q.normalizePriority(newPriority)

	cutoff := time.Now().Add(-age)

	// Collect first to avoid mutating lists while walking the B-Tree.
	toMove := make([]*QueueItem, 0, 64)
	q.timeIndex.AscendLessThan(
		timeIndexItem{when: cutoff, seq: 1 << 62}, // iterate all items older than cutoff
		func(it btree.Item) bool {
			ti := it.(timeIndexItem)
			if ti.item != nil && ti.item.Ad.Priority != targetPriority {
				toMove = append(toMove, ti.item)
			}
			return true
		},
	)

	// Move in ascending enqueue order (preserves global FIFO among moved items).
	for _, item := range toMove {
		src := q.queueMap[item.Ad.Priority]
		if src == nil || src.Size == 0 {
			continue
		}
		src.Remove(item)
		item.Ad.Priority = targetPriority
		q.insertIntoPriorityByTime(item, targetPriority)
	}
}
