package queue

import (
	"icetea/priority_queue/config"
	"icetea/priority_queue/internal/ads"
	"math"
	"sync"
	"time"

	"github.com/google/btree"
)

// QueueItem is a node that lives inside a priority list and indices.
type QueueItem struct {
	Ad        *ads.Ad
	EnqueueAt time.Time
	Next      *QueueItem
	Prev      *QueueItem
	seq       int64 // unique per enqueue for stable ordering/deletes
}

type PriorityDist struct {
	Priority int
	Count    int
	Percent  float64 // 0..100
}

type VideoProcessingQueue struct {
	mu                   sync.Mutex
	queueMap             map[int]*DList // priority -> queue
	priorities           []int          // descending: highest first
	totalPriority        int
	enableAntiStarvation bool
	maximumWaitTime      int
	gameFamilyIndex      map[string]map[*QueueItem]struct{}
	timeIndex            *btree.BTree // ordered by EnqueueAt
	nextSeq              int64
	timeBoost            float64
}

// New creates a new queue. maximumWait caps per-ad MaxWaitTime.
func New(
	totalPriority int,
	enableStarvation bool,
	maximumWait int,
	btreeDegree int,
	timeBoost float64,
) *VideoProcessingQueue {
	if totalPriority <= 0 {
		totalPriority = 2
	}
	if btreeDegree <= 0 {
		btreeDegree = 16
	}
	if timeBoost <= 0 {
		timeBoost = 1
	}
	priorities := make([]int, totalPriority)
	for i := 0; i < totalPriority; i++ {
		priorities[i] = totalPriority - i // Descending
	}
	return &VideoProcessingQueue{
		queueMap:             make(map[int]*DList),
		priorities:           priorities,
		totalPriority:        totalPriority,
		enableAntiStarvation: enableStarvation,
		maximumWaitTime:      maximumWait,
		gameFamilyIndex:      make(map[string]map[*QueueItem]struct{}),
		timeIndex:            btree.New(btreeDegree),
		timeBoost:            timeBoost,
	}
}

func NewFromConfig(cfg config.Config) *VideoProcessingQueue {
	return New(
		cfg.TotalPriority,
		cfg.EnableAntiStarvation,
		cfg.MaximumWaitSeconds,
		cfg.BTreeDegree,
		cfg.TimeBoost,
	)
}

func (q *VideoProcessingQueue) normalizePriority(p int) int {
	if p < 1 {
		return 1
	}
	if p > q.totalPriority {
		return q.totalPriority
	}
	return p
}

func (q *VideoProcessingQueue) removeFromFamilyIndex(item *QueueItem) {
	if familyItems, ok := q.gameFamilyIndex[item.Ad.GameFamily]; ok {
		delete(familyItems, item)
		if len(familyItems) == 0 {
			delete(q.gameFamilyIndex, item.Ad.GameFamily)
		}
	}
}

func (q *VideoProcessingQueue) removeFromTimeIndex(item *QueueItem) {
	if item == nil || q.timeIndex == nil {
		return
	}
	q.timeIndex.Delete(timeIndexItem{when: item.EnqueueAt, seq: item.seq, item: item})
}

// insertIntoPriorityByTime inserts 'item' into the DList for priority 'p'
// keeping ascending EnqueueAt order. It uses the global timeIndex to find the
// nearest same-priority neighbor, so insertion is O(log N) + O(1) splice.
//
// PRECONDITIONS:
// - 'item' has already been removed from its old priority list
// - item.Ad.Priority does NOT need to be 'p' yet (we pass 'p' explicitly)
func (q *VideoProcessingQueue) insertIntoPriorityByTime(item *QueueItem, p int) {
	queue, ok := q.queueMap[p]
	if !ok {
		queue = &DList{}
		q.queueMap[p] = queue
	}

	if queue.Size == 0 {
		queue.Head = item
		queue.Tail = item
		queue.Size = 1
		return
	}

	// Find nearest same-priority neighbor via timeIndex.
	var prevItem *QueueItem // greatest (time,seq) <= (item.time, +inf) with same priority p

	// predecessor (older or equal)
	q.timeIndex.DescendLessOrEqual(
		timeIndexItem{when: item.EnqueueAt, seq: math.MaxInt64},
		func(it btree.Item) bool {
			ti := it.(timeIndexItem)
			if item != ti.item && ti.item.Ad.Priority == p {
				prevItem = ti.item
				return false // found nearest
			}
			return true
		},
	)

	if prevItem == nil {
		item.Next = queue.Head
		item.Prev = nil
		queue.Head.Prev = item
		queue.Head = item
	} else {
		nxt := prevItem.Next
		item.Prev = prevItem
		item.Next = nxt
		prevItem.Next = item
		if nxt != nil {
			nxt.Prev = item
		} else {
			queue.Tail = item
		}
	}
	queue.Size++
}
