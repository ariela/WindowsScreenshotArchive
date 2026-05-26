package fileops_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yuki/media-archive-converter/internal/fileops"
)

func TestCopyFileContent(t *testing.T) {
	src := filepath.Join(t.TempDir(), "source.avif")
	dst := filepath.Join(t.TempDir(), "dest.avif")
	if err := os.WriteFile(src, []byte("avif content"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := fileops.CopyFile(src, dst); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(dst)
	if string(data) != "avif content" {
		t.Errorf("unexpected content: %q", data)
	}
}

func TestCopyFileCreatesIntermediateDirs(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src.avif")
	dst := filepath.Join(t.TempDir(), "a", "b", "c", "dst.avif")
	if err := os.WriteFile(src, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := fileops.CopyFile(src, dst); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dst); err != nil {
		t.Errorf("dst not created: %v", err)
	}
}

func TestCopyFileMissingSrc(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "out.avif")
	if err := fileops.CopyFile("/nonexistent/src.avif", dst); err == nil {
		t.Error("expected error for missing source")
	}
}
