package logger_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuki/media-archive-converter/internal/logger"
)

func TestLoggerWritesToPrimaryWriter(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, nil)
	l.Info("CONV", "foo.avif")

	if !strings.Contains(buf.String(), "CONV") {
		t.Errorf("expected CONV in output, got: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "foo.avif") {
		t.Errorf("expected foo.avif in output, got: %q", buf.String())
	}
}

func TestLoggerWritesToBothWriters(t *testing.T) {
	var stdout, file bytes.Buffer
	l := logger.New(&stdout, &file)
	l.Info("SKIP", "bar.avif")

	if !strings.Contains(stdout.String(), "SKIP") {
		t.Errorf("stdout missing SKIP")
	}
	if !strings.Contains(file.String(), "SKIP") {
		t.Errorf("file missing SKIP")
	}
}

func TestLoggerError(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, nil)
	l.Error("source.jpg", "ffmpeg exited 1")

	if !strings.Contains(buf.String(), "ERROR") {
		t.Errorf("expected ERROR level, got: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "source.jpg") {
		t.Errorf("expected source path in error line: %q", buf.String())
	}
}

func TestLoggerSummary(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, nil)
	l.Summary(10, 3, 1)

	out := buf.String()
	for _, want := range []string{"processed=10", "skipped=3", "errors=1"} {
		if !strings.Contains(out, want) {
			t.Errorf("summary missing %q: %q", want, out)
		}
	}
}
