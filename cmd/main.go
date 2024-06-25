package main

import (
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

var version = "v2-test"

func main() {
	app := &cli.App{
		Name:    "feishu2md",
		Version: strings.TrimSpace(string(version)),
		Usage:   "Download feishu/larksuite document to markdown file",
		Action: func(ctx *cli.Context) error {
			cli.ShowAppHelp(ctx)
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "config",
				Usage: "Read config file or set field(s) if provided",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "appId",
						Value:       "",
						Usage:       "Set app id for the OPEN API",
						Destination: &configOpts.appId,
					},
					&cli.StringFlag{
						Name:        "appSecret",
						Value:       "",
						Usage:       "Set app secret for the OPEN API",
						Destination: &configOpts.appSecret,
					},
				},
				Action: func(ctx *cli.Context) error {
					return handleConfigCommand(&configOpts)
				},
			},
			{
				Name:    "download",
				Aliases: []string{"dl"},
				Usage:   "Download feishu/larksuite document to markdown file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "output",
						Aliases:     []string{"o"},
						Value:       "./",
						Usage:       "Specify the output directory for the markdown files",
						Destination: &downloadOpts.outputDir,
					},
					&cli.BoolFlag{
						Name:        "dump",
						Value:       false,
						Usage:       "Dump json response of the OPEN API",
						Destination: &downloadOpts.dump,
					},
				},
				ArgsUsage: "<url>",
				Action: func(ctx *cli.Context) error {
					if ctx.NArg() == 0 {
						return cli.Exit("Please specify the document url", 1)
					} else {
						url := ctx.Args().First()
						return handleDownloadCommand(url, &downloadOpts)
					}
				},
			},
			{
				Name:    "batch",
				Aliases: []string{"b"},
				Usage:   "Download multiple feishu/larksuite documents to markdown files",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "baseDir",
						Aliases:     []string{"b"},
						Value:       "",
						Usage:       "Specify the base directory for the document urls",
						Destination: &batchDownloadOpts.baseDir,
					},
					&cli.StringFlag{
						Name:        "output",
						Aliases:     []string{"o"},
						Value:       "./",
						Usage:       "Specify the output directory for the markdown files",
						Destination: &batchDownloadOpts.outputDir,
					},
					&cli.StringFlag{
						Name:        "Lark space URL",
						Aliases:     []string{"l"},
						Value:       "xxxxxxxx.larksuite.com",
						Usage:       "Specify the base URL for your larksuite space without https:// prefix",
						Destination: &batchDownloadOpts.larkSpaceURL,
					},
				},
				ArgsUsage: "<base directory> <output directory> <lark space URL>",
				Action: func(ctx *cli.Context) error {
					if ctx.NArg() == 0 {
						return cli.Exit("Please specify the base directory for the document urls", 1)
					} else {
						baseDir := ctx.Args().Get(0)
						output := ctx.Args().Get(1)
						larkSpaceURL := ctx.Args().Get(2)
						return handleBatchDownloadCommand(baseDir, output, larkSpaceURL, &batchDownloadOpts)
					}
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
