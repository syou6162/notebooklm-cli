package main

import (
	"fmt"
	"strings"
	"time"
)

const (
	maxDeleteAttempts = 20
)

// Service はNotebookLM高レベルワークフロー
type Service struct {
	client      *Client
	notebookURL string
	mapping     *MappingStore
}

// NewService は新しいServiceを作成する
func NewService(client *Client, notebookURL string, mapping *MappingStore) *Service {
	return &Service{
		client:      client,
		notebookURL: notebookURL,
		mapping:     mapping,
	}
}

// EnsureNotebookPage はノートブックページにいることを確認し、必要であればナビゲートする
func (s *Service) EnsureNotebookPage() error {
	tabIndex, found := s.client.FindNotebookTab()
	if !found {
		newTab, err := s.client.OpenURLInNewTab(s.notebookURL)
		if err != nil {
			return err
		}
		s.client.tabIndex = newTab
		time.Sleep(3 * time.Second)
		return nil
	}

	s.client.tabIndex = tabIndex
	currentURL, err := s.client.GetCurrentURL()
	if err != nil {
		return err
	}
	if !strings.Contains(currentURL, s.notebookURL) {
		if err := s.client.NavigateTo(s.notebookURL); err != nil {
			return err
		}
		time.Sleep(3 * time.Second)
	}
	return nil
}

// DeleteAllSources は既存ソースをすべて削除する
func (s *Service) DeleteAllSources() error {
	s.client.CloseSourceViewerIfOpen()

	for range maxDeleteAttempts {
		count := s.client.GetElementCount(sourceCheckbox)
		if count <= 0 {
			return nil
		}

		if err := s.client.ClickButton(moreButtonAria); err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)

		if err := s.client.ClickMenuItem(deleteMenuText); err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)

		if err := s.client.ClickOverlayButton(confirmDeleteText); err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
	}

	return &BrowserError{
		Message: fmt.Sprintf("ソースの削除に失敗しました（%d回試行後もソースが残存しています）", maxDeleteAttempts),
	}
}

// AddSourceText はテキストをノートブックのソースとして追加する
func (s *Service) AddSourceText(text string) error {
	s.client.CloseSourceViewerIfOpen()

	if err := s.client.ClickButton(sourceAddButtonAria); err != nil {
		return err
	}
	time.Sleep(1 * time.Second)

	if err := s.client.ClickButtonByText(pasteTextIcon); err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)

	if err := s.client.activateChrome(); err != nil {
		return err
	}
	if err := s.client.FocusElement("." + textareaClass); err != nil {
		return err
	}
	time.Sleep(300 * time.Millisecond)

	if err := s.client.clipboardPaste(text); err != nil {
		return err
	}
	time.Sleep(1 * time.Second)

	if err := s.client.ClickButtonByText(insertButtonText); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)

	return nil
}

// AddSource はテキストからソースを追加する（マッピング管理込み）
func (s *Service) AddSource(text string) error {
	if strings.TrimSpace(text) == "" {
		return &BrowserError{Message: "テキストが空です"}
	}

	hash := ComputeSHA256(text)
	if _, found := s.mapping.LookupNotebook(hash); found {
		fmt.Println("同一入力のノートブックが既に存在します。スキップします。")
		return nil
	}

	if err := s.EnsureNotebookPage(); err != nil {
		return err
	}

	if err := s.AddSourceText(text); err != nil {
		return err
	}

	currentURL, err := s.client.GetCurrentURL()
	if err != nil {
		return err
	}

	if err := s.mapping.SaveMapping(hash, currentURL); err != nil {
		return fmt.Errorf("マッピングの保存に失敗しました: %w", err)
	}

	return nil
}

// DeleteSource はノートブック内の全ソースを削除し、マッピングも削除する
func (s *Service) DeleteSource() error {
	if err := s.EnsureNotebookPage(); err != nil {
		return err
	}
	if err := s.DeleteAllSources(); err != nil {
		return err
	}
	return s.mapping.DeleteByURL(s.notebookURL)
}

// DeleteAllInfographics は既存のインフォグラフィックをすべて削除する
func (s *Service) DeleteAllInfographics() error {
	for range maxDeleteAttempts {
		if s.client.CountInfographicCards() <= 0 {
			return nil
		}

		if err := s.client.ClickMoreButtonOnFirstInfographicCard(); err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)

		if err := s.client.ClickMenuItem(deleteMenuText); err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)

		if err := s.client.ClickOverlayButton(confirmDeleteText); err != nil {
			return err
		}
		time.Sleep(1500 * time.Millisecond)
	}

	return &BrowserError{
		Message: fmt.Sprintf("インフォグラフィックの削除に失敗しました（%d回試行後もカードが残存しています）", maxDeleteAttempts),
	}
}

// GenerateInfographic は既存インフォグラフィックを削除してから生成を開始する（即返し）
func (s *Service) GenerateInfographic() error {
	if err := s.EnsureNotebookPage(); err != nil {
		return err
	}

	if err := s.DeleteAllInfographics(); err != nil {
		return err
	}

	return s.client.ClickButton(infographicButtonAria)
}

// StatusInfographic はインフォグラフィックの生成状態を返す
func (s *Service) StatusInfographic() (string, error) {
	if err := s.EnsureNotebookPage(); err != nil {
		return "", err
	}

	if s.client.PageContainsText(generatingText) {
		return "generating", nil
	}

	if s.client.CountInfographicCards() > 0 {
		return "done", nil
	}

	return "none", nil
}

// DownloadInfographic はインフォグラフィックをダウンロードする
func (s *Service) DownloadInfographic() error {
	if err := s.EnsureNotebookPage(); err != nil {
		return err
	}

	if s.client.CountInfographicCards() <= 0 {
		return &BrowserError{Message: "インフォグラフィックカードが見つかりませんでした"}
	}

	if err := s.client.ClickMoreButtonOnFirstInfographicCard(); err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)

	return s.client.ClickMenuItem(downloadMenuText)
}
