package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	SyncConfigFileName = ".feishu2md.sync.json"
	SyncConfigVersion  = "1.0"

	// 源类型常量
	SourceTypeFolder = "folder"
	SourceTypeWiki   = "wiki"
)

// SyncConfig 同步配置，保存在输出目录下
type SyncConfig struct {
	Version     string    `json:"version"`
	SourceURL   string    `json:"source_url"`
	SourceType  string    `json:"source_type"` // "folder" | "wiki"
	Include     []string  `json:"include,omitempty"`
	Exclude     []string  `json:"exclude,omitempty"`
	Concurrency int       `json:"concurrency"`
	LastSync    time.Time `json:"last_sync"`
}

// NewSyncConfig 创建新的同步配置
func NewSyncConfig(sourceURL, sourceType string) *SyncConfig {
	return &SyncConfig{
		Version:     SyncConfigVersion,
		SourceURL:   sourceURL,
		SourceType:  sourceType,
		Include:     nil,
		Exclude:     nil,
		Concurrency: 5,
		LastSync:    time.Time{},
	}
}

// GetSyncConfigPath 获取同步配置文件路径
func GetSyncConfigPath(outputDir string) string {
	return filepath.Join(outputDir, SyncConfigFileName)
}

// LoadSyncConfig 从输出目录加载同步配置
func LoadSyncConfig(outputDir string) (*SyncConfig, error) {
	configPath := GetSyncConfigPath(outputDir)

	file, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // 配置不存在，返回 nil
		}
		return nil, err
	}

	var config SyncConfig
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Save 保存同步配置到输出目录
func (c *SyncConfig) Save(outputDir string) error {
	// 确保目录存在
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	// 更新最后同步时间
	c.LastSync = time.Now()

	configPath := GetSyncConfigPath(outputDir)

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0o644)
}

// Update 更新同步配置（合并命令行参数）
func (c *SyncConfig) Update(include, exclude []string, concurrency int) {
	if len(include) > 0 {
		c.Include = include
	}
	if len(exclude) > 0 {
		c.Exclude = exclude
	}
	if concurrency > 0 {
		c.Concurrency = concurrency
	}
}
