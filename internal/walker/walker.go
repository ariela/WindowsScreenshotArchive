package walker

import (
	"io/fs"
	"path/filepath"
	"strings"
)

type Kind int

const (
	KindJPEG Kind = iota
	KindPNG
	KindMP4
	KindAVIF
	KindMKV
)

type FileInfo struct {
	AbsPath string
	RelPath string
	Kind    Kind
}

func Walk(root string) ([]FileInfo, error) {
	var files []FileInfo
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		k, ok := classify(filepath.Ext(path))
		if !ok {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, FileInfo{AbsPath: path, RelPath: rel, Kind: k})
		return nil
	})
	return files, err
}

func classify(ext string) (Kind, bool) {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return KindJPEG, true
	case ".png":
		return KindPNG, true
	case ".mp4":
		return KindMP4, true
	case ".avif":
		return KindAVIF, true
	case ".mkv":
		return KindMKV, true
	default:
		return 0, false
	}
}
