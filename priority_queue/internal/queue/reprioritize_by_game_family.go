package queue

func (q *VideoProcessingQueue) ReprioritizeByGameFamily(family string, newPriority int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	targetPriority := q.normalizePriority(newPriority)

	items, found := q.gameFamilyIndex[family]
	if !found {
		return
	}

	// Preserve enqueue order by appending in the order we walk old queues.
	for item := range items {
		if item.Ad.Priority == targetPriority {
			continue
		}
		q.queueMap[item.Ad.Priority].Remove(item)
		item.Ad.Priority = targetPriority
		q.insertIntoPriorityByTime(item, targetPriority)
	}
}
