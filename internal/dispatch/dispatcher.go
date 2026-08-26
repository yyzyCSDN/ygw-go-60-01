package dispatch

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"hookrelay/internal/clock"
	"hookrelay/internal/deadletter"
	"hookrelay/internal/dedup"
	"hookrelay/internal/model"
	"hookrelay/internal/offset"
	"hookrelay/internal/queue"
	"hookrelay/internal/retry"
	"hookrelay/internal/route"
)

// Config 聚合投递器依赖与策略参数。
type Config struct {
	Registry    *route.Registry
	Queue       *queue.Queue
	Offsets     *offset.Store
	Dedup       *dedup.Store
	Retry       *retry.Scheduler
	DeadLetter  *deadletter.Store
	Clock       clock.Clock
	Client      *CallbackClient
	Timeout     time.Duration
	Secret      string
	DedupTTL    time.Duration
	Logger      *slog.Logger
	Metrics     *Metrics
	Health      *route.HealthTracker
}

// Dispatcher 协调一次事件投递的完整生命周期：路由、去重、签名、
// 调用回调、确认位点、失败重试与死信。
type Dispatcher struct {
	registry   *route.Registry
	queue      *queue.Queue
	offsets    *offset.Store
	dedup      *dedup.Store
	retry      *retry.Scheduler
	deadletter *deadletter.Store
	clock      clock.Clock
	client     *CallbackClient
	timeout    time.Duration
	secret     string
	dedupTTL   time.Duration
	logger     *slog.Logger
	signer     *Signer
	retryDue   map[string]retryEntry
	attempts   map[string]int
	tasks      map[string]*model.DeliveryTask
	metrics    *Metrics
	health     *route.HealthTracker
}

// retryEntry 记录一条回调投递的退避时间，key 为 eventID/callbackID。
type retryEntry struct {
	callbackID string
	nextAt     time.Time
}

// NewDispatcher 组装投递器。
func NewDispatcher(cfg Config) (*Dispatcher, error) {
	if cfg.Registry == nil || cfg.Queue == nil || cfg.Offsets == nil ||
		cfg.Dedup == nil || cfg.Retry == nil || cfg.DeadLetter == nil ||
		cfg.Clock == nil || cfg.Client == nil {
		return nil, errors.New("dispatcher dependencies are incomplete")
	}
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}
	metrics := cfg.Metrics
	if metrics == nil {
		metrics = NewMetrics()
	}
	health := cfg.Health
	if health == nil {
		health = route.NewHealthTracker()
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ttl := cfg.DedupTTL
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &Dispatcher{
		registry:   cfg.Registry,
		queue:      cfg.Queue,
		offsets:    cfg.Offsets,
		dedup:      cfg.Dedup,
		retry:      cfg.Retry,
		deadletter: cfg.DeadLetter,
		clock:      cfg.Clock,
		client:     cfg.Client,
		timeout:    timeout,
		secret:     cfg.Secret,
		dedupTTL:   ttl,
		logger:     logger,
		signer:     NewSigner(cfg.Clock, cfg.Secret),
		retryDue:   make(map[string]retryEntry),
		attempts:   make(map[string]int),
		tasks:      make(map[string]*model.DeliveryTask),
		metrics:    metrics,
		health:     health,
	}, nil
}

// deliverOne 对单个回调执行一次完整投递：去重、签名、调用、确认、
// 失败重试或死信。
func (d *Dispatcher) deliverOne(ctx context.Context, event *model.Event, cb *model.Callback) error {
	key := dedup.DedupKey(cb.ID, event.Seq)
	if d.dedup.Check(key) {
		d.logger.Debug("callback delivery skipped by dedup window", "callback", cb.ID, "event", event.ID)
		d.metrics.RecordSkipped()
		return nil
	}
	d.dedup.Mark(cb.ID, event.Seq, d.dedupTTL)
	task := d.taskFor(event.ID, cb.ID)
	task.MarkDelivering(d.clock.Now())
	attemptCtx, cancel := d.retry.Deadline(ctx, d.timeout)
	defer cancel()
	request, err := d.buildRequest(attemptCtx, event, cb)
	if err != nil {
		return err
	}
	started := d.clock.Now()
	body, status, err := d.client.Send(attemptCtx, request)
	result := &Result{StatusCode: status, Body: body, Err: err, Duration: d.clock.Now().Sub(started)}
	if result.Success() {
		if err := d.offsets.Acknowledge(cb.ID, event.Seq); err != nil {
			return err
		}
		d.clearRetry(event.ID, cb.ID)
		task.MarkDelivered(d.clock.Now())
		d.metrics.RecordDelivered()
		d.health.Record(cb.ID, true, "delivered", d.clock.Now().Unix())
		d.logger.Info("callback delivered", "callback", cb.ID, "event", event.ID, "seq", event.Seq)
		return nil
	}
	d.dedup.Remove(key)
	d.metrics.RecordFailed()
	d.health.Record(cb.ID, false, result.Error(), d.clock.Now().Unix())
	return d.handleFailure(event, cb, result)
}

// DeliverTo 把事件投递给指定回调，用于死信重放等定向场景。
func (d *Dispatcher) DeliverTo(ctx context.Context, event *model.Event, callbackID string) error {
	cb, ok := d.registry.Get(callbackID)
	if !ok {
		return errors.New("callback not found")
	}
	return d.deliverOne(ctx, event, cb)
}

