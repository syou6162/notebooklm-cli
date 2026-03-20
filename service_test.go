package main

import (
	"path/filepath"
	"testing"
	"time"
)

// mockBrowser はテスト用のBrowserモック
type mockBrowser struct {
	tabIndex         int
	currentURL       string
	findTabResult    int
	findTabFound     bool
	elementCount     int
	infographicCards int
	pageText         string
	sourceNames      []string
	clickedButtons   []string
	clickedMenuItems []string
	navigatedURLs    []string
	openedURLs       []string
	pastedTexts      []string
	focusedSelectors []string
}

func (m *mockBrowser) FindNotebookTab() (int, bool) { return m.findTabResult, m.findTabFound }
func (m *mockBrowser) OpenURLInNewTab(url string) (int, error) {
	m.openedURLs = append(m.openedURLs, url)
	return 1, nil
}
func (m *mockBrowser) NavigateTo(url string) error {
	m.navigatedURLs = append(m.navigatedURLs, url)
	return nil
}
func (m *mockBrowser) GetCurrentURL() (string, error) { return m.currentURL, nil }
func (m *mockBrowser) ClickButton(ariaLabel string) error {
	m.clickedButtons = append(m.clickedButtons, ariaLabel)
	return nil
}
func (m *mockBrowser) ClickButtonByText(text string) error {
	m.clickedButtons = append(m.clickedButtons, text)
	return nil
}
func (m *mockBrowser) ClickMenuItem(text string) error {
	m.clickedMenuItems = append(m.clickedMenuItems, text)
	return nil
}
func (m *mockBrowser) ClickOverlayButton(_ string) error { return nil }
func (m *mockBrowser) GetElementCount(_ string) int      { return m.elementCount }
func (m *mockBrowser) FocusElement(selector string) error {
	m.focusedSelectors = append(m.focusedSelectors, selector)
	return nil
}
func (m *mockBrowser) PageContainsText(text string) bool {
	return m.pageText != "" && m.pageText == text
}
func (m *mockBrowser) CloseSourceViewerIfOpen()                     {}
func (m *mockBrowser) CountInfographicCards() int                   { return m.infographicCards }
func (m *mockBrowser) ClickMoreButtonOnFirstInfographicCard() error { return nil }
func (m *mockBrowser) ListSourceNames() ([]string, error)           { return m.sourceNames, nil }
func (m *mockBrowser) SetTabIndex(index int)                        { m.tabIndex = index }
func (m *mockBrowser) ActivateChrome() error                        { return nil }
func (m *mockBrowser) ClipboardPaste(text string) error {
	m.pastedTexts = append(m.pastedTexts, text)
	return nil
}

func noSleep(_ time.Duration) {}

func TestAddSource_SkipsWhenMappingExists(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))
	text := "existing input"
	hash := ComputeSHA256(text)
	if err := store.SaveMapping(hash, "https://notebooklm.google.com/notebook/existing", text); err != nil {
		t.Fatal(err)
	}

	browser := &mockBrowser{}
	service := NewService(browser, "", store)
	service.sleep = noSleep

	err := service.AddSource(text)
	if err != nil {
		t.Fatalf("AddSource() error = %v", err)
	}
	if len(browser.clickedButtons) > 0 {
		t.Error("expected no browser interaction when mapping exists")
	}
}

func TestAddSource_CreatesNotebookWhenURLEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))

	browser := &mockBrowser{
		findTabFound: false,
		currentURL:   "https://notebooklm.google.com/notebook/new-123",
	}
	service := NewService(browser, "", store)
	service.sleep = noSleep

	err := service.AddSource("new input text")
	if err != nil {
		t.Fatalf("AddSource() error = %v", err)
	}

	// ノートブック作成ボタンがクリックされたか
	foundCreate := false
	for _, btn := range browser.clickedButtons {
		if btn == createNotebookButtonText {
			foundCreate = true
			break
		}
	}
	if !foundCreate {
		t.Errorf("expected create notebook button click, got buttons: %v", browser.clickedButtons)
	}

	// マッピングが保存されたか
	hash := ComputeSHA256("new input text")
	entry, found := store.LookupEntry(hash)
	if !found {
		t.Error("expected mapping to be saved")
	}
	if entry.URL != "https://notebooklm.google.com/notebook/new-123" {
		t.Errorf("entry.URL = %q, want notebook URL", entry.URL)
	}
}

func TestAddSource_UsesExistingURLWhenProvided(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))

	browser := &mockBrowser{
		findTabFound:  true,
		findTabResult: 1,
		currentURL:    "https://notebooklm.google.com/notebook/existing-456",
	}
	service := NewService(browser, "https://notebooklm.google.com/notebook/existing-456", store)
	service.sleep = noSleep

	err := service.AddSource("input with url")
	if err != nil {
		t.Fatalf("AddSource() error = %v", err)
	}

	// ノートブック作成ボタンはクリックされていないこと
	for _, btn := range browser.clickedButtons {
		if btn == createNotebookButtonText {
			t.Error("should not create notebook when URL is provided")
		}
	}
}

