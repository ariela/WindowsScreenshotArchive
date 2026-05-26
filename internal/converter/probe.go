package converter

import (
	"context"
	"fmt"
	"strings"

	"github.com/yuki/media-archive-converter/internal/runner"
)

type Prober struct {
	r runner.Runner
}

func NewProber(r runner.Runner) *Prober {
	return &Prober{r: r}
}

func (p *Prober) IsAV1(ctx context.Context, path string) (bool, error) {
	stdout, _, err := p.r.Run(ctx,
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=codec_name",
		"-of", "default=noprint_wrappers=1",
		path,
	)
	if err != nil {
		return false, fmt.Errorf("ffprobe: %w", err)
	}
	return strings.Contains(stdout, "codec_name=av1"), nil
}
