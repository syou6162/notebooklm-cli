package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const defaultTimeout = 30 * time.Second

// BrowserError はNotebookLM Chrome操作エラー
type BrowserError struct {
	Message string
}

func (e *BrowserError) Error() string {
	return e.Message
}

// Client は低レベルChrome操作クライアント
type Client struct {
	tabIndex int
}

// NewClient は新しいClientを作成する
func NewClient(tabIndex int) *Client {
	return &Client{tabIndex: tabIndex}
}

func (c *Client) runAppleScript(script string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "osascript", "-")
	cmd.Stdin = strings.NewReader(script)

	out, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		return "", &BrowserError{Message: fmt.Sprintf("AppleScriptがタイムアウトしました（%v）", timeout)}
	}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := strings.TrimRight(string(exitErr.Stderr), "\n")
			if stderr != "" {
				return "", &BrowserError{Message: stderr}
			}
			return "", &BrowserError{Message: fmt.Sprintf("AppleScript失敗: exit code=%d", exitErr.ExitCode())}
		}
		return "", &BrowserError{Message: err.Error()}
	}

	return strings.TrimRight(string(out), "\n"), nil
}

func (c *Client) executeJS(jsCode string, timeout time.Duration) (string, error) {
	jsOneLine := strings.Join(strings.Fields(jsCode), " ")
	script := fmt.Sprintf(`tell application "Google Chrome"
  tell tab %d of front window
    set theResult to execute javascript "%s"
  end tell
end tell
if theResult is missing value then
  return ""
else
  return theResult
end if`, c.tabIndex, jsOneLine)
	return c.runAppleScript(script, timeout)
}

func (c *Client) activateChrome() error {
	script := fmt.Sprintf(`tell application "Google Chrome"
  activate
  set active tab index of front window to %d
end tell`, c.tabIndex)
	_, err := c.runAppleScript(script, defaultTimeout)
	return err
}

func (c *Client) clipboardPaste(text string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pbcopy")
	cmd.Stdin = strings.NewReader(text)
	if err := cmd.Run(); err != nil {
		return &BrowserError{Message: fmt.Sprintf("クリップボードへのコピーに失敗しました: %v", err)}
	}

	pasteScript := `tell application "Google Chrome"
  activate
end tell
delay 0.3
tell application "System Events"
  keystroke "v" using command down
end tell`
	_, err := c.runAppleScript(pasteScript, defaultTimeout)
	return err
}

// FindNotebookTab はChromeでNotebookLMのタブを検索してインデックスを返す
func (c *Client) FindNotebookTab() (int, bool) {
	countScript := `tell application "Google Chrome"
  return count of tabs of front window
end tell`
	countStr, err := c.runAppleScript(countScript, defaultTimeout)
	if err != nil {
		return 0, false
	}

	var tabCount int
	if _, err := fmt.Sscanf(countStr, "%d", &tabCount); err != nil {
		return 0, false
	}

	for i := 1; i <= tabCount; i++ {
		urlScript := fmt.Sprintf(`tell application "Google Chrome"
  return URL of tab %d of front window
end tell`, i)
		url, err := c.runAppleScript(urlScript, defaultTimeout)
		if err != nil {
			continue
		}
		if strings.Contains(url, notebookLMHost) {
			return i, true
		}
	}

	return 0, false
}

// NavigateTo は現在のタブを指定URLにナビゲートする
func (c *Client) NavigateTo(url string) error {
	safeURL := strings.ReplaceAll(url, `"`, "")
	script := fmt.Sprintf(`tell application "Google Chrome"
  set URL of tab %d of front window to "%s"
end tell`, c.tabIndex, safeURL)
	_, err := c.runAppleScript(script, defaultTimeout)
	return err
}

// GetCurrentURL は現在のタブのURLを取得する
func (c *Client) GetCurrentURL() (string, error) {
	script := fmt.Sprintf(`tell application "Google Chrome"
  return URL of tab %d of front window
end tell`, c.tabIndex)
	return c.runAppleScript(script, defaultTimeout)
}

// ClickButton はaria-labelでボタンをクリックする
func (c *Client) ClickButton(ariaLabel string) error {
	safeLabel := strings.ReplaceAll(ariaLabel, "'", "\\'")
	jsCode := fmt.Sprintf(`(() => {
		var els = document.querySelectorAll('[aria-label]');
		for (var i = 0; i < els.length; i++) {
			if ((els[i].getAttribute('aria-label') || '') === '%s') {
				els[i].click(); return 'CLICKED';
			}
		}
		return 'NOT_FOUND';
	})()`, safeLabel)
	result, err := c.executeJS(jsCode, defaultTimeout)
	if err != nil {
		return err
	}
	if result != "CLICKED" {
		return &BrowserError{Message: fmt.Sprintf("ボタンが見つかりませんでした: aria-label='%s'", ariaLabel)}
	}
	return nil
}

