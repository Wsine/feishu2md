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
  c := core.NewDocClient(appID, appSecret, "feishu.cn")
  if c == nil {
    t.Errorf("Error creating DocClient")
  }
}

func TestGetContent(t *testing.T) {
  appID, appSecret := getIdAndSecretFromEnv()
  c := core.NewDocClient(appID, appSecret, "feishu.cn")
  content := c.GetContent(context.Background(), "doccnZhhDCLiCP6LJa1nrjDhRSc")
  fmt.Println(content)
  if content == "" {
    t.Errorf("Error: content is empty")
  }
}