func TestAddSource_RejectsEmptyText(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))
	browser := &mockBrowser{}
	service := NewService(browser, "", store)
	service.sleep = noSleep

	err := service.AddSource("")
	if err == nil {
		t.Fatal("expected error for empty text")
	}
}

func TestCreateNotebook_ClicksCreateButton(t *testing.T) {
	browser := &mockBrowser{
		findTabFound: false,
		currentURL:   "https://notebooklm.google.com/notebook/created-789",
	}
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))
	service := NewService(browser, "", store)
	service.sleep = noSleep

	url, err := service.CreateNotebook()
	if err != nil {
		t.Fatalf("CreateNotebook() error = %v", err)
	}
	if url != "https://notebooklm.google.com/notebook/created-789" {
		t.Errorf("CreateNotebook() = %q, want created URL", url)
	}

	// ホームページが開かれたか
	if len(browser.openedURLs) == 0 || browser.openedURLs[0] != notebookLMHomeURL {
		t.Errorf("expected home URL to be opened, got %v", browser.openedURLs)
	}
}

func TestDeleteSource_DeletesMappingToo(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))
	notebookURL := "https://notebooklm.google.com/notebook/to-delete"
	hash := ComputeSHA256("delete me")
	if err := store.SaveMapping(hash, notebookURL, "delete me"); err != nil {
		t.Fatal(err)
	}

	browser := &mockBrowser{
		findTabFound:  true,
		findTabResult: 1,
		currentURL:    notebookURL,
		elementCount:  0,
	}
	service := NewService(browser, notebookURL, store)
	service.sleep = noSleep

	if err := service.DeleteSource(); err != nil {
		t.Fatalf("DeleteSource() error = %v", err)
	}

	_, found := store.LookupEntry(hash)
	if found {
		t.Error("expected mapping to be deleted after DeleteSource")
	}
}

func TestStatusInfographic_ReturnsGenerating(t *testing.T) {
	browser := &mockBrowser{
		findTabFound:  true,
		findTabResult: 1,
		currentURL:    "https://notebooklm.google.com/notebook/test",
		pageText:      generatingText,
	}
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))
	service := NewService(browser, "https://notebooklm.google.com/notebook/test", store)
	service.sleep = noSleep

	status, err := service.StatusInfographic()
	if err != nil {
		t.Fatalf("StatusInfographic() error = %v", err)
	}
	if status != "generating" {
		t.Errorf("StatusInfographic() = %q, want \"generating\"", status)
	}
}

func TestStatusInfographic_ReturnsDone(t *testing.T) {
	browser := &mockBrowser{
		findTabFound:     true,
		findTabResult:    1,
		currentURL:       "https://notebooklm.google.com/notebook/test",
		infographicCards: 1,
	}
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))
	service := NewService(browser, "https://notebooklm.google.com/notebook/test", store)
	service.sleep = noSleep

	status, err := service.StatusInfographic()
	if err != nil {
		t.Fatalf("StatusInfographic() error = %v", err)
	}
	if status != "done" {
		t.Errorf("StatusInfographic() = %q, want \"done\"", status)
	}
}

func TestStatusInfographic_ReturnsNone(t *testing.T) {
	browser := &mockBrowser{
		findTabFound:     true,
		findTabResult:    1,
		currentURL:       "https://notebooklm.google.com/notebook/test",
		infographicCards: 0,
	}
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))
	service := NewService(browser, "https://notebooklm.google.com/notebook/test", store)
	service.sleep = noSleep

	status, err := service.StatusInfographic()
	if err != nil {
		t.Fatalf("StatusInfographic() error = %v", err)
	}
	if status != "none" {
		t.Errorf("StatusInfographic() = %q, want \"none\"", status)
	}
}

func TestListSources_ReturnsSourceNames(t *testing.T) {
	browser := &mockBrowser{
		findTabFound:  true,
		findTabResult: 1,
		currentURL:    "https://notebooklm.google.com/notebook/test",
		sourceNames:   []string{"ソース1", "ソース2"},
	}
	tmpDir := t.TempDir()
	store := NewMappingStore(filepath.Join(tmpDir, "mapping.yaml"))
	service := NewService(browser, "https://notebooklm.google.com/notebook/test", store)
	service.sleep = noSleep

	names, err := service.ListSources()
	if err != nil {
		t.Fatalf("ListSources() error = %v", err)
	}
	if len(names) != 2 {
		t.Errorf("ListSources() returned %d names, want 2", len(names))
	}
}
