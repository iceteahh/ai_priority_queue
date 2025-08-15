package queue

import (
	"icetea/priority_queue/internal/ads"
	"testing"
	"time"
)

func makeItem(id string) *QueueItem {
	return &QueueItem{
		Ad:        &ads.Ad{AdID: id, GameFamily: "G", Priority: 1, MaxWaitTime: 10},
		EnqueueAt: time.Now(),
	}
}

func TestDList_PushBackPopFrontSize(t *testing.T) {
	var dl DList
	if got := dl.Front(); got != nil {
		t.Fatalf("Front on empty list = %v, want nil", got)
	}
	if dl.Size != 0 {
		t.Fatalf("Size = %d, want 0", dl.Size)
	}

	a := makeItem("a")
	b := makeItem("b")
	c := makeItem("c")

	dl.PushBack(a)
	dl.PushBack(b)
	dl.PushBack(c)

	if dl.Size != 3 {
		t.Fatalf("Size after pushes = %d, want 3", dl.Size)
	}
	if dl.Front() != a {
		t.Fatalf("Front = %v, want a", dl.Front())
	}

	// Pop order FIFO
	if got := dl.PopFront(); got != a {
		t.Fatalf("PopFront #1 = %v, want a", got)
	}
	if got := dl.PopFront(); got != b {
		t.Fatalf("PopFront #2 = %v, want b", got)
	}
	if got := dl.PopFront(); got != c {
		t.Fatalf("PopFront #3 = %v, want c", got)
	}
	if got := dl.PopFront(); got != nil {
		t.Fatalf("PopFront on empty = %v, want nil", got)
	}
	if dl.Size != 0 {
		t.Fatalf("Size after pops = %d, want 0", dl.Size)
	}
}

func TestDList_Remove_Head_Middle_Tail(t *testing.T) {
	var dl DList
	a := makeItem("a")
	b := makeItem("b")
	c := makeItem("c")
	dl.PushBack(a)
	dl.PushBack(b)
	dl.PushBack(c)

	// Remove head
	dl.Remove(a)
	if dl.Size != 2 || dl.Head != b || dl.Tail != c || b.Prev != nil {
		t.Fatalf("remove head failed: size=%d head=%v tail=%v", dl.Size, dl.Head, dl.Tail)
	}

	// Remove middle
	dl.Remove(b)
	if dl.Size != 1 || dl.Head != c || dl.Tail != c || c.Prev != nil || c.Next != nil {
		t.Fatalf("remove middle failed: size=%d head=%v tail=%v", dl.Size, dl.Head, dl.Tail)
	}

	// Remove tail (also the only element)
	dl.Remove(c)
	if dl.Size != 0 || dl.Head != nil || dl.Tail != nil {
		t.Fatalf("remove tail failed: size=%d head=%v tail=%v", dl.Size, dl.Head, dl.Tail)
	}
}
