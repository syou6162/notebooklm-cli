package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_ReturnsDefaultWhenFileNotExists(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Downloads.Infographic != "" {
		t.Errorf("expected empty default, got %q", cfg.Downloads.Infographic)
	}
}

func TestLoadConfig_ReadsDownloadsConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	content := `downloads:
  infographic: "~/Documents/notebooklm/infographics"
  audio: "~/.local/share/podself/episodes"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.Downloads.Infographic != "~/Documents/notebooklm/infographics" {
		t.Errorf("Infographic = %q, want ~/Documents/notebooklm/infographics", cfg.Downloads.Infographic)
	}
	if cfg.Downloads.Audio != "~/.local/share/podself/episodes" {
		t.Errorf("Audio = %q, want ~/.local/share/podself/episodes", cfg.Downloads.Audio)
	}
}

func TestConfig_ResolveDownloadDir_UsesConfigValue(t *testing.T) {
	cfg := &Config{
		Downloads: DownloadsConfig{
			Infographic: "~/Documents/infographics",
		},
	}
	dir := cfg.ResolveDownloadDir("infographic", "")
	homeDir, _ := os.UserHomeDir()
	want := filepath.Join(homeDir, "Documents", "infographics")
	if dir != want {
		t.Errorf("ResolveDownloadDir() = %q, want %q", dir, want)
	}
}

func TestConfig_ResolveDownloadDir_FlagOverridesConfig(t *testing.T) {
	cfg := &Config{
		Downloads: DownloadsConfig{
			Infographic: "~/Documents/infographics",
		},
	}
	dir := cfg.ResolveDownloadDir("infographic", "/tmp/override")
	if dir != "/tmp/override" {
		t.Errorf("ResolveDownloadDir() = %q, want /tmp/override", dir)
	}
}

func TestConfig_ResolveDownloadDir_EmptyBoth(t *testing.T) {
	cfg := &Config{}
	dir := cfg.ResolveDownloadDir("infographic", "")
	if dir != "" {
		t.Errorf("ResolveDownloadDir() = %q, want empty", dir)
	}
}

func TestExpandPath_ExpandsTilde(t *testing.T) {
	result := expandPath("~/Documents")
	homeDir, _ := os.UserHomeDir()
	want := filepath.Join(homeDir, "Documents")
	if result != want {
		t.Errorf("expandPath() = %q, want %q", result, want)
	}
}

func TestExpandPath_EmptyReturnsEmpty(t *testing.T) {
	result := expandPath("")
	if result != "" {
		t.Errorf("expandPath() = %q, want empty", result)
	}
}
