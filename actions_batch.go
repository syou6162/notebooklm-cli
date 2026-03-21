package main

import (
	"fmt"
	"io"
	"time"
)

func batchInfographicAction(xdg *XDGPaths, reader io.Reader, outputFlag string) error {
	cfg, err := LoadConfig(xdg.ConfigFile())
	if err != nil {
		return fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
	}
	outputDir := cfg.ResolveDownloadDir("infographic", outputFlag)
	if outputDir == "" {
		return fmt.Errorf("出力先が指定されていません。--outputフラグまたはconfigのdownloads.infographicを設定してください")
	}

	if err := xdg.EnsureDirectories(); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
	}
	text, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("stdinの読み取りに失敗しました: %w", err)
	}

	mapping := NewMappingStore(xdg.MappingFile())
	client := NewClient(1)
	service := NewService(client, "", mapping, NewClaudeMetadataGenerator())

	fmt.Println("[1/4] ソースを追加中...")
	if err := service.AddSource(string(text)); err != nil {
		return err
	}

	hash := ComputeSHA256(string(text))
	entry, found := mapping.LookupEntry(hash)
	if !found || entry.URL == "" {
		return fmt.Errorf("マッピングからノートブックURLを取得できませんでした")
	}
	service.notebookURL = entry.URL

	fmt.Println("[2/4] インフォグラフィックを生成中...")
	if err := service.GenerateInfographic(); err != nil {
		return err
	}

	fmt.Print("[3/4] 生成完了を待機中...")
	for {
		status, err := service.StatusInfographic()
		if err != nil {
			return err
		}
		if status == "done" {
			fmt.Println(" 完了")
			break
		}
		fmt.Print(".")
		time.Sleep(5 * time.Second)
	}

	fmt.Println("[4/4] ダウンロード中...")
	if err := service.DownloadInfographic(); err != nil {
		return err
	}

	dest, err := downloadAndMove(mapping, hash, entry, outputDir, "infographic", FindDownloadedInfographic, 60*time.Second)
	if err != nil {
		return err
	}

	fmt.Println(dest)
	return nil
}

func batchAudioAction(xdg *XDGPaths, reader io.Reader, outputFlag string) error {
	cfg, err := LoadConfig(xdg.ConfigFile())
	if err != nil {
		return fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
	}
	outputDir := cfg.ResolveDownloadDir("audio", outputFlag)
	if outputDir == "" {
		return fmt.Errorf("出力先が指定されていません。--outputフラグまたはconfigのdownloads.audioを設定してください")
	}

	if err := xdg.EnsureDirectories(); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
	}
	text, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("stdinの読み取りに失敗しました: %w", err)
	}

	mapping := NewMappingStore(xdg.MappingFile())
	client := NewClient(1)
	service := NewService(client, "", mapping, NewClaudeMetadataGenerator())

	fmt.Println("[1/4] ソースを追加中...")
	if err := service.AddSource(string(text)); err != nil {
		return err
	}

	hash := ComputeSHA256(string(text))
	entry, found := mapping.LookupEntry(hash)
	if !found || entry.URL == "" {
		return fmt.Errorf("マッピングからノートブックURLを取得できませんでした")
	}
	service.notebookURL = entry.URL

	fmt.Println("[2/4] 音声解説を生成中...")
	if err := service.GenerateAudio(); err != nil {
		return err
	}

	fmt.Print("[3/4] 生成完了を待機中...")
	for {
		status, err := service.StatusAudio()
		if err != nil {
			return err
		}
		if status == "done" {
			fmt.Println(" 完了")
			break
		}
		fmt.Print(".")
		time.Sleep(10 * time.Second)
	}

	fmt.Println("[4/4] ダウンロード中...")
	if err := service.DownloadAudio(); err != nil {
		return err
	}

	dest, err := downloadAndMove(mapping, hash, entry, outputDir, "audio", FindDownloadedAudio, 120*time.Second)
	if err != nil {
		return err
	}

	fmt.Println(dest)
	return nil
}
