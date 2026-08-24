// HookRelay 是 Webhook 异步投递与重试中心的可执行入口。
// 启动后监听 HTTP 端口，提供事件接入、健康检查、监控页面与
// 后台投递 worker。
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hookrelay/internal/clock"
	"hookrelay/internal/deadletter"
	"hookrelay/internal/dedup"
	"hookrelay/internal/dispatch"
	"hookrelay/internal/offset"
	"hookrelay/internal/queue"
	"hookrelay/internal/retry"
	"hookrelay/internal/route"
)

func main() {
	addr := flag.String("addr", ":8787", "HTTP listen address")
	secret := flag.String("secret", "hookrelay-dev-secret", "HMAC signing secret")
	timeout := flag.Duration("timeout", 5*time.Second, "per-callback HTTP timeout")
	backoffBase := flag.Duration("backoff-base", time.Second, "retry backoff base interval")
	backoffMax := flag.Duration("backoff-max", 30*time.Second, "retry backoff max interval")
	maxAttempts := flag.Int("max-attempts", 3, "max delivery attempts before dead letter")
	workerInterval := flag.Duration("worker-interval", 500*time.Millisecond, "retry worker poll interval")
	dataFile := flag.String("data", "hookrelay-state.json", "state file for queue/offset/dead-letter persistence")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	registry := route.NewRegistry()
	loader := route.NewLoader(registry)
	loaded, err := loader.LoadDefaults()
	if err != nil {
		logger.Error("load default callbacks failed", "error", err)
		os.Exit(1)
	}
	logger.Info("default callbacks loaded", "count", loaded)

	eventQueue := queue.NewQueue()
	offsetStore := offset.NewStore()
	dedupStore := dedup.NewStore()
	deadStore := deadletter.NewStore()
	health := route.NewHealthTracker()
	source := clock.NewSource(clock.SystemClock{})
	policy := retry.NewPolicy(*backoffBase, *backoffMax, *maxAttempts)
	scheduler := retry.NewScheduler(policy, source)
	client := dispatch.NewCallbackClient(*timeout)

	dispatcher, err := dispatch.NewDispatcher(dispatch.Config{
		Registry:   registry,
		Queue:      eventQueue,
		Offsets:    offsetStore,
		Dedup:      dedupStore,
		Retry:      scheduler,
		DeadLetter: deadStore,
		Clock:      source,
		Client:     client,
		Timeout:    *timeout,
		Secret:     *secret,
		DedupTTL:   10 * time.Minute,
		Logger:     logger,
		Health:     health,
	})
	if err != nil {
		logger.Error("create dispatcher failed", "error", err)
		os.Exit(1)
	}

	server := NewServer(serverConfig{
		addr:       *addr,
		registry:   registry,
		queue:      eventQueue,
		offsets:    offsetStore,
		dedup:      dedupStore,
		deadletter: deadStore,
		dispatcher: dispatcher,
		health:     health,
		logger:     logger,
	})
	worker := NewWorker(dispatcher, deadStore, dedupStore, *workerInterval, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dataPath := *dataFile
	if err := loadState(dataPath, eventQueue, offsetStore, deadStore); err != nil {
		logger.Error("load state failed", "path", dataPath, "error", err)
		os.Exit(1)
	}
	var lowest uint64
	for _, seq := range offsetStore.All() {
		if lowest == 0 || seq < lowest {
			lowest = seq
		}
	}
	if remaining := len(eventQueue.ResumeFrom(lowest)); remaining > 0 {
		logger.Info("queue resume pending", "count", remaining, "from", lowest)
	}

	worker.Start()
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("hookrelay listening", "addr", *addr)
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server failed", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("http server shutdown failed", "error", err)
	}
	worker.Stop()
	if err := saveState(dataPath, eventQueue, offsetStore, deadStore); err != nil {
		logger.Error("save state failed", "path", dataPath, "error", err)
	} else {
		logger.Info("state saved", "path", dataPath)
	}
	client.CloseIdle()
	logger.Info("hookrelay stopped")
}
