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

func TestVideoConverterPassesNVENCArgs(t *testing.T) {
	fake := &runner.FakeRunner{}
	c := converter.NewVideoConverter(fake, t.TempDir())

	_, err := c.Convert(context.Background(), "/src/clip.mp4")
	if err != nil {
		t.Fatal(err)
	}
	args := fake.Calls[0]
	for _, want := range []string{"av1_nvenc", "p7", "25", "main"} {
		found := false
		for _, a := range args {
			if a == want {
				found = true
			}
		}
		if !found {
			t.Errorf("missing %q in args: %v", want, args)
		}
	}
}

func TestVideoConverterOutputIsMKV(t *testing.T) {
	tmpDir := t.TempDir()
	fake := &runner.FakeRunner{}
	c := converter.NewVideoConverter(fake, tmpDir)

	out, err := c.Convert(context.Background(), "/src/clip.mp4")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Ext(out) != ".mkv" {
		t.Errorf("expected .mkv, got %q", filepath.Ext(out))
	}
	if filepath.Base(out) != "clip.mkv" {
		t.Errorf("expected clip.mkv, got %q", filepath.Base(out))
	}
}

func TestVideoConverterErrorIncludesStderr(t *testing.T) {
	fake := &runner.FakeRunner{Err: errors.New("exit 1"), Stderr: "NVENC not available"}
	c := converter.NewVideoConverter(fake, t.TempDir())

	_, err := c.Convert(context.Background(), "/src/clip.mp4")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "NVENC not available") {
		t.Errorf("error should include stderr: %v", err)
	}
}
