package main

import (
	"io"
	"net/http"
	"os"
)

// monitorPath 是监控页面相对工作目录的路径。
const monitorPath = "web/monitor.html"

// handleMonitor 提供投递监控页面。页面通过 fetch 轮询 /api/v1/stats
// 渲染队列深度、位点、死信与重试状态。
func (s *Server) handleMonitor(w http.ResponseWriter, _ *http.Request) {
	file, err := os.Open(monitorPath)
	if err != nil {
		writeError(w, http.StatusNotFound, "monitor page not found")
		return
	}
	defer file.Close()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.Copy(w, file)
}
