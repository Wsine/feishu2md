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
)

type SyncOpts struct {
	outputDir   string
	incremental bool   // 增量同步，默认开启
	force       bool   // 强制重新下载
	include     string // 仅下载匹配的目录（白名单，逗号分隔）
	exclude     string // 排除匹配的目录（黑名单，逗号分隔）
	concurrency int    // 并发数
	dump        bool   // 导出 JSON 响应
}

var syncOpts = SyncOpts{}
var syncConfig core.Config

// syncDocument 同步单个文档
func syncDocument(ctx context.Context, client *core.Client, url string, opts *SyncOpts, cacheManager *core.CacheManager) error {
	// Validate the url to download
	docType, docToken, err := utils.ValidateDocumentURL(url)
	if err != nil {
		return err
	}

	// for a wiki page, we need to renew docType and docToken first
	if docType == "wiki" {
		node, err := client.GetWikiNodeInfo(ctx, docToken)
		if err != nil {
			err = fmt.Errorf("GetWikiNodeInfo err: %v for %v", err, url)
		}
		utils.CheckErr(err)
		docType = node.ObjType
		docToken = node.ObjToken
	}

	// Process the download
	docx, blocks, err := client.GetDocxContent(ctx, docToken)
	utils.CheckErr(err)

	title := docx.Title
	revisionID := docx.RevisionID

	// 确定输出文件名
	var mdName string
	if syncConfig.Output.TitleAsFilename {
		mdName = fmt.Sprintf("%s.md", utils.SanitizeFileName(title))
	} else {
		mdName = fmt.Sprintf("%s.md", docToken)
	}
	outputPath := filepath.Join(opts.outputDir, mdName)

	// 增量下载逻辑：检查是否需要下载
	if opts.incremental && !opts.force && cacheManager != nil {
		shouldDownload, skipReason := cacheManager.ShouldDownload(
			docToken,
			revisionID,
			outputPath,
		)

		if !shouldDownload {
			fmt.Printf("⊘ 跳过: %s - %s\n", title, skipReason)
			// 即使跳过下载，也要更新缓存（用于建立缓存映射）
			cacheManager.UpdateDocument(
				docToken,
				revisionID,
				title,
				mdName,
				docType,
			)
			return nil
		}
	}

	// 继续执行下载流程
	parser := core.NewParser(syncConfig.Output)

	markdown := parser.ParseDocxContent(docx, blocks)

	if !syncConfig.Output.SkipImgDownload {
		for _, imgToken := range parser.ImgTokens {
			localLink, err := client.DownloadImage(
				ctx, imgToken, filepath.Join(opts.outputDir, syncConfig.Output.ImageDir),
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

	if opts.dump {
		jsonName := fmt.Sprintf("%s.json", docToken)
		jsonOutputPath := filepath.Join(opts.outputDir, jsonName)
		data := struct {
			Document *lark.DocxDocument `json:"document"`
			Blocks   []*lark.DocxBlock  `json:"blocks"`
		}{
			Document: docx,
			Blocks:   blocks,
		}
		pdata := utils.PrettyPrint(data)

		if err = os.WriteFile(jsonOutputPath, []byte(pdata), 0o644); err != nil {
			return err
		}
		fmt.Printf("Dumped json response to %s\n", jsonOutputPath)
	}

	// Write to markdown file
	if err = os.WriteFile(outputPath, []byte(result), 0o644); err != nil {
		return err
	}
	fmt.Printf("✓ 已同步: %s\n", outputPath)

	// 更新缓存
	if cacheManager != nil {
		cacheManager.UpdateDocument(
			docToken,
			revisionID,
			title,
			mdName,
			docType,
		)
	}

	return nil
}

// syncFolder 同步云盘文件夹
func syncFolder(ctx context.Context, client *core.Client, url string, opts *SyncOpts, cacheManager *core.CacheManager, filter *core.NodeFilter) error {
	// Validate the url to download
	folderToken, err := utils.ValidateFolderURL(url)
	if err != nil {
		return err
	}
	fmt.Println("Captured folder token:", folderToken)

	// Error channel and wait group
	errChan := make(chan error)
	wg := sync.WaitGroup{}
	semaphore := make(chan struct{}, opts.concurrency)

	// Recursively go through the folder and download the documents
	var processFolder func(ctx context.Context, folderPath, folderToken string) error
	processFolder = func(ctx context.Context, folderPath, folderToken string) error {
		files, err := client.GetDriveFolderFileList(ctx, nil, &folderToken)
		if err != nil {
			return err
		}
		docOpts := &SyncOpts{
			outputDir:   folderPath,
			dump:        opts.dump,
			incremental: opts.incremental,
			force:       opts.force,
			concurrency: opts.concurrency,
		}
		for _, file := range files {
			if file.Type == "folder" {
				// 检查文件夹是否应该被下载
				if filter != nil && !filter.ShouldDownloadFolder(folderPath, file.Name) {
					fmt.Printf("⊘ 跳过文件夹: %s\n", file.Name)
					continue
				}
				_folderPath := filepath.Join(folderPath, file.Name)
				if err := processFolder(ctx, _folderPath, file.Token); err != nil {
					return err
				}
			} else if file.Type == "docx" {
				// 检查文档的父目录是否被排除
				if filter != nil && !filter.ShouldDownloadDocument(folderPath) {
					continue
				}
				// concurrently download the document
				wg.Add(1)
				semaphore <- struct{}{}
				go func(_url string) {
					defer func() {
						wg.Done()
						<-semaphore
					}()
					if err := syncDocument(ctx, client, _url, docOpts, cacheManager); err != nil {
						errChan <- err
					}
				}(file.URL)
			}
		}
		return nil
	}
	if err := processFolder(ctx, opts.outputDir, folderToken); err != nil {
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

// syncWiki 同步知识库
func syncWiki(ctx context.Context, client *core.Client, url string, opts *SyncOpts, cacheManager *core.CacheManager, filter *core.NodeFilter) error {
	prefixURL, spaceID, err := utils.ValidateWikiURL(url)
	if err != nil {
		return err
	}

	folderPath, err := client.GetWikiName(ctx, spaceID)
	if err != nil {
		return err
	}
	if folderPath == "" {
		return fmt.Errorf("failed to GetWikiName")
	}

	errChan := make(chan error)

	wg := sync.WaitGroup{}
	semaphore := make(chan struct{}, opts.concurrency)

	var downloadWikiNode func(ctx context.Context,
		client *core.Client,
		spaceID string,
		parentPath string,
		parentNodeToken *string) error

	downloadWikiNode = func(ctx context.Context,
		client *core.Client,
		spaceID string,
		folderPath string,
		parentNodeToken *string) error {
		nodes, err := client.GetWikiNodeList(ctx, spaceID, parentNodeToken)
		if err != nil {
			return err
		}
		for _, n := range nodes {
			if n.HasChild {
				// 检查目录是否应该被下载
				if filter != nil {
					include, skippedByParent := filter.ShouldIncludeNode(folderPath, n.Title)
					if !include {
						if skippedByParent {
							fmt.Printf("⊘ 跳过目录（父目录已排除）: %s\n", n.Title)
						} else {
							fmt.Printf("⊘ 跳过目录: %s\n", n.Title)
						}
						continue
					}
				}
				_folderPath := filepath.Join(folderPath, n.Title)
				if err := downloadWikiNode(ctx, client,
					spaceID, _folderPath, &n.NodeToken); err != nil {
					return err
				}
			}
			if n.ObjType == "docx" {
				// 检查文档的父目录是否被排除
				if filter != nil && !filter.ShouldDownloadDocument(folderPath) {
					continue
				}
				docOpts := &SyncOpts{
					outputDir:   folderPath,
					dump:        opts.dump,
					incremental: opts.incremental,
					force:       opts.force,
					concurrency: opts.concurrency,
				}
				wg.Add(1)
				semaphore <- struct{}{}
				go func(_url string) {
					defer func() {
						wg.Done()
						<-semaphore
					}()
					if err := syncDocument(ctx, client, _url, docOpts, cacheManager); err != nil {
						errChan <- err
					}
				}(prefixURL + "/wiki/" + n.NodeToken)
			}
		}
		return nil
	}

	if err = downloadWikiNode(ctx, client, spaceID, folderPath, nil); err != nil {
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

// detectURLType 检测 URL 类型
func detectURLType(url string) (string, error) {
	// 尝试作为 Wiki URL
	if _, _, err := utils.ValidateWikiURL(url); err == nil {
		return core.SourceTypeWiki, nil
	}

	// 尝试作为文件夹 URL
	if _, err := utils.ValidateFolderURL(url); err == nil {
		return core.SourceTypeFolder, nil
	}

	return "", fmt.Errorf("URL 格式不正确，sync 命令仅支持文件夹或知识库 URL")
}

func handleSyncCommand(url string) error {
	// Load config
	configPath, err := core.GetConfigFilePath()
	if err != nil {
		return err
	}
	config, err := core.ReadConfigFromFile(configPath)
	if err != nil {
		return err
	}
	syncConfig = *config

	// 尝试加载已有的同步配置
	existingSyncConfig, err := core.LoadSyncConfig(syncOpts.outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "警告: 加载同步配置失败: %v\n", err)
	}

	// 如果没有提供 URL，尝试从同步配置中读取
	if url == "" {
		if existingSyncConfig != nil && existingSyncConfig.SourceURL != "" {
			url = existingSyncConfig.SourceURL
			fmt.Printf("使用已保存的同步配置: %s\n", url)
		} else {
			return fmt.Errorf("请提供要同步的 URL，或在已有同步配置的目录中运行")
		}
	}

	// 检测 URL 类型
	sourceType, err := detectURLType(url)
	if err != nil {
		return err
	}
	fmt.Printf("检测到源类型: %s\n", sourceType)

	// 创建或更新同步配置
	var currentSyncConfig *core.SyncConfig
	if existingSyncConfig != nil {
		currentSyncConfig = existingSyncConfig
		currentSyncConfig.SourceURL = url
		currentSyncConfig.SourceType = sourceType
	} else {
		currentSyncConfig = core.NewSyncConfig(url, sourceType)
	}

	// 更新同步配置（合并命令行参数）
	currentSyncConfig.Update(
		core.ParsePatterns(syncOpts.include),
		core.ParsePatterns(syncOpts.exclude),
		syncOpts.concurrency,
	)

	// Instantiate the client
	client := core.NewClient(syncConfig.Feishu)
	ctx := context.Background()

	// 初始化缓存管理器（sync 命令总是启用缓存）
	cacheManager, err := core.NewCacheManager(syncOpts.outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "警告: 无法初始化缓存管理器: %v\n", err)
		cacheManager = nil
	}

	if syncOpts.force {
		fmt.Println("强制模式: 将重新下载所有文档并更新缓存")
	} else if syncOpts.incremental {
		fmt.Println("增量模式: 将跳过未修改的文档")
	}

	// 初始化目录过滤器
	var nodeFilter *core.NodeFilter
	includePatterns := currentSyncConfig.Include
	excludePatterns := currentSyncConfig.Exclude

	// 命令行参数优先
	if syncOpts.include != "" {
		includePatterns = core.ParsePatterns(syncOpts.include)
	}
	if syncOpts.exclude != "" {
		excludePatterns = core.ParsePatterns(syncOpts.exclude)
	}

	if len(includePatterns) > 0 || len(excludePatterns) > 0 {
		filterConfig := core.FilterConfig{
			IncludePatterns: includePatterns,
			ExcludePatterns: excludePatterns,
		}
		nodeFilter = core.NewNodeFilter(filterConfig)

		fmt.Println("目录过滤已启用:")
		if len(filterConfig.IncludePatterns) > 0 {
			fmt.Printf("  包含: %v\n", filterConfig.IncludePatterns)
		}
		if len(filterConfig.ExcludePatterns) > 0 {
			fmt.Printf("  排除: %v\n", filterConfig.ExcludePatterns)
		}
	}

	// 设置并发数
	if syncOpts.concurrency <= 0 {
		syncOpts.concurrency = currentSyncConfig.Concurrency
	}
	fmt.Printf("并发数: %d\n", syncOpts.concurrency)

	// 执行同步
	var syncErr error
	switch sourceType {
	case core.SourceTypeFolder:
		syncErr = syncFolder(ctx, client, url, &syncOpts, cacheManager, nodeFilter)
	case core.SourceTypeWiki:
		syncErr = syncWiki(ctx, client, url, &syncOpts, cacheManager, nodeFilter)
	}

	// 保存缓存
	if cacheManager != nil {
		if err := cacheManager.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "警告: 缓存保存失败: %v\n", err)
		} else {
			fmt.Println("✓ 缓存已更新")
		}
	}

	// 保存同步配置
	if syncErr == nil {
		if err := currentSyncConfig.Save(syncOpts.outputDir); err != nil {
			fmt.Fprintf(os.Stderr, "警告: 同步配置保存失败: %v\n", err)
		} else {
			fmt.Printf("✓ 同步配置已保存到 %s\n", core.GetSyncConfigPath(syncOpts.outputDir))
		}
	}

	return syncErr
}
