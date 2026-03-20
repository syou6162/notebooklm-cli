package main

import (
	"fmt"
	"strings"
)

// ResolveSource はテキストからSHA256マッピングを引いてノートブックURLを返す
func ResolveSource(text string, store *MappingStore) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("テキストが空です")
	}

	hash := ComputeSHA256(text)
	url, found := store.LookupNotebook(hash)
	if !found {
		return "", fmt.Errorf("マッピングが見つかりません（hash=%s）", hash[:12])
	}

	return url, nil
}