// ClickButtonByText はテキストでボタンをクリックする（CDKオーバーレイ優先）
func (c *Client) ClickButtonByText(text string) error {
	safeText := strings.ReplaceAll(text, "'", "\\'")
	jsCode := fmt.Sprintf(`(() => {
		var roots = [document.querySelector('.cdk-overlay-container'), document.body];
		for (var r = 0; r < roots.length; r++) {
			if (!roots[r]) continue;
			var buttons = roots[r].querySelectorAll('button');
			for (var i = 0; i < buttons.length; i++) {
				if ((buttons[i].textContent || '').indexOf('%s') !== -1) {
					buttons[i].click(); return 'CLICKED';
				}
			}
		}
		return 'NOT_FOUND';
	})()`, safeText)
	result, err := c.executeJS(jsCode, defaultTimeout)
	if err != nil {
		return err
	}
	if result != "CLICKED" {
		return &BrowserError{Message: fmt.Sprintf("ボタンが見つかりませんでした: text='%s'", text)}
	}
	return nil
}

// ClickMenuItem はCDKオーバーレイのメニュー項目をテキストでクリックする
func (c *Client) ClickMenuItem(text string) error {
	safeText := strings.ReplaceAll(text, "'", "\\'")
	jsCode := fmt.Sprintf(`(() => {
		var overlay = document.querySelector('.cdk-overlay-container');
		if (!overlay) return 'NO_OVERLAY';
		var items = overlay.querySelectorAll('[role=menuitem]');
		for (var i = 0; i < items.length; i++) {
			if (items[i].textContent.indexOf('%s') !== -1) {
				items[i].click(); return 'CLICKED';
			}
		}
		return 'NOT_FOUND';
	})()`, safeText)
	result, err := c.executeJS(jsCode, defaultTimeout)
	if err != nil {
		return err
	}
	if result != "CLICKED" {
		return &BrowserError{Message: fmt.Sprintf("メニュー項目が見つかりませんでした: text='%s' (result=%s)", text, result)}
	}
	return nil
}

// ClickOverlayButton はCDKオーバーレイのボタンをテキストで完全一致クリックする
func (c *Client) ClickOverlayButton(text string) error {
	safeText := strings.ReplaceAll(text, "'", "\\'")
	jsCode := fmt.Sprintf(`(() => {
		var overlay = document.querySelector('.cdk-overlay-container');
		if (!overlay) return 'NO_OVERLAY';
		var buttons = overlay.querySelectorAll('button');
		for (var i = 0; i < buttons.length; i++) {
			var t = (buttons[i].textContent || '').trim();
			if (t === '%s') {
				buttons[i].click(); return 'CLICKED';
			}
		}
		return 'NOT_FOUND';
	})()`, safeText)
	result, err := c.executeJS(jsCode, defaultTimeout)
	if err != nil {
		return err
	}
	if result != "CLICKED" {
		return &BrowserError{Message: fmt.Sprintf("オーバーレイボタンが見つかりませんでした: text='%s' (result=%s)", text, result)}
	}
	return nil
}

// GetElementCount はCSSセレクターに一致する要素数を返す
func (c *Client) GetElementCount(selector string) int {
	safeSelector := strings.ReplaceAll(selector, "'", "\\'")
	jsCode := fmt.Sprintf(`(() => {
		var section = document.querySelector('section');
		if (!section) return '0';
		var els = section.querySelectorAll('%s');
		return '' + els.length;
	})()`, safeSelector)
	result, err := c.executeJS(jsCode, defaultTimeout)
	if err != nil {
		return 0
	}
	var count int
	if _, err := fmt.Sscanf(result, "%d", &count); err != nil {
		return 0
	}
	return count
}

// FocusElement はCSSセレクターで要素にフォーカスする
func (c *Client) FocusElement(selector string) error {
	safeSelector := strings.ReplaceAll(selector, "'", "\\'")
	jsCode := fmt.Sprintf(`(() => {
		var el = document.querySelector('%s');
		if (!el) return 'NOT_FOUND';
		el.focus(); return 'FOCUSED';
	})()`, safeSelector)
	result, err := c.executeJS(jsCode, defaultTimeout)
	if err != nil {
		return err
	}
	if result != "FOCUSED" {
		return &BrowserError{Message: fmt.Sprintf("要素が見つかりませんでした: selector='%s'", selector)}
	}
	return nil
}

// PageContainsText はページに指定テキストが含まれるか確認する
func (c *Client) PageContainsText(text string) bool {
	safeText := strings.ReplaceAll(text, "'", "\\'")
	jsCode := fmt.Sprintf(`(() => {
		var body = document.body;
		if (!body) return 'false';
		return body.innerText.indexOf('%s') !== -1 ? 'true' : 'false';
	})()`, safeText)
	result, err := c.executeJS(jsCode, defaultTimeout)
	if err != nil {
		return false
	}
	return result == "true"
}

