package core

import (
  "os"
  "path"

  "github.com/Wsine/feishu2md/utils"
)

type Config struct {
  Feishu    FeishuConfig `json:"feishu"`
  Output    OutputConfig `json:"output"`
}

type FeishuConfig struct {
  AppId     string `json:"app_id"`
  AppSecret string `json:"app_secret"`
}

type OutputConfig struct {
  ImageDir  string `json:"image_dir"`
}

func GetConfigFilePath() (string, error) {
  configPath, err := os.UserConfigDir()
  utils.CheckErr(err)
  configFilePath := path.Join(configPath, "feishu2md", "config.json")
  return configFilePath, nil
}
