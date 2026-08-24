package main

import (
	"net/http"
	"time"
)

// handleStats 返回投递中心完整运行状态：队列深度、指标、健康度、
// 位点、死信与重试数量。
func (s *Server) statsPayload() map[string]any {
	offsets := s.config.offsets.All()
	return map[string]any{
		"queue_depth":     s.config.queue.Len(),
		"callbacks":       s.config.registry.Len(),
		"offsets":         offsets,
		"offset_callbacks": s.config.offsets.CallbackCount(),
		"dedup_keys":      s.config.dedup.Len(),
		"dead_letters":    s.config.deadletter.Len(),
		"pending_retries": s.config.dispatcher.PendingRetries(),
		"metrics":         s.config.dispatcher.MetricsSnapshot(),
		"tasks":           s.config.dispatcher.TaskCounts(),
		"health":          s.config.health.List(),
		"now":             time.Now().UTC().Format(time.RFC3339),
	}
}

func (s *Server) writeStats(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, s.statsPayload())
}
