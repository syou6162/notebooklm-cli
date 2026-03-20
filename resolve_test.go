package main

import (
	"path/filepath"
	"testing"
)

func TestResolveSource_Found(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))

	text := "test input text"
	hash := ComputeSHA256(text)
	notebookURL := "https://notebooklm.google.com/notebook/abc-123"
	if err := store.SaveMapping(hash, notebookURL); err != nil {
		t.Fatalf("SaveMapping() error = %v", err)
	}

	got, err := ResolveSource(text, store)
	if err != nil {
		t.Fatalf("ResolveSource() error = %v", err)
	}
	if got != notebookURL {
		t.Errorf("ResolveSource() = %q, want %q", got, notebookURL)
	}
}

func TestResolveSource_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))

	_, err := ResolveSource("unknown input", store)
	if err == nil {
		t.Fatal("ResolveSource() expected error for unknown input, got nil")
	}
}

func TestResolveSource_EmptyText(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))

	_, err := ResolveSource("", store)
	if err == nil {
		t.Fatal("ResolveSource() expected error for empty text, got nil")
	}
}
