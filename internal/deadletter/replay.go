package deadletter

import (
	"errors"
	"time"

	"hookrelay/internal/model"
)

// MarkReplaying 把死信状态从 pending 推进到 replaying，进入重放流程。
func (s *Store) MarkReplaying(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	letter, ok := s.byID[id]
	if !ok {
		return errors.New("dead letter not found")
	}
	if letter.Status != model.DeadLetterPending {
		return errors.New("only pending dead letters can be replayed")
	}
	letter.Status = model.DeadLetterReplaying
	return nil
}

// Resolve 把死信标记为已解决。重放成功后调用，记录解决时间。
func (s *Store) Resolve(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	letter, ok := s.byID[id]
	if !ok {
		return errors.New("dead letter not found")
	}
	if letter.Status != model.DeadLetterReplaying {
		return errors.New("only replaying dead letters can be resolved")
	}
	now := time.Now().UTC()
	letter.Status = model.DeadLetterResolved
	letter.ResolvedAt = &now
	return nil
}
