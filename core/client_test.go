package core_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Wsine/feishu2md/core"
	"github.com/Wsine/feishu2md/utils"
)

func getIdAndSecretFromEnv() (string, string) {
  utils.LoadEnv()
  appID := os.Getenv("FEISHU_APP_ID")
  appSecret := os.Getenv("FEISHU_APP_SECRET")
  return appID, appSecret
}

func TestNewClient(t *testing.T) {
  appID, appSecret := getIdAndSecretFromEnv()
  c := core.NewClient(appID, appSecret, "feishu.cn")
  if c == nil {
    t.Errorf("Error creating DocClient")
  }
}

func TestGetContent(t *testing.T) {
  appID, appSecret := getIdAndSecretFromEnv()
  c := core.NewClient(appID, appSecret, "feishu.cn")
  content, err := c.GetDocContent(context.Background(), "doccnZhhDCLiCP6LJa1nrjDhRSc")
  if err != nil {
    t.Error(err)
  }
  title := content.Title.Elements[0].TextRun.Text
  fmt.Println(title)
  if title == "" {
    t.Errorf("Error: parsed title is empty")
  }
}

func TestDownloadImage(t *testing.T) {
  appID, appSecret := getIdAndSecretFromEnv()
  c := core.NewClient(appID, appSecret, "feishu.cn")
  imgToken := "boxcnsXaIKcbwVmvGAwopgu2pre"
  filename, err := c.DownloadImage(context.Background(), imgToken)
  if err != nil {
    t.Error(err)
  }
  if filename != "static/" + imgToken + ".png" {
    fmt.Println(filename)
    t.Errorf("Error: not expected file extension")
  }
}
