package offset

import (
	"encoding/json"
	"os"

	"hookrelay/internal/model"
)

// SaveSnapshots 把位点快照原子写入 JSON 文件，供服务关闭时落盘。
func SaveSnapshots(path string, snapshots []model.OffsetSnapshot) error {
	payload, err := json.MarshalIndent(snapshots, "", "  ")
	if err != nil {
		return err
	}
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, payload, 0o644); err != nil {
		return err
	}
	return os.Rename(temporary, path)
}

// LoadSnapshots 从 JSON 文件读取位点快照。文件不存在时返回空列表，
// 让服务以全新状态启动。
func LoadSnapshots(path string) ([]model.OffsetSnapshot, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var snapshots []model.OffsetSnapshot
	if err := json.Unmarshal(payload, &snapshots); err != nil {
		return nil, err
	}
	return snapshots, nil
}
