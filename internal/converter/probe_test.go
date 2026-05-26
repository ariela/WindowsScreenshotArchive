package converter_test

import (
	"context"
	"testing"

	"github.com/yuki/media-archive-converter/internal/converter"
	"github.com/yuki/media-archive-converter/internal/runner"
)

func TestProberDetectsAV1(t *testing.T) {
	fake := &runner.FakeRunner{Stdout: "codec_name=av1\n"}
	p := converter.NewProber(fake)

	is, err := p.IsAV1(context.Background(), "video.mkv")
	if err != nil {
		t.Fatal(err)
	}
	if !is {
		t.Error("expected true for av1")
	}
}

func TestProberDetectsNonAV1(t *testing.T) {
	fake := &runner.FakeRunner{Stdout: "codec_name=h264\n"}
	p := converter.NewProber(fake)

	is, err := p.IsAV1(context.Background(), "video.mkv")
	if err != nil {
		t.Fatal(err)
	}
	if is {
		t.Error("expected false for h264")
	}
}

func TestProberPassesInputPath(t *testing.T) {
	fake := &runner.FakeRunner{Stdout: "codec_name=av1\n"}
	p := converter.NewProber(fake)
	_, _ = p.IsAV1(context.Background(), "myvideo.mkv")

	found := false
	for _, a := range fake.Calls[0] {
		if a == "myvideo.mkv" {
			found = true
		}
	}
	if !found {
		t.Errorf("input path not in args: %v", fake.Calls[0])
	}
}
