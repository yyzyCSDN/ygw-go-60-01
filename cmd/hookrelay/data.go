package main

import (
	"encoding/json"
	"os"
	"time"

	"hookrelay/internal/deadletter"
	"hookrelay/internal/model"
	"hookrelay/internal/offset"
	"hookrelay/internal/queue"
)

// stateFile 描述服务状态的落盘结构：队列事件、位点快照与死信记录。
// 服务启动时加载，关闭时保存，实现重启后的状态恢复。
type stateFile struct {
	Events     []queue.EventSnapshot     `json:"events"`
	DeadLetter []deadLetterSnapshot      `json:"dead_letters"`
}

type deadLetterSnapshot struct {
	ID         string               `json:"id"`
	EventID    string               `json:"event_id"`
	CallbackID string               `json:"callback_id"`
	Reason     string               `json:"reason"`
	Payload    []byte               `json:"payload"`
	Status     model.DeadLetterStatus `json:"status"`
	CreatedAt  time.Time            `json:"created_at"`
}

// saveState 把队列、位点与死信状态原子写入数据文件。
func saveState(path string, eventQueue *queue.Queue, offsetStore *offset.Store, deadStore *deadletter.Store) error {
	payload := stateFile{
		Events: eventQueue.Snapshot(),
	}
	for _, letter := range deadStore.List() {
		payload.DeadLetter = append(payload.DeadLetter, deadLetterSnapshot{
			ID:         letter.ID,
			EventID:    letter.EventID,
			CallbackID: letter.CallbackID,
			Reason:     letter.Reason,
			Payload:    append([]byte(nil), letter.Payload...),
			Status:     letter.Status,
			CreatedAt:  letter.CreatedAt,
		})
	}
	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, encoded, 0o644); err != nil {
		return err
	}
	if err := os.Rename(temporary, path); err != nil {
		return err
	}
	return offset.SaveSnapshots(path+".offsets", offsetStore.Snapshot())
}

// loadState 从数据文件恢复队列、位点与死信记录。文件不存在时静默返回。
func loadState(path string, eventQueue *queue.Queue, offsetStore *offset.Store, deadStore *deadletter.Store) error {
	payload, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var file stateFile
	if err := json.Unmarshal(payload, &file); err != nil {
		return err
	}
	if err := eventQueue.Restore(file.Events); err != nil {
		return err
	}
	snapshots, err := offset.LoadSnapshots(path + ".offsets")
	if err != nil {
		return err
	}
	if err := offsetStore.Restore(snapshots); err != nil {
		return err
	}
	for _, snapshot := range file.DeadLetter {
		letter := &model.DeadLetter{
			ID:         snapshot.ID,
			EventID:    snapshot.EventID,
			CallbackID: snapshot.CallbackID,
			Reason:     snapshot.Reason,
			Payload:    append([]byte(nil), snapshot.Payload...),
			Status:     snapshot.Status,
			CreatedAt:  snapshot.CreatedAt,
		}
		if err := deadStore.Write(letter); err != nil {
			return err
		}
	}
	return nil
}
