package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"hookrelay/internal/dispatch"
)

// handleDeleteCallback 停用并移除一条回调注册。
func (s *Server) handleDeleteCallback(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.config.registry.Unregister(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"removed": id})
}

// handleReplayDeadLetter 把一条死信重新投递给原回调，成功后将死信
// 标记为已解决。
func (s *Server) handleReplayDeadLetter(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	letter, ok := s.config.deadletter.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "dead letter not found")
		return
	}
	if err := s.config.deadletter.MarkReplaying(id); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	event, ok := s.config.queue.GetByID(letter.EventID)
	if !ok {
		writeError(w, http.StatusNotFound, "original event is no longer in queue")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	if err := s.config.dispatcher.DeliverTo(ctx, event, letter.CallbackID); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if err := s.config.deadletter.Resolve(id); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"replayed": id})
}

// signatureVerifyRequest 是签名校验接口的请求体。
type signatureVerifyRequest struct {
	Secret    string `json:"secret"`
	Timestamp string `json:"timestamp"`
	Signature string `json:"signature"`
	Body      []byte `json:"body"`
}

// handleVerifySignature 供运维校验下游回调签名是否正确。
func (s *Server) handleVerifySignature(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "read body failed")
		return
	}
	var request signatureVerifyRequest
	if err := json.Unmarshal(body, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := dispatch.VerifySignature(request.Secret, request.Body, request.Timestamp, request.Signature); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"valid": true})
}
