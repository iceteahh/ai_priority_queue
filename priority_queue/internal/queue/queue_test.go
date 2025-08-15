package queue

import (
	"testing"
	"time"

	"icetea/priority_queue/config"
	"icetea/priority_queue/internal/ads"
)

func newAd(id, family string, prio, maxWait int) *ads.Ad {
	return &ads.Ad{
		AdID:        id,
		Title:       "t-" + id,
		GameFamily:  family,
		Priority:    prio,
		MaxWaitTime: maxWait,
	}
}

func mustTime(t time.Time, d time.Duration) time.Time { return t.Add(d) }

// Helper: pull next k via Dequeue (consumes)
func takeDequeue(q *VideoProcessingQueue, k int) (out []string) {
	for i := 0; i < k; i++ {
		ad := q.Dequeue()
		if ad == nil {
			break
		}
		out = append(out, ad.AdID)
	}
	return
}

// Helper: PeekNext IDs (non-consuming)
func peekIDs(q *VideoProcessingQueue, k int) []string {
	var ids []string
	for _, ad := range q.PeekNext(k) {
		ids = append(ids, ad.AdID)
	}
	return ids
}

// === 1) Potential bug: insert at front when earlier than current head ===
// This test reveals incorrect behavior if the code inserts AFTER head
// when the new node should be at the very front (earliest time).
func TestInsertIntoPriorityByTime_ShouldGoToFrontWhenEarlier(t *testing.T) {
	queueConfig := config.Config{
		TotalPriority:        5,
		EnableAntiStarvation: true,
		MaximumWaitSeconds:   600,
		BTreeDegree:          16,
		TimeBoost:            2,
	}
	q := NewFromConfig(queueConfig)
	base := time.Now().Add(-1 * time.Hour)

	// Existing same-priority items at P=4 with times T0 < T1
	a := newAd("A", "F", 4, 600)
	b := newAd("B", "F", 4, 600)
	q.EnqueueWithTime(a, mustTime(base, 10*time.Minute)) // T0
	q.EnqueueWithTime(b, mustTime(base, 20*time.Minute)) // T1

	// New item C currently at priority 2, older than head(A), same family
	c := newAd("C", "F", 2, 600)
	q.EnqueueWithTime(c, mustTime(base, 5*time.Minute)) // T-5 (earliest)

	// Reprioritize family "F" to P=4; C must land BEFORE A (front of P=4)
	q.ReprioritizeByGameFamily("F", 4)

	// Dequeue first three: should be C, A, B (by time ascending within p=4)
	got := takeDequeue(q, 3)
	want := []string{"C", "A", "B"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("wrong order: got %v, want %v", got, want)
		}
	}
}

// === 2) Map iteration non-determinism in ReprioritizeByGameFamily ===
// If you iterate the map[*QueueItem]struct{} directly, order can scramble.
// This test expects stable ascending-by-time order after reprioritization.
func TestReprioritizeByGameFamily_OrderShouldBeStableByTime(t *testing.T) {
	queueConfig := config.Config{
		TotalPriority:        5,
		EnableAntiStarvation: true,
		MaximumWaitSeconds:   600,
		BTreeDegree:          16,
		TimeBoost:            2,
	}
	q := NewFromConfig(queueConfig)
	base := time.Now().Add(-2 * time.Hour)

	// Three items of the same family F, scattered across priorities and times.
	x := newAd("X", "F", 5, 600)
	y := newAd("Y", "F", 3, 600)
	z := newAd("Z", "F", 1, 600)

	q.EnqueueWithTime(x, mustTime(base, 30*time.Minute)) // older
	q.EnqueueWithTime(y, mustTime(base, 60*time.Minute))
	q.EnqueueWithTime(z, mustTime(base, 90*time.Minute)) // newest

	// Reprioritize all F to priority 4. Expected order inside P=4: X, Y, Z.
	q.ReprioritizeByGameFamily("F", 4)

	// Peek next 3 (no other items): should be X,Y,Z if stable; map-iteration may break this.
	got := peekIDs(q, 3)
	want := []string{"X", "Y", "Z"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("non-deterministic order from map iteration; got %v, want %v", got, want)
		}
	}
}

