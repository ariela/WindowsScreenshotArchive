package converter

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/yuki/media-archive-converter/internal/runner"
)

type VideoConverter struct {
	r      runner.Runner
	tmpDir string
}

func NewVideoConverter(r runner.Runner, tmpDir string) *VideoConverter {
	return &VideoConverter{r: r, tmpDir: tmpDir}
}

func (c *VideoConverter) Convert(ctx context.Context, src string) (string, error) {
	dst := filepath.Join(c.tmpDir, stem(src)+".mkv")
	_, stderr, err := c.r.Run(ctx,
		"-i", src,
		"-c:v", "av1_nvenc",
		"-preset", "p7",
		"-cq", "25",
		"-profile:v", "main",
		"-c:a", "copy",
		"-y", dst,
	)
	if err != nil {
		return "", fmt.Errorf("ffmpeg video: %w: %s", err, stderr)
	}
	return dst, nil
}
