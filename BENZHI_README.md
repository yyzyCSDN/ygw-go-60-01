# HookRelay Webhook 异步投递与重试中心

HookRelay 是通用的 Webhook 异步投递基础设施：事件按回调注册表路由到
目标端点，投递队列持久化事件，投递失败按退避策略重试，超过上限进入
死信队列；回调请求带 HMAC 签名与幂等去重，投递位点保证重启后不重复
投递已确认事件。

## 功能

- 事件接入与回调注册 HTTP API
- 批量投递窗口与投递位点管理
- 指数退避重试与死信队列
- HMAC 签名、幂等去重
- 浏览器投递监控页面

## 本地运行

```bash
go build -mod=vendor ./...
go test -mod=vendor ./...
go vet -mod=vendor ./...
go run ./cmd/hookrelay -addr :8787
```

启动后访问：

- 健康检查：http://127.0.0.1:8787/healthz
- 事件接入：POST http://127.0.0.1:8787/api/v1/events
- 监控页面：http://127.0.0.1:8787/monitor

## Docker 构建

```bash
bash build_benzhi_docker.sh hookrelay linux/amd64
bash build_benzhi_docker.sh hookrelay linux/arm64
```

容器内可执行 go build ./...、go test ./...、go vet ./...（依赖离线
vendor，GOPROXY=off）。
