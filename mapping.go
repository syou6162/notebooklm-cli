package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ComputeSHA256 はテキストのSHA256ハッシュを16進文字列で返す
func ComputeSHA256(text string) string {
	h := sha256.Sum256([]byte(text))
	return fmt.Sprintf("%x", h)
}

// MappingStore はSHA256→ノートブックURLのマッピングを管理する
type MappingStore struct {
	path string
}

// NewMappingStore は指定パスのマッピングストアを作成する
func NewMappingStore(path string) *MappingStore {
	return &MappingStore{path: path}
}

// LookupNotebook はハッシュに対応するノートブックURLを返す
func (m *MappingStore) LookupNotebook(hash string) (string, bool) {
	data, err := m.load()
	if err != nil {
		return "", false
	}
	url, ok := data[hash]
	return url, ok
}

// SaveMapping はハッシュとノートブックURLの対応を保存する
func (m *MappingStore) SaveMapping(hash, notebookURL string) error {
	data, err := m.load()
	if err != nil {
		data = make(map[string]string)
	}
	data[hash] = notebookURL

	if err := os.MkdirAll(filepath.Dir(m.path), 0755); err != nil {
		return err
	}

	out, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(m.path, out, 0644)
}

func (m *MappingStore) load() (map[string]string, error) {
	raw, err := os.ReadFile(m.path)
	if err != nil {
		return nil, err
	}
	var data map[string]string
	if err := yaml.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	return data, nil
}
