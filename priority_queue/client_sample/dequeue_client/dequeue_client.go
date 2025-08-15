package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type Ad struct {
	AdID       string `json:"adId"`
	GameFamily string `json:"gameFamily"`
	Priority   int    `json:"priority"`
	CreatedAt  string `json:"createdAt"`
}

// Simple shared rate limiter (stdlib only).
// Allows ~rate tokens per second with a burst capacity.
type limiter struct {
	tokens chan struct{}
	stop   chan struct{}
}

func newLimiter(rate int, burst int) *limiter {
	if rate <= 0 {
		rate = 1
	}
	if burst <= 0 {
		burst = 1
	}
	l := &limiter{
		tokens: make(chan struct{}, burst),
		stop:   make(chan struct{}),
	}
	// Fill initial burst
	for i := 0; i < burst; i++ {
		l.tokens <- struct{}{}
	}
	// Refill goroutine: 1 token every 1/rate seconds
	go func() {
		t := time.NewTicker(time.Second / time.Duration(rate))
		defer t.Stop()
		for {
			select {
			case <-t.C:
				select {
				case l.tokens <- struct{}{}:
				default: // bucket full
				}
			case <-l.stop:
				return
			}
		}
	}()
	return l
}

func (l *limiter) Wait() {
	<-l.tokens
}
func (l *limiter) Stop() { close(l.stop) }

func main() {
	// Flags
	base := flag.String("base", "http://localhost:8080", "Queue server base URL")
	workers := flag.Int("workers", 1, "Concurrent dequeue workers")
	rate := flag.Int("rate", 5, "Dequeue rate (ads per second, shared across all workers)")
	burst := flag.Int("burst", 5, "Burst capacity tokens")
	flag.Parse()

	dequeueURL := *base + "/dequeue"
	client := &http.Client{Timeout: 5 * time.Second}

	log.Printf("Dequeue target: %s | workers=%d | rate=%d ads/s | burst=%d",
		dequeueURL, *workers, *rate, *burst)

	var lim *limiter
	if *rate > 0 {
		lim = newLimiter(*rate, *burst)
		defer lim.Stop()
	}

	var wg sync.WaitGroup

	wg.Add(*workers)
	for i := 0; i < *workers; i++ {
		id := i + 1
		go func() {
			defer wg.Done()
			for {
				lim.Wait()
				req, err := http.NewRequest("POST", dequeueURL, nil)
				if err != nil {
					log.Printf("[w%d] new request error: %v", id, err)
					continue
				}
				resp, err := client.Do(req)
				if err != nil {
					log.Printf("[w%d] http error: %v", id, err)
					continue
				}
				func() {
					defer resp.Body.Close()
					if resp.StatusCode == http.StatusOK {
						b, _ := io.ReadAll(resp.Body)
						var ad Ad
						if err := json.Unmarshal(b, &ad); err != nil {
							log.Printf("[w%d] JSON error: %v body=%s", id, err, string(b))
							return
						}
						fmt.Printf("[w%d][priority=%d][created_at=%s] dequeued: id=%s family=%s\n", id, ad.Priority, ad.CreatedAt, ad.AdID, ad.GameFamily)
						return
					}
					if resp.StatusCode != http.StatusNotFound {
						b, _ := io.ReadAll(resp.Body)
						log.Printf("[w%d] status %d: %s", id, resp.StatusCode, string(b))
					}
				}()
			}
		}()
	}
	wg.Wait()
}
