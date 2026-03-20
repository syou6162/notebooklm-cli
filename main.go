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
								Name:     "output",
								Usage:    "出力先ディレクトリ",
								Required: true,
							},
						},
						Action: func(_ context.Context, cmd *cli.Command) error {
							return downloadInfographicAction(cmd.String("notebook-url"), cmd.String("output"))
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
	service := NewService(client, notebookURL, mapping)
	service.metadataGen = NewClaudeMetadataGenerator()

	return service.AddSource(string(text))
}

func listSourceAction(notebookURL string) error {
	client := NewClient(1)
	mapping := NewMappingStore("")
	service := NewService(client, notebookURL, mapping)

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
	service := NewService(client, notebookURL, mapping)

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
	service := NewService(client, notebookURL, mapping)

	return service.GenerateInfographic()
}

func statusInfographicAction(notebookURL string) error {
	client := NewClient(1)
	mapping := NewMappingStore("")
	service := NewService(client, notebookURL, mapping)

	status, err := service.StatusInfographic()
	if err != nil {
		return err
	}

	fmt.Println(status)
	return nil
}

func downloadInfographicAction(notebookURL, outputDir string) error {
	client := NewClient(1)
	mapping := NewMappingStore("")
	service := NewService(client, notebookURL, mapping)

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
		found, err := FindDownloadedInfographic(downloadsDir, startTime, 1_000_000)
		if err == nil {
			downloaded = found
			break
		}
		time.Sleep(3 * time.Second)
	}

	if downloaded == "" {
		return fmt.Errorf("ダウンロードがタイムアウトしました: %s/unnamed*.png", downloadsDir)
	}

	dest, err := CopyToOutput(downloaded, outputDir)
	if err != nil {
		return fmt.Errorf("ファイルのコピーに失敗しました: %w", err)
	}

	fmt.Println(dest)
	return nil
}
