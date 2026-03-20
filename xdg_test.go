package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestXDGPaths_ConfigDir(t *testing.T) {
	xdg := NewXDGPaths()
	want := filepath.Join(xdg.ConfigHome, "notebooklm-cli")
	if got := xdg.ConfigDir(); got != want {
		t.Errorf("ConfigDir() = %q, want %q", got, want)
	}
}

func TestXDGPaths_DataDir(t *testing.T) {
	xdg := NewXDGPaths()
	want := filepath.Join(xdg.DataHome, "notebooklm-cli")
	if got := xdg.DataDir(); got != want {
		t.Errorf("DataDir() = %q, want %q", got, want)
	}
}

func TestXDGPaths_MappingFile(t *testing.T) {
	xdg := NewXDGPaths()
	want := filepath.Join(xdg.DataHome, "notebooklm-cli", "mapping.yaml")
	if got := xdg.MappingFile(); got != want {
		t.Errorf("MappingFile() = %q, want %q", got, want)
	}
}

func TestXDGPaths_ConfigFile(t *testing.T) {
	xdg := NewXDGPaths()
	want := filepath.Join(xdg.ConfigHome, "notebooklm-cli", "config.yaml")
	if got := xdg.ConfigFile(); got != want {
		t.Errorf("ConfigFile() = %q, want %q", got, want)
	}
}

func TestXDGPaths_RespectsEnvVars(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/test-config")
	t.Setenv("XDG_DATA_HOME", "/tmp/test-data")

	xdg := NewXDGPaths()

	if xdg.ConfigHome != "/tmp/test-config" {
		t.Errorf("ConfigHome = %q, want /tmp/test-config", xdg.ConfigHome)
	}
	if xdg.DataHome != "/tmp/test-data" {
		t.Errorf("DataHome = %q, want /tmp/test-data", xdg.DataHome)
	}
}

func TestXDGPaths_DefaultPaths(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")

	xdg := NewXDGPaths()
	homeDir, _ := os.UserHomeDir()

	wantConfig := filepath.Join(homeDir, ".config")
	if xdg.ConfigHome != wantConfig {
		t.Errorf("ConfigHome = %q, want %q", xdg.ConfigHome, wantConfig)
	}

	wantData := filepath.Join(homeDir, ".local", "share")
	if xdg.DataHome != wantData {
		t.Errorf("DataHome = %q, want %q", xdg.DataHome, wantData)
	}
}

func TestXDGPaths_EnsureDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))

	xdg := NewXDGPaths()
	if err := xdg.EnsureDirectories(); err != nil {
		t.Fatalf("EnsureDirectories() error = %v", err)
	}

	dirs := []string{xdg.ConfigDir(), xdg.DataDir()}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("directory not created: %s", dir)
		}
	}
}
