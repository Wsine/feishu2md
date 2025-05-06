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
					return handleConfigCommand()
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
						Destination: &dlOpts.outputDir,
					},
					&cli.StringFlag{
						Name:        "file",
						Aliases:     []string{"f"},
						Value:       "",
						Usage:       "Specify the output filename for the markdown files",
						Destination: &downloadOpts.outputFile,
					},
					&cli.BoolFlag{
						Name:        "dump",
						Value:       false,
						Usage:       "Dump json response of the OPEN API",
						Destination: &dlOpts.dump,
					},
					&cli.BoolFlag{
						Name:        "batch",
						Value:       false,
						Usage:       "Download all documents under a folder",
						Destination: &dlOpts.batch,
					},
					&cli.BoolFlag{
						Name:        "wiki",
						Value:       false,
						Usage:       "Download all documents within the wiki.",
						Destination: &dlOpts.wiki,
					},
				},
				ArgsUsage: "<url>",
				Action: func(ctx *cli.Context) error {
					if ctx.NArg() == 0 {
						return cli.Exit("Please specify the document/folder/wiki url", 1)
					} else {
						url := ctx.Args().First()
						return handleDownloadCommand(url)
					}
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
