package core_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Wsine/feishu2md/core"
)

func getConfigFromEnv(t *testing.T) core.FeishuConfig {
	configPath, err := core.GetConfigFilePath()
	if err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return core.FeishuConfig{
			AppId:     os.Getenv("FEISHU_APP_ID"),
			AppSecret: os.Getenv("FEISHU_APP_SECRET"),
			AuthType:  core.AuthTypeApp,
		}
	} else {
		config, err := core.ReadConfigFromFile(configPath)
		if err != nil {
			t.Error(err)
		}
		return config.Feishu
	}
}

func TestNewClient(t *testing.T) {
	config := getConfigFromEnv(t)
	c := core.NewClient(config)
	if c == nil {
		t.Errorf("Error creating DocClient")
	}
}

func TestDownloadImage(t *testing.T) {
	config := getConfigFromEnv(t)
	c := core.NewClient(config)
	imgToken := "boxcnA1QKPanfMhLxzF1eMhoArM"
	filename, err := c.DownloadImage(
		context.Background(),
		imgToken,
		"static",
	)
	if err != nil {
		t.Error(err)
	}
	if filename != "static/"+imgToken+".png" {
		fmt.Println(filename)
		t.Errorf("Error: not expected file extension")
	}
	if err := os.RemoveAll("static"); err != nil {
		t.Errorf("Error: failed to clean up the folder")
	}
}

func TestGetDocxContent(t *testing.T) {
	config := getConfigFromEnv(t)
	c := core.NewClient(config)
	docx, blocks, err := c.GetDocxContent(
		context.Background(),
		"doxcnXhd93zqoLnmVPGIPTy7AFe",
	)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(docx.Title)
	if docx.Title == "" {
		t.Errorf("Error: parsed title is empty")
	}
	fmt.Printf("number of blocks: %d\n", len(blocks))
	if len(blocks) == 0 {
		t.Errorf("Error: parsed blocks are empty")
	}
}

func TestGetWikiNodeInfo(t *testing.T) {
	config := getConfigFromEnv(t)
	c := core.NewClient(config)
	const token = "wikcnLgRX9AMtvaB5x1cl57Yuah"
	node, err := c.GetWikiNodeInfo(context.Background(), token)
	if err != nil {
		t.Error(err)
	}
	if node.ObjType != "docx" {
		t.Errorf("Error: node type incorrect")
	}
}

func TestGetDriveFolderFileList(t *testing.T) {
	config := getConfigFromEnv(t)
	c := core.NewClient(config)
	folderToken := "G15mfSfIHlyquudfhq5cg9kdnjg"
	files, err := c.GetDriveFolderFileList(
		context.Background(), nil, &folderToken)
	if err != nil {
		t.Error(err)
	}
	if len(files) == 0 {
		t.Errorf("Error: no files found")
	}
}

func TestGetWikiNodeList(t *testing.T) {
	config := getConfigFromEnv(t)
	c := core.NewClient(config)
	wikiToken := "7376995595006787612"
	nodes, err := c.GetWikiNodeList(context.Background(), wikiToken, nil)
	if err != nil {
		t.Error(err)
	}
	if len(nodes) == 0 {
		t.Errorf("Error: no nodes found")
	}
}
