package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/88250/lute"
	"github.com/Wsine/feishu2md/core"
	"github.com/Wsine/feishu2md/utils"
	"github.com/pkg/errors"
)

func handleUrlArgument(url string) error {
	configPath, err := core.GetConfigFilePath()
	utils.CheckErr(err)
	config, err := core.ReadConfigFromFile(configPath)
	utils.CheckErr(err)

	reg := regexp.MustCompile("^https://[a-zA-Z0-9-]+.(feishu.cn|larksuite.com)/(docx|wiki)/([a-zA-Z0-9]+)")
	matchResult := reg.FindStringSubmatch(url)
	if matchResult == nil || len(matchResult) != 4 {
		return errors.Errorf("Invalid feishu/larksuite URL format")
	}

	domain := matchResult[1]
	docType := matchResult[2]
	docToken := matchResult[3]
	fmt.Println("Captured document token:", docToken)

	ctx := context.Background()

	client := core.NewClient(
		config.Feishu.AppId, config.Feishu.AppSecret, domain,
	)

	// for a wiki page, we need to renew docType and docToken first
	if docType == "wiki" {
		node, err := client.GetWikiNodeInfo(ctx, docToken)
		utils.CheckErr(err)
		docType = node.ObjType
		docToken = node.ObjToken
	}

	docx, blocks, err := client.GetDocxContent(ctx, docToken)
	utils.CheckErr(err)

	parser := core.NewParser(ctx)

	title := docx.Title
	markdown := parser.ParseDocxContent(docx, blocks)

	for _, imgToken := range parser.ImgTokens {
		localLink, err := client.DownloadImage(ctx, imgToken, config.Output.ImageDir)
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
