package queue

func (q *VideoProcessingQueue) SetEnableAntiStarvation(enable bool) {
	q.mu.Lock()
	q.enableAntiStarvation = enable
	q.mu.Unlock()
}

func (q *VideoProcessingQueue) IsEnableAntiStarvation() bool {
	return q.enableAntiStarvation
}
