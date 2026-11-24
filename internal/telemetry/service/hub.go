package service

import (
	"sync"

	"ptmuda.tech/backend/internal/telemetry/model"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[chan model.TelemetrySnapshot]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[chan model.TelemetrySnapshot]struct{}),
	}
}

func (h *Hub) Subscribe() chan model.TelemetrySnapshot {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch := make(chan model.TelemetrySnapshot, 16)
	h.clients[ch] = struct{}{}
	return ch
}

func (h *Hub) Unsubscribe(ch chan model.TelemetrySnapshot) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[ch]; ok {
		delete(h.clients, ch)
		close(ch)
	}
}

func (h *Hub) Broadcast(snapshot model.TelemetrySnapshot) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.clients {
		select {
		case ch <- snapshot:
		default:
			// Non-blocking: skip if client buffer is full
		}
	}
}
