package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSyncConfig(t *testing.T) {
	config := NewSyncConfig("https://example.feishu.cn/wiki/xxx", SourceTypeWiki)

	assert.Equal(t, SyncConfigVersion, config.Version)
	assert.Equal(t, "https://example.feishu.cn/wiki/xxx", config.SourceURL)
	assert.Equal(t, SourceTypeWiki, config.SourceType)
	assert.Equal(t, 5, config.Concurrency)
	assert.Nil(t, config.Include)
	assert.Nil(t, config.Exclude)
	assert.True(t, config.LastSync.IsZero())
}

func TestSyncConfigSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建并保存配置
	config := NewSyncConfig("https://example.feishu.cn/wiki/xxx", SourceTypeWiki)
	config.Include = []string{"*文档*", "API*"}
	config.Exclude = []string{"*草稿*"}
	config.Concurrency = 10

	err := config.Save(tmpDir)
	assert.NoError(t, err)

	// 验证文件存在
	configPath := GetSyncConfigPath(tmpDir)
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// 加载配置
	loaded, err := LoadSyncConfig(tmpDir)
	assert.NoError(t, err)
	assert.NotNil(t, loaded)

	assert.Equal(t, SyncConfigVersion, loaded.Version)
	assert.Equal(t, "https://example.feishu.cn/wiki/xxx", loaded.SourceURL)
	assert.Equal(t, SourceTypeWiki, loaded.SourceType)
	assert.Equal(t, []string{"*文档*", "API*"}, loaded.Include)
	assert.Equal(t, []string{"*草稿*"}, loaded.Exclude)
	assert.Equal(t, 10, loaded.Concurrency)
	assert.False(t, loaded.LastSync.IsZero())
}

func TestLoadSyncConfig_NotExist(t *testing.T) {
	tmpDir := t.TempDir()

	// 加载不存在的配置
	config, err := LoadSyncConfig(tmpDir)
	assert.NoError(t, err)
	assert.Nil(t, config)
}

func TestSyncConfigUpdate(t *testing.T) {
	config := NewSyncConfig("https://example.feishu.cn/wiki/xxx", SourceTypeWiki)

	// 更新配置
	config.Update(
		[]string{"include1", "include2"},
		[]string{"exclude1"},
		8,
	)

	assert.Equal(t, []string{"include1", "include2"}, config.Include)
	assert.Equal(t, []string{"exclude1"}, config.Exclude)
	assert.Equal(t, 8, config.Concurrency)
}

func TestSyncConfigUpdate_EmptyValues(t *testing.T) {
	config := NewSyncConfig("https://example.feishu.cn/wiki/xxx", SourceTypeWiki)
	config.Include = []string{"original"}
	config.Exclude = []string{"original"}
	config.Concurrency = 5

	// 空值不应该覆盖现有值
	config.Update(nil, nil, 0)

	assert.Equal(t, []string{"original"}, config.Include)
	assert.Equal(t, []string{"original"}, config.Exclude)
	assert.Equal(t, 5, config.Concurrency)
}

func TestSyncConfigUpdate_PartialUpdate(t *testing.T) {
	config := NewSyncConfig("https://example.feishu.cn/wiki/xxx", SourceTypeWiki)
	config.Include = []string{"original_include"}
	config.Exclude = []string{"original_exclude"}

	// 只更新 include
	config.Update([]string{"new_include"}, nil, 0)

	assert.Equal(t, []string{"new_include"}, config.Include)
	assert.Equal(t, []string{"original_exclude"}, config.Exclude)
}

func TestGetSyncConfigPath(t *testing.T) {
	path := GetSyncConfigPath("/tmp/output")
	expected := filepath.Join("/tmp/output", SyncConfigFileName)
	assert.Equal(t, expected, path)
}

func TestSyncConfigSave_UpdatesLastSync(t *testing.T) {
	tmpDir := t.TempDir()

	config := NewSyncConfig("https://example.feishu.cn/wiki/xxx", SourceTypeWiki)
	assert.True(t, config.LastSync.IsZero())

	before := time.Now()
	err := config.Save(tmpDir)
	assert.NoError(t, err)
	after := time.Now()

	// LastSync 应该被更新到当前时间
	assert.False(t, config.LastSync.IsZero())
	assert.True(t, config.LastSync.After(before) || config.LastSync.Equal(before))
	assert.True(t, config.LastSync.Before(after) || config.LastSync.Equal(after))
}

func TestSyncConfigSourceTypes(t *testing.T) {
	// 测试 Wiki 类型
	wikiConfig := NewSyncConfig("https://example.feishu.cn/wiki/xxx", SourceTypeWiki)
	assert.Equal(t, SourceTypeWiki, wikiConfig.SourceType)

	// 测试 Folder 类型
	folderConfig := NewSyncConfig("https://example.feishu.cn/drive/folder/xxx", SourceTypeFolder)
	assert.Equal(t, SourceTypeFolder, folderConfig.SourceType)
}
