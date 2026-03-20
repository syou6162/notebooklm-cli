package main

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DownloadsConfig はダウンロード先の設定
type DownloadsConfig struct {
	Infographic string `yaml:"infographic,omitempty"`
	Audio       string `yaml:"audio,omitempty"`
	Slide       string `yaml:"slide,omitempty"`
	Video       string `yaml:"video,omitempty"`
}

// Config はアプリケーションの設定
type Config struct {
	Downloads DownloadsConfig `yaml:"downloads"`
}

// LoadConfig は設定ファイルを読み込む。ファイルが存在しない場合はデフォルト設定を返す。
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// ResolveDownloadDir はダウンロード先ディレクトリを解決する。
// flagValueが指定されていればそちらを優先、なければconfigの値を使う。
func (c *Config) ResolveDownloadDir(downloadType, flagValue string) string {
	if flagValue != "" {
		return flagValue
	}

	var configValue string
	switch downloadType {
	case "infographic":
		configValue = c.Downloads.Infographic
	case "audio":
		configValue = c.Downloads.Audio
	case "slide":
		configValue = c.Downloads.Slide
	case "video":
		configValue = c.Downloads.Video
	}

	return expandPath(configValue)
}

// expandPath はパスを展開して正規化する
func expandPath(path string) string {
	if path == "" {
		return ""
	}

	if len(path) >= 2 && path[:2] == "~/" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, path[2:])
		}
	}

	path = os.ExpandEnv(path)

	return filepath.Clean(path)
}
