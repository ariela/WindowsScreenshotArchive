package processor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/yuki/media-archive-converter/internal/converter"
	"github.com/yuki/media-archive-converter/internal/fileops"
	"github.com/yuki/media-archive-converter/internal/logger"
	"github.com/yuki/media-archive-converter/internal/skipper"
	"github.com/yuki/media-archive-converter/internal/walker"
)

type Config struct {
	Src     string
	Dst     string
	Workers int
	DryRun  bool
}

type Stats struct {
	Processed int
	Skipped   int
	Errors    int
}

type Processor struct {
	cfg     Config
	log     *logger.Logger
	imgConv *converter.ImageConverter
	vidConv *converter.VideoConverter
	prober  *converter.Prober
}

func New(cfg Config, log *logger.Logger,
	imgConv *converter.ImageConverter,
	vidConv *converter.VideoConverter,
	prober *converter.Prober,
) *Processor {
	return &Processor{cfg: cfg, log: log, imgConv: imgConv, vidConv: vidConv, prober: prober}
}

func (p *Processor) Run(ctx context.Context) (Stats, error) {
	files, err := walker.Walk(p.cfg.Src)
	if err != nil {
		return Stats{}, err
	}

	var stats Stats
	var images, videos, copies []walker.FileInfo

	for _, f := range files {
		dstDir := filepath.Join(p.cfg.Dst, filepath.Dir(f.RelPath))
		st := stem(f.AbsPath)
		skip, err := skipper.HasSameStem(dstDir, st)
		if err != nil {
			p.log.Error(f.AbsPath, err.Error())
			stats.Errors++
			continue
		}
		if skip {
			p.log.Info("SKIP", filepath.Join(dstDir, st)+" (already exists on dst)")
			stats.Skipped++
			continue
		}
		switch f.Kind {
		case walker.KindJPEG, walker.KindPNG:
			images = append(images, f)
		case walker.KindMP4:
			videos = append(videos, f)
		case walker.KindAVIF:
			copies = append(copies, f)
		case walker.KindMKV:
			isAV1, err := p.prober.IsAV1(ctx, f.AbsPath)
			if err != nil {
				p.log.Error(f.AbsPath, err.Error())
				stats.Errors++
				continue
			}
			if isAV1 {
				copies = append(copies, f)
			}
			// 非AV1 MKV はスコープ外: 無視
		}
	}

	// 画像を worker pool で並列変換
	type imgResult struct {
		f   walker.FileInfo
		err error
	}
	jobs := make(chan walker.FileInfo, len(images))
	results := make(chan imgResult, len(images))

	workers := p.cfg.Workers
	if workers < 1 {
		workers = 1
	}
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range jobs {
				results <- imgResult{f: f, err: p.handleImage(ctx, f)}
			}
		}()
	}
	for _, f := range images {
		jobs <- f
	}
	close(jobs)
	wg.Wait()
	close(results)

	for r := range results {
		if r.err != nil {
			p.log.Error(r.f.AbsPath, r.err.Error())
			stats.Errors++
		} else {
			stats.Processed++
		}
	}

	// AVIF 画像 / AV1 動画を直接コピー
	for _, f := range copies {
		if err := p.handleCopy(f); err != nil {
			p.log.Error(f.AbsPath, err.Error())
			stats.Errors++
		} else {
			stats.Processed++
		}
	}

	// 動画をシリアル変換
	for _, f := range videos {
		if err := p.handleVideo(ctx, f); err != nil {
			p.log.Error(f.AbsPath, err.Error())
			stats.Errors++
		} else {
			stats.Processed++
		}
	}

	p.log.Summary(stats.Processed, stats.Skipped, stats.Errors)
	return stats, nil
}

func (p *Processor) handleImage(ctx context.Context, f walker.FileInfo) error {
	dstPath := filepath.Join(p.cfg.Dst, filepath.Dir(f.RelPath), stem(f.AbsPath)+".avif")
	if p.cfg.DryRun {
		p.log.Info("DRYRUN", f.AbsPath+" → "+dstPath)
		return nil
	}
	tmpOut, err := p.imgConv.Convert(ctx, f.AbsPath)
	if err != nil {
		return err
	}
	defer os.Remove(tmpOut)
	if err := fileops.CopyFile(tmpOut, dstPath); err != nil {
		return err
	}
	if err := os.Remove(f.AbsPath); err != nil {
		return err
	}
	p.log.Info("CONV", f.AbsPath+" → "+dstPath)
	return nil
}

func (p *Processor) handleVideo(ctx context.Context, f walker.FileInfo) error {
	dstPath := filepath.Join(p.cfg.Dst, filepath.Dir(f.RelPath), stem(f.AbsPath)+".mkv")
	if p.cfg.DryRun {
		p.log.Info("DRYRUN", f.AbsPath+" → "+dstPath)
		return nil
	}
	tmpOut, err := p.vidConv.Convert(ctx, f.AbsPath)
	if err != nil {
		return err
	}
	defer os.Remove(tmpOut)
	if err := fileops.CopyFile(tmpOut, dstPath); err != nil {
		return err
	}
	if err := os.Remove(f.AbsPath); err != nil {
		return err
	}
	p.log.Info("CONV", f.AbsPath+" → "+dstPath)
	return nil
}

func (p *Processor) handleCopy(f walker.FileInfo) error {
	dstPath := filepath.Join(p.cfg.Dst, f.RelPath)
	if p.cfg.DryRun {
		p.log.Info("DRYRUN", f.AbsPath+" → "+dstPath)
		return nil
	}
	if err := fileops.CopyFile(f.AbsPath, dstPath); err != nil {
		return err
	}
	if err := os.Remove(f.AbsPath); err != nil {
		return err
	}
	p.log.Info("COPY", f.AbsPath+" → "+dstPath)
	return nil
}

func stem(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}
