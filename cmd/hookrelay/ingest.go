package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"hookrelay/internal/model"
)

// ingestRequest 是事件接入接口的请求体。
type ingestRequest struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func (s *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 4<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "read body failed")
		return
	}
	var request ingestRequest
	if err := json.Unmarshal(body, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if request.Type == "" {
		writeError(w, http.StatusBadRequest, "event type is required")
		return
	}
	event := model.NewEvent(request.ID, request.Type, request.Payload)
	seq, err := s.config.queue.Enqueue(event)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	s.config.logger.Info("event enqueued", "event", event.ID, "type", event.Type, "seq", seq)
	writeJSON(w, http.StatusAccepted, map[string]any{
		"accepted": true,
		"event_id": event.ID,
		"seq":      seq,
		"content_hash": fmt.Sprintf("%016x", event.ContentHash()),
	})
}

func (s *Server) handleStats(w http.ResponseWriter, _ *http.Request) {
	s.writeStats(w)
}

func (s *Server) handleCallbacks(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"callbacks": s.config.registry.List(),
	})
}

type registerCallbackRequest struct {
	ID          string `json:"id"`
	EventType   string `json:"event_type"`
	URL         string `json:"url"`
	Secret      string `json:"secret"`
	MaxAttempts int    `json:"max_attempts"`
}

func (s *Server) handleRegisterCallback(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "read body failed")
		return
	}
	var request registerCallbackRequest
	if err := json.Unmarshal(body, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	cb := model.NewCallback(request.ID, request.EventType, request.URL, request.Secret)
	if request.MaxAttempts > 0 {
		cb.MaxAttempts = request.MaxAttempts
	}
	if err := s.config.registry.Register(cb); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"registered": cb.ID})
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, map[string]any{"error": message})
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
