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

	// まずos.Renameを試みる（同一デバイスなら高速）
	if err := os.Rename(src, dest); err == nil {
		return dest, nil
	}

	// クロスデバイスの場合はコピー+削除
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

	_ = os.Remove(src)

	return dest, nil
}
