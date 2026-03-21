package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func generateAudioAction(notebookURL string) error {
	client := NewClient(1)
	mapping := NewMappingStore("")
	service := NewService(client, notebookURL, mapping, nil)

	return service.GenerateAudio()
}

func statusAudioAction(notebookURL string) error {
	client := NewClient(1)
	mapping := NewMappingStore("")
	service := NewService(client, notebookURL, mapping, nil)

	status, err := service.StatusAudio()
	if err != nil {
		return err
	}

	fmt.Println(status)
	return nil
}

func downloadAudioAction(xdg *XDGPaths, notebookURL, outputFlag string) error {
	cfg, err := LoadConfig(xdg.ConfigFile())
	if err != nil {
		return fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
	}

	outputDir := cfg.ResolveDownloadDir("audio", outputFlag)
	if outputDir == "" {
		return fmt.Errorf("出力先が指定されていません。--outputフラグまたはconfigのdownloads.audioを設定してください")
	}

	mapping := NewMappingStore(xdg.MappingFile())

	entry, hash, found := mapping.LookupByURL(notebookURL)
	var title string
	if found {
		title = entry.Title
	}

	client := NewClient(1)
	service := NewService(client, notebookURL, mapping, nil)

	if err := service.DownloadAudio(); err != nil {
		return err
	}

	homeDir, _ := os.UserHomeDir()
	downloadsDir := filepath.Join(homeDir, "Downloads")
	startTime := time.Now().Add(-30 * time.Second)

	var downloaded string
	deadline := time.Now().Add(120 * time.Second)
	for time.Now().Before(deadline) {
		f, err := FindDownloadedAudio(downloadsDir, startTime, 1_000_000)
		if err == nil {
			downloaded = f
			break
		}
		time.Sleep(3 * time.Second)
	}

	if downloaded == "" {
		return fmt.Errorf("ダウンロードがタイムアウトしました: %s/*.m4a", downloadsDir)
	}

	dest, err := MoveToOutput(downloaded, outputDir, title)
	if err != nil {
		return fmt.Errorf("ファイルの移動に失敗しました: %w", err)
	}

	if found {
		if err := mapping.UpdateDownload(hash, "audio", dest); err != nil {
			fmt.Printf("マッピングの更新に失敗しました: %v\n", err)
		}
		if entry.Description != "" {
			if err := SetFinderComment(dest, entry.Description); err != nil {
				fmt.Printf("Finderコメントの設定に失敗しました: %v\n", err)
			}
		}
	}

	fmt.Println(dest)
	return nil
}
