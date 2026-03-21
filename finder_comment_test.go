package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetFinderComment(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.m4a")
	if err := os.WriteFile(testFile, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	comment := "テストコメント: ポッドキャストの説明文"
	if err := SetFinderComment(testFile, comment); err != nil {
		t.Fatalf("SetFinderComment() error = %v", err)
	}

	got, err := GetFinderComment(testFile)
	if err != nil {
		t.Fatalf("GetFinderComment() error = %v", err)
	}
	if got != comment {
		t.Errorf("GetFinderComment() = %q, want %q", got, comment)
	}
}

func TestSetFinderComment_FileNotExists(t *testing.T) {
	err := SetFinderComment("/nonexistent/file.m4a", "comment")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestGetFinderComment_NoComment(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "no_comment.m4a")
	if err := os.WriteFile(testFile, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := GetFinderComment(testFile)
	if err != nil {
		t.Fatalf("GetFinderComment() error = %v", err)
	}
	if got != "" {
		t.Errorf("GetFinderComment() = %q, want empty", got)
	}
}
