package main

import "time"

// Browser はChrome操作のインターフェース（テスト用モック作成に使用）
type Browser interface {
	FindNotebookTab() (int, bool)
	OpenURLInNewTab(url string) (int, error)
	NavigateTo(url string) error
	GetCurrentURL() (string, error)
	ClickButton(ariaLabel string) error
	ClickButtonByText(text string) error
	ClickMenuItem(text string) error
	ClickOverlayButton(text string) error
	GetElementCount(selector string) int
	FocusElement(selector string) error
	PageContainsText(text string) bool
	CloseSourceViewerIfOpen()
	CountInfographicCards() int
	ClickMoreButtonOnFirstInfographicCard() error
	ListSourceNames() ([]string, error)

	// 内部メソッド（Serviceから呼ばれる）
	SetTabIndex(index int)
	ActivateChrome() error
	ClipboardPaste(text string) error
}

// compile-time check
var _ Browser = (*Client)(nil)

// SetTabIndex はtabIndexを設定する
func (c *Client) SetTabIndex(index int) {
	c.tabIndex = index
}

// ActivateChrome はChromeをアクティブにする（Browserインターフェース用エクスポート）
func (c *Client) ActivateChrome() error {
	return c.activateChrome()
}

// ClipboardPaste はクリップボード経由でテキストを貼り付ける（Browserインターフェース用エクスポート）
func (c *Client) ClipboardPaste(text string) error {
	return c.clipboardPaste(text)
}

// Sleeper は時間待機の抽象化（テスト時に差し替え可能）
type Sleeper func(d time.Duration)

// RealSleep は実際のtime.Sleepを使うSleeper
func RealSleep(d time.Duration) {
	time.Sleep(d)
}
