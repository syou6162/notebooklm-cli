package main

import (
	"fmt"
	"io"
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

// CopyToOutput はファイルを出力ディレクトリにコピーする
func CopyToOutput(src, outputDir string) (string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", err
	}

	dest := filepath.Join(outputDir, filepath.Base(src))

	srcFile, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer func() { _ = srcFile.Close() }()

	destFile, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer func() { _ = destFile.Close() }()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return "", err
	}

	return dest, nil
}
