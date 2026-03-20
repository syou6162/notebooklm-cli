package main

// NotebookLM DOM セレクター定数
//
// NotebookLMはAngular Material UIを使用しているため、
// aria-label・テキスト内容ベースのセレクターを優先する。
// UIが変更された場合はこのファイルのみ修正すること。

const (
	// NotebookLMホスト
	notebookLMHost = "notebooklm.google.com"

	// ソースパネル
	sourceAddButtonAria      = "ソースを追加"
	sourceViewerCloseTooltip = "ソース表示を閉じる"
	pasteTextIcon            = "content_paste"
	textareaClass            = "copied-text-input-textarea"
	insertButtonText         = "挿入"

	// メニュー操作
	moreButtonAria = "もっと見る"

	// ソース削除
	deleteMenuText    = "削除"
	confirmDeleteText = "削除"

	// ソースカウント
	sourceCheckbox        = "mat-checkbox"
	selectAllSourcesLabel = "すべてのソースを選択"

	// ノートブック作成
	notebookLMHomeURL        = "https://notebooklm.google.com/"
	createNotebookButtonText = "ノートブックを新規作成"

	// インフォグラフィック生成
	infographicButtonAria      = "インフォグラフィック"
	infographicCardDescription = "インフォグラフィック"
	moreButtonDescription      = "もっと見る"
	generatingText             = "インフォグラフィックを生成しています"

	// ダウンロード
	downloadMenuText = "ダウンロード"
)
