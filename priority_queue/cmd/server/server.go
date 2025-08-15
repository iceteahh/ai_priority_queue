package main

import (
	"context"
	"encoding/json"
	"icetea/priority_queue/config"
	"icetea/priority_queue/internal/httpapi"
	"icetea/priority_queue/internal/queue"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		panic(err)
	}

	// Log the loaded config as JSON (pretty printed)
	if b, err := json.MarshalIndent(cfg, "", "  "); err == nil {
		log.Printf("Loaded config:\n%s", string(b))
	} else {
		log.Printf("Loaded config (could not marshal): %+v", cfg)
	}

	q := queue.NewFromConfig(cfg)

	h := &httpapi.Handler{Q: q}
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           h.Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("HTTP server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
	log.Println("server stopped")
}
