# CLAUDE.md

本文件为 Claude Code (claude.ai/code) 在处理此仓库代码时提供指导。

## 项目概述

**feishu2md** 是一个基于 Go 的工具,用于将飞书(LarkSuite)文档转换为 Markdown 格式。它支持 CLI 和 Web 服务两种部署模式。该项目使用飞书开放 API(通过 `chyroc/lark` SDK)获取文档内容,并将飞书基于块的文档结构转换为 Markdown。

## 开发命令

### 构建
```bash
make build          # 构建 CLI 二进制文件(输出: feishu2md)
make server         # 构建 Web 服务二进制文件(输出: feishu2md4web)
make all            # 构建 CLI 和 Web 二进制文件
```

CLI 二进制文件通过 `-ldflags="-X main.version=v2-<hash>"` 在版本字符串中嵌入 git commit hash。

### 测试
```bash
make test           # 运行所有测试: go test ./...
make format         # 格式化代码: gofmt -l -w .
```

测试需要飞书凭证(FEISHU_APP_ID 和 FEISHU_APP_SECRET 环境变量)。测试固件位于 `testdata/` 目录,包含配对的 JSON(API 响应)和 MD(预期输出)文件。

### Docker
```bash
make image          # 构建 Docker 镜像(标签: feishu2md)
make docker         # 在端口 8080 上运行 Docker 容器
```

Web 服务需要环境变量: `FEISHU_APP_ID`、`FEISHU_APP_SECRET`、`GIN_MODE`。

## 架构

代码库遵循三层架构:

### 1. 接口层(CLI 或 Web)
- **cmd/** - 使用 `urfave/cli/v2` 框架的 CLI 应用
  - `cmd/main.go`: 入口点,包含 config 和 download 命令
  - `cmd/config.go`: 读写配置到 `~/.config/feishu2md/config.json`
  - `cmd/download.go`: 协调单个/批量/wiki 下载,支持并发
- **web/** - 使用 `gin-gonic/gin` 的 Web 服务
  - `web/main.go`: 端口 8080 上的 HTTP 服务器
  - `web/download.go`: `/download?url=<encoded_url>` 处理器,返回 ZIP 或 MD 文件

### 2. 核心业务逻辑
- **core/client.go**: 通过 `chyroc/lark` SDK 封装飞书 API 调用
  - `GetDocxContent()`: 获取文档元数据和所有块(处理分页)
  - `GetWikiNodeInfo()`: 检索 wiki 节点元数据
  - `GetWikiNodeList()`: 递归列出所有子节点(跟踪 `previousPageToken` 以防止无限循环)
  - `GetDriveFolderFileList()`: 分页列出文件夹内容
  - `DownloadImage()`: 通过 token 下载图片到文件系统
- **core/parser.go**: 将飞书块结构转换为 Markdown
  - `Parse()`: 主入口点,构建块映射并递归处理
  - 处理嵌套块(标题、列表、表格、代码、图片等)
  - 在 `parser.ImgTokens` 中收集图片 token 以供批量下载
  - 使用 `github.com/88250/lute` 进行最终的 Markdown 格式化
- **core/config.go**: 配置结构和文件持久化

### 3. 数据流

```
用户输入(URL) → 验证(utils/url.go)
                ↓
          客户端(飞书 API 调用,带速率限制)
                ↓
          解析器(飞书块 → Markdown)
                ↓
          输出(文件或带图片的 ZIP)
```

### 关键架构模式

**分页处理**: 所有返回列表的 API 调用都使用 `HasMore` 和 `PageToken` 字段处理分页。Wiki 节点列表跟踪 `previousPageToken` 以防止无限循环(最近的 bug 修复)。

**速率限制**: 客户端使用 `lark_rate_limiter.Wait(4, 4)` 中间件,限制为每 4 秒 4 个请求。

**并发**:
- 批量文件夹下载: 使用 `sync.WaitGroup` 和 goroutine 并行下载文档
- Wiki 下载: 使用信号量模式,最大并发数为 10,以防止 API 限流

**块解析**: 飞书文档是块的树结构。解析器:
1. 构建 `blockMap`(ID → block)以实现 O(1) 子块查找
2. 递归处理每种块类型(Page → Heading → Text → List 等)
3. 处理文本块内的内联元素(粗体、斜体、链接、公式、提及)
4. 将表格转换为 HTML(支持 colspan/rowspan)
5. 在解析过程中收集图片 token,之后下载

## 重要实现细节

### 配置
- **CLI**: 配置存储在 `~/.config/feishu2md/config.json`(XDG 配置目录)
- **Web**: 环境变量(`FEISHU_APP_ID`、`FEISHU_APP_SECRET`)
- 配置结构包括 `feishu`(凭证)和 `output`(image_dir、title_as_filename、use_html_tags、skip_img_download)

### URL 模式
工具通过 `utils/url.go` 中的正则表达式验证三种 URL 类型:
- 文档: `https://domain.feishu.cn/docx/<token>`
- 文件夹: `https://domain.feishu.cn/drive/folder/<token>`
- Wiki: `https://domain.feishu.cn/wiki/settings/<space_id>`

### 文档块类型
解析器处理以下飞书块类型(参见 `core/parser.go`):
- Page、Heading(h1-h9)、Text、Code、Quote、Todo、Callout
- Bullet/Ordered/Numbered 列表(带嵌套子项)
- Table(渲染为 HTML,使用 tablewriter)
- Image(基于 token,单独下载)
- Divider、Grid(列布局)

内联元素: TextRun(样式文本)、Mention(用户/文档)、Equation(LaTeX)

### 文件名清理
当使用 `title_as_filename` 选项时,`utils.SanitizeFileName()` 会从文档标题中删除无效字符(`/\:*?"<>|`)。

## CI/CD

`.github/workflows/` 中的 GitHub Actions 工作流:
- **unittest.yaml**(PR 时): gofmt 检查、测试、构建验证
- **release.yaml**(发布时): 跨平台构建(Linux/Windows/Darwin amd64+arm64)、Docker Hub 推送(`wwwsine/feishu2md`)

使用 `wangyoucao577/go-release-action` 进行矩阵构建,支持 UPX 压缩。

## 依赖项

主要外部依赖:
- `github.com/chyroc/lark@v0.0.98` - 飞书 SDK(API 封装)
- `github.com/chyroc/lark_rate_limiter@v0.1.0` - 速率限制中间件
- `github.com/urfave/cli/v2@v2.6.0` - CLI 框架
- `github.com/gin-gonic/gin@v1.9.0` - Web 框架
- `github.com/88250/lute@v1.7.3` - Markdown 格式化器
- `github.com/olekukonko/tablewriter@v0.0.5` - ASCII 表格渲染

## 项目状态

社区维护项目(原作者不再使用飞书)。最近的修复包括:
- Wiki 分页无限循环防护(#139)
- 速率限制增加(#137)
- 文件名清理(#117)

欢迎提交 PR,活跃的维护者可能成为协调者。
