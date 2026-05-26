package logger

import (
	"fmt"
	"io"
	"sync"
	"time"
)

type Logger struct {
	mu        sync.Mutex
	primary   io.Writer
	secondary io.Writer
}

func New(primary, secondary io.Writer) *Logger {
	return &Logger{primary: primary, secondary: secondary}
}

func (l *Logger) Info(level, msg string) {
	l.write(level, msg)
}

func (l *Logger) Error(src, detail string) {
	l.write("ERROR", fmt.Sprintf("%s: %s", src, detail))
}

func (l *Logger) Summary(processed, skipped, errors int) {
	l.write("DONE", fmt.Sprintf("processed=%d skipped=%d errors=%d", processed, skipped, errors))
}

func (l *Logger) write(level, msg string) {
	line := fmt.Sprintf("[%s] %-6s %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		level,
		msg,
	)
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprint(l.primary, line)
	if l.secondary != nil {
		fmt.Fprint(l.secondary, line)
	}
}
