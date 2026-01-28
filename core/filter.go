package core

import (
	"path/filepath"
	"strings"
)

// FilterConfig 目录过滤配置
type FilterConfig struct {
	IncludePatterns []string // 包含模式列表（白名单）
	ExcludePatterns []string // 排除模式列表（黑名单）
}

// NodeFilter 节点过滤器
type NodeFilter struct {
	config        FilterConfig
	excludedPaths map[string]bool // 已排除的路径缓存（用于跟踪父目录）
	includedPaths map[string]bool // 已包含的路径缓存（用于 Include 白名单检查）
}

// NewNodeFilter 创建节点过滤器
func NewNodeFilter(config FilterConfig) *NodeFilter {
	return &NodeFilter{
		config:        config,
		excludedPaths: make(map[string]bool),
		includedPaths: make(map[string]bool),
	}
}

// ParsePatterns 解析逗号分隔的模式字符串
func ParsePatterns(patterns string) []string {
	if patterns == "" {
		return nil
	}
	parts := strings.Split(patterns, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// matchPattern 检查名称是否匹配单个模式
// 支持 Go 的 filepath.Match 语法：*, ?, [abc], [a-z]
func matchPattern(name, pattern string) bool {
	// 使用 filepath.Match 支持通配符
	matched, err := filepath.Match(pattern, name)
	if err != nil {
		return false
	}
	return matched
}

// matchAnyPattern 检查名称是否匹配任一模式
func matchAnyPattern(name string, patterns []string) bool {
	for _, pattern := range patterns {
		if matchPattern(name, pattern) {
			return true
		}
	}
	return false
}

// ShouldIncludeNode 判断节点是否应该被包含
// parentPath: 父目录的完整路径（用于检查父目录是否已被排除）
// nodeName: 当前节点名称
// 返回: (是否包含, 是否为因父目录被排除而跳过)
func (f *NodeFilter) ShouldIncludeNode(parentPath, nodeName string) (include bool, skippedByParent bool) {
	// 规范化路径，确保路径分隔符一致
	parentPath = filepath.Clean(parentPath)
	currentPath := filepath.Join(parentPath, nodeName)

	// 检查父目录是否已被排除
	if f.isParentExcluded(parentPath) {
		f.excludedPaths[currentPath] = true
		return false, true
	}

	// 无过滤条件时，默认包含
	if len(f.config.IncludePatterns) == 0 && len(f.config.ExcludePatterns) == 0 {
		return true, false
	}

	// 如果父目录已经被包含，子目录自动被包含（不再检查 include 模式）
	// 但仍然检查 exclude 模式
	parentIncluded := f.isParentIncluded(parentPath)

	// 检查 include（白名单）
	if len(f.config.IncludePatterns) > 0 {
		// 如果父目录已包含，则子目录自动包含；否则检查当前节点是否匹配
		if !parentIncluded && !matchAnyPattern(nodeName, f.config.IncludePatterns) {
			f.excludedPaths[currentPath] = true
			return false, false
		}
	}

	// 再检查 exclude（黑名单）- 即使父目录被包含，也要检查排除规则
	if len(f.config.ExcludePatterns) > 0 {
		if matchAnyPattern(nodeName, f.config.ExcludePatterns) {
			f.excludedPaths[currentPath] = true
			return false, false
		}
	}

	// 记录已包含的路径
	f.includedPaths[currentPath] = true
	return true, false
}

// isParentExcluded 检查父路径是否已被排除
func (f *NodeFilter) isParentExcluded(path string) bool {
	if path == "" || path == "." {
		return false
	}

	// 规范化路径
	path = filepath.Clean(path)

	// 递归检查所有父路径
	current := path
	for current != "" && current != "." {
		if f.excludedPaths[current] {
			return true
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return false
}

// isParentIncluded 检查路径或其祖先路径是否被包含
func (f *NodeFilter) isParentIncluded(path string) bool {
	if path == "" || path == "." {
		return false
	}

	path = filepath.Clean(path)
	current := path
	for current != "" && current != "." {
		if f.includedPaths[current] {
			return true
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return false
}

// ShouldDownloadFolder 判断文件夹是否应该下载（用于 batch 模式）
func (f *NodeFilter) ShouldDownloadFolder(parentPath, folderName string) bool {
	include, _ := f.ShouldIncludeNode(parentPath, folderName)
	return include
}

// ShouldDownloadDocument 判断文档是否应该下载
// 文档本身不受过滤影响，只有其父目录被排除时才跳过
// 当配置了 Include 白名单时，只有父目录在已包含路径中的文档才会被下载
func (f *NodeFilter) ShouldDownloadDocument(parentPath string) bool {
	parentPath = filepath.Clean(parentPath)

	// 检查父目录是否被排除
	if f.isParentExcluded(parentPath) || f.excludedPaths[parentPath] {
		return false
	}

	// 如果配置了 Include 白名单，检查父目录是否在已包含的路径中
	if len(f.config.IncludePatterns) > 0 {
		return f.isParentIncluded(parentPath)
	}

	return true
}

// HasFilters 检查是否配置了过滤条件
func (f *NodeFilter) HasFilters() bool {
	return len(f.config.IncludePatterns) > 0 || len(f.config.ExcludePatterns) > 0
}

// Reset 重置过滤器状态（用于新的下载任务）
func (f *NodeFilter) Reset() {
	f.excludedPaths = make(map[string]bool)
	f.includedPaths = make(map[string]bool)
}
