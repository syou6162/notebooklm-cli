package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFindDownloadedInfographic_FindsUnnamedPNG(t *testing.T) {
	tmpDir := t.TempDir()
	startTime := time.Now().Add(-1 * time.Second)
	pngFile := filepath.Join(tmpDir, "unnamed.png")
	if err := os.WriteFile(pngFile, make([]byte, 2*1024*1024), 0644); err != nil {
		t.Fatal(err)
	}
	result, err := FindDownloadedInfographic(tmpDir, startTime, 1_000_000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != pngFile {
		t.Errorf("got %q, want %q", result, pngFile)
	}
}

func TestFindDownloadedInfographic_FindsUnnamedWithSuffix(t *testing.T) {
	tmpDir := t.TempDir()
	startTime := time.Now().Add(-1 * time.Second)
	pngFile := filepath.Join(tmpDir, "unnamed (1).png")
	if err := os.WriteFile(pngFile, make([]byte, 2*1024*1024), 0644); err != nil {
		t.Fatal(err)
	}
	result, err := FindDownloadedInfographic(tmpDir, startTime, 1_000_000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != pngFile {
		t.Errorf("got %q, want %q", result, pngFile)
	}
}

func TestFindDownloadedInfographic_IgnoresOldFiles(t *testing.T) {
	tmpDir := t.TempDir()
	pngFile := filepath.Join(tmpDir, "unnamed.png")
	if err := os.WriteFile(pngFile, make([]byte, 2*1024*1024), 0644); err != nil {
		t.Fatal(err)
	}
	startTime := time.Now().Add(10 * time.Second)
	_, err := FindDownloadedInfographic(tmpDir, startTime, 1_000_000)
	if err == nil {
		t.Fatal("expected error for old file, got nil")
	}
}

func TestFindDownloadedInfographic_IgnoresSmallFiles(t *testing.T) {
	tmpDir := t.TempDir()
	startTime := time.Now().Add(-1 * time.Second)
	pngFile := filepath.Join(tmpDir, "unnamed.png")
	if err := os.WriteFile(pngFile, make([]byte, 100), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := FindDownloadedInfographic(tmpDir, startTime, 1_000_000)
	if err == nil {
		t.Fatal("expected error for small file, got nil")
	}
}

func TestFindDownloadedInfographic_ReturnsLatest(t *testing.T) {
	tmpDir := t.TempDir()
	startTime := time.Now().Add(-1 * time.Second)

	old := filepath.Join(tmpDir, "unnamed.png")
	if err := os.WriteFile(old, make([]byte, 2*1024*1024), 0644); err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	newer := filepath.Join(tmpDir, "unnamed (1).png")
	if err := os.WriteFile(newer, make([]byte, 2*1024*1024), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := FindDownloadedInfographic(tmpDir, startTime, 1_000_000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != newer {
		t.Errorf("got %q, want %q", result, newer)
	}
}

func TestMoveToOutput_MovesWithRename(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "unnamed.png")
	data := []byte("PNG data")
	if err := os.WriteFile(src, data, 0644); err != nil {
		t.Fatal(err)
	}

	outputDir := filepath.Join(tmpDir, "output")
	dest, err := MoveToOutput(src, outputDir, "テストタイトル")
	if err != nil {
		t.Fatalf("MoveToOutput() error = %v", err)
	}

	if filepath.Base(dest) != "テストタイトル.png" {
		t.Errorf("filename = %q, want テストタイトル.png", filepath.Base(dest))
	}

	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("reading dest: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("file content mismatch")
	}

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("expected source file to be removed after move")
	}
}

func TestMoveToOutput_EmptyNameUsesOriginal(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "unnamed.png")
	if err := os.WriteFile(src, []byte("PNG"), 0644); err != nil {
		t.Fatal(err)
	}

	outputDir := filepath.Join(tmpDir, "output")
	dest, err := MoveToOutput(src, outputDir, "")
	if err != nil {
		t.Fatalf("MoveToOutput() error = %v", err)
	}

	if filepath.Base(dest) != "unnamed.png" {
		t.Errorf("filename = %q, want unnamed.png", filepath.Base(dest))
	}
}

func TestMoveToOutput_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "unnamed.png")
	if err := os.WriteFile(src, []byte("PNG"), 0644); err != nil {
		t.Fatal(err)
	}

	outputDir := filepath.Join(tmpDir, "new", "subdir")
	_, err := MoveToOutput(src, outputDir, "test")
	if err != nil {
		t.Fatalf("MoveToOutput() error = %v", err)
	}

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		t.Error("output directory not created")
	}
}