// handleFailure 按退避策略调度重试，耗尽后写入死信。
func (d *Dispatcher) handleFailure(event *model.Event, cb *model.Callback, result *Result) error {
	reason := result.Error()
	if snippet := result.BodyText(); snippet != "" {
		reason += ": " + snippet
	}
	if result.Err != nil && errors.Is(result.Err, context.DeadlineExceeded) {
		reason = "callback timeout: " + reason
	}
	key := taskKey(event.ID, cb.ID)
	attempt := d.attempts[key] + 1
	d.attempts[key] = attempt
	task := d.taskFor(event.ID, cb.ID)
	if d.retry.ShouldRetry(attempt) {
		next := d.retry.NextAttemptAt(attempt, d.clock.Now())
		d.retryDue[key] = retryEntry{callbackID: cb.ID, nextAt: next}
		task.MarkRetrying(next, reason, d.clock.Now())
		d.metrics.RecordRetried()
		d.logger.Warn("callback retry scheduled", "callback", cb.ID, "event", event.ID, "attempt", attempt, "next", next)
		return nil
	}
	letter := model.NewDeadLetter(event.ID, cb.ID, reason, event.Body())
	if err := d.deadletter.Write(letter); err != nil {
		return err
	}
	delete(d.retryDue, key)
	delete(d.attempts, key)
	task.MarkDead(reason, d.clock.Now())
	d.metrics.RecordDead()
	d.logger.Error("callback moved to dead letter", "callback", cb.ID, "event", event.ID, "reason", reason)
	return nil
}

// DeliverDue 重投所有已到退避时间的回调，返回成功数量。
func (d *Dispatcher) DeliverDue(ctx context.Context, now time.Time) (int, error) {
	delivered := 0
	for key, entry := range d.retryDue {
		if entry.nextAt.After(now) {
			continue
		}
		eventID, callbackID := splitTaskKey(key)
		event, ok := d.queue.GetByID(eventID)
		if !ok {
			delete(d.retryDue, key)
			continue
		}
		cb, ok := d.registry.Get(callbackID)
		if !ok {
			delete(d.retryDue, key)
			continue
		}
		if err := d.deliverOne(ctx, event, cb); err != nil {
			return delivered, err
		}
		delivered++
	}
	return delivered, nil
}

// PendingRetries 返回当前等待退避重试的数量。
func (d *Dispatcher) PendingRetries() int {
	return len(d.retryDue)
}

// Sweep 对每条启用回调按位点切出批量投递窗口并投递窗口内匹配事件。
// 返回成功投递次数。批量窗口边界由队列与位点共同保证，窗口重叠
// 会导致同一事件重复投递。
func (d *Dispatcher) Sweep(ctx context.Context, batchSize int, now time.Time) (int, error) {
	delivered := 0
	for _, cb := range d.registry.List() {
		if !cb.Enabled {
			continue
		}
		slot := d.queue.NextSlot(d.offsets.Current(cb.ID), batchSize)
		if slot == nil {
			continue
		}
		for _, event := range slot.Events {
			if !cb.Matches(event.Type) {
				continue
			}
			if err := d.deliverOne(ctx, event, cb); err != nil {
				return delivered, err
			}
			delivered++
		}
	}
	return delivered, nil
}

// TaskCounts 返回投递任务状态机的分布统计。
func (d *Dispatcher) TaskCounts() map[model.TaskStatus]int {
	counts := make(map[model.TaskStatus]int)
	for _, task := range d.tasks {
		counts[task.Status]++
	}
	return counts
}

// MetricsSnapshot 返回投递指标快照。
func (d *Dispatcher) MetricsSnapshot() map[string]uint64 {
	return d.metrics.Snapshot()
}

func (d *Dispatcher) taskFor(eventID, callbackID string) *model.DeliveryTask {
	key := taskKey(eventID, callbackID)
	task, ok := d.tasks[key]
	if !ok {
		task = model.NewTask(eventID, callbackID)
		d.tasks[key] = task
	}
	return task
}

func (d *Dispatcher) clearRetry(eventID, callbackID string) {
	key := taskKey(eventID, callbackID)
	delete(d.retryDue, key)
	delete(d.attempts, key)
}

func taskKey(eventID, callbackID string) string {
	return eventID + "/" + callbackID
}

func splitTaskKey(key string) (string, string) {
	sep := len(key) - 1
	for i := len(key) - 1; i >= 0; i-- {
		if key[i] == '/' {
			sep = i
			break
		}
	}
	return key[:sep], key[sep+1:]
}

// buildRequest 构造带签名头的回调请求。
func (d *Dispatcher) buildRequest(ctx context.Context, event *model.Event, cb *model.Callback) (*http.Request, error) {
	body := event.Body()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, cb.URL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	signature, stamp := d.signer.Sign(body)
	tsHeader, sigHeader := HeaderNames()
	contentType := "application/json"
	if len(body) > 0 && body[0] == '{' {
		contentType = "application/json; charset=utf-8"
	}
	request.Header.Set("Content-Type", contentType)
	request.Header.Set("X-Hook-Event", event.Type)
	request.Header.Set(tsHeader, FormatTimestamp(stamp))
	request.Header.Set(sigHeader, signature)
	request.ContentLength = int64(len(body))
	request.Close = true
	request.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	return request, nil
}
