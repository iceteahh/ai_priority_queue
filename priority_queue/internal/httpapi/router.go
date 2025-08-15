package httpapi

import (
	"net/http"
)

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("GET /healthz", h.Health)

	// Core queue operations
	mux.HandleFunc("POST /enqueue", h.Enqueue)
	mux.HandleFunc("POST /dequeue", h.Dequeue)
	mux.HandleFunc("GET /peek", h.Peek)
	mux.HandleFunc("GET /distribution", h.Distribution)
	mux.HandleFunc("GET /waiting", h.Waiting)

	// Admin / maintenance
	mux.HandleFunc("POST /reprioritize/family", h.ReprioritizeFamily)
	mux.HandleFunc("POST /reprioritize/age", h.ReprioritizeAge)
	mux.HandleFunc("POST /settings/antiStarvation", h.SetAntiStarvation)
	mux.HandleFunc("POST /settings/maximumWait", h.SetMaximumWait)

	return mux
}
