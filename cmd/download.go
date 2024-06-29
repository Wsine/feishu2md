package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/88250/lute"
	"github.com/Wsine/feishu2md/core"
	"github.com/Wsine/feishu2md/utils"
	"github.com/chyroc/lark"
	"github.com/pkg/errors"
)

type DownloadOpts struct {
	outputDir string
	dump      bool
	batch     bool
}

var downloadOpts = DownloadOpts{}

func downloadDocument(url string, outputDir string, dump bool, config *core.Config) error {
	// Validate the url to download
	docType, docToken, err := utils.ValidateDownloadURL(url)
	// This might be a sheet or other kind of URL, we don't support downloading it yet so we just skip it
	if err != nil {
		return nil
	}
	fmt.Println("Captured document token:", docToken)

	// Create client with context
	ctx := context.WithValue(context.Background(), "output", config.Output)

	client := core.NewClient(
		config.Feishu.AppId, config.Feishu.AppSecret,
	)

	// for a wiki page, we need to renew docType and docToken first
	if docType == "wiki" {
		node, err := client.GetWikiNodeInfo(ctx, docToken)
		utils.CheckErr(err)
		docType = node.ObjType
		docToken = node.ObjToken
	}
	if docType == "docs" {
		return errors.Errorf("Feishu Docs is no longer supported. Please refer to the Readme/Release for v1_support.")
	}

	// Process the download
	docx, blocks, err := client.GetDocxContent(ctx, docToken)
	utils.CheckErr(err)

	parser := core.NewParser(ctx)

	title := docx.Title
	markdown := parser.ParseDocxContent(docx, blocks)

	if !config.Output.SkipImgDownload {
		for _, imgToken := range parser.ImgTokens {
			localLink, err := client.DownloadImage(
				ctx, imgToken, filepath.Join(outputDir, config.Output.ImageDir),
			)
			if utils.CheckErr(err) != nil {
				return err
			}
			markdown = strings.Replace(markdown, imgToken, localLink, 1)
		}
	}

	// Format the markdown document
	engine := lute.New(func(l *lute.Lute) {
		l.RenderOptions.AutoSpace = true
	})
	result := engine.FormatStr("md", markdown)

	// Handle the output directory and name
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return err
		}
	}

	if dump {
		jsonName := fmt.Sprintf("%s.json", docToken)
		outputPath := filepath.Join(outputDir, jsonName)
		data := struct {
			Document *lark.DocxDocument `json:"document"`
			Blocks   []*lark.DocxBlock  `json:"blocks"`
		}{
			Document: docx,
			Blocks:   blocks,
		}
		pdata := utils.PrettyPrint(data)

		if err = os.WriteFile(outputPath, []byte(pdata), 0o644); err != nil {
			return err
		}
		fmt.Printf("Dumped json response to %s\n", outputPath)
	}

	// Write to markdown file
	mdName := fmt.Sprintf("%s.md", docToken)
	if config.Output.TitleAsFilename {
		mdName = fmt.Sprintf("%s.md", title)
	}
	outputPath := filepath.Join(outputDir, mdName)
	if err = os.WriteFile(outputPath, []byte(result), 0o644); err != nil {
		return err
	}
	fmt.Printf("Downloaded markdown file to %s\n", outputPath)

	return nil
}

func handleDownloadCommand(url string, opts *DownloadOpts) error {
	// Load config
	configPath, err := core.GetConfigFilePath()
	utils.CheckErr(err)
	config, err := core.ReadConfigFromFile(configPath)
	utils.CheckErr(err)

	if opts.batch {
		return batchDownloadFolder(opts.outputDir, url, config)
	}

	return downloadDocument(url, opts.outputDir, opts.dump, config)
}
