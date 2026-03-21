package main

import (
	"fmt"
	"time"
)

func generateInfographicAction(notebookURL string) error {
	client := NewClient(1)
	mapping := NewMappingStore("")
	service := NewService(client, notebookURL, mapping, nil)

	return service.GenerateInfographic()
}

func statusInfographicAction(notebookURL string) error {
	client := NewClient(1)
	mapping := NewMappingStore("")
	service := NewService(client, notebookURL, mapping, nil)

	status, err := service.StatusInfographic()
	if err != nil {
		return err
	}

	fmt.Println(status)
	return nil
}

func downloadInfographicAction(xdg *XDGPaths, notebookURL, outputFlag string) error {
	cfg, err := LoadConfig(xdg.ConfigFile())
	if err != nil {
		return fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
	}

	outputDir := cfg.ResolveDownloadDir("infographic", outputFlag)
	if outputDir == "" {
		return fmt.Errorf("出力先が指定されていません。--outputフラグまたはconfigのdownloads.infographicを設定してください")
	}

	mapping := NewMappingStore(xdg.MappingFile())
	entry, hash, found := mapping.LookupByURL(notebookURL)

	client := NewClient(1)
	service := NewService(client, notebookURL, mapping, nil)

	if err := service.DownloadInfographic(); err != nil {
		return err
	}

	var e *MappingEntry
	var h string
	if found {
		e = entry
		h = hash
	}

	dest, err := downloadAndMove(mapping, h, e, outputDir, "infographic", FindDownloadedInfographic, 60*time.Second)
	if err != nil {
		return err
	}

	fmt.Println(dest)
	return nil
}
