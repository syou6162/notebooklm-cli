package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v3"
)

var notebookURLFlag = &cli.StringFlag{
	Name:     "notebook-url",
	Usage:    "操作対象のノートブックURL",
	Required: true,
}

func main() {
	xdg := NewXDGPaths()

	app := &cli.Command{
		Name:  "notebooklm-cli",
		Usage: "Google NotebookLMのブラウザ操作を自動化するmacOS専用CLIツール",
		Commands: []*cli.Command{
			{
				Name:  "add",
				Usage: "リソースを追加する",
				Commands: []*cli.Command{
					{
						Name:  "source",
						Usage: "stdinからテキストを読み取りNotebookLMにソースとして追加する",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "notebook-url",
								Usage: "ソースを追加するノートブックのURL",
							},
						},
						Action: func(_ context.Context, cmd *cli.Command) error {
							return addSourceAction(xdg, os.Stdin, cmd.String("notebook-url"))
						},
					},
				},
			},
			{
				Name:  "delete",
				Usage: "リソースを削除する",
				Commands: []*cli.Command{
					{
						Name:  "source",
						Usage: "ノートブック内の全ソースを削除する",
						Flags: []cli.Flag{notebookURLFlag},
						Action: func(_ context.Context, cmd *cli.Command) error {
							return deleteSourceAction(xdg, cmd.String("notebook-url"))
						},
					},
				},
			},
			{
				Name:  "list",
				Usage: "リソースの一覧を表示する",
				Commands: []*cli.Command{
					{
						Name:  "source",
						Usage: "ノートブック内のソース一覧を表示する",
						Flags: []cli.Flag{notebookURLFlag},
						Action: func(_ context.Context, cmd *cli.Command) error {
							return listSourceAction(cmd.String("notebook-url"))
						},
					},
				},
			},
			{
				Name:  "resolve",
				Usage: "マッピングからリソースを解決する",
				Commands: []*cli.Command{
					{
						Name:  "source",
						Usage: "stdinからテキストを読み取りSHA256マッピングでノートブックURLを返す",
						Action: func(_ context.Context, _ *cli.Command) error {
							return resolveSourceAction(xdg, os.Stdin)
						},
					},
				},
			},
			{
				Name:  "generate",
				Usage: "生成物を作成する",
				Commands: []*cli.Command{
					{
						Name:  "infographic",
						Usage: "インフォグラフィックの生成を開始する（即返し）",
						Flags: []cli.Flag{notebookURLFlag},
						Action: func(_ context.Context, cmd *cli.Command) error {
							return generateInfographicAction(cmd.String("notebook-url"))
						},
					},
					{
						Name:  "audio",
						Usage: "音声解説の生成を開始する（即返し）",
						Flags: []cli.Flag{notebookURLFlag},
						Action: func(_ context.Context, cmd *cli.Command) error {
							return generateAudioAction(cmd.String("notebook-url"))
						},
					},
				},
			},
			{
				Name:  "status",
				Usage: "生成状態を確認する",
				Commands: []*cli.Command{
					{
						Name:  "infographic",
						Usage: "インフォグラフィックの生成状態を確認する",
						Flags: []cli.Flag{notebookURLFlag},
						Action: func(_ context.Context, cmd *cli.Command) error {
							return statusInfographicAction(cmd.String("notebook-url"))
						},
					},
					{
						Name:  "audio",
						Usage: "音声解説の生成状態を確認する",
						Flags: []cli.Flag{notebookURLFlag},
						Action: func(_ context.Context, cmd *cli.Command) error {
							return statusAudioAction(cmd.String("notebook-url"))
						},
					},
				},
			},
			{
				Name:  "download",
				Usage: "生成物をダウンロードする",
				Commands: []*cli.Command{
					{
						Name:  "infographic",
						Usage: "インフォグラフィックをダウンロードする",
						Flags: []cli.Flag{
							notebookURLFlag,
							&cli.StringFlag{
								Name:  "output",
								Usage: "出力先ディレクトリ（未指定時はconfigの値を使用）",
							},
						},
						Action: func(_ context.Context, cmd *cli.Command) error {
							return downloadInfographicAction(xdg, cmd.String("notebook-url"), cmd.String("output"))
						},
					},
					{
						Name:  "audio",
						Usage: "音声解説をダウンロードする",
						Flags: []cli.Flag{
							notebookURLFlag,
							&cli.StringFlag{
								Name:  "output",
								Usage: "出力先ディレクトリ（未指定時はconfigの値を使用）",
							},
						},
						Action: func(_ context.Context, cmd *cli.Command) error {
							return downloadAudioAction(xdg, cmd.String("notebook-url"), cmd.String("output"))
						},
					},
				},
			},
			{
				Name:  "batch",
				Usage: "stdin→download一気通貫で実行する",
				Commands: []*cli.Command{
					{
						Name:  "infographic",
						Usage: "stdinからテキストを読み取り、ソース追加→インフォグラフィック生成→ダウンロードを一気通貫で実行する",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "output",
								Usage: "出力先ディレクトリ（未指定時はconfigの値を使用）",
							},
						},
						Action: func(_ context.Context, cmd *cli.Command) error {
							return batchInfographicAction(xdg, os.Stdin, cmd.String("output"))
						},
					},
					{
						Name:  "audio",
						Usage: "stdinからテキストを読み取り、ソース追加→音声解説生成→ダウンロードを一気通貫で実行する",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "output",
								Usage: "出力先ディレクトリ（未指定時はconfigの値を使用）",
							},
						},
						Action: func(_ context.Context, cmd *cli.Command) error {
							return batchAudioAction(xdg, os.Stdin, cmd.String("output"))
						},
					},
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		os.Exit(1)
	}
}

