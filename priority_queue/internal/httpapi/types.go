package httpapi

import (
	"time"
)

// Requests

type EnqueueRequest struct {
	Ad struct {
		AdID           string   `json:"adId"`
		Title          string   `json:"title"`
		GameFamily     string   `json:"gameFamily"`
		TargetAudience []string `json:"targetAudience"`
		Priority       int      `json:"priority"`
		CreatedAt      string   `json:"createdAt"`
		MaxWaitTime    int      `json:"maxWaitTime"`
	} `json:"ad"`
	// Optional. If set, server uses this time instead of Now.
	EnqueueAt *time.Time `json:"enqueueAt,omitempty"`
}

type PeekRequest struct {
	N int `json:"n"`
}

type ReprioritizeFamilyRequest struct {
	Family      string `json:"family"`
	NewPriority int    `json:"newPriority"`
}

type ReprioritizeAgeRequest struct {
	// Duration string like "5s", "3m", "1h"
	Age         string `json:"age"`
	NewPriority int    `json:"newPriority"`
}

type WaitingRequest struct {
	// Duration string like "5s", "3m"
	Age string `json:"age"`
}

type AntiStarvationRequest struct {
	Enable bool `json:"enable"`
}

type MaximumWaitRequest struct {
	MaximumWait int `json:"maximumWait"`
}

// Responses

type ErrorResponse struct {
	Error string `json:"error"`
}

type OKResponse struct {
	OK bool `json:"ok"`
}
