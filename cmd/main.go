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
		Usage:   "download feishu/larksuite document to markdown file",
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() > 0 {
				url := ctx.Args().Get(0)
				return handleUrlArgument(url)
			} else {
				cli.ShowAppHelp(ctx)
			}
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "config",
				Usage: "Read config file or set field(s) if provided",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "appId",
						Value: "",
						Usage: "Set app id for the OPEN API",
					},
					&cli.StringFlag{
						Name:  "appSecret",
						Value: "",
						Usage: "Set app secret for the OPEN API",
					},
				},
				Action: func(ctx *cli.Context) error {
					return handleConfigCommand(
						ctx.String("appId"), ctx.String("appSecret"),
					)
				},
			},
			{
				Name:  "dump",
				Usage: "Dump json response of the OPEN API",
				Action: func(ctx *cli.Context) error {
					if ctx.NArg() > 0 {
						url := ctx.Args().Get(0)
						return handleDumpCommand(url)
					} else {
						cli.ShowCommandHelp(ctx, "dump")
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
