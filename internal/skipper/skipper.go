package skipper

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// HasSameStem は dir 内に stem と同一のベース名（拡張子除く）を持つ
// 非ディレクトリファイルが存在するか返す。dir が存在しない場合は false を返す。
func HasSameStem(dir, stem string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		s := strings.TrimSuffix(name, filepath.Ext(name))
		if s == stem {
			return true, nil
		}
	}
	return false, nil
}
