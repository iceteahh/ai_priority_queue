package httpapi

import (
	"encoding/json"
	"icetea/priority_queue/internal/ads"
	"icetea/priority_queue/internal/queue"
	"net/http"
	"strconv"
	"time"
)

type Handler struct {
	Q *queue.VideoProcessingQueue
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, ErrorResponse{Error: msg})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, OKResponse{OK: true})
}

func (h *Handler) Enqueue(w http.ResponseWriter, r *http.Request) {
	var req EnqueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	ad := &ads.Ad{
		AdID:           req.Ad.AdID,
		Title:          req.Ad.Title,
		GameFamily:     req.Ad.GameFamily,
		TargetAudience: req.Ad.TargetAudience,
		Priority:       req.Ad.Priority,
		CreatedAt:      req.Ad.CreatedAt,
		MaxWaitTime:    req.Ad.MaxWaitTime,
	}
	if req.EnqueueAt != nil {
		h.Q.EnqueueWithTime(ad, *req.EnqueueAt)
	} else {
		h.Q.Enqueue(ad)
	}
	writeJSON(w, http.StatusCreated, ad)
}

func (h *Handler) Dequeue(w http.ResponseWriter, r *http.Request) {
	ad := h.Q.Dequeue()
	if ad == nil {
		writeErr(w, http.StatusNotFound, "queue empty")
		return
	}
	writeJSON(w, http.StatusOK, ad)
}

func (h *Handler) Peek(w http.ResponseWriter, r *http.Request) {
	nStr := r.URL.Query().Get("n")
	n := 1
	if nStr != "" {
		if v, err := strconv.Atoi(nStr); err == nil && v > 0 {
			n = v
		} else {
			writeErr(w, http.StatusBadRequest, "invalid n")
			return
		}
	}
	ads := h.Q.PeekNext(n)
	writeJSON(w, http.StatusOK, ads)
}

func (h *Handler) Distribution(w http.ResponseWriter, r *http.Request) {
	dist, total := h.Q.DistributionByPriority()
	resp := struct {
		Total                int                  `json:"total"`
		Dist                 []queue.PriorityDist `json:"distribution"`
		EnableAntiStarvation bool                 `json:"enable_anti_starvation"`
	}{
		Total:                total,
		Dist:                 dist,
		EnableAntiStarvation: h.Q.IsEnableAntiStarvation(),
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) Waiting(w http.ResponseWriter, r *http.Request) {
	ageStr := r.URL.Query().Get("age")
	if ageStr == "" {
		var req WaitingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, "missing age in query or body")
			return
		}
		ageStr = req.Age
	}
	d, err := time.ParseDuration(ageStr)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid duration: "+err.Error())
		return
	}
	list := h.Q.ListWaitingLongerThan(d)
	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) ReprioritizeFamily(w http.ResponseWriter, r *http.Request) {
	var req ReprioritizeFamilyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.Family == "" || req.NewPriority == 0 {
		writeErr(w, http.StatusBadRequest, "family and newPriority required")
		return
	}
	h.Q.ReprioritizeByGameFamily(req.Family, req.NewPriority)
	writeJSON(w, http.StatusOK, OKResponse{OK: true})
}

func (h *Handler) ReprioritizeAge(w http.ResponseWriter, r *http.Request) {
	var req ReprioritizeAgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.Age == "" || req.NewPriority == 0 {
		writeErr(w, http.StatusBadRequest, "age and newPriority required")
		return
	}
	d, err := time.ParseDuration(req.Age)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid duration: "+err.Error())
		return
	}
	h.Q.ReprioritizeByAgeOlderThan(d, req.NewPriority)
	writeJSON(w, http.StatusOK, OKResponse{OK: true})
}

func (h *Handler) SetAntiStarvation(w http.ResponseWriter, r *http.Request) {
	var req AntiStarvationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	h.Q.SetEnableAntiStarvation(req.Enable)
	writeJSON(w, http.StatusOK, OKResponse{OK: true})
}

func (h *Handler) SetMaximumWait(w http.ResponseWriter, r *http.Request) {
	var req MaximumWaitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.MaximumWait <= 0 {
		writeErr(w, http.StatusBadRequest, "maximumWait must be > 0")
		return
	}
	h.Q.SetMaximumWaitTime(req.MaximumWait)
	writeJSON(w, http.StatusOK, OKResponse{OK: true})
}
