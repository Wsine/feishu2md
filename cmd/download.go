package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

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

var dlOpts = DownloadOpts{}
var dlConfig core.Config

func downloadDocument(ctx context.Context, client *core.Client, url string, opts *DownloadOpts) error {
	// Validate the url to download
	docType, docToken, err := utils.ValidateDocumentURL(url)
	if err != nil {
		return err
	}
	fmt.Println("Captured document token:", docToken)

	// for a wiki page, we need to renew docType and docToken first
	if docType == "wiki" {
		node, err := client.GetWikiNodeInfo(ctx, docToken)
		utils.CheckErr(err)
		docType = node.ObjType
		docToken = node.ObjToken
	}
	if docType == "docs" {
		return errors.Errorf(
			`Feishu Docs is no longer supported. ` +
				`Please refer to the Readme/Release for v1_support.`)
	}

	// Process the download
	docx, blocks, err := client.GetDocxContent(ctx, docToken)
	utils.CheckErr(err)

	parser := core.NewParser(dlConfig.Output)

	title := docx.Title
	markdown := parser.ParseDocxContent(docx, blocks)

	if !dlConfig.Output.SkipImgDownload {
		for _, imgToken := range parser.ImgTokens {
			localLink, err := client.DownloadImage(
				ctx, imgToken, filepath.Join(opts.outputDir, dlConfig.Output.ImageDir),
			)
			if err != nil {
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
	if _, err := os.Stat(opts.outputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(opts.outputDir, 0o755); err != nil {
			return err
		}
	}

	if dlOpts.dump {
		jsonName := fmt.Sprintf("%s.json", docToken)
		outputPath := filepath.Join(opts.outputDir, jsonName)
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
	if dlConfig.Output.TitleAsFilename {
		mdName = fmt.Sprintf("%s.md", title)
	}
	outputPath := filepath.Join(opts.outputDir, mdName)
	if err = os.WriteFile(outputPath, []byte(result), 0o644); err != nil {
		return err
	}
	fmt.Printf("Downloaded markdown file to %s\n", outputPath)

	return nil
}

func downloadDocuments(ctx context.Context, client *core.Client, url string) error {
	// Validate the url to download
	folderToken, err := utils.ValidateFolderURL(url)
	if err != nil {
		return err
	}
	fmt.Println("Captured folder token:", folderToken)

	// Error channel and wait group
	errChan := make(chan error)
	wg := sync.WaitGroup{}

	// Recursively go through the folder and download the documents
	var processFolder func(ctx context.Context, folderPath, folderToken string) error
	processFolder = func(ctx context.Context, folderPath, folderToken string) error {
		files, err := client.GetDriveFolderFileList(ctx, nil, &folderToken)
		if err != nil {
			return err
		}
		opts := DownloadOpts{outputDir: folderPath, dump: dlOpts.dump, batch: false}
		for _, file := range files {
			if file.Type == "folder" {
				_folderPath := filepath.Join(folderPath, file.Name)
				if err := processFolder(ctx, _folderPath, file.Token); err != nil {
					return err
				}
			} else if file.Type == "docx" {
				// concurrently download the document
				wg.Add(1)
				go func(_url string) {
					if err := downloadDocument(ctx, client, _url, &opts); err != nil {
						errChan <- err
					}
					wg.Done()
				}(file.URL)
			}
		}
		return nil
	}
	if err := processFolder(ctx, dlOpts.outputDir, folderToken); err != nil {
		return err
	}

	// Wait for all the downloads to finish
	go func() {
		wg.Wait()
		close(errChan)
	}()
	for err := range errChan {
		return err
	}
	return nil
}

func handleDownloadCommand(url string) error {
	// Load config
	configPath, err := core.GetConfigFilePath()
	if err != nil {
		return err
	}
	config, err := core.ReadConfigFromFile(configPath)
	if err != nil {
		return err
	}
	dlConfig = *config

	// Instantiate the client
	client := core.NewClient(
		dlConfig.Feishu.AppId, dlConfig.Feishu.AppSecret,
	)
	ctx := context.Background()

	if dlOpts.batch {
		return downloadDocuments(ctx, client, url)
	}

	return downloadDocument(ctx, client, url, &dlOpts)
}
