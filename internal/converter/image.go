package converter

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/yuki/media-archive-converter/internal/runner"
)

type ImageConverter struct {
	r      runner.Runner
	tmpDir string
}

func NewImageConverter(r runner.Runner, tmpDir string) *ImageConverter {
	return &ImageConverter{r: r, tmpDir: tmpDir}
}

func (c *ImageConverter) Convert(ctx context.Context, src string) (string, error) {
	dst := filepath.Join(c.tmpDir, stem(src)+".avif")
	_, stderr, err := c.r.Run(ctx, "-i", src, "-c:v", "libaavif", "-still-picture", "1", "-y", dst)
	if err != nil {
		return "", fmt.Errorf("ffmpeg image: %w: %s", err, stderr)
	}
	return dst, nil
}

func stem(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}
