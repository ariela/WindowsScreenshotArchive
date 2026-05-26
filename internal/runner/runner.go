package runner

import (
	"bytes"
	"context"
	"os/exec"
)

// Runner は ffmpeg / ffprobe のサブプロセス実行を抽象化するインターフェース。
type Runner interface {
	Run(ctx context.Context, args ...string) (stdout, stderr string, err error)
}

// ExecRunner は実際のバイナリを呼び出す。
type ExecRunner struct {
	bin string
}

func NewExec(bin string) *ExecRunner {
	return &ExecRunner{bin: bin}
}

func (r *ExecRunner) Run(ctx context.Context, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, r.bin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// FakeRunner はテスト用のモック実装。Calls に呼び出し引数が記録される。
type FakeRunner struct {
	Stdout string
	Stderr string
	Err    error
	Calls  [][]string
}

func (f *FakeRunner) Run(_ context.Context, args ...string) (string, string, error) {
	cp := make([]string, len(args))
	copy(cp, args)
	f.Calls = append(f.Calls, cp)
	return f.Stdout, f.Stderr, f.Err
}
