package main

import (
	"context"
	"fmt"
	"os"

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
