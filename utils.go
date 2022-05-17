package main

import (
	"os"
	"path"
)

func checkErr(e error) {
  if e != nil {
    panic(e)
  }
}

func getConfigFilePath() (string, error) {
  configPath, err := os.UserConfigDir()
  checkErr(err)
  configFilePath := path.Join(configPath, "feishu2md", "config.json")
  return configFilePath, nil
}
