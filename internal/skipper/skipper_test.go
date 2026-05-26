package skipper_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yuki/media-archive-converter/internal/skipper"
)

func TestHasSameStemFindsExisting(t *testing.T) {
	dir := t.TempDir()
	f, _ := os.Create(filepath.Join(dir, "photo.avif"))
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	found, err := skipper.HasSameStem(dir, "photo")
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Error("expected true, got false")
	}
}

func TestHasSameStemReturnsFalseForMismatch(t *testing.T) {
	dir := t.TempDir()
	f, _ := os.Create(filepath.Join(dir, "other.avif"))
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	found, err := skipper.HasSameStem(dir, "photo")
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Error("expected false, got true")
	}
}

func TestHasSameStemNonexistentDir(t *testing.T) {
	found, err := skipper.HasSameStem(t.TempDir()+"/no-such-dir", "photo")
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Error("expected false for nonexistent dir")
	}
}

func TestHasSameStemMatchesDifferentExtension(t *testing.T) {
	dir := t.TempDir()
	f, _ := os.Create(filepath.Join(dir, "shot.png"))
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	found, err := skipper.HasSameStem(dir, "shot")
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Error("expected true when same stem with different ext")
	}
}

func TestHasSameStemIgnoresDirectories(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "photo"), 0755); err != nil {
		t.Fatal(err)
	}

	found, err := skipper.HasSameStem(dir, "photo")
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Error("should not match directories")
	}
}