func addSourceAction(xdg *XDGPaths, reader io.Reader, notebookURL string) error {
	if err := xdg.EnsureDirectories(); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
	}

	text, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("stdinの読み取りに失敗しました: %w", err)
	}

	mapping := NewMappingStore(xdg.MappingFile())
	client := NewClient(1)
	service := NewService(client, notebookURL, mapping, NewClaudeMetadataGenerator())

	return service.AddSource(string(text))
}

func listSourceAction(notebookURL string) error {
	client := NewClient(1)
	mapping := NewMappingStore("")
	service := NewService(client, notebookURL, mapping, nil)

	names, err := service.ListSources()
	if err != nil {
		return err
	}

	for _, name := range names {
		fmt.Println(name)
	}
	return nil
}

func deleteSourceAction(xdg *XDGPaths, notebookURL string) error {
	client := NewClient(1)
	mapping := NewMappingStore(xdg.MappingFile())
	service := NewService(client, notebookURL, mapping, nil)

	return service.DeleteSource()
}

func resolveSourceAction(xdg *XDGPaths, reader io.Reader) error {
	text, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("stdinの読み取りに失敗しました: %w", err)
	}

	mapping := NewMappingStore(xdg.MappingFile())
	url, err := ResolveSource(string(text), mapping)
	if err != nil {
		return err
	}

	fmt.Println(url)
	return nil
}

func generateInfographicAction(notebookURL string) error {
	client := NewClient(1)
	mapping := NewMappingStore("")
	service := NewService(client, notebookURL, mapping, nil)

	return service.GenerateInfographic()
}

func statusInfographicAction(notebookURL string) error {
	client := NewClient(1)
	mapping := NewMappingStore("")
	service := NewService(client, notebookURL, mapping, nil)

	status, err := service.StatusInfographic()
	if err != nil {
		return err
	}

	fmt.Println(status)
	return nil
}

