package queue

func (q *VideoProcessingQueue) SetMaximumWaitTime(maxWait int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.maximumWaitTime = maxWait
	for _, queue := range q.queueMap {
		for item := queue.Head; item != nil; item = item.Next {
			if item.Ad.MaxWaitTime > maxWait {
				item.Ad.MaxWaitTime = maxWait
			}
		}
	}
}
