package main

import (
	"fmt"
	"os"

	"github.com/Wsine/feishu2md/core"
	"github.com/Wsine/feishu2md/utils"
)

type ConfigOpts struct {
	appId     string
	appSecret string
}

var configOpts = ConfigOpts{}

func handleConfigCommand() error {
	configPath, err := core.GetConfigFilePath()
	if err != nil {
		return err
	}

	fmt.Println("Configuration file on: " + configPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := core.NewConfig(configOpts.appId, configOpts.appSecret)
		if err = config.WriteConfig2File(configPath); err != nil {
			return err
		}
		fmt.Println(utils.PrettyPrint(config))
	} else {
		config, err := core.ReadConfigFromFile(configPath)
		if err != nil {
			return err
		}
		if configOpts.appId != "" {
			config.Feishu.AppId = configOpts.appId
		}
		if configOpts.appSecret != "" {
			config.Feishu.AppSecret = configOpts.appSecret
		}
		if configOpts.appId != "" || configOpts.appSecret != "" {
			if err = config.WriteConfig2File(configPath); err != nil {
				return err
			}
		}
		fmt.Println(utils.PrettyPrint(config))
	}
	return nil
}
