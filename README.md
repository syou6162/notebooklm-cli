# notebooklm-cli - Google NotebookLMのブラウザ操作を自動化するCLIツール

notebooklm-cliは、Google NotebookLMのブラウザ操作をAppleScript経由で自動化するmacOS専用CLIツールです。テキストを流し込むだけで、インフォグラフィックや音声解説（Audio Overview）の生成・ダウンロードまでを一気通貫で実行できます。

## クイックスタート

```bash
# 1. インストール
go install github.com/syou6162/notebooklm-cli@latest

# 2. テキストからインフォグラフィックを一気通貫で生成
cat article.txt | notebooklm-cli batch infographic --output ./images/

# 3. テキストから音声解説を一気通貫で生成
cat article.txt | notebooklm-cli batch audio
# → ~/.local/share/podself/episodes/ に保存（config設定済みの場合）
```

## 特徴

- 🤖 テキストを流すだけでNotebookLMの操作を完全自動化
- 📊 インフォグラフィック・音声解説の生成→ダウンロードまで一気通貫
- 🧠 `claude -p`によるタイトル・要約の自動生成
- 🎙️ [podself](https://github.com/syou6162/podself)連携（ファイル名=エピソードタイトル、Finderコメント=shownote）
- 📁 XDG Base Directory準拠の設定・データ管理
- 🔄 同一入力の重複防止（SHA256マッピング）

## 必要な環境

- **macOS専用**（AppleScript、osascript、pbcopyを使用）
- **Google Chrome**（Googleアカウントでログイン済み）
- **Go 1.24以上**
- **Claude CLI**（`claude -p`コマンド。タイトル・要約の自動生成に使用）

## インストール

### Go installを使用（推奨）

```bash
go install github.com/syou6162/notebooklm-cli@latest
```

### ソースからビルド

```bash
git clone https://github.com/syou6162/notebooklm-cli.git
cd notebooklm-cli
go build -o notebooklm-cli
```

## 初回セットアップ

### 設定ファイルの作成（オプション）

`--output`を毎回指定したくない場合は、設定ファイルでダウンロード先を指定します：

```bash
mkdir -p ~/.config/notebooklm-cli
cp config_example.yaml ~/.config/notebooklm-cli/config.yaml
vi ~/.config/notebooklm-cli/config.yaml
```

設定例：

```yaml
downloads:
  infographic: "~/Documents/notebooklm/infographics"
  audio: "~/.local/share/podself/episodes"
```

> **podself連携**: `downloads.audio`をpodselfの`episodes_dir`と同じにしておくと、ダウンロードした音声がそのままポッドキャストのエピソードになります。

## 使い方

### batch（一気通貫）

最も簡単な使い方です。テキストを流すだけで、ソース追加→生成→ダウンロードまでを自動実行します。

```bash
# インフォグラフィック生成
cat article.txt | notebooklm-cli batch infographic --output ./images/

# 音声解説生成（config設定済みなら--output省略可）
cat article.txt | notebooklm-cli batch audio
```

batchコマンドは以下の4ステップを自動実行します：

1. **ソース追加**: テキストをNotebookLMに追加（同一入力ならスキップ）
2. **生成**: インフォグラフィックまたは音声解説の生成を開始
3. **待機**: 生成完了をポーリングで待機
4. **ダウンロード**: 生成物をダウンロードし、タイトルでリネーム

### 個別コマンド

デバッグや細かい制御が必要な場合は、個別コマンドを使います。

#### ソース管理

```bash
# ソース追加（新規ノートブックが自動作成される）
cat article.txt | notebooklm-cli add source

# 既存ノートブックにソース追加
cat article.txt | notebooklm-cli add source --notebook-url "https://notebooklm.google.com/notebook/xxx"

# ソース一覧
notebooklm-cli list source --notebook-url "$URL"

# 全ソース削除
notebooklm-cli delete source --notebook-url "$URL"
```

#### マッピング解決

```bash
# テキストからノートブックURLを取得（ブラウザ操作なし）
URL=$(cat article.txt | notebooklm-cli resolve source)
```

#### インフォグラフィック

```bash
notebooklm-cli generate infographic --notebook-url "$URL"
notebooklm-cli status infographic --notebook-url "$URL"   # → generating / done / none
notebooklm-cli download infographic --notebook-url "$URL" --output ./images/
```

#### 音声解説

```bash
notebooklm-cli generate audio --notebook-url "$URL"
notebooklm-cli status audio --notebook-url "$URL"          # → generating / done / none
notebooklm-cli download audio --notebook-url "$URL" --output ./podcasts/
```

## podself連携

[podself](https://github.com/syou6162/podself)と組み合わせると、テキストからポッドキャスト配信までを自動化できます。

```bash
# 1. テキストから音声解説を生成・ダウンロード
cat article.txt | notebooklm-cli batch audio
# → ~/.local/share/podself/episodes/記事タイトル.m4a として保存
# → Finderコメント（shownote）が自動設定される

# 2. Google Driveにアップロード（手動）

# 3. RSSフィード生成
podself generate-rss

# 4. Overcastで聴く
```

notebooklm-cliが自動で行うこと：

- **ファイル名**: `claude -p`で生成したタイトルがファイル名になる → podselfのエピソードタイトル
- **Finderコメント**: `claude -p`で生成した要約がFinderコメントに設定される → podselfのshownote
- **重複防止**: 同一テキストなら同じノートブックを再利用

## マッピング

notebooklm-cliは入力テキストのSHA256ハッシュとノートブックの情報を紐付けて管理します。

保存先: `~/.local/share/notebooklm-cli/mapping.yaml`

```yaml
entries:
  a3f2b1c4...:
    url: "https://notebooklm.google.com/notebook/xxx"
    title: "データストレージ設計パターンの全体像"
    description: "本章では、データエンジニアリングにおける..."
    downloads:
      infographic: "/abs/path/データストレージ設計パターンの全体像.png"
      audio: "/abs/path/データストレージ設計パターンの全体像.m4a"
```

## コマンド一覧

| コマンド | 説明 |
|---|---|
| `batch infographic` | stdin→インフォグラフィック生成→ダウンロードを一気通貫 |
| `batch audio` | stdin→音声解説生成→ダウンロードを一気通貫 |
| `add source` | stdinからテキストを読み取りソース追加 |
| `delete source` | ノートブック内の全ソース削除 |
| `list source` | ソース一覧表示 |
| `resolve source` | テキストからノートブックURLを解決 |
| `generate infographic` | インフォグラフィック生成開始（即返し） |
| `generate audio` | 音声解説生成開始（即返し） |
| `status infographic` | インフォグラフィック生成状態確認 |
| `status audio` | 音声解説生成状態確認 |
| `download infographic` | インフォグラフィックダウンロード |
| `download audio` | 音声解説ダウンロード |

## 既知の制約

- macOS以外では動作しない（AppleScript依存）
- Google Chromeが必要（Googleアカウントでログイン済みであること）
- NotebookLMのUI変更でDOMセレクターが壊れるリスクがある
- ブラウザ操作のため同時実行不可
- `claude -p`（Claude CLI）が必要

## ライセンス

MIT License

## 作者

[@syou6162](https://github.com/syou6162)
