package deadletter

import (
	"context"

	"hookrelay/internal/model"
)

// Sink 抽象死信的外部持久化目标。内存存储之外可替换为文件或
// 对象存储实现；Save 返回错误时写入必须向调用方传播。
type Sink interface {
	Save(ctx context.Context, letter *model.DeadLetter) error
}

// MemorySink 是内存模式的持久化实现：记录已经保存在 Store 内存中，
// Save 只负责确认持久化动作成功。
type MemorySink struct{}

// NewMemorySink 创建内存 sink。
func NewMemorySink() *MemorySink {
	return &MemorySink{}
}

// Save 确认内存模式持久化成功。
func (s *MemorySink) Save(ctx context.Context, letter *model.DeadLetter) error {
	return nil
}

// WriteWithSink 通过指定 sink 写入死信并原样传播错误。
// 调用方依赖返回错误来决定事件是否保留重试，因此这里绝不吞错。
func (s *Store) WriteWithSink(ctx context.Context, letter *model.DeadLetter, sink Sink) error {
	if sink != nil {
		if err := sink.Save(ctx, letter); err != nil {
			return err
		}
	}
	return s.writeRecord(letter)
}
