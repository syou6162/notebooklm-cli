package main

import (
	"fmt"
	"io"
)

func addSourceAction(xdg *XDGPaths, reader io.Reader, notebookURL string) error {
	if err := xdg.EnsureDirectories(); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
	}

	text, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("stdinの読み取りに失敗しました: %w", err)
	}

	mapping := NewMappingStore(xdg.MappingFile())
	client := NewClient(1)
	service := NewService(client, notebookURL, mapping, NewClaudeMetadataGenerator())

	return service.AddSource(string(text))
}

func listSourceAction(notebookURL string) error {
	client := NewClient(1)
	mapping := NewMappingStore("")
	service := NewService(client, notebookURL, mapping, nil)

	names, err := service.ListSources()
	if err != nil {
		return err
	}

	for _, name := range names {
		fmt.Println(name)
	}
	return nil
}

func deleteSourceAction(xdg *XDGPaths, notebookURL string) error {
	client := NewClient(1)
	mapping := NewMappingStore(xdg.MappingFile())
	service := NewService(client, notebookURL, mapping, nil)

	return service.DeleteSource()
}

func resolveSourceAction(xdg *XDGPaths, reader io.Reader) error {
	text, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("stdinの読み取りに失敗しました: %w", err)
	}

	mapping := NewMappingStore(xdg.MappingFile())
	url, err := ResolveSource(string(text), mapping)
	if err != nil {
		return err
	}

	fmt.Println(url)
	return nil
}
