package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"
)

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
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "notebook-url",
								Usage:    "操作対象のノートブックURL",
								Required: true,
							},
						},
						Action: func(_ context.Context, cmd *cli.Command) error {
							return deleteSourceAction(cmd.String("notebook-url"))
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

	return service.AddSource(string(text))
}

func deleteSourceAction(notebookURL string) error {
	client := NewClient(1)
	mapping := NewMappingStore("")
	service := NewService(client, notebookURL, mapping)

	return service.DeleteSource()
}
