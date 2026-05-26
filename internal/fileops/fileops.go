package fileops

import (
	"io"
	"os"
	"path/filepath"
)

// CopyFile は src を dst へコピーする。dst の中間ディレクトリを自動作成する。
// 同一ボリューム内では tmpファイル + rename でアトミックに書き込む。
// クロスドライブの rename 失敗時は copy + delete にフォールバックする。
func CopyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	tmp, err := os.CreateTemp(filepath.Dir(dst), ".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	if _, err := io.Copy(tmp, in); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, dst); err != nil {
		// クロスドライブの rename 失敗: コピー + 削除にフォールバック
		if err2 := copyThenDelete(tmpPath, dst); err2 != nil {
			os.Remove(tmpPath)
			return err2
		}
	}
	return nil
}

func copyThenDelete(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(dst)
		return err
	}
	if err := out.Close(); err != nil {
		os.Remove(dst)
		return err
	}
	return os.Remove(src)
}
