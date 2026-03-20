package main

import (
	"os"
	"path/filepath"
)

// XDGPaths はXDG Base Directory仕様に基づくパスを管理する
type XDGPaths struct {
	ConfigHome string
	DataHome   string
}

// NewXDGPaths はXDGPathsの新しいインスタンスを作成する
func NewXDGPaths() *XDGPaths {
	return &XDGPaths{
		ConfigHome: getXDGConfigHome(),
		DataHome:   getXDGDataHome(),
	}
}

func getXDGConfigHome() string {
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
		return v
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config")
}

func getXDGDataHome() string {
	if v := os.Getenv("XDG_DATA_HOME"); v != "" {
		return v
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".local", "share")
}

// ConfigDir はnotebooklm-cliの設定ディレクトリパスを返す
func (x *XDGPaths) ConfigDir() string {
	return filepath.Join(x.ConfigHome, "notebooklm-cli")
}

// ConfigFile はnotebooklm-cliの設定ファイルパスを返す
func (x *XDGPaths) ConfigFile() string {
	return filepath.Join(x.ConfigDir(), "config.yaml")
}

// DataDir はnotebooklm-cliのデータディレクトリパスを返す
func (x *XDGPaths) DataDir() string {
	return filepath.Join(x.DataHome, "notebooklm-cli")
}

// MappingFile はSHA256→ノートブックURLのマッピングファイルパスを返す
func (x *XDGPaths) MappingFile() string {
	return filepath.Join(x.DataDir(), "mapping.yaml")
}

// EnsureDirectories は必要なディレクトリを作成する
func (x *XDGPaths) EnsureDirectories() error {
	dirs := []string{
		x.ConfigDir(),
		x.DataDir(),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}
