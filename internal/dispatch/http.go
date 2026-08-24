package dispatch

import (
	"context"
	"io"
	"net/http"
	"time"
)

// CallbackClient 封装回调 HTTP 调用。Send 负责完整读取响应体并
// 关闭响应连接，避免每次回调都泄漏一个连接导致连接池耗尽。
type CallbackClient struct {
	httpClient *http.Client
	transport  *http.Transport
}

// NewCallbackClient 创建带连接池的回调客户端。
func NewCallbackClient(timeout time.Duration) *CallbackClient {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 16,
		IdleConnTimeout:     30 * time.Second,
	}
	return &CallbackClient{
		httpClient: &http.Client{Transport: transport},
		transport:  transport,
	}
}

// NewCallbackClientWithTransport 创建带自定义传输层的回调客户端，
// 供测试注入记录型 transport 或特殊网络策略使用。
func NewCallbackClientWithTransport(transport http.RoundTripper) *CallbackClient {
	httpClient := &http.Client{Transport: transport}
	return &CallbackClient{
		httpClient: httpClient,
		transport:  &http.Transport{},
	}
}

// Send 执行一次回调请求，返回响应体与状态码。无论成功失败，
// 响应体都会被完整读取并关闭，确保底层连接归还连接池而不泄漏。
func (c *CallbackClient) Send(ctx context.Context, req *http.Request) ([]byte, int, error) {
	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

// CloseIdle 关闭空闲连接，供服务优雅关闭时调用。
func (c *CallbackClient) CloseIdle() {
	c.transport.CloseIdleConnections()
}
