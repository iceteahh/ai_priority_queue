package queue

import (
	"icetea/priority_queue/internal/ads"
	"time"

	"github.com/google/btree"
)

// ListWaitingLongerThan: O(logN + K) using the time btree (ascending).
func (q *VideoProcessingQueue) ListWaitingLongerThan(age time.Duration) []*ads.Ad {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.timeIndex == nil {
		return nil
	}

	cutoff := time.Now().Add(-age)
	var out []*ads.Ad
	q.timeIndex.AscendLessThan(timeIndexItem{when: cutoff, seq: 1 << 62}, func(it btree.Item) bool {
		ti := it.(timeIndexItem)
		out = append(out, ti.item.Ad)
		return true
	})
	return out
}
