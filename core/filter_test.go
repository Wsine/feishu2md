package core

import (
	"path/filepath"
	"testing"
)

func TestParsePatterns(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", nil},
		{"test", []string{"test"}},
		{"test,demo", []string{"test", "demo"}},
		{"test, demo, ", []string{"test", "demo"}},
		{"*测试*,*草稿*", []string{"*测试*", "*草稿*"}},
		{"  spaced  ,  patterns  ", []string{"spaced", "patterns"}},
	}

	for _, tt := range tests {
		result := ParsePatterns(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("ParsePatterns(%q) = %v, want %v", tt.input, result, tt.expected)
			continue
		}
		for i, v := range result {
			if v != tt.expected[i] {
				t.Errorf("ParsePatterns(%q)[%d] = %q, want %q", tt.input, i, v, tt.expected[i])
			}
		}
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    bool
	}{
		{"test", "test", true},
		{"testing", "test*", true},
		{"mytest", "*test", true},
		{"mytesting", "*test*", true},
		{"测试文档", "*测试*", true},
		{"文档", "*测试*", false},
		{"draft-v1", "draft-*", true},
		{"abc", "???", true},
		{"ab", "???", false},
		{"file.txt", "*.txt", true},
		{"file.md", "*.txt", false},
		{"doc", "doc", true},
		{"docs", "doc", false},
	}

	for _, tt := range tests {
		got := matchPattern(tt.name, tt.pattern)
		if got != tt.want {
			t.Errorf("matchPattern(%q, %q) = %v, want %v",
				tt.name, tt.pattern, got, tt.want)
		}
	}
}

func TestMatchAnyPattern(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		want     bool
	}{
		{"test", []string{"test", "demo"}, true},
		{"demo", []string{"test", "demo"}, true},
		{"other", []string{"test", "demo"}, false},
		{"testing", []string{"test*", "demo*"}, true},
		{"文档", []string{"*测试*", "*文档*"}, true},
		{"any", []string{}, false},
	}

	for _, tt := range tests {
		got := matchAnyPattern(tt.name, tt.patterns)
		if got != tt.want {
			t.Errorf("matchAnyPattern(%q, %v) = %v, want %v",
				tt.name, tt.patterns, got, tt.want)
		}
	}
}

func TestNodeFilter_ShouldIncludeNode_IncludeOnly(t *testing.T) {
	filter := NewNodeFilter(FilterConfig{
		IncludePatterns: []string{"*文档*", "API*"},
	})

	tests := []struct {
		parent  string
		name    string
		include bool
	}{
		{"", "技术文档", true},
		{"", "API指南", true},
		{"", "草稿", false},
		{"", "其他资料", false},
	}

	for _, tt := range tests {
		include, _ := filter.ShouldIncludeNode(tt.parent, tt.name)
		if include != tt.include {
			t.Errorf("ShouldIncludeNode(%q, %q) = %v, want %v",
				tt.parent, tt.name, include, tt.include)
		}
	}
}

func TestNodeFilter_ShouldIncludeNode_ExcludeOnly(t *testing.T) {
	filter := NewNodeFilter(FilterConfig{
		ExcludePatterns: []string{"*草稿*", "*测试*"},
	})

	tests := []struct {
		parent  string
		name    string
		include bool
	}{
		{"", "正式文档", true},
		{"", "测试文档", false},
		{"", "草稿", false},
		{"", "产品草稿", false},
	}

	for _, tt := range tests {
		include, _ := filter.ShouldIncludeNode(tt.parent, tt.name)
		if include != tt.include {
			t.Errorf("ShouldIncludeNode(%q, %q) = %v, want %v",
				tt.parent, tt.name, include, tt.include)
		}
	}
}

func TestNodeFilter_ShouldIncludeNode_IncludeAndExclude(t *testing.T) {
	filter := NewNodeFilter(FilterConfig{
		IncludePatterns: []string{"*文档*"},
		ExcludePatterns: []string{"*草稿*"},
	})

	tests := []struct {
		parent  string
		name    string
		include bool
	}{
		{"", "技术文档", true},
		{"", "草稿文档", false}, // 被 exclude 排除
		{"", "其他资料", false}, // 不匹配 include
	}

	for _, tt := range tests {
		include, _ := filter.ShouldIncludeNode(tt.parent, tt.name)
		if include != tt.include {
			t.Errorf("ShouldIncludeNode(%q, %q) = %v, want %v",
				tt.parent, tt.name, include, tt.include)
		}
	}
}

