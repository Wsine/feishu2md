package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/88250/lute"
	"github.com/Wsine/feishu2md/core"
	"github.com/Wsine/feishu2md/utils"
	"github.com/urfave/cli/v2"
)

func handleConfigCommand(appId, appSecret string) error {
	configPath, err := core.GetConfigFilePath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := core.NewConfig(appId, appSecret)
		if err = config.WriteConfig2File(configPath); err != nil {
			return err
		}
		fmt.Println(utils.PrettyPrint(config))
	} else {
		config, err := core.ReadConfigFromFile(configPath)
		if err != nil {
			return err
		}
		if appId != "" {
			config.Feishu.AppId = appId
		}
		if appSecret != "" {
			config.Feishu.AppSecret = appSecret
		}
		if appId != "" || appSecret != "" {
			if err = config.WriteConfig2File(configPath); err != nil {
				return err
			}
		}
		fmt.Println(utils.PrettyPrint(config))
	}
	return nil
}

func handleUrlArgument(url string, verbose bool) error {
	configPath, err := core.GetConfigFilePath()
	if err != nil {
		return err
	}
	config, err := core.ReadConfigFromFile(configPath)
	if err != nil {
		return err
	}

	reg := regexp.MustCompile("^https://[a-zA-Z0-9-]+.(feishu.cn|larksuite.com)/(docs|docx|wiki)/([a-zA-Z0-9]+)")
	matchResult := reg.FindStringSubmatch(url)
	if matchResult == nil || len(matchResult) != 4 {
		return fmt.Errorf("Invalid feishu/larksuite URL format\n")
	}

	domain := matchResult[1]
	docType := matchResult[2]
	docToken := matchResult[3]
	fmt.Println("Captured document token:", docToken)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "Verbose", verbose)
	ctx = context.WithValue(ctx, "OutputConfig", config.Output)

	client := core.NewClient(
		config.Feishu.AppId, config.Feishu.AppSecret, domain,
	)

	parser := core.NewParser(ctx)
	title := ""
	markdown := ""

	// for a wiki page, we need to renew docType and docToken first
	if docType == "wiki" {
		node, err := client.GetWikiNodeInfo(ctx, docToken)
		if err != nil {
			return err
		}
		docType = node.ObjType
		docToken = node.ObjToken
	}

	if docType == "docx" {
		docx, blocks, err := client.GetDocxContent(ctx, docToken)
		if err != nil {
			return err
		}
		markdown = parser.ParseDocxContent(docx, blocks)
		title = docx.Title
	} else {
		doc, err := client.GetDocContent(ctx, docToken)
		if err != nil {
			return err
		}
		markdown = parser.ParseDocContent(doc)
		title = doc.Title.Elements[0].TextRun.Text
	}

	for _, imgToken := range parser.ImgTokens {
		localLink, err := client.DownloadImage(ctx, imgToken)
		if err != nil {
			return err
		}
		markdown = strings.Replace(markdown, imgToken, localLink, 1)
	}

	engine := lute.New(func(l *lute.Lute) {
		l.RenderOptions.AutoSpace = true
	})
	result := engine.FormatStr("md", markdown)

	mdName := fmt.Sprintf("%s.md", docToken)
	if config.Output.TitleAsFilename {
		mdName = fmt.Sprintf("%s.md", title)
	}
	if err = os.WriteFile(mdName, []byte(result), 0o644); err != nil {
		return err
	}
	fmt.Printf("Downloaded markdown file to %s\n", mdName)

	return nil
}

func main() {
	app := &cli.App{
		Name:    "feishu2md",
		Version: "v1.2.2",
		Usage:   "download feishu/larksuite document to markdown file",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"vv"},
				Usage:   "verbose the intermediate output",
			},
		},
		Action: func(ctx *cli.Context) error {
			verbose := ctx.Bool("verbose")
			if ctx.NArg() > 0 {
				url := ctx.Args().Get(0)
				return handleUrlArgument(url, verbose)
			} else {
				cli.ShowAppHelp(ctx)
			}
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "config",
				Usage: "read config file or set field(s) if provided",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "appId",
						Value: "",
						Usage: "set app id for the OPEN api",
					},
					&cli.StringFlag{
						Name:  "appSecret",
						Value: "",
						Usage: "set app secret for the OPEN api",
					},
				},
				Action: func(ctx *cli.Context) error {
					return handleConfigCommand(
						ctx.String("appId"), ctx.String("appSecret"),
					)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
