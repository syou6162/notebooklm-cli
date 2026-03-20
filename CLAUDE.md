# notebooklm-cli

## プロジェクト概要

Google NotebookLMのブラウザ操作を自動化するmacOS専用CLIツール。
AppleScript + Chrome JavaScript実行でNotebookLMを操作する。

## システム要件

- **対応OS**: macOS専用（AppleScript、osascript、pbcopyを使用）
- **ブラウザ**: Google Chrome必須（Googleアカウントでログイン済み）
- **Goバージョン**: 1.24以上

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

## 開発方針

- **開発手法**: TDD（テスト駆動開発）- t_wada式
- **CLIフレームワーク**: `urfave/cli/v3`（2階層サブコマンド対応のため）
- **設定管理**: XDG Base Directory準拠（`~/.config/notebooklm-cli/config.yaml`）
- **データ保存**: `$XDG_DATA_HOME/notebooklm-cli/`（マッピングファイル等）

## サブコマンド体系

「動詞 + 対象」パターン:

```bash
notebooklm-cli add source          # ソース追加（stdin）
notebooklm-cli delete source       # ソース削除
notebooklm-cli list source         # ソース一覧
notebooklm-cli rename source       # ソースリネーム
notebooklm-cli list notebook       # ノートブック一覧
notebooklm-cli rename notebook     # ノートブックタイトル変更
notebooklm-cli generate infographic # インフォグラフィック生成
notebooklm-cli status infographic  # 生成状態確認
notebooklm-cli download infographic # ダウンロード
notebooklm-cli generate audio      # 音声解説生成
notebooklm-cli status audio        # 生成状態確認
notebooklm-cli download audio      # ダウンロード
```

## 既知の制約

- macOS以外では動作しない（AppleScript依存）
- Google Chromeが必要（Googleアカウントでログイン済みであること）
- NotebookLMのUI変更でDOMセレクター（selectors.go）が壊れるリスクがある
- ブラウザ操作のため同時実行不可
