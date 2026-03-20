package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// ComputeSHA256 はテキストのSHA256ハッシュを16進文字列で返す
func ComputeSHA256(text string) string {
	h := sha256.Sum256([]byte(text))
	return fmt.Sprintf("%x", h)
}

// MappingEntry はノートブックのメタデータ
type MappingEntry struct {
	URL         string            `yaml:"url"`
	Title       string            `yaml:"title"`
	Description string            `yaml:"description,omitempty"`
	Downloads   map[string]string `yaml:"downloads,omitempty"`
}

// MappingData はマッピングファイルのトップレベル構造
type MappingData struct {
	Entries map[string]*MappingEntry `yaml:"entries"`
}

// MappingStore はSHA256→ノートブックメタデータのマッピングを管理する
type MappingStore struct {
	path string
}

// NewMappingStore は指定パスのマッピングストアを作成する
func NewMappingStore(path string) *MappingStore {
	return &MappingStore{path: path}
}

// LookupEntry はハッシュに対応するエントリを返す
func (m *MappingStore) LookupEntry(hash string) (*MappingEntry, bool) {
	data, err := m.load()
	if err != nil {
		return nil, false
	}
	entry, ok := data.Entries[hash]
	return entry, ok
}

// SaveEntry はハッシュとエントリの対応を保存する
func (m *MappingStore) SaveEntry(hash string, entry *MappingEntry) error {
	data, err := m.load()
	if err != nil {
		data = &MappingData{Entries: make(map[string]*MappingEntry)}
	}
	data.Entries[hash] = entry

	return m.save(data)
}

// SaveMapping はハッシュとノートブックURLの対応を保存する（タイトルは決め打ちで自動生成）
func (m *MappingStore) SaveMapping(hash, notebookURL, inputText string) error {
	title := GenerateDefaultTitle(inputText)
	entry := &MappingEntry{
		URL:   notebookURL,
		Title: title,
	}
	return m.SaveEntry(hash, entry)
}

// UpdateDownload はダウンロード済みファイルパスを記録する
func (m *MappingStore) UpdateDownload(hash, downloadType, filePath string) error {
	data, err := m.load()
	if err != nil {
		return err
	}
	entry, ok := data.Entries[hash]
	if !ok {
		return fmt.Errorf("マッピングが見つかりません: hash=%s", hash[:12])
	}
	if entry.Downloads == nil {
		entry.Downloads = make(map[string]string)
	}
	entry.Downloads[downloadType] = filePath
	return m.save(data)
}

// DeleteByURL はノートブックURLに一致するマッピングを削除する
func (m *MappingStore) DeleteByURL(notebookURL string) error {
	data, err := m.load()
	if err != nil {
		return nil
	}

	for hash, entry := range data.Entries {
		if entry.URL == notebookURL {
			delete(data.Entries, hash)
		}
	}

	return m.save(data)
}

func (m *MappingStore) save(data *MappingData) error {
	if err := os.MkdirAll(filepath.Dir(m.path), 0755); err != nil {
		return err
	}
	out, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(m.path, out, 0644)
}

func (m *MappingStore) load() (*MappingData, error) {
	raw, err := os.ReadFile(m.path)
	if err != nil {
		return nil, err
	}
	var data MappingData
	if err := yaml.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	if data.Entries == nil {
		data.Entries = make(map[string]*MappingEntry)
	}
	return &data, nil
}

// GenerateDefaultTitle は入力テキスト先頭30文字+タイムスタンプで決め打ちタイトルを生成する
func GenerateDefaultTitle(text string) string {
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "/", "_")

	maxRunes := 30
	if utf8.RuneCountInString(text) > maxRunes {
		runes := []rune(text)
		text = string(runes[:maxRunes])
	}

	timestamp := time.Now().Format("20060102")
	if text == "" {
		return timestamp
	}
	return text + "_" + timestamp
}
