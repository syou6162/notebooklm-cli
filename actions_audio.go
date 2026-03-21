package main

import (
	"fmt"
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

	client := NewClient(1)
	service := NewService(client, notebookURL, mapping, nil)

	if err := service.DownloadAudio(); err != nil {
		return err
	}

	var e *MappingEntry
	var h string
	if found {
		e = entry
		h = hash
	}

	dest, err := downloadAndMove(mapping, h, e, outputDir, "audio", FindDownloadedAudio, 120*time.Second)
	if err != nil {
		return err
	}

	fmt.Println(dest)
	return nil
}
