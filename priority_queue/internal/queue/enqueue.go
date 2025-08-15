package queue

import (
	"icetea/priority_queue/internal/ads"
	"time"
)

func (q *VideoProcessingQueue) EnqueueWithTime(ad *ads.Ad, enqueuedAt time.Time) {
	q.mu.Lock()
	defer q.mu.Unlock()

	ad.Priority = q.normalizePriority(ad.Priority)
	if ad.MaxWaitTime > q.maximumWaitTime {
		ad.MaxWaitTime = q.maximumWaitTime
	}

	q.nextSeq++
	item := &QueueItem{
		Ad:        ad,
		EnqueueAt: enqueuedAt,
		seq:       q.nextSeq, // IMPORTANT: set seq before indexing
	}
	q.insertIntoPriorityByTime(item, ad.Priority)

	if _, ok := q.gameFamilyIndex[ad.GameFamily]; !ok {
		q.gameFamilyIndex[ad.GameFamily] = make(map[*QueueItem]struct{})
	}
	q.gameFamilyIndex[ad.GameFamily][item] = struct{}{}

	q.timeIndex.ReplaceOrInsert(timeIndexItem{when: item.EnqueueAt, seq: item.seq, item: item})
}

func (q *VideoProcessingQueue) Enqueue(ad *ads.Ad) {
	q.mu.Lock()
	defer q.mu.Unlock()

	ad.Priority = q.normalizePriority(ad.Priority)
	if ad.MaxWaitTime > q.maximumWaitTime {
		ad.MaxWaitTime = q.maximumWaitTime
	}
	queue, ok := q.queueMap[ad.Priority]
	if !ok {
		queue = &DList{}
		q.queueMap[ad.Priority] = queue
	}

	q.nextSeq++
	item := &QueueItem{
		Ad:        ad,
		EnqueueAt: time.Now(),
		seq:       q.nextSeq, // IMPORTANT: set seq before indexing
	}
	queue.PushBack(item)

	if _, ok := q.gameFamilyIndex[ad.GameFamily]; !ok {
		q.gameFamilyIndex[ad.GameFamily] = make(map[*QueueItem]struct{})
	}
	q.gameFamilyIndex[ad.GameFamily][item] = struct{}{}

	q.timeIndex.ReplaceOrInsert(timeIndexItem{when: item.EnqueueAt, seq: item.seq, item: item})
}
