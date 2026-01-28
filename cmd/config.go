package main

import (
	"fmt"
	"os"

	"github.com/Wsine/feishu2md/core"
	"github.com/Wsine/feishu2md/utils"
)

type ConfigOpts struct {
	appId           string
	appSecret       string
	userAccessToken string
	authType        string
}

var configOpts = ConfigOpts{}

func handleConfigCommand() error {
	configPath, err := core.GetConfigFilePath()
	if err != nil {
		return err
	}

	fmt.Println("Configuration file on: " + configPath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 创建新配置
		config := core.NewConfig(configOpts.appId, configOpts.appSecret)

		// 设置用户鉴权相关字段
		if configOpts.userAccessToken != "" {
			config.Feishu.UserAccessToken = configOpts.userAccessToken
		}
		if configOpts.authType != "" {
			config.Feishu.AuthType = configOpts.authType
		}

		// 验证配置
		if err = config.Feishu.Validate(); err != nil {
			return err
		}

		if err = config.WriteConfig2File(configPath); err != nil {
			return err
		}
		fmt.Println(utils.PrettyPrint(config))
	} else {
		// 更新现有配置
		config, err := core.ReadConfigFromFile(configPath)
		if err != nil {
			return err
		}

		// 更新字段
		if configOpts.appId != "" {
			config.Feishu.AppId = configOpts.appId
		}
		if configOpts.appSecret != "" {
			config.Feishu.AppSecret = configOpts.appSecret
		}
		if configOpts.userAccessToken != "" {
			config.Feishu.UserAccessToken = configOpts.userAccessToken
		}
		if configOpts.authType != "" {
			config.Feishu.AuthType = configOpts.authType
		}

		// 验证配置
		if err = config.Feishu.Validate(); err != nil {
			return err
		}

		// 如果有任何字段被修改,保存配置
		if configOpts.appId != "" || configOpts.appSecret != "" ||
			configOpts.userAccessToken != "" || configOpts.authType != "" {
			if err = config.WriteConfig2File(configPath); err != nil {
				return err
			}
		}
		fmt.Println(utils.PrettyPrint(config))
	}
	return nil
}
