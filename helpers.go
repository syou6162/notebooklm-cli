package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// FindDownloadedInfographic はダウンロードされたインフォグラフィックPNGを同定して返す
func FindDownloadedInfographic(downloadsDir string, startTime time.Time, minSize int64) (string, error) {
	matches, err := filepath.Glob(filepath.Join(downloadsDir, "unnamed*.png"))
	if err != nil {
		return "", err
	}

	type candidate struct {
		path  string
		mtime time.Time
	}

	var candidates []candidate
	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if info.ModTime().Before(startTime) {
			continue
		}
		if info.Size() < minSize {
			continue
		}
		candidates = append(candidates, candidate{path: path, mtime: info.ModTime()})
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("ダウンロード済みのインフォグラフィックが見つかりません: %s/unnamed*.png", downloadsDir)
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].mtime.Before(candidates[j].mtime)
	})

	return candidates[len(candidates)-1].path, nil
}

// MoveToOutput はファイルを出力ディレクトリにリネームして移動する
// nameが空の場合は元のファイル名を使う。拡張子は元ファイルから引き継ぐ。
func MoveToOutput(src, outputDir, name string) (string, error) {
	absDir, err := filepath.Abs(outputDir)
	if err != nil {
		return "", err
	}
	outputDir = absDir

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", err
	}

	ext := filepath.Ext(src)
	var destName string
	if name != "" {
		destName = name + ext
	} else {
		destName = filepath.Base(src)
	}
	dest := filepath.Join(outputDir, destName)

	if err := os.Rename(src, dest); err != nil {
		return "", err
	}

	return dest, nil
}
