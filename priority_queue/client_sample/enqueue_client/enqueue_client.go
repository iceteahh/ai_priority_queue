package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// Ad represents the ad structure expected by your queue server.
type EnqueueRequest struct {
	Ad Ad `json:"ad"`
	// Optional. If set, server uses this time instead of Now.
	EnqueueAt *time.Time `json:"enqueueAt,omitempty"`
}

type Ad struct {
	AdID           string   `json:"adId"`
	Title          string   `json:"title"`
	GameFamily     string   `json:"gameFamily"`
	TargetAudience []string `json:"targetAudience"`
	Priority       int      `json:"priority"`
	CreatedAt      string   `json:"createdAt"`
	MaxWaitTime    int      `json:"maxWaitTime"`
}

func main() {
	// Config

	rate := flag.Int("rate", 5, "Enqueue ad per second")
	total := flag.Int("total", 10, "Total ads")
	flag.Parse()

	endpoint := "http://localhost:8080/enqueue" 

	interval := time.Second / time.Duration(*rate)
	client := &http.Client{Timeout: 5 * time.Second}

	gameFamilies := []string{"RPG-Fantasy", "Shooter", "Puzzle", "Sports"}

	fmt.Printf("Sending %d ads at %d ads/sec to %s\n", *total, *rate, endpoint)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	sent := 0
	for range ticker.C {
		enqueueRequest := EnqueueRequest{
			Ad: Ad{
				AdID:        fmt.Sprintf("ad_%d", time.Now().UnixNano()),
				Title:       fmt.Sprintf("Game %d", rand.Intn(1000)),
				GameFamily:  gameFamilies[rand.Intn(len(gameFamilies))],
				Priority:    rand.Intn(3) + 1, // 1..3
				CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
				MaxWaitTime: 10,
			},
		}

		body, _ := json.Marshal(enqueueRequest)
		req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
		if err != nil {
			log.Printf("Error creating request: %v", err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error sending ad: %v", err)
			continue
		}
		resp.Body.Close()

		log.Printf("Enqueued: %s (%s, P%d)", enqueueRequest.Ad.AdID, enqueueRequest.Ad.GameFamily, enqueueRequest.Ad.Priority)

		sent++
		if sent >= *total {
			break
		}
	}

	fmt.Println("Done sending ads.")
}
