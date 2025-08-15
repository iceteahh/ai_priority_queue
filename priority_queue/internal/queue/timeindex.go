package queue

import (
	"time"

	"github.com/google/btree"
)

// timeIndexItem makes entries unique and stably ordered in timeIndex.
type timeIndexItem struct {
	when time.Time
	seq  int64       // unique/enqueue order tie-breaker
	item *QueueItem  // back-pointer to the queue node
}

func (a timeIndexItem) Less(b btree.Item) bool {
	x := b.(timeIndexItem)
	if a.when.Before(x.when) {
		return true
	}
	if a.when.After(x.when) {
		return false
	}
	return a.seq < x.seq
}
