package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CacheVersion 缓存文件格式版本
const CacheVersion = "1.0"

// DocumentCache 单个文档的缓存信息
type DocumentCache struct {
	RevisionID   int64     `json:"revision_id"`   // 文档版本号
	Title        string    `json:"title"`         // 文档标题
	FileName     string    `json:"file_name"`     // 实际保存的文件名
	LastDownload time.Time `json:"last_download"` // 上次下载时间
	DocType      string    `json:"doc_type"`      // 文档类型 (docx/wiki)
}

// CacheManager 缓存管理器
type CacheManager struct {
	Version   string                    `json:"version"`    // 缓存格式版本
	UpdatedAt time.Time                 `json:"updated_at"` // 缓存更新时间
	Documents map[string]*DocumentCache `json:"documents"`  // 文档token -> 缓存信息映射

	filePath string       // 缓存文件路径
	mutex    sync.RWMutex // 读写锁保护并发访问
	dirty    bool         // 标记是否有修改未保存
}

// NewCacheManager 创建新的缓存管理器
func NewCacheManager(outputDir string) (*CacheManager, error) {
	cachePath := filepath.Join(outputDir, ".feishu2md.cache.json")

	cm := &CacheManager{
		Version:   CacheVersion,
		UpdatedAt: time.Now(),
		Documents: make(map[string]*DocumentCache),
		filePath:  cachePath,
		dirty:     false,
	}

	// 尝试加载现有缓存
	if err := cm.Load(); err != nil && !os.IsNotExist(err) {
		// 缓存文件损坏或无法读取，返回空缓存但不报错
		return cm, nil
	}

	return cm, nil
}

// Load 从文件加载缓存
func (cm *CacheManager) Load() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	data, err := os.ReadFile(cm.filePath)
	if err != nil {
		return err
	}

	// 尝试解析JSON
	if err := json.Unmarshal(data, cm); err != nil {
		return err
	}

	// 版本兼容性检查
	if cm.Version != CacheVersion {
		// 未来可能需要处理版本迁移
		cm.Version = CacheVersion
		cm.dirty = true
	}

	cm.dirty = false
	return nil
}

// Save 保存缓存到文件
func (cm *CacheManager) Save() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if !cm.dirty {
		return nil // 没有修改不需要保存
	}

	cm.UpdatedAt = time.Now()

	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(cm.filePath), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cm, "", "  ")
	if err != nil {
		return err
	}

	// 原子写入：先写临时文件再重命名
	tmpPath := cm.filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, cm.filePath); err != nil {
		os.Remove(tmpPath) // 清理临时文件
		return err
	}

	cm.dirty = false
	return nil
}

// ShouldDownload 判断文档是否需要下载
// 返回: (需要下载, 跳过原因)
func (cm *CacheManager) ShouldDownload(
	docToken string,
	remoteRevisionID int64,
	fileName string,
) (bool, string) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	cache, exists := cm.Documents[docToken]

	// 1. 缓存中不存在
	if !exists {
		// 检查文件是否已存在（可能是旧版本下载的）
		if _, err := os.Stat(fileName); err == nil {
			// 文件存在，跳过下载，但外层会更新缓存建立映射
			return false, fmt.Sprintf("文件已存在，建立缓存映射 (版本: %d)", remoteRevisionID)
		}
		// 文件也不存在，需要下载
		return true, ""
	}

	// 2. 版本号不同，需要下载
	if cache.RevisionID != remoteRevisionID {
		return true, ""
	}

	// 3. 版本号相同，检查文件是否存在
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return true, ""
	}

	// 4. 版本相同且文件存在，跳过
	return false, fmt.Sprintf("文档未修改 (版本: %d)", remoteRevisionID)
}

// UpdateDocument 更新文档缓存信息
func (cm *CacheManager) UpdateDocument(
	docToken string,
	revisionID int64,
	title string,
	fileName string,
	docType string,
) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.Documents[docToken] = &DocumentCache{
		RevisionID:   revisionID,
		Title:        title,
		FileName:     fileName,
		LastDownload: time.Now(),
		DocType:      docType,
	}
	cm.dirty = true
}

// GetDocumentCache 获取文档缓存信息（只读）
func (cm *CacheManager) GetDocumentCache(docToken string) (*DocumentCache, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	cache, exists := cm.Documents[docToken]
	return cache, exists
}

// RemoveDocument 从缓存中移除文档
func (cm *CacheManager) RemoveDocument(docToken string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if _, exists := cm.Documents[docToken]; exists {
		delete(cm.Documents, docToken)
		cm.dirty = true
	}
}

// GetStats 获取缓存统计信息
func (cm *CacheManager) GetStats() (totalDocs int, oldestDownload time.Time) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	totalDocs = len(cm.Documents)
	oldestDownload = time.Now()

	for _, doc := range cm.Documents {
		if doc.LastDownload.Before(oldestDownload) {
			oldestDownload = doc.LastDownload
		}
	}

	return
}
