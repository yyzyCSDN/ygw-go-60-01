// Package deadletter 保存重试耗尽仍无法投递的事件，支持写入、
// 查询、重放与解决。写入失败必须向调用方传播，否则事件会既未
// 投递成功也未进死信而直接丢失。
package deadletter

import (
	"context"
	"errors"
	"sort"
	"sync"

	"hookrelay/internal/model"
)

// Store 是死信存储。letters 保存全部死信记录，byID 提供索引。
type Store struct {
	mu      sync.Mutex
	letters []*model.DeadLetter
	byID    map[string]*model.DeadLetter
	sink    Sink
}

// NewStore 创建空死信存储。
func NewStore() *Store {
	return &Store{
		byID: make(map[string]*model.DeadLetter),
		sink: NewMemorySink(),
	}
}

// Write 追加一条死信记录。重复 ID 或空事件会返回错误。
func (s *Store) Write(letter *model.DeadLetter) error {
	return s.WriteWithSink(context.Background(), letter, s.sink)
}

// SetSink 替换死信的外部持久化 sink。
func (s *Store) SetSink(sink Sink) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sink = sink
}

func (s *Store) writeRecord(letter *model.DeadLetter) error {
	if letter == nil || letter.ID == "" {
		return errors.New("dead letter id is required")
	}
	if letter.EventID == "" || letter.CallbackID == "" {
		return errors.New("dead letter must reference an event and callback")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.byID[letter.ID]; exists {
		return errors.New("dead letter already exists")
	}
	s.letters = append(s.letters, letter)
	s.byID[letter.ID] = letter
	return nil
}

// Get 按 ID 查询死信记录。
func (s *Store) Get(id string) (*model.DeadLetter, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	letter, ok := s.byID[id]
	return letter, ok
}

// Len 返回死信记录总数。
func (s *Store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.letters)
}

// List 返回全部死信记录，按创建时间排序。
func (s *Store) List() []*model.DeadLetter {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]*model.DeadLetter, len(s.letters))
	copy(result, s.letters)
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result
}
