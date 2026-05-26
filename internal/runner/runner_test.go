package runner_test

import (
	"context"
	"errors"
	"testing"

	"github.com/yuki/media-archive-converter/internal/runner"
)

func TestFakeRunnerRecordsCalls(t *testing.T) {
	fake := &runner.FakeRunner{Stdout: "hello"}
	stdout, stderr, err := fake.Run(context.Background(), "-i", "input.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if stdout != "hello" {
		t.Errorf("stdout: got %q, want %q", stdout, "hello")
	}
	if stderr != "" {
		t.Errorf("unexpected stderr: %q", stderr)
	}
	if len(fake.Calls) != 1 || fake.Calls[0][0] != "-i" {
		t.Errorf("calls not recorded: %v", fake.Calls)
	}
}

func TestFakeRunnerReturnsError(t *testing.T) {
	fake := &runner.FakeRunner{Err: errors.New("exit 1"), Stderr: "bad input"}
	_, stderr, err := fake.Run(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if stderr != "bad input" {
		t.Errorf("stderr: got %q", stderr)
	}
}
