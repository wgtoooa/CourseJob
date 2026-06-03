package http

import (
	"CourseJob/internal/transport/http/dto"
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"sync"
	"time"
)

type scheduleUpdatedEvent struct {
	Type      string                  `json:"type"`
	UpdatedAt string                  `json:"updated_at"`
	Chunk     dto.WeekScheduleRequest `json:"chunk"`
}

type scheduleSSEBroker struct {
	mu          sync.RWMutex
	subscribers map[chan string]struct{}
}

func newScheduleSSEBroker() *scheduleSSEBroker {
	return &scheduleSSEBroker{
		subscribers: make(map[chan string]struct{}),
	}
}

func (b *scheduleSSEBroker) subscribe() chan string {
	ch := make(chan string, 16)
	b.mu.Lock()
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *scheduleSSEBroker) unsubscribe(ch chan string) {
	b.mu.Lock()
	if _, exists := b.subscribers[ch]; exists {
		delete(b.subscribers, ch)
		close(ch)
	}
	b.mu.Unlock()
}

func (b *scheduleSSEBroker) publish(event scheduleUpdatedEvent) {
	body, err := json.Marshal(event)
	if err != nil {
		return
	}

	msg := string(body)
	b.mu.RLock()
	defer b.mu.RUnlock()

	for subscriber := range b.subscribers {
		select {
		case subscriber <- msg:
		default:
			// Do not block global publishing on slow subscribers.
		}
	}
}

// ScheduleSSE streams schedule update events via Server-Sent Events.
// @Summary Subscribe to schedule updates
// @Description Opens an SSE stream and sends `schedule_updated` events when schedule import succeeds.
// @Tags Schedule
// @Produce text/event-stream
// @Success 200 {string} string "SSE stream"
// @Failure 405 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/sse/schedule [get]
func (h *Handler) ScheduleSSE(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodGet {
		writeJSON(w, nethttp.StatusMethodNotAllowed, jsonResponse{
			"status": "error",
			"error":  "method not allowed",
		})
		return
	}

	flusher, ok := w.(nethttp.Flusher)
	if !ok {
		writeJSON(w, nethttp.StatusInternalServerError, jsonResponse{
			"status": "error",
			"error":  "streaming is not supported",
		})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	subscriber := h.scheduleEvents.subscribe()
	defer h.scheduleEvents.unsubscribe(subscriber)

	if _, err := fmt.Fprint(w, ": connected\n\n"); err != nil {
		return
	}
	flusher.Flush()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case payload := <-subscriber:
			if _, err := fmt.Fprintf(w, "event: schedule_updated\ndata: %s\n\n", payload); err != nil {
				return
			}
			flusher.Flush()
		case <-heartbeat.C:
			if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
