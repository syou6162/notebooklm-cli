package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// downloadAndMove はブラウザダウンロード後のファイル検出→リネーム→マッピング更新→Finderコメント設定を行う
func downloadAndMove(
	mapping *MappingStore,
	hash string,
	entry *MappingEntry,
	outputDir string,
	downloadType string,
	findFunc func(dir string, startTime time.Time, minSize int64) (string, error),
	timeout time.Duration,
) (string, error) {
	homeDir, _ := os.UserHomeDir()
	downloadsDir := filepath.Join(homeDir, "Downloads")
	startTime := time.Now().Add(-30 * time.Second)

	var downloaded string
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		f, err := findFunc(downloadsDir, startTime, 1_000_000)
		if err == nil {
			downloaded = f
			break
		}
		time.Sleep(3 * time.Second)
	}

	if downloaded == "" {
		return "", fmt.Errorf("ダウンロードがタイムアウトしました")
	}

	title := ""
	if entry != nil {
		title = entry.Title
	}

	dest, err := MoveToOutput(downloaded, outputDir, title)
	if err != nil {
		return "", fmt.Errorf("ファイルの移動に失敗しました: %w", err)
	}

	if hash != "" {
		if err := mapping.UpdateDownload(hash, downloadType, dest); err != nil {
			fmt.Printf("マッピングの更新に失敗しました: %v\n", err)
		}
	}

	if downloadType == "audio" && entry != nil && entry.Description != "" {
		if err := SetFinderComment(dest, entry.Description); err != nil {
			fmt.Printf("Finderコメントの設定に失敗しました: %v\n", err)
		}
	}

	return dest, nil
}
