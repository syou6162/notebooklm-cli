package main

import (
	"path/filepath"
	"testing"
)

func TestComputeSHA256(t *testing.T) {
	hash := ComputeSHA256("hello world")
	want := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if hash != want {
		t.Errorf("ComputeSHA256(\"hello world\") = %q, want %q", hash, want)
	}
}

func TestComputeSHA256_DifferentInputs(t *testing.T) {
	h1 := ComputeSHA256("input A")
	h2 := ComputeSHA256("input B")
	if h1 == h2 {
		t.Error("different inputs should produce different hashes")
	}
}

func TestComputeSHA256_SameInput(t *testing.T) {
	h1 := ComputeSHA256("same text")
	h2 := ComputeSHA256("same text")
	if h1 != h2 {
		t.Error("same inputs should produce same hashes")
	}
}

func TestMappingStore_LookupNotebook_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))

	url, found := store.LookupNotebook("nonexistent-hash")
	if found {
		t.Error("expected not found for empty store")
	}
	if url != "" {
		t.Errorf("expected empty url, got %q", url)
	}
}

func TestMappingStore_SaveAndLookup(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))

	hash := ComputeSHA256("test input")
	notebookURL := "https://notebooklm.google.com/notebook/abc-123"

	if err := store.SaveMapping(hash, notebookURL); err != nil {
		t.Fatalf("SaveMapping() error = %v", err)
	}

	url, found := store.LookupNotebook(hash)
	if !found {
		t.Error("expected found after save")
	}
	if url != notebookURL {
		t.Errorf("LookupNotebook() = %q, want %q", url, notebookURL)
	}
}

func TestMappingStore_PersistsToFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "mapping.yaml")

	store1 := NewMappingStore(path)
	hash := ComputeSHA256("persist test")
	notebookURL := "https://notebooklm.google.com/notebook/def-456"
	if err := store1.SaveMapping(hash, notebookURL); err != nil {
		t.Fatalf("SaveMapping() error = %v", err)
	}

	// 新しいインスタンスでファイルから読み込み
	store2 := NewMappingStore(path)
	url, found := store2.LookupNotebook(hash)
	if !found {
		t.Error("expected found after reload from file")
	}
	if url != notebookURL {
		t.Errorf("LookupNotebook() after reload = %q, want %q", url, notebookURL)
	}
}

func TestMappingStore_MultipleEntries(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))

	entries := map[string]string{
		ComputeSHA256("input 1"): "https://notebooklm.google.com/notebook/aaa",
		ComputeSHA256("input 2"): "https://notebooklm.google.com/notebook/bbb",
		ComputeSHA256("input 3"): "https://notebooklm.google.com/notebook/ccc",
	}

	for hash, url := range entries {
		if err := store.SaveMapping(hash, url); err != nil {
			t.Fatalf("SaveMapping() error = %v", err)
		}
	}

	for hash, wantURL := range entries {
		gotURL, found := store.LookupNotebook(hash)
		if !found {
			t.Errorf("expected found for hash %q", hash)
		}
		if gotURL != wantURL {
			t.Errorf("LookupNotebook(%q) = %q, want %q", hash, gotURL, wantURL)
		}
	}
}

func TestMappingStore_DeleteByURL(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))

	hash := ComputeSHA256("delete test")
	url := "https://notebooklm.google.com/notebook/to-delete"
	if err := store.SaveMapping(hash, url); err != nil {
		t.Fatalf("SaveMapping() error = %v", err)
	}

	if err := store.DeleteByURL(url); err != nil {
		t.Fatalf("DeleteByURL() error = %v", err)
	}

	_, found := store.LookupNotebook(hash)
	if found {
		t.Error("expected not found after DeleteByURL")
	}
}

func TestMappingStore_DeleteByURL_PreservesOtherEntries(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))

	hash1 := ComputeSHA256("keep this")
	url1 := "https://notebooklm.google.com/notebook/keep"
	hash2 := ComputeSHA256("delete this")
	url2 := "https://notebooklm.google.com/notebook/delete"

	if err := store.SaveMapping(hash1, url1); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveMapping(hash2, url2); err != nil {
		t.Fatal(err)
	}

	if err := store.DeleteByURL(url2); err != nil {
		t.Fatalf("DeleteByURL() error = %v", err)
	}

	_, found1 := store.LookupNotebook(hash1)
	if !found1 {
		t.Error("expected keep entry to still exist")
	}

	_, found2 := store.LookupNotebook(hash2)
	if found2 {
		t.Error("expected deleted entry to be gone")
	}
}