func downloadInfographicAction(xdg *XDGPaths, notebookURL, outputFlag string) error {
	cfg, err := LoadConfig(xdg.ConfigFile())
	if err != nil {
		return fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
	}

	outputDir := cfg.ResolveDownloadDir("infographic", outputFlag)
	if outputDir == "" {
		return fmt.Errorf("出力先が指定されていません。--outputフラグまたはconfigのdownloads.infographicを設定してください")
	}

	mapping := NewMappingStore(xdg.MappingFile())

	// マッピングからtitleを取得
	entry, hash, found := mapping.LookupByURL(notebookURL)
	var title string
	if found {
		title = entry.Title
	}

	client := NewClient(1)
	service := NewService(client, notebookURL, mapping, nil)

	if err := service.DownloadInfographic(); err != nil {
		return err
	}

	// ダウンロード完了を待機してファイルを検出
	homeDir, _ := os.UserHomeDir()
	downloadsDir := filepath.Join(homeDir, "Downloads")
	startTime := time.Now().Add(-30 * time.Second)

	var downloaded string
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		f, err := FindDownloadedInfographic(downloadsDir, startTime, 1_000_000)
		if err == nil {
			downloaded = f
			break
		}
		time.Sleep(3 * time.Second)
	}

	if downloaded == "" {
		return fmt.Errorf("ダウンロードがタイムアウトしました: %s/unnamed*.png", downloadsDir)
	}

	dest, err := MoveToOutput(downloaded, outputDir, title)
	if err != nil {
		return fmt.Errorf("ファイルの移動に失敗しました: %w", err)
	}

	// マッピングにダウンロード済みパスを記録
	if found {
		if err := mapping.UpdateDownload(hash, "infographic", dest); err != nil {
			fmt.Printf("マッピングの更新に失敗しました: %v\n", err)
		}
	}

	fmt.Println(dest)
	return nil
}

func generateAudioAction(notebookURL string) error {
	client := NewClient(1)
	mapping := NewMappingStore("")
	service := NewService(client, notebookURL, mapping, nil)

	return service.GenerateAudio()
}

func statusAudioAction(notebookURL string) error {
	client := NewClient(1)
	mapping := NewMappingStore("")
	service := NewService(client, notebookURL, mapping, nil)

	status, err := service.StatusAudio()
	if err != nil {
		return err
	}

	fmt.Println(status)
	return nil
}

func downloadAudioAction(xdg *XDGPaths, notebookURL, outputFlag string) error {
	cfg, err := LoadConfig(xdg.ConfigFile())
	if err != nil {
		return fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
	}

	outputDir := cfg.ResolveDownloadDir("audio", outputFlag)
	if outputDir == "" {
		return fmt.Errorf("出力先が指定されていません。--outputフラグまたはconfigのdownloads.audioを設定してください")
	}

	mapping := NewMappingStore(xdg.MappingFile())

	entry, hash, found := mapping.LookupByURL(notebookURL)
	var title string
	if found {
		title = entry.Title
	}

	client := NewClient(1)
	service := NewService(client, notebookURL, mapping, nil)

	if err := service.DownloadAudio(); err != nil {
		return err
	}

	homeDir, _ := os.UserHomeDir()
	downloadsDir := filepath.Join(homeDir, "Downloads")
	startTime := time.Now().Add(-30 * time.Second)

	var downloaded string
	deadline := time.Now().Add(120 * time.Second)
	for time.Now().Before(deadline) {
		f, err := FindDownloadedAudio(downloadsDir, startTime, 1_000_000)
		if err == nil {
			downloaded = f
			break
		}
		time.Sleep(3 * time.Second)
	}

	if downloaded == "" {
		return fmt.Errorf("ダウンロードがタイムアウトしました: %s/*.m4a", downloadsDir)
	}

	dest, err := MoveToOutput(downloaded, outputDir, title)
	if err != nil {
		return fmt.Errorf("ファイルの移動に失敗しました: %w", err)
	}

	if found {
		if err := mapping.UpdateDownload(hash, "audio", dest); err != nil {
			fmt.Printf("マッピングの更新に失敗しました: %v\n", err)
		}
	}

	fmt.Println(dest)
	return nil
}

