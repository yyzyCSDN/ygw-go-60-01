package route

import (
	"fmt"
	"time"

	"hookrelay/internal/model"
)

// Loader 负责把默认回调注册灌入注册表，让投递中心开箱即可演示
// 事件路由到不同下游端点的完整链路。
type Loader struct {
	registry *Registry
	now      func() time.Time
}

// NewLoader 创建默认回调加载器。
func NewLoader(registry *Registry) *Loader {
	return &Loader{registry: registry, now: time.Now}
}

// LoadDefaults 注册三条示例回调：订单创建、支付成功与事件审计。
// 返回实际加载的数量，供启动日志展示。
func (l *Loader) LoadDefaults() (int, error) {
	defaults := []*model.Callback{
		model.NewCallback("cb-order-created", "order.created", "http://127.0.0.1:9091/hook/order", "order-secret"),
		model.NewCallback("cb-payment-succeeded", "payment.succeeded", "http://127.0.0.1:9092/hook/payment", "payment-secret"),
		model.NewCallback("cb-audit-trail", "audit.trail", "http://127.0.0.1:9093/hook/audit", "audit-secret"),
	}
	loaded := 0
	for _, cb := range defaults {
		cb.CreatedAt = l.now().UTC()
		if err := l.registry.Register(cb); err != nil {
			return loaded, fmt.Errorf("load default callback %s: %w", cb.ID, err)
		}
		loaded++
	}
	return loaded, nil
}
