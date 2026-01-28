package core

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFeishuConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  FeishuConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "有效的应用鉴权配置",
			config: FeishuConfig{
				AppId:     "cli_test123",
				AppSecret: "secret123",
				AuthType:  AuthTypeApp,
			},
			wantErr: false,
		},
		{
			name: "有效的用户鉴权配置",
			config: FeishuConfig{
				UserAccessToken: "u-test-token",
				AuthType:        AuthTypeUser,
			},
			wantErr: false,
		},
		{
			name: "AuthType 默认值",
			config: FeishuConfig{
				AppId:     "cli_test123",
				AppSecret: "secret123",
				AuthType:  "", // 空值应默认为 "app"
			},
			wantErr: false,
		},
		{
			name: "无效的 AuthType",
			config: FeishuConfig{
				AuthType: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid auth_type",
		},
		{
			name: "应用鉴权缺少 AppId",
			config: FeishuConfig{
				AppSecret: "secret123",
				AuthType:  AuthTypeApp,
			},
			wantErr: true,
			errMsg:  "app_id and app_secret are required",
		},
		{
			name: "用户鉴权缺少 UserAccessToken",
			config: FeishuConfig{
				AuthType: AuthTypeUser,
			},
			wantErr: true,
			errMsg:  "user_access_token is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigReadWrite(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// 测试应用鉴权配置
	appConfig := NewConfig("test_app_id", "test_app_secret")
	err := appConfig.WriteConfig2File(configPath)
	assert.NoError(t, err)

	readConfig, err := ReadConfigFromFile(configPath)
	assert.NoError(t, err)
	assert.Equal(t, "test_app_id", readConfig.Feishu.AppId)
	assert.Equal(t, "test_app_secret", readConfig.Feishu.AppSecret)
	assert.Equal(t, AuthTypeApp, readConfig.Feishu.AuthType)

	// 测试用户鉴权配置
	userConfig := NewConfig("", "")
	userConfig.Feishu.UserAccessToken = "u-test-token"
	userConfig.Feishu.AuthType = AuthTypeUser
	err = userConfig.WriteConfig2File(configPath)
	assert.NoError(t, err)

	readConfig2, err := ReadConfigFromFile(configPath)
	assert.NoError(t, err)
	assert.Equal(t, "u-test-token", readConfig2.Feishu.UserAccessToken)
	assert.Equal(t, AuthTypeUser, readConfig2.Feishu.AuthType)
}

func TestConfigMigrate(t *testing.T) {
	tests := []struct {
		name         string
		config       Config
		wantMigrated bool
		wantVersion  string
		wantAuthType string
	}{
		{
			name: "新配置无需迁移",
			config: Config{
				Version: ConfigVersion,
				Feishu: FeishuConfig{
					AppId:     "test_id",
					AppSecret: "test_secret",
					AuthType:  AuthTypeApp,
				},
			},
			wantMigrated: false,
			wantVersion:  ConfigVersion,
			wantAuthType: AuthTypeApp,
		},
		{
			name: "旧配置需要迁移版本号",
			config: Config{
				Version: "",
				Feishu: FeishuConfig{
					AppId:     "test_id",
					AppSecret: "test_secret",
					AuthType:  AuthTypeApp,
				},
			},
			wantMigrated: true,
			wantVersion:  ConfigVersion,
			wantAuthType: AuthTypeApp,
		},
		{
			name: "旧配置需要迁移AuthType",
			config: Config{
				Version: ConfigVersion,
				Feishu: FeishuConfig{
					AppId:     "test_id",
					AppSecret: "test_secret",
					AuthType:  "",
				},
			},
			wantMigrated: true,
			wantVersion:  ConfigVersion,
			wantAuthType: AuthTypeApp,
		},
		{
			name: "旧配置需要迁移版本号和AuthType",
			config: Config{
				Version: "",
				Feishu: FeishuConfig{
					AppId:     "test_id",
					AppSecret: "test_secret",
					AuthType:  "",
				},
			},
			wantMigrated: true,
			wantVersion:  ConfigVersion,
			wantAuthType: AuthTypeApp,
		},
		{
			name: "用户鉴权配置无需迁移AuthType",
			config: Config{
				Version: "",
				Feishu: FeishuConfig{
					UserAccessToken: "u-test",
					AuthType:        AuthTypeUser,
				},
			},
			wantMigrated: true,
			wantVersion:  ConfigVersion,
			wantAuthType: AuthTypeUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migrated := tt.config.Migrate()
			assert.Equal(t, tt.wantMigrated, migrated)
			assert.Equal(t, tt.wantVersion, tt.config.Version)
			assert.Equal(t, tt.wantAuthType, tt.config.Feishu.AuthType)
		})
	}
}

func TestNewConfigHasVersion(t *testing.T) {
	config := NewConfig("test_id", "test_secret")
	assert.Equal(t, ConfigVersion, config.Version)
	assert.Equal(t, AuthTypeApp, config.Feishu.AuthType)
}