func TestNodeFilter_ShouldIncludeNode_ParentExcluded(t *testing.T) {
	filter := NewNodeFilter(FilterConfig{
		ExcludePatterns: []string{"草稿"},
	})

	// 排除"草稿"目录
	include, skipped := filter.ShouldIncludeNode("wiki", "草稿")
	if include {
		t.Error("Expected '草稿' to be excluded")
	}
	if skipped {
		t.Error("Expected skippedByParent to be false for first exclusion")
	}

	// 子目录应该因父目录被排除而跳过
	// 使用 filepath.Join 确保路径分隔符一致
	parentPath := filepath.Join("wiki", "草稿")
	include, skipped = filter.ShouldIncludeNode(parentPath, "子文件夹")
	if include {
		t.Error("Expected '子文件夹' to be excluded because parent is excluded")
	}
	if !skipped {
		t.Error("Expected skippedByParent to be true")
	}

	// 深层嵌套也应该被跳过
	deepParentPath := filepath.Join("wiki", "草稿", "子文件夹")
	include, skipped = filter.ShouldIncludeNode(deepParentPath, "深层文件夹")
	if include {
		t.Error("Expected '深层文件夹' to be excluded because ancestor is excluded")
	}
	if !skipped {
		t.Error("Expected skippedByParent to be true for deeply nested")
	}
}

func TestNodeFilter_NoFilters(t *testing.T) {
	filter := NewNodeFilter(FilterConfig{})

	include, _ := filter.ShouldIncludeNode("", "任意目录")
	if !include {
		t.Error("Expected all nodes to be included when no filters")
	}

	if filter.HasFilters() {
		t.Error("Expected HasFilters() to return false")
	}
}

func TestNodeFilter_HasFilters(t *testing.T) {
	tests := []struct {
		config FilterConfig
		want   bool
	}{
		{FilterConfig{}, false},
		{FilterConfig{IncludePatterns: []string{"test"}}, true},
		{FilterConfig{ExcludePatterns: []string{"test"}}, true},
		{FilterConfig{IncludePatterns: []string{"a"}, ExcludePatterns: []string{"b"}}, true},
	}

	for _, tt := range tests {
		filter := NewNodeFilter(tt.config)
		if got := filter.HasFilters(); got != tt.want {
			t.Errorf("HasFilters() with %+v = %v, want %v", tt.config, got, tt.want)
		}
	}
}

func TestNodeFilter_ShouldDownloadFolder(t *testing.T) {
	filter := NewNodeFilter(FilterConfig{
		ExcludePatterns: []string{"*draft*"},
	})

	if !filter.ShouldDownloadFolder("", "docs") {
		t.Error("Expected 'docs' folder to be downloadable")
	}

	if filter.ShouldDownloadFolder("", "draft-v1") {
		t.Error("Expected 'draft-v1' folder to be excluded")
	}
}

func TestNodeFilter_ShouldDownloadDocument(t *testing.T) {
	filter := NewNodeFilter(FilterConfig{
		ExcludePatterns: []string{"草稿"},
	})

	// 排除"草稿"目录
	filter.ShouldIncludeNode("wiki", "草稿")

	// 文档在正常路径下应该可以下载
	normalPath := filepath.Join("wiki", "正式")
	if !filter.ShouldDownloadDocument(normalPath) {
		t.Error("Expected document in 'wiki/正式' to be downloadable")
	}

	// 文档在被排除的路径下不应该下载
	excludedPath := filepath.Join("wiki", "草稿")
	if filter.ShouldDownloadDocument(excludedPath) {
		t.Error("Expected document in 'wiki/草稿' to be excluded")
	}
}

func TestNodeFilter_Reset(t *testing.T) {
	filter := NewNodeFilter(FilterConfig{
		ExcludePatterns: []string{"test"},
	})

	// 排除一个目录
	filter.ShouldIncludeNode("", "test")

	// 子目录应该被跳过（因为父目录被排除）
	include, skipped := filter.ShouldIncludeNode("test", "child")
	if include {
		t.Error("Expected child to be excluded before reset")
	}
	if !skipped {
		t.Error("Expected child to be skipped by parent before reset")
	}

	// 重置
	filter.Reset()

	// 重置后，"test" 目录仍然会被排除（因为匹配规则），但不是因为缓存
	include, _ = filter.ShouldIncludeNode("", "test")
	if include {
		t.Error("Expected 'test' to still be excluded by pattern after reset")
	}

	// 现在 "test/child" 会因为父目录被排除而跳过
	include, skipped = filter.ShouldIncludeNode("test", "child")
	if include {
		t.Error("Expected child to be excluded because parent matches exclude pattern")
	}
	if !skipped {
		t.Error("Expected child to be skipped by parent after reset")
	}
}

func TestNodeFilter_ShouldDownloadDocument_WithIncludePatterns(t *testing.T) {
	filter := NewNodeFilter(FilterConfig{
		IncludePatterns: []string{"*管理与规范*"},
	})

	// 模拟 Wiki 结构:
	// wiki/                       <- 根目录
	// ├── 首页                    <- 不匹配，被排除
	// ├── 测试管理                <- 不匹配，被排除
	// ├── 测试管理与规范          <- 匹配，被包含
	// └── doc.md                  <- 在根目录下的文档

	// 处理目录
	include, _ := filter.ShouldIncludeNode("wiki", "首页")
	if include {
		t.Error("Expected '首页' to be excluded")
	}

	include, _ = filter.ShouldIncludeNode("wiki", "测试管理")
	if include {
		t.Error("Expected '测试管理' to be excluded")
	}

	include, _ = filter.ShouldIncludeNode("wiki", "测试管理与规范")
	if !include {
		t.Error("Expected '测试管理与规范' to be included")
	}

	// 测试文档下载
	// 根目录下的文档不应该下载（因为配置了 Include 白名单，但根目录不在 includedPaths 中）
	if filter.ShouldDownloadDocument("wiki") {
		t.Error("Expected document in root 'wiki' to NOT be downloaded when include patterns are set")
	}

	// 被排除目录下的文档不应该下载
	excludedPath := filepath.Join("wiki", "首页")
	if filter.ShouldDownloadDocument(excludedPath) {
		t.Error("Expected document in excluded path to NOT be downloaded")
	}

	// 被包含目录下的文档应该下载
	includedPath := filepath.Join("wiki", "测试管理与规范")
	if !filter.ShouldDownloadDocument(includedPath) {
		t.Error("Expected document in included path '测试管理与规范' to be downloaded")
	}

	// 被包含目录的子目录下的文档也应该下载
	nestedPath := filepath.Join("wiki", "测试管理与规范", "子目录")
	if !filter.ShouldDownloadDocument(nestedPath) {
		t.Error("Expected document in nested path under included directory to be downloaded")
	}
}

