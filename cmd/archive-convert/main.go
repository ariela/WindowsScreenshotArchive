package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/yuki/media-archive-converter/internal/converter"
	"github.com/yuki/media-archive-converter/internal/logger"
	"github.com/yuki/media-archive-converter/internal/processor"
	"github.com/yuki/media-archive-converter/internal/runner"
)

func main() {
	var (
		src         = flag.String("src", "", "変換元ディレクトリ (必須)")
		dst         = flag.String("dst", "", "コピー先ディレクトリ (必須)")
		workers     = flag.Int("workers", runtime.NumCPU(), "画像変換の並列数")
		logFile     = flag.String("log", "", "ログファイルパス (省略時: stdout のみ)")
		dryRun      = flag.Bool("dry-run", false, "実際の変換/コピー/削除を行わずに表示のみ")
		ffmpegPath  = flag.String("ffmpeg-path", "ffmpeg", "ffmpeg バイナリパス")
		ffprobePath = flag.String("ffprobe-path", "ffprobe", "ffprobe バイナリパス")
	)
	flag.Parse()

	if *src == "" || *dst == "" {
		fmt.Fprintln(os.Stderr, "error: --src and --dst are required")
		flag.Usage()
		os.Exit(1)
	}

	var secondary *os.File
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: cannot open log file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		secondary = f
	}

	tmpDir, err := os.MkdirTemp("", "archive-convert-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	log := logger.New(os.Stdout, secondary)

	ffmpegRunner := runner.NewExec(*ffmpegPath)
	ffprobeRunner := runner.NewExec(*ffprobePath)

	p := processor.New(
		processor.Config{
			Src:     *src,
			Dst:     *dst,
			Workers: *workers,
			DryRun:  *dryRun,
		},
		log,
		converter.NewImageConverter(ffmpegRunner, tmpDir),
		converter.NewVideoConverter(ffmpegRunner, tmpDir),
		converter.NewProber(ffprobeRunner),
	)

	stats, err := p.Run(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if stats.Errors > 0 {
		os.Exit(2)
	}
}