// === 3) Anti-starvation: lower priority should preempt when it exceeds MaxWaitTime ===
func TestDequeue_AntiStarvationPreempts(t *testing.T) {
	queueConfig := config.Config{
		TotalPriority:        3,
		EnableAntiStarvation: true,
		MaximumWaitSeconds:   600,
		BTreeDegree:          16,
		TimeBoost:            2,
	}
	q := NewFromConfig(queueConfig)
	now := time.Now()

	// High priority fresh
	h := newAd("H", "G", 3, 600)
	q.EnqueueWithTime(h, now.Add(-10*time.Second))

	// Low priority but starved beyond max wait
	l := newAd("L", "G", 1, 1) // 1 second MaxWaitTime
	q.EnqueueWithTime(l, now.Add(-10*time.Second))

	got := q.Dequeue()
	if got == nil || got.AdID != "L" {
		t.Fatalf("expected starved low-priority L to preempt, got %#v", got)
	}
}

// === 4) PeekNext should mirror Dequeue sequence (non-mutating) ===
func TestPeekNext_MatchesDequeue(t *testing.T) {
	queueConfig := config.Config{
		TotalPriority:        3,
		EnableAntiStarvation: true,
		MaximumWaitSeconds:   600,
		BTreeDegree:          16,
		TimeBoost:            2,
	}
	q := NewFromConfig(queueConfig)
	base := time.Now().Add(-30 * time.Minute)

	ids := []string{"A1", "A2", "B1", "C1", "C2", "C3"}
	prios := []int{3, 3, 2, 1, 1, 1}
	for i := range ids {
		q.EnqueueWithTime(newAd(ids[i], "X", prios[i], 600), base.Add(time.Duration(i)*time.Minute))
	}
	peek := peekIDs(q, 6)
	deq := takeDequeue(q, 6)
	if len(peek) != len(deq) {
		t.Fatalf("peek vs dequeue length mismatch: %v vs %v", peek, deq)
	}
	for i := range peek {
		if peek[i] != deq[i] {
			t.Fatalf("peek sequence diverges from dequeue at %d: peek=%v dequeue=%v", i, peek, deq)
		}
	}
}

// === 5) ListWaitingLongerThan should return exactly those older than cutoff ===
func TestListWaitingLongerThan(t *testing.T) {
	queueConfig := config.Config{
		TotalPriority:        3,
		EnableAntiStarvation: true,
		MaximumWaitSeconds:   600,
		BTreeDegree:          16,
		TimeBoost:            2,
	}
	q := NewFromConfig(queueConfig)
	now := time.Now()

	old := newAd("OLD", "A", 2, 600)
	neww := newAd("NEW", "A", 2, 600)
	q.EnqueueWithTime(old, now.Add(-10*time.Minute))
	q.EnqueueWithTime(neww, now.Add(-1*time.Minute))

	res := q.ListWaitingLongerThan(5 * time.Minute)
	ids := map[string]bool{}
	for _, ad := range res {
		ids[ad.AdID] = true
	}
	if !ids["OLD"] || ids["NEW"] {
		t.Fatalf("ListWaitingLongerThan mismatch, got %v", ids)
	}
}

// === 6) SetMaximumWaitTime caps existing items ===
func TestSetMaximumWaitTime_CapsExisting(t *testing.T) {
	queueConfig := config.Config{
		TotalPriority:        3,
		EnableAntiStarvation: true,
		MaximumWaitSeconds:   600,
		BTreeDegree:          16,
		TimeBoost:            2,
	}
	q := NewFromConfig(queueConfig)
	a := newAd("A", "K", 2, 10_000) // very large
	q.EnqueueWithTime(a, time.Now())

	q.SetMaximumWaitTime(120)
	if a.MaxWaitTime != 120 {
		t.Fatalf("expected cap to 120, got %d", a.MaxWaitTime)
	}
}

