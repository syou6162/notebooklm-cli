package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Metadata はノートブックのタイトルと要約
type Metadata struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// MetadataGenerator はテキストからメタデータを生成するインターフェース
type MetadataGenerator interface {
	Generate(text string) (*Metadata, error)
}

// ClaudeMetadataGenerator はclaude -pを使ってメタデータを生成する
type ClaudeMetadataGenerator struct {
	Timeout time.Duration
}

// NewClaudeMetadataGenerator は新しいClaudeMetadataGeneratorを作成する
func NewClaudeMetadataGenerator() *ClaudeMetadataGenerator {
	return &ClaudeMetadataGenerator{
		Timeout: 60 * time.Second,
	}
}

const claudeJSONSchema = `{"type":"object","properties":{"title":{"type":"string","description":"30文字以内の簡潔なタイトル"},"description":{"type":"string","description":"1000文字程度の要約"}},"required":["title","description"]}`

const claudeSystemPrompt = "与えられたテキストに対してタイトルと要約を生成してください。タイトルは30文字以内で簡潔に。要約は1000文字程度でテキストの内容を的確にまとめてください。"

// Generate はclaude -pを呼び出してメタデータを生成する
func (g *ClaudeMetadataGenerator) Generate(text string) (*Metadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "claude", "-p",
		"--setting-sources", "user",
		"--system-prompt", claudeSystemPrompt,
		"--output-format", "json",
		"--json-schema", claudeJSONSchema,
		"--no-session-persistence",
		"以下のテキストに対してタイトルと要約を生成してください",
	)
	cmd.Stdin = strings.NewReader(text)

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("claude -p の実行に失敗しました: %w", err)
	}

	return parseClaudeResponse(out)
}

// claudeResponseEntry はclaude -p --output-format jsonの出力の各要素
type claudeResponseEntry struct {
	Type             string    `json:"type"`
	StructuredOutput *Metadata `json:"structured_output,omitempty"`
}

// parseClaudeResponse はclaude -pのJSON配列出力からstructured_outputを取り出す
func parseClaudeResponse(data []byte) (*Metadata, error) {
	var entries []claudeResponseEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("claude出力のパースに失敗しました: %w", err)
	}

	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].Type == "result" && entries[i].StructuredOutput != nil {
			return entries[i].StructuredOutput, nil
		}
	}

	return nil, fmt.Errorf("claude出力にstructured_outputが見つかりませんでした")
}
