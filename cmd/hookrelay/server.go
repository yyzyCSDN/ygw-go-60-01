package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"hookrelay/internal/deadletter"
	"hookrelay/internal/dedup"
	"hookrelay/internal/dispatch"
	"hookrelay/internal/offset"
	"hookrelay/internal/queue"
	"hookrelay/internal/route"
)

type serverConfig struct {
	addr       string
	registry   *route.Registry
	queue      *queue.Queue
	offsets    *offset.Store
	dedup      *dedup.Store
	deadletter *deadletter.Store
	dispatcher *dispatch.Dispatcher
	health     *route.HealthTracker
	logger     *slog.Logger
}

// Server 暴露投递中心的 HTTP 接口。
type Server struct {
	config     serverConfig
	httpServer *http.Server
}

// NewServer 组装 HTTP 路由。
func NewServer(cfg serverConfig) *Server {
	mux := http.NewServeMux()
	server := &Server{config: cfg}
	mux.HandleFunc("GET /healthz", server.handleHealth)
	mux.HandleFunc("POST /api/v1/events", server.handleIngest)
	mux.HandleFunc("GET /api/v1/stats", server.handleStats)
	mux.HandleFunc("GET /api/v1/callbacks", server.handleCallbacks)
	mux.HandleFunc("POST /api/v1/callbacks", server.handleRegisterCallback)
	mux.HandleFunc("DELETE /api/v1/callbacks/{id}", server.handleDeleteCallback)
	mux.HandleFunc("POST /api/v1/deadletters/{id}/replay", server.handleReplayDeadLetter)
	mux.HandleFunc("POST /api/v1/signature/verify", server.handleVerifySignature)
	mux.HandleFunc("GET /monitor", server.handleMonitor)
	server.httpServer = &http.Server{
		Addr:              cfg.addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return server
}

// ListenAndServe 启动 HTTP 服务。
func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown 优雅关闭 HTTP 服务。
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