// === 7) ReprioritizeByAgeOlderThan keeps ascending time order among moved items ===
func TestReprioritizeByAgeOlderThan_OrderStable(t *testing.T) {
	queueConfig := config.Config{
		TotalPriority:        5,
		EnableAntiStarvation: true,
		MaximumWaitSeconds:   600,
		BTreeDegree:          16,
		TimeBoost:            2,
	}
	q := NewFromConfig(queueConfig)
	now := time.Now().Add(-2 * time.Hour)

	// Mixed ages: O1,O2 older; N1,N2 newer
	o1 := newAd("O1", "F1", 2, 600)
	o2 := newAd("O2", "F2", 3, 600)
	n1 := newAd("N1", "F3", 4, 600)
	n2 := newAd("N2", "F4", 5, 600)

	q.EnqueueWithTime(o1, now.Add(10*time.Minute))
	q.EnqueueWithTime(o2, now.Add(20*time.Minute))
	q.EnqueueWithTime(n1, time.Now().Add(-5*time.Minute))
	q.EnqueueWithTime(n2, time.Now().Add(-1*time.Minute))

	// Move items older than 30m ago to priority 4
	q.ReprioritizeByAgeOlderThan(30*time.Minute, 4)

	// Peek first 2 from priority 4 should be O1 then O2 (ascending by time)
	// We can't peek by priority, so just PeekNext and filter the first two from prio=4
	var seen []string
	for _, ad := range q.PeekNext(10) {
		if ad.Priority == 4 {
			seen = append(seen, ad.AdID)
		}
	}
	if len(seen) < 2 || seen[0] != "O1" || seen[1] != "O2" {
		t.Fatalf("moved order not stable, got %v want [O1 O2 ...]", seen)
	}
}

// === 8) After Dequeue, indices should no longer expose the item ===
func TestIndicesMaintenance_AfterDequeue(t *testing.T) {
	queueConfig := config.Config{
		TotalPriority:        3,
		EnableAntiStarvation: true,
		MaximumWaitSeconds:   600,
		BTreeDegree:          16,
		TimeBoost:            2,
	}
	q := NewFromConfig(queueConfig)
	now := time.Now().Add(-10 * time.Minute)

	a := newAd("A", "Fam", 2, 600)
	q.EnqueueWithTime(a, now)

	// Dequeue it away
	deq := q.Dequeue()
	if deq == nil || deq.AdID != "A" {
		t.Fatalf("dequeue failed")
	}

	// Should not show up in time-based listing anymore
	res := q.ListWaitingLongerThan(1 * time.Minute)
	for _, ad := range res {
		if ad.AdID == "A" {
			t.Fatalf("dequeued ad still present in time index via ListWaitingLongerThan")
		}
	}

	// Reprioritizing its family should do nothing / not panic
	q.ReprioritizeByGameFamily("Fam", 3)
}

// === 9) Stable tie-breaking when EnqueueAt equal (uses seq for index, list order for FIFO) ===
func TestStableOrdering_SameEnqueueAt(t *testing.T) {
	queueConfig := config.Config{
		TotalPriority:        3,
		EnableAntiStarvation: true,
		MaximumWaitSeconds:   600,
		BTreeDegree:          16,
		TimeBoost:            2,
	}
	q := NewFromConfig(queueConfig)
	t0 := time.Now().Add(-1 * time.Hour)

	a1 := newAd("A1", "T", 3, 600)
	a2 := newAd("A2", "T", 3, 600)
	a3 := newAd("A3", "T", 3, 600)

	// Same EnqueueAt but different seq (enqueue order A1, A2, A3)
	q.EnqueueWithTime(a1, t0)
	q.EnqueueWithTime(a2, t0)
	q.EnqueueWithTime(a3, t0)

	got := takeDequeue(q, 3)
	want := []string{"A1", "A2", "A3"} // FIFO stability within same timestamp
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unstable tie-breaking with equal EnqueueAt: got %v, want %v", got, want)
		}
	}
}

// === 10) DistributionByPriority sanity ===
func TestDistributionByPriority(t *testing.T) {
	queueConfig := config.Config{
		TotalPriority:        4,
		EnableAntiStarvation: true,
		MaximumWaitSeconds:   600,
		BTreeDegree:          16,
		TimeBoost:            2,
	}
	q := NewFromConfig(queueConfig)
	base := time.Now().Add(-10 * time.Minute)

	q.EnqueueWithTime(newAd("P4a", "D", 4, 600), base)
	q.EnqueueWithTime(newAd("P4b", "D", 4, 600), base)
	q.EnqueueWithTime(newAd("P3a", "D", 3, 600), base)
	q.EnqueueWithTime(newAd("P1a", "D", 1, 600), base)

	dist, total := q.DistributionByPriority()
	if total != 4 {
		t.Fatalf("total=%d, want 4", total)
	}
	// priorities slice is descending in New
	wantCounts := map[int]int{4: 2, 3: 1, 2: 0, 1: 1}
	for _, d := range dist {
		if d.Count != wantCounts[d.Priority] {
			t.Fatalf("priority %d count=%d want=%d", d.Priority, d.Count, wantCounts[d.Priority])
		}
	}
}