func TestNodeFilter_ShouldDownloadDocument_NoFilters(t *testing.T) {
	filter := NewNodeFilter(FilterConfig{})

	// 无过滤条件时，所有文档都应该下载
	if !filter.ShouldDownloadDocument("wiki") {
		t.Error("Expected document to be downloadable when no filters")
	}

	if !filter.ShouldDownloadDocument(filepath.Join("wiki", "any", "path")) {
		t.Error("Expected document to be downloadable when no filters")
	}
}

func TestNodeFilter_ShouldDownloadDocument_ExcludeOnly(t *testing.T) {
	filter := NewNodeFilter(FilterConfig{
		ExcludePatterns: []string{"草稿"},
	})

	// 排除"草稿"目录
	filter.ShouldIncludeNode("wiki", "草稿")
	filter.ShouldIncludeNode("wiki", "正式")

	// 只配置 Exclude 时，正常路径下的文档应该可以下载
	normalPath := filepath.Join("wiki", "正式")
	if !filter.ShouldDownloadDocument(normalPath) {
		t.Error("Expected document in non-excluded path to be downloadable")
	}

	// 被排除路径下的文档不应该下载
	excludedPath := filepath.Join("wiki", "草稿")
	if filter.ShouldDownloadDocument(excludedPath) {
		t.Error("Expected document in excluded path to NOT be downloaded")
	}

	// 根目录下的文档应该可以下载（因为没有配置 Include 白名单）
	if !filter.ShouldDownloadDocument("wiki") {
		t.Error("Expected document in root to be downloadable when only exclude patterns are set")
	}
}

func TestNodeFilter_ChildDirectoriesAutoIncluded(t *testing.T) {
	filter := NewNodeFilter(FilterConfig{
		IncludePatterns: []string{"*测试*"},
	})

	// 模拟 Wiki 结构:
	// wiki/
	// ├── 测试管理/              <- 匹配 *测试*，被包含
	// │   ├── 能力建设规划/      <- 不匹配 *测试*，但父目录已包含，应该自动包含
	// │   │   └── doc.md
	// │   └── 子目录A/
	// │       └── 深层目录/
	// │           └── doc.md
	// └── 其他目录/              <- 不匹配，被排除

	// 一级目录：测试管理 匹配
	include, _ := filter.ShouldIncludeNode("wiki", "测试管理")
	if !include {
		t.Error("Expected '测试管理' to be included (matches pattern)")
	}

	// 一级目录：其他目录 不匹配
	include, _ = filter.ShouldIncludeNode("wiki", "其他目录")
	if include {
		t.Error("Expected '其他目录' to be excluded (doesn't match pattern)")
	}

	// 二级目录：能力建设规划 不匹配模式，但父目录已包含
	parentPath := filepath.Join("wiki", "测试管理")
	include, _ = filter.ShouldIncludeNode(parentPath, "能力建设规划")
	if !include {
		t.Error("Expected '能力建设规划' to be auto-included (parent is included)")
	}

	// 三级目录：深层目录 也应该自动包含
	deepParentPath := filepath.Join("wiki", "测试管理", "能力建设规划")
	include, _ = filter.ShouldIncludeNode(deepParentPath, "深层目录")
	if !include {
		t.Error("Expected '深层目录' to be auto-included (ancestor is included)")
	}

	// 验证文档下载
	// 被包含目录的子目录下的文档应该可以下载
	docPath := filepath.Join("wiki", "测试管理", "能力建设规划")
	if !filter.ShouldDownloadDocument(docPath) {
		t.Error("Expected document in auto-included directory to be downloadable")
	}

	deepDocPath := filepath.Join("wiki", "测试管理", "能力建设规划", "深层目录")
	if !filter.ShouldDownloadDocument(deepDocPath) {
		t.Error("Expected document in deeply nested auto-included directory to be downloadable")
	}

	// 被排除目录下的文档不应该下载
	excludedDocPath := filepath.Join("wiki", "其他目录")
	if filter.ShouldDownloadDocument(excludedDocPath) {
		t.Error("Expected document in excluded directory to NOT be downloadable")
	}
}
