package main

import (
	"fmt"
	"path/filepath"
	"testing"
)

type mockMetadataGenerator struct {
	title       string
	description string
	err         error
	called      bool
}

func (m *mockMetadataGenerator) Generate(text string) (*Metadata, error) {
	m.called = true
	if m.err != nil {
		return nil, m.err
	}
	return &Metadata{Title: m.title, Description: m.description}, nil
}

func TestGenerateMetadata_ReturnsResult(t *testing.T) {
	gen := &mockMetadataGenerator{
		title:       "テストタイトル",
		description: "テストの要約です",
	}

	meta, err := gen.Generate("入力テキスト")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if meta.Title != "テストタイトル" {
		t.Errorf("Title = %q, want %q", meta.Title, "テストタイトル")
	}
	if meta.Description != "テストの要約です" {
		t.Errorf("Description = %q, want %q", meta.Description, "テストの要約です")
	}
}

func TestClaudeMetadataGenerator_ParseResponse(t *testing.T) {
	// claude -p --output-format json の出力をパースするテスト
	jsonOutput := `[{"type":"system","subtype":"init"},{"type":"result","structured_output":{"title":"AIの最新動向","description":"大規模言語モデルの進化について..."}}]`

	meta, err := parseClaudeResponse([]byte(jsonOutput))
	if err != nil {
		t.Fatalf("parseClaudeResponse() error = %v", err)
	}
	if meta.Title != "AIの最新動向" {
		t.Errorf("Title = %q, want %q", meta.Title, "AIの最新動向")
	}
	if meta.Description != "大規模言語モデルの進化について..." {
		t.Errorf("Description = %q, want %q", meta.Description, "大規模言語モデルの進化について...")
	}
}

func TestClaudeMetadataGenerator_ParseResponse_NoResult(t *testing.T) {
	jsonOutput := `[{"type":"system","subtype":"init"}]`

	_, err := parseClaudeResponse([]byte(jsonOutput))
	if err == nil {
		t.Fatal("expected error when no result in response")
	}
}

func TestClaudeMetadataGenerator_ParseResponse_NoStructuredOutput(t *testing.T) {
	jsonOutput := `[{"type":"result","result":"text only"}]`

	_, err := parseClaudeResponse([]byte(jsonOutput))
	if err == nil {
		t.Fatal("expected error when no structured_output in result")
	}
}

func TestAddSource_WithMetadataGenerator(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))

	browser := &mockBrowser{
		findTabFound: false,
		currentURL:   "https://notebooklm.google.com/notebook/new-123",
	}
	gen := &mockMetadataGenerator{
		title:       "生成されたタイトル",
		description: "生成された要約",
	}

	service := NewService(browser, "", store, gen)
	service.sleep = noSleep
	service.metadataGen = gen

	err := service.AddSource("テスト入力テキスト")
	if err != nil {
		t.Fatalf("AddSource() error = %v", err)
	}

	hash := ComputeSHA256("テスト入力テキスト")
	entry, found := store.LookupEntry(hash)
	if !found {
		t.Fatal("expected mapping to be saved")
	}
	if entry.Title != "生成されたタイトル" {
		t.Errorf("Title = %q, want %q", entry.Title, "生成されたタイトル")
	}
	if entry.Description != "生成された要約" {
		t.Errorf("Description = %q, want %q", entry.Description, "生成された要約")
	}
}

func TestAddSource_FailsWhenMetadataGenerationFails(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))

	browser := &mockBrowser{
		findTabFound: false,
		currentURL:   "https://notebooklm.google.com/notebook/new-456",
	}
	gen := &mockMetadataGenerator{
		err: fmt.Errorf("claude -p failed"),
	}

	service := NewService(browser, "", store, gen)
	service.sleep = noSleep

	err := service.AddSource("テスト入力")
	if err == nil {
		t.Fatal("expected error when metadata generation fails")
	}
}
