# notebooklm-cli

## プロジェクト概要

Google NotebookLMのブラウザ操作を自動化するmacOS専用CLIツール。
AppleScript + Chrome JavaScript実行でNotebookLMを操作する。

## システム要件

- **対応OS**: macOS専用（AppleScript、osascript、pbcopyを使用）
- **ブラウザ**: Google Chrome必須（Googleアカウントでログイン済み）
- **Goバージョン**: 1.24以上
- **Claude CLI**: `claude -p`コマンド（タイトル・要約の自動生成に使用）

## 開発コマンド

```bash
# ビルド
go build -o notebooklm-cli

# テスト実行
go test -v -race ./...

# リント（pre-commit経由）
pre-commit run --all-files

# 単一テストの実行
go test -v -run TestFunctionName
```

## レイアウト

フラットレイアウト（全てのGoファイルをルートディレクトリに配置、package main）。

## ファイル構成

| ファイル | 責務 |
|---|---|
| `main.go` | CLIエントリポイント、サブコマンド定義 |
| `actions_source.go` | add/delete/list/resolve sourceのアクション関数 |
| `actions_infographic.go` | generate/status/download infographicのアクション関数 |
| `actions_audio.go` | generate/status/download audioのアクション関数 |
| `actions_batch.go` | batch infographic/audioのアクション関数 |
| `actions_download.go` | ダウンロード後の共通処理（ファイル検出→リネーム→マッピング更新→Finderコメント） |
| `client.go` | AppleScript + JS Chrome操作の低レベルクライアント |
| `browser.go` | Browserインターフェース定義（テスト用モック作成に使用） |
| `service.go` | ワークフロー統合（ソース管理、インフォグラフィック、音声解説） |
| `selectors.go` | NotebookLM DOMセレクター定数（UIが変更された場合はここのみ修正） |
| `config.go` | XDG準拠の設定ファイル読み込み |
| `xdg.go` | XDG Base Directoryパス解決 |
| `mapping.go` | SHA256→ノートブックメタデータのマッピング管理 |
| `metadata.go` | claude -pによるタイトル・要約自動生成 |
| `resolve.go` | マッピングからノートブックURLを解決 |
| `helpers.go` | ダウンロード済みファイル検出、ファイル移動 |
| `finder_comment.go` | Finderコメントの読み書き（xattr + plist） |

## 開発方針

- **開発手法**: TDD（テスト駆動開発）- t_wada式
- **CLIフレームワーク**: `urfave/cli/v3`（2階層サブコマンド対応のため）
- **設定管理**: XDG Base Directory準拠（`~/.config/notebooklm-cli/config.yaml`）
- **データ保存**: `$XDG_DATA_HOME/notebooklm-cli/`（マッピングファイル）

## サブコマンド体系

「動詞 + 対象」パターン:

```bash
notebooklm-cli add source          # ソース追加（stdin）
notebooklm-cli delete source       # ソース削除
notebooklm-cli list source         # ソース一覧
notebooklm-cli resolve source      # マッピングからURL解決
notebooklm-cli generate infographic # インフォグラフィック生成
notebooklm-cli generate audio      # 音声解説生成
notebooklm-cli status infographic  # 生成状態確認
notebooklm-cli status audio        # 生成状態確認
notebooklm-cli download infographic # ダウンロード
notebooklm-cli download audio      # ダウンロード
notebooklm-cli batch infographic   # stdin→download一気通貫
notebooklm-cli batch audio         # stdin→download一気通貫
```

## 依存関係

- `github.com/urfave/cli/v3` — CLIフレームワーク
- `gopkg.in/yaml.v3` — YAML読み書き（マッピング、設定）
- `golang.org/x/sys` — macOS拡張属性（Finderコメント）
- `howett.net/plist` — バイナリplist（Finderコメント）

## 既知の制約

- macOS以外では動作しない（AppleScript依存）
- Google Chromeが必要（Googleアカウントでログイン済みであること）
- NotebookLMのUI変更でDOMセレクター（selectors.go）が壊れるリスクがある
- ブラウザ操作のため同時実行不可
- `claude -p`（Claude CLI）が必要（タイトル・要約の自動生成）
