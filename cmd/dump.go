package main

import (
	"context"
	"fmt"
	"regexp"

	"github.com/Wsine/feishu2md/core"
	"github.com/Wsine/feishu2md/utils"
	"github.com/chyroc/lark"
	"github.com/pkg/errors"
)

func handleDumpCommand(url string) error {
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

	data := struct {
		Document *lark.DocxDocument `json:"document"`
		Blocks   []*lark.DocxBlock  `json:"blocks"`
	}{
		Document: docx,
		Blocks:   blocks,
	}
	pdata := utils.PrettyPrint(data)
	fmt.Println(pdata)

	return nil
}
