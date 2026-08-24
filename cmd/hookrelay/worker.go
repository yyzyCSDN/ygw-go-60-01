package main

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"hookrelay/internal/deadletter"
	"hookrelay/internal/dedup"
	"hookrelay/internal/dispatch"
)

// Worker 周期扫描退避队列，把到期的重试任务重新投递。
type Worker struct {
	dispatcher *dispatch.Dispatcher
	deadStore  *deadletter.Store
	dedupStore *dedup.Store
	interval   time.Duration
	logger     *slog.Logger
	stopCh     chan struct{}
	ticks      int
	once       sync.Once
	wg         sync.WaitGroup
}

// NewWorker 创建重试 worker。
func NewWorker(dispatcher *dispatch.Dispatcher, deadStore *deadletter.Store, dedupStore *dedup.Store, interval time.Duration, logger *slog.Logger) *Worker {
	return &Worker{
		dispatcher: dispatcher,
		deadStore:  deadStore,
		dedupStore: dedupStore,
		interval:   interval,
		logger:     logger,
		stopCh:     make(chan struct{}),
	}
}

// Start 启动后台循环。
func (w *Worker) Start() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		for {
			select {
			case <-w.stopCh:
				return
			case now := <-ticker.C:
				w.ticks++
				if w.ticks%40 == 0 {
					purged := w.deadStore.PurgeResolved(100)
					if purged > 0 {
						w.logger.Info("resolved dead letters purged", "count", purged)
					}
					expired := w.dedupStore.CleanExpired(now)
					if expired > 0 {
						w.logger.Info("expired dedup keys cleaned", "count", expired)
					}
				}
				ctx, cancel := context.WithTimeout(context.Background(), w.interval)
				dueDelivered, dueErr := w.dispatcher.DeliverDue(ctx, now)
				swept, sweepErr := w.dispatcher.Sweep(ctx, 16, now)
				cancel()
				delivered := dueDelivered + swept
				err := dueErr
				if err == nil {
					err = sweepErr
				}
				if err != nil {
					w.logger.Warn("retry sweep failed", "error", err)
					continue
				}
				if delivered > 0 {
					w.logger.Info("retry sweep delivered", "count", delivered)
				}
			}
		}
	}()
}

// Stop 停止后台循环并等待退出。
func (w *Worker) Stop() {
	w.once.Do(func() {
		close(w.stopCh)
	})
	w.wg.Wait()
}
