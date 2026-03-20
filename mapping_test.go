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

func TestMappingStore_LookupEntry_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))

	entry, found := store.LookupEntry("nonexistent-hash")
	if found {
		t.Error("expected not found for empty store")
	}
	if entry != nil {
		t.Errorf("expected nil entry, got %v", entry)
	}
}

func TestMappingStore_SaveAndLookup(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))

	hash := ComputeSHA256("test input")
	notebookURL := "https://notebooklm.google.com/notebook/abc-123"

	if err := store.SaveMapping(hash, notebookURL, "test input"); err != nil {
		t.Fatalf("SaveMapping() error = %v", err)
	}

	entry, found := store.LookupEntry(hash)
	if !found {
		t.Error("expected found after save")
	}
	if entry.URL != notebookURL {
		t.Errorf("entry.URL = %q, want %q", entry.URL, notebookURL)
	}
	if entry.Title == "" {
		t.Error("expected non-empty title")
	}
}

func TestMappingStore_PersistsToFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "mapping.yaml")

	store1 := NewMappingStore(path)
	hash := ComputeSHA256("persist test")
	notebookURL := "https://notebooklm.google.com/notebook/def-456"
	if err := store1.SaveMapping(hash, notebookURL, "persist test"); err != nil {
		t.Fatalf("SaveMapping() error = %v", err)
	}

	store2 := NewMappingStore(path)
	entry, found := store2.LookupEntry(hash)
	if !found {
		t.Error("expected found after reload from file")
	}
	if entry.URL != notebookURL {
		t.Errorf("entry.URL after reload = %q, want %q", entry.URL, notebookURL)
	}
}

func TestMappingStore_MultipleEntries(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))

	inputs := []struct {
		text string
		url  string
	}{
		{"input 1", "https://notebooklm.google.com/notebook/aaa"},
		{"input 2", "https://notebooklm.google.com/notebook/bbb"},
		{"input 3", "https://notebooklm.google.com/notebook/ccc"},
	}

	for _, in := range inputs {
		hash := ComputeSHA256(in.text)
		if err := store.SaveMapping(hash, in.url, in.text); err != nil {
			t.Fatalf("SaveMapping() error = %v", err)
		}
	}

	for _, in := range inputs {
		hash := ComputeSHA256(in.text)
		entry, found := store.LookupEntry(hash)
		if !found {
			t.Errorf("expected found for hash of %q", in.text)
		}
		if entry.URL != in.url {
			t.Errorf("entry.URL = %q, want %q", entry.URL, in.url)
		}
	}
}

func TestMappingStore_DeleteByURL(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))

	hash := ComputeSHA256("delete test")
	url := "https://notebooklm.google.com/notebook/to-delete"
	if err := store.SaveMapping(hash, url, "delete test"); err != nil {
		t.Fatalf("SaveMapping() error = %v", err)
	}

	if err := store.DeleteByURL(url); err != nil {
		t.Fatalf("DeleteByURL() error = %v", err)
	}

	_, found := store.LookupEntry(hash)
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

	if err := store.SaveMapping(hash1, url1, "keep this"); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveMapping(hash2, url2, "delete this"); err != nil {
		t.Fatal(err)
	}

	if err := store.DeleteByURL(url2); err != nil {
		t.Fatalf("DeleteByURL() error = %v", err)
	}

	_, found1 := store.LookupEntry(hash1)
	if !found1 {
		t.Error("expected keep entry to still exist")
	}

	_, found2 := store.LookupEntry(hash2)
	if found2 {
		t.Error("expected deleted entry to be gone")
	}
}

func TestMappingStore_UpdateDownload(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))

	hash := ComputeSHA256("download test")
	if err := store.SaveMapping(hash, "https://notebooklm.google.com/notebook/dl", "download test"); err != nil {
		t.Fatal(err)
	}

	if err := store.UpdateDownload(hash, "infographic", "./images/test.png"); err != nil {
		t.Fatalf("UpdateDownload() error = %v", err)
	}

	entry, _ := store.LookupEntry(hash)
	if entry.Downloads["infographic"] != "./images/test.png" {
		t.Errorf("Downloads[infographic] = %q, want ./images/test.png", entry.Downloads["infographic"])
	}
}

func TestGenerateDefaultTitle(t *testing.T) {
	title := GenerateDefaultTitle("AIの最新動向について解説します。これは長いテキストなので30文字で切られるはずです。")
	if len([]rune(title)) > 40 { // 30文字 + _YYYYMMDD(9文字)
		t.Errorf("title too long: %q (%d runes)", title, len([]rune(title)))
	}
}

func TestGenerateDefaultTitle_EmptyText(t *testing.T) {
	title := GenerateDefaultTitle("")
	if title == "" {
		t.Error("expected non-empty title even for empty input")
	}
}

func TestGenerateDefaultTitle_ShortText(t *testing.T) {
	title := GenerateDefaultTitle("短いテキスト")
	if title == "" {
		t.Error("expected non-empty title")
	}
}
