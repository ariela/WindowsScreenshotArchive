package converter_test

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuki/media-archive-converter/internal/converter"
	"github.com/yuki/media-archive-converter/internal/runner"
)

func TestImageConverterPassesCorrectArgs(t *testing.T) {
	fake := &runner.FakeRunner{}
	c := converter.NewImageConverter(fake, t.TempDir())

	_, err := c.Convert(context.Background(), "/src/photo.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if len(fake.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(fake.Calls))
	}
	args := fake.Calls[0]
	if args[0] != "-i" || args[1] != "/src/photo.jpg" {
		t.Errorf("unexpected input args: %v", args)
	}
	found := false
	for _, a := range args {
		if a == "libaavif" {
			found = true
		}
	}
	if !found {
		t.Errorf("libaavif not in ffmpeg args: %v", args)
	}
}

func TestImageConverterOutputPath(t *testing.T) {
	tmpDir := t.TempDir()
	fake := &runner.FakeRunner{}
	c := converter.NewImageConverter(fake, tmpDir)

	out, err := c.Convert(context.Background(), "/src/photo.png")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(out) != "photo.avif" {
		t.Errorf("expected photo.avif, got %q", filepath.Base(out))
	}
	if !strings.HasPrefix(out, tmpDir) {
		t.Errorf("output not in tmpDir: %q", out)
	}
}

func TestImageConverterErrorIncludesStderr(t *testing.T) {
	fake := &runner.FakeRunner{Err: errors.New("exit 1"), Stderr: "libaavif not found"}
	c := converter.NewImageConverter(fake, t.TempDir())

	_, err := c.Convert(context.Background(), "/src/broken.jpg")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "libaavif not found") {
		t.Errorf("error should include stderr: %v", err)
	}
}
