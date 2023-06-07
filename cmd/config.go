package main

import (
	"fmt"
	"os"

	"github.com/Wsine/feishu2md/core"
	"github.com/Wsine/feishu2md/utils"
)

func handleConfigCommand(appId, appSecret string) error {
	configPath, err := core.GetConfigFilePath()
	if err != nil {
		return err
	}
	fmt.Println("Configuration file on: " + configPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := core.NewConfig(appId, appSecret)
		if err = config.WriteConfig2File(configPath); err != nil {
			return err
		}
		fmt.Println(utils.PrettyPrint(config))
	} else {
		config, err := core.ReadConfigFromFile(configPath)
		if err != nil {
			return err
		}
		if appId != "" {
			config.Feishu.AppId = appId
		}
		if appSecret != "" {
			config.Feishu.AppSecret = appSecret
		}
		if appId != "" || appSecret != "" {
			if err = config.WriteConfig2File(configPath); err != nil {
				return err
			}
		}
		fmt.Println(utils.PrettyPrint(config))
	}
	return nil
}
