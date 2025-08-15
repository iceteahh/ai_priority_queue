package queue

// DistributionByPriority returns distribution (ordered by q.priorities) and total count.
func (q *VideoProcessingQueue) DistributionByPriority() ([]PriorityDist, int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	dist := make([]PriorityDist, 0, len(q.priorities))
	total := 0

	counts := make(map[int]int, len(q.priorities))
	for _, p := range q.priorities {
		if queue := q.queueMap[p]; queue != nil {
			counts[p] = queue.Size
		} else {
			counts[p] = 0
		}
		total += counts[p]
	}

	for _, p := range q.priorities {
		c := counts[p]
		var pct float64
		if total > 0 {
			pct = float64(c) * 100.0 / float64(total)
		}
		dist = append(dist, PriorityDist{Priority: p, Count: c, Percent: pct})
	}
	return dist, total
}