// OpenURLInNewTab は新タブでURLを開いてタブインデックスを返す
func (c *Client) OpenURLInNewTab(url string) (int, error) {
	safeURL := strings.ReplaceAll(url, `"`, "")
	script := fmt.Sprintf(`tell application "Google Chrome"
  if (count windows) = 0 then
    make new document with properties {URL:"%s"}
  else
    tell front window
      make new tab with properties {URL:"%s"}
    end tell
  end if
  return count of tabs of front window
end tell`, safeURL, safeURL)
	result, err := c.runAppleScript(script, defaultTimeout)
	if err != nil {
		return 0, err
	}
	var idx int
	if _, err := fmt.Sscanf(result, "%d", &idx); err != nil {
		return 0, &BrowserError{Message: fmt.Sprintf("新タブのインデックスを取得できませんでした: result='%s'", result)}
	}
	return idx, nil
}

// CountInfographicCards はStudioパネルのインフォグラフィックカード数を取得する
func (c *Client) CountInfographicCards() int {
	jsCode := fmt.Sprintf(`(() => {
		var buttons = document.querySelectorAll('button[aria-description]');
		var count = 0;
		for (var i = 0; i < buttons.length; i++) {
			var desc = buttons[i].getAttribute('aria-description') || '';
			if (desc === '%s') count++;
		}
		return '' + count;
	})()`, infographicCardDescription)
	result, err := c.executeJS(jsCode, defaultTimeout)
	if err != nil {
		return 0
	}
	var count int
	if _, err := fmt.Sscanf(result, "%d", &count); err != nil {
		return 0
	}
	return count
}

// ClickMoreButtonOnFirstInfographicCard は最初のインフォグラフィックカードの「もっと見る」をクリックする
func (c *Client) ClickMoreButtonOnFirstInfographicCard() error {
	jsCode := fmt.Sprintf(`(() => {
		var cards = document.querySelectorAll('button[aria-description]');
		var card = null;
		for (var i = 0; i < cards.length; i++) {
			if (cards[i].getAttribute('aria-description') === '%s') { card = cards[i]; break; }
		}
		if (!card) return 'NO_CARD';
		var ancestor = card.parentElement;
		for (var d = 0; d < 5 && ancestor; d++) {
			var buttons = ancestor.querySelectorAll('button[mattooltip]');
			for (var j = 0; j < buttons.length; j++) {
				if (buttons[j].getAttribute('mattooltip') === '%s') {
					buttons[j].click(); return 'CLICKED';
				}
			}
			ancestor = ancestor.parentElement;
		}
		return 'NOT_FOUND';
	})()`, infographicCardDescription, moreButtonDescription)
	result, err := c.executeJS(jsCode, defaultTimeout)
	if err != nil {
		return err
	}
	if result != "CLICKED" {
		return &BrowserError{Message: fmt.Sprintf("インフォグラフィックの「もっと見る」ボタンが見つかりませんでした (result=%s)", result)}
	}
	return nil
}

// CloseSourceViewerIfOpen はソースビューアが開いている場合に閉じる（冪等）
func (c *Client) CloseSourceViewerIfOpen() {
	jsCode := fmt.Sprintf(`(() => {
		var buttons = document.querySelectorAll('button');
		for (var i = 0; i < buttons.length; i++) {
			var tip = buttons[i].getAttribute('mattooltip') || '';
			if (tip === '%s') {
				buttons[i].click(); return 'CLICKED';
			}
		}
		return 'NOT_FOUND';
	})()`, sourceViewerCloseTooltip)
	_, _ = c.executeJS(jsCode, defaultTimeout)
}

// ListSourceNames はソースパネルからソース名一覧を取得する
func (c *Client) ListSourceNames() ([]string, error) {
	jsCode := fmt.Sprintf(`(() => {
		var checkboxes = document.querySelectorAll('mat-checkbox');
		var names = [];
		for (var i = 0; i < checkboxes.length; i++) {
			var input = checkboxes[i].querySelector('input');
			if (!input) continue;
			var label = input.getAttribute('aria-label') || '';
			if (label && label !== '%s') {
				names.push(label);
			}
		}
		return names.join('\\n');
	})()`, selectAllSourcesLabel)
	result, err := c.executeJS(jsCode, defaultTimeout)
	if err != nil {
		return nil, err
	}
	if result == "" {
		return nil, nil
	}
	var names []string
	for _, name := range strings.Split(result, "\n") {
		if name != "" {
			names = append(names, name)
		}
	}
	return names, nil
}
