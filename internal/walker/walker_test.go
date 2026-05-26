package walker_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yuki/media-archive-converter/internal/walker"
)

func TestWalkCollectsTargetFiles(t *testing.T) {
	dir := t.TempDir()
	must(t, os.MkdirAll(filepath.Join(dir, "sub"), 0755))
	touch(t, filepath.Join(dir, "img.jpg"))
	touch(t, filepath.Join(dir, "img.png"))
	touch(t, filepath.Join(dir, "sub", "video.mp4"))
	touch(t, filepath.Join(dir, "already.avif"))
	touch(t, filepath.Join(dir, "encoded.mkv"))
	touch(t, filepath.Join(dir, "ignore.txt"))

	files, err := walker.Walk(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 5 {
		t.Fatalf("expected 5 files, got %d", len(files))
	}
}

func TestWalkRelativePaths(t *testing.T) {
	dir := t.TempDir()
	must(t, os.MkdirAll(filepath.Join(dir, "2024", "01"), 0755))
	touch(t, filepath.Join(dir, "2024", "01", "shot.png"))

	files, err := walker.Walk(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("2024", "01", "shot.png")
	if files[0].RelPath != want {
		t.Errorf("RelPath: got %q, want %q", files[0].RelPath, want)
	}
}

func TestWalkKindClassification(t *testing.T) {
	dir := t.TempDir()
	touch(t, filepath.Join(dir, "a.jpg"))
	touch(t, filepath.Join(dir, "b.JPEG"))
	touch(t, filepath.Join(dir, "c.png"))
	touch(t, filepath.Join(dir, "d.mp4"))
	touch(t, filepath.Join(dir, "e.avif"))
	touch(t, filepath.Join(dir, "f.mkv"))

	files, err := walker.Walk(dir)
	if err != nil {
		t.Fatal(err)
	}
	kinds := make(map[string]walker.Kind)
	for _, f := range files {
		kinds[filepath.Base(f.AbsPath)] = f.Kind
	}
	cases := map[string]walker.Kind{
		"a.jpg":  walker.KindJPEG,
		"b.JPEG": walker.KindJPEG,
		"c.png":  walker.KindPNG,
		"d.mp4":  walker.KindMP4,
		"e.avif": walker.KindAVIF,
		"f.mkv":  walker.KindMKV,
	}
	for name, want := range cases {
		if got := kinds[name]; got != want {
			t.Errorf("%s: want kind %d, got %d", name, want, got)
		}
	}
}

func touch(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
