package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

// 鉴权类型常量
const (
	AuthTypeApp  = "app"
	AuthTypeUser = "user"
)

// 配置版本
const ConfigVersion = "2.0"

type Config struct {
	Version string       `json:"version,omitempty"`
	Feishu  FeishuConfig `json:"feishu"`
	Output  OutputConfig `json:"output"`
}

type FeishuConfig struct {
	AppId     string `json:"app_id"`
	AppSecret string `json:"app_secret"`

	// 用户鉴权相关字段
	UserAccessToken string `json:"user_access_token,omitempty"`

	// 鉴权类型选择: "app" 或 "user"
	// 默认为 "app",保持向后兼容
	AuthType string `json:"auth_type,omitempty"`
}

type OutputConfig struct {
	ImageDir        string `json:"image_dir"`
	TitleAsFilename bool   `json:"title_as_filename"`
	UseHTMLTags     bool   `json:"use_html_tags"`
	SkipImgDownload bool   `json:"skip_img_download"`
}

func NewConfig(appId, appSecret string) *Config {
	return &Config{
		Version: ConfigVersion,
		Feishu: FeishuConfig{
			AppId:     appId,
			AppSecret: appSecret,
			AuthType:  AuthTypeApp, // 默认应用鉴权
		},
		Output: OutputConfig{
			ImageDir:        "static",
			TitleAsFilename: false,
			UseHTMLTags:     false,
			SkipImgDownload: false,
		},
	}
}

// Validate 验证配置的有效性
func (fc *FeishuConfig) Validate() error {
	// 设置默认值
	if fc.AuthType == "" {
		fc.AuthType = AuthTypeApp
	}

	// 验证 AuthType 的有效性
	if fc.AuthType != AuthTypeApp && fc.AuthType != AuthTypeUser {
		return fmt.Errorf("invalid auth_type: %s, must be 'app' or 'user'", fc.AuthType)
	}

	// 验证必需字段
	if fc.AuthType == AuthTypeApp {
		if fc.AppId == "" || fc.AppSecret == "" {
			return fmt.Errorf("app_id and app_secret are required for app authentication")
		}
	} else if fc.AuthType == AuthTypeUser {
		if fc.UserAccessToken == "" {
			return fmt.Errorf("user_access_token is required for user authentication")
		}
	}

	return nil
}

func GetConfigFilePath() (string, error) {
	configPath, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	configFilePath := path.Join(configPath, "feishu2md", "config.json")
	return configFilePath, nil
}

func ReadConfigFromFile(configPath string) (*Config, error) {
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	config := NewConfig("", "")
	err = json.Unmarshal([]byte(file), &config)
	if err != nil {
		return nil, err
	}

	// 迁移旧配置
	migrated := config.Migrate()

	// 验证配置
	if err = config.Feishu.Validate(); err != nil {
		return nil, err
	}

	// 如果发生迁移，自动保存新配置
	if migrated {
		fmt.Println("配置文件已自动迁移到 v2.0 格式")
		if err = config.WriteConfig2File(configPath); err != nil {
			fmt.Printf("警告：保存迁移后的配置失败: %v\n", err)
		}
	}

	return config, nil
}

// Migrate 迁移旧版本配置到新版本，返回是否发生迁移
func (c *Config) Migrate() bool {
	migrated := false

	// 检测旧配置（无 version 字段）
	if c.Version == "" {
		c.Version = ConfigVersion
		migrated = true
	}

	// 确保 AuthType 有默认值
	if c.Feishu.AuthType == "" {
		c.Feishu.AuthType = AuthTypeApp
		migrated = true
	}

	return migrated
}

func (conf *Config) WriteConfig2File(configPath string) error {
	err := os.MkdirAll(filepath.Dir(configPath), 0o755)
	if err != nil {
		return err
	}
	file, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(configPath, file, 0o644)
	return err
}