func batchInfographicAction(xdg *XDGPaths, reader io.Reader, outputFlag string) error {
	cfg, err := LoadConfig(xdg.ConfigFile())
	if err != nil {
		return fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
	}
	outputDir := cfg.ResolveDownloadDir("infographic", outputFlag)
	if outputDir == "" {
		return fmt.Errorf("出力先が指定されていません。--outputフラグまたはconfigのdownloads.infographicを設定してください")
	}

	if err := xdg.EnsureDirectories(); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
	}
	text, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("stdinの読み取りに失敗しました: %w", err)
	}

	mapping := NewMappingStore(xdg.MappingFile())
	client := NewClient(1)
	service := NewService(client, "", mapping, NewClaudeMetadataGenerator())

	fmt.Println("[1/4] ソースを追加中...")
	if err := service.AddSource(string(text)); err != nil {
		return err
	}

	hash := ComputeSHA256(string(text))
	entry, found := mapping.LookupEntry(hash)
	if !found || entry.URL == "" {
		return fmt.Errorf("マッピングからノートブックURLを取得できませんでした")
	}
	service.notebookURL = entry.URL

	fmt.Println("[2/4] インフォグラフィックを生成中...")
	if err := service.GenerateInfographic(); err != nil {
		return err
	}

	fmt.Print("[3/4] 生成完了を待機中...")
	for {
		status, err := service.StatusInfographic()
		if err != nil {
			return err
		}
		if status == "done" {
			fmt.Println(" 完了")
			break
		}
		fmt.Print(".")
		time.Sleep(5 * time.Second)
	}

	fmt.Println("[4/4] ダウンロード中...")
	if err := service.DownloadInfographic(); err != nil {
		return err
	}

	homeDir, _ := os.UserHomeDir()
	downloadsDir := filepath.Join(homeDir, "Downloads")
	startTime := time.Now().Add(-30 * time.Second)

	var downloaded string
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		f, err := FindDownloadedInfographic(downloadsDir, startTime, 1_000_000)
		if err == nil {
			downloaded = f
			break
		}
		time.Sleep(3 * time.Second)
	}
	if downloaded == "" {
		return fmt.Errorf("ダウンロードがタイムアウトしました")
	}

	dest, err := MoveToOutput(downloaded, outputDir, entry.Title)
	if err != nil {
		return fmt.Errorf("ファイルの移動に失敗しました: %w", err)
	}

	if err := mapping.UpdateDownload(hash, "infographic", dest); err != nil {
		fmt.Printf("マッピングの更新に失敗しました: %v\n", err)
	}

	fmt.Println(dest)
	return nil
}

func batchAudioAction(xdg *XDGPaths, reader io.Reader, outputFlag string) error {
	cfg, err := LoadConfig(xdg.ConfigFile())
	if err != nil {
		return fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
	}
	outputDir := cfg.ResolveDownloadDir("audio", outputFlag)
	if outputDir == "" {
		return fmt.Errorf("出力先が指定されていません。--outputフラグまたはconfigのdownloads.audioを設定してください")
	}

	if err := xdg.EnsureDirectories(); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
	}
	text, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("stdinの読み取りに失敗しました: %w", err)
	}

	mapping := NewMappingStore(xdg.MappingFile())
	client := NewClient(1)
	service := NewService(client, "", mapping, NewClaudeMetadataGenerator())

	fmt.Println("[1/4] ソースを追加中...")
	if err := service.AddSource(string(text)); err != nil {
		return err
	}

	hash := ComputeSHA256(string(text))
	entry, found := mapping.LookupEntry(hash)
	if !found || entry.URL == "" {
		return fmt.Errorf("マッピングからノートブックURLを取得できませんでした")
	}
	service.notebookURL = entry.URL

	fmt.Println("[2/4] 音声解説を生成中...")
	if err := service.GenerateAudio(); err != nil {
		return err
	}

	fmt.Print("[3/4] 生成完了を待機中...")
	for {
		status, err := service.StatusAudio()
		if err != nil {
			return err
		}
		if status == "done" {
			fmt.Println(" 完了")
			break
		}
		fmt.Print(".")
		time.Sleep(10 * time.Second)
	}

	fmt.Println("[4/4] ダウンロード中...")
	if err := service.DownloadAudio(); err != nil {
		return err
	}

	homeDir, _ := os.UserHomeDir()
	downloadsDir := filepath.Join(homeDir, "Downloads")
	startTime := time.Now().Add(-30 * time.Second)

	var downloaded string
	deadline := time.Now().Add(120 * time.Second)
	for time.Now().Before(deadline) {
		f, err := FindDownloadedAudio(downloadsDir, startTime, 1_000_000)
		if err == nil {
			downloaded = f
			break
		}
		time.Sleep(3 * time.Second)
	}
	if downloaded == "" {
		return fmt.Errorf("ダウンロードがタイムアウトしました")
	}

	dest, err := MoveToOutput(downloaded, outputDir, entry.Title)
	if err != nil {
		return fmt.Errorf("ファイルの移動に失敗しました: %w", err)
	}

	if err := mapping.UpdateDownload(hash, "audio", dest); err != nil {
		fmt.Printf("マッピングの更新に失敗しました: %v\n", err)
	}

	fmt.Println(dest)
	return nil
}
