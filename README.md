# feishu2md

[![Golang - feishu2md](https://img.shields.io/github/go-mod/go-version/wsine/feishu2md?color=%2376e1fe&logo=go)](https://go.dev/)
[![Unittest](https://github.com/Wsine/feishu2md/actions/workflows/unittest.yaml/badge.svg)](https://github.com/Wsine/feishu2md/actions/workflows/unittest.yaml)
[![Release](https://img.shields.io/github/v/release/wsine/feishu2md?color=orange&logo=github)](https://github.com/Wsine/feishu2md/releases)
[![Docker - feishu2md](https://img.shields.io/badge/Docker-feishu2md-2496ed?logo=docker&logoColor=white)](https://hub.docker.com/r/wwwsine/feishu2md)
[![Render - feishu2md](https://img.shields.io/badge/Render-feishu2md-4cfac9?logo=render&logoColor=white)](https://feishu2md.onrender.com)
![Last Review](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fbadge-last-review.wsine.workers.dev%2FWsine%2Ffeishu2md&query=%24.reviewed_at&label=last%20review)

这是一个下载飞书文档为 Markdown 文件的工具，使用 Go 语言实现。

**请看这里：由于原作者已不再使用飞书文档，项目转为社区维护，欢迎 PR，有能力的维护者会被选择为主协调员。**

## 动机

[《一日一技 | 我开发的这款小工具，轻松助你将飞书文档转为 Markdown》](https://sspai.com/post/73386)

## 获取 API Token

配置文件需要填写 APP ID 和 APP SECRET 信息，请参考 [飞书官方文档](https://open.feishu.cn/document/ukTMukTMukTM/ukDNz4SO0MjL5QzM/get-) 获取。推荐设置为

- 进入飞书[开发者后台](https://open.feishu.cn/app)
- 创建企业自建应用（个人版），信息随意填写
- （重要）打开权限管理，开通以下必要的权限（可点击以下链接参考 API 调试台->权限配置字段）
  - [获取文档基本信息](https://open.feishu.cn/document/server-docs/docs/docs/docx-v1/document/get)，「查看新版文档」权限 `docx:document:readonly`
  - [获取文档所有块](https://open.feishu.cn/document/server-docs/docs/docs/docx-v1/document/list)，「查看新版文档」权限 `docx:document:readonly`
  - [下载素材](https://open.feishu.cn/document/server-docs/docs/drive-v1/media/download)，「下载云文档中的图片和附件」权限 `docs:document.media:download`
  - [获取文件夹中的文件清单](https://open.feishu.cn/document/server-docs/docs/drive-v1/folder/list)，「查看、评论、编辑和管理云空间中所有文件」权限 `drive:file:readonly`
  - [获取知识空间节点信息](https://open.feishu.cn/document/server-docs/docs/wiki-v2/space-node/get_node)，「查看知识库」权限 `wiki:wiki:readonly`
- 打开凭证与基础信息，获取 App ID 和 App Secret

## 鉴权方式

feishu2md 支持两种鉴权方式：**应用鉴权**和**用户鉴权**。

### 应用鉴权（默认）

使用飞书应用的 App ID 和 App Secret 进行鉴权。

**适用场景**：
- 批量下载多个文档
- 长期使用的自动化任务
- 访问应用有权限的文档

**配置方法**：
```bash
feishu2md config --appId "cli_xxxxx" --appSecret "xxxxx" --authType "app"
```

**配置文件示例**：
```json
{
  "feishu": {
    "app_id": "cli_xxxxx",
    "app_secret": "xxxxx",
    "auth_type": "app"
  },
  "output": {
    "image_dir": "static",
    "title_as_filename": false,
    "use_html_tags": false,
    "skip_img_download": false
  }
}
```

### 用户鉴权

使用个人用户访问令牌（User Access Token）进行鉴权。

**适用场景**：
- 访问个人私有文档
- 机器人权限不足时的替代方案
- 临时下载任务

**配置方法**：
```bash
# 设置用户访问令牌和鉴权类型
feishu2md config --userAccessToken "u-xxxxx" --authType "user"

# 或使用简写
feishu2md config --uat "u-xxxxx" --authType "user"
```

**配置文件示例**：
```json
{
  "feishu": {
    "user_access_token": "u-xxxxx",
    "auth_type": "user"
  },
  "output": {
    "image_dir": "static",
    "title_as_filename": false,
    "use_html_tags": false,
    "skip_img_download": false
  }
}
```

**获取用户访问令牌**：

用户访问令牌需要通过飞书开放平台的 OAuth 2.0 授权流程获取。详细步骤请参考：
- [飞书开放平台 - 获取 user_access_token](https://open.feishu.cn/document/server-docs/api-call-guide/calling-process/get-access-token)

**注意事项**：
- 用户访问令牌有效期较短（通常为 2 小时），过期后需要重新获取
- 令牌过期时，使用 `feishu2md config --uat "new-token"` 更新配置
- 不要将包含令牌的配置文件提交到版本控制系统

### 切换鉴权方式

配置文件可以同时保存两种鉴权方式的凭证，通过 `auth_type` 字段灵活切换：

```bash
# 切换到用户鉴权
feishu2md config --authType "user"

# 切换回应用鉴权
feishu2md config --authType "app"

# 查看当前配置
feishu2md config
```

## 如何使用

注意：飞书旧版文档的下载工具已决定不再维护，但分支 [v1_support](https://github.com/Wsine/feishu2md/tree/v1_support) 仍可使用，对应的归档为 [v1.4.0](https://github.com/Wsine/feishu2md/releases/tag/v1.4.0)，请知悉。

<details>
  <summary>命令行版本</summary>

  借助 Go 语言跨平台的特性，已编译好了主要平台的可执行文件，可以在 [Release](https://github.com/Wsine/feishu2md/releases) 中下载，并将相应平台的 feishu2md 可执行文件放置在 PATH 路径中即可。

   **查阅帮助文档**

   ```bash
   $ feishu2md -h
   NAME:
     feishu2md - Download feishu/larksuite document to markdown file

   USAGE:
     feishu2md [global options] command [command options] [arguments...]

   VERSION:
     v2-0e25fa5

   COMMANDS:
     config        Read config file or set field(s) if provided
     download, dl  Download feishu/larksuite document to markdown file
     help, h       Shows a list of commands or help for one command

   GLOBAL OPTIONS:
     --help, -h     show help (default: false)
     --version, -v  print the version (default: false)

   $ feishu2md config -h
   NAME:
      feishu2md config - Read config file or set field(s) if provided

   USAGE:
      feishu2md config [command options] [arguments...]

   OPTIONS:
      --appId value      Set app id for the OPEN API
      --appSecret value  Set app secret for the OPEN API
      --help, -h         show help (default: false)

   $ feishu2md dl -h
   NAME:
     feishu2md download - Download feishu/larksuite document to markdown file

   USAGE:
     feishu2md download [command options] <url>

   OPTIONS:
     --output value, -o value  Specify the output directory for the markdown files (default: "./")
     --dump                    Dump json response of the OPEN API (default: false)
     --batch                   Download all documents under a folder (default: false)
     --wiki                    Download all documents within the wiki. (default: false)
     --help, -h                show help (default: false)

   $ feishu2md sync -h
   NAME:
     feishu2md sync - Sync feishu/larksuite folder or wiki to local directory

   USAGE:
     feishu2md sync [command options] [url]

   OPTIONS:
     --output value, -o value     Specify the output directory (default: "./")
     --incremental, -i            Enable incremental sync (default: true)
     --force, -f                  Force re-download all documents (default: false)
     --include value              Only sync directories matching patterns (comma-separated)
     --exclude value              Exclude directories matching patterns (comma-separated)
     --concurrency value, -c      Maximum concurrent downloads (default: 5)
     --dump                       Dump json response of the OPEN API (default: false)
     --help, -h                   show help (default: false)

   ```

   **生成配置文件**

   通过 `feishu2md config --appId <your_id> --appSecret <your_secret>` 命令即可生成该工具的配置文件。

   通过 `feishu2md config` 命令可以查看配置文件路径以及是否成功配置。

   更多的配置选项请手动打开配置文件更改。

   **下载单个文档为 Markdown**

   通过 `feishu2md dl <your feishu docx url>` 直接下载，文档链接可以通过 **分享 > 开启链接分享 > 互联网上获得链接的人可阅读 > 复制链接** 获得。

   示例：

   ```bash
   $ feishu2md dl "https://domain.feishu.cn/docx/docxtoken"
   ```

  **批量下载某文件夹内的全部文档为 Markdown**

  此功能暂时不支持Docker版本

  通过`feishu2md dl --batch <your feishu folder url>` 直接下载，文件夹链接可以通过 **分享 > 开启链接分享 > 互联网上获得链接的人可阅读 > 复制链接** 获得。

  示例：

  ```bash
  $ feishu2md dl --batch -o output_directory "https://domain.feishu.cn/drive/folder/foldertoken"
  ```

  **批量下载某知识库的全部文档为 Markdown**

  通过`feishu2md dl --wiki <your feishu wiki setting url>` 直接下载，wiki settings链接可以通过 打开知识库设置获得。

  示例：

  ```bash
  $ feishu2md dl --wiki -o output_directory "https://domain.feishu.cn/wiki/settings/123456789101112"
  ```

</details>

<details>
  <summary>sync 命令（推荐）</summary>

  `sync` 命令是推荐的批量同步方式，支持增量下载、目录过滤等高级功能。

  **基本用法**

  ```bash
  # 同步知识库（自动检测 URL 类型）
  $ feishu2md sync "https://domain.feishu.cn/wiki/settings/123456789101112"

  # 同步云盘文件夹
  $ feishu2md sync "https://domain.feishu.cn/drive/folder/foldertoken"

  # 指定输出目录
  $ feishu2md sync -o ./output "https://domain.feishu.cn/wiki/settings/xxx"
  ```

  **增量同步**

  sync 命令默认开启增量模式，只下载有更新的文档：

  ```bash
  # 增量同步（默认行为，跳过未修改的文档）
  $ feishu2md sync "https://domain.feishu.cn/wiki/settings/xxx"

  # 强制重新下载所有文档
  $ feishu2md sync -f "https://domain.feishu.cn/wiki/settings/xxx"
  ```

  **目录过滤**

  支持通过 `--include` 和 `--exclude` 参数过滤目录：

  ```bash
  # 仅同步包含"文档"的目录
  $ feishu2md sync --include "*文档*" "https://domain.feishu.cn/wiki/settings/xxx"

  # 排除草稿和测试目录
  $ feishu2md sync --exclude "*草稿*,*测试*" "https://domain.feishu.cn/wiki/settings/xxx"

  # 组合使用：仅同步技术相关目录，但排除草稿
  $ feishu2md sync --include "技术*,API*" --exclude "*草稿*" "https://domain.feishu.cn/wiki/settings/xxx"
  ```

  通配符语法：
  - `*` 匹配任意字符
  - `?` 匹配单个字符
  - `[abc]` 匹配指定字符

  **并发控制**

  ```bash
  # 设置并发数为 10（默认为 5）
  $ feishu2md sync -c 10 "https://domain.feishu.cn/wiki/settings/xxx"
  ```

  **同步配置持久化**

  sync 命令会自动在输出目录保存同步配置（`.feishu2md.sync.json`），后续可以省略 URL：

  ```bash
  # 首次同步
  $ feishu2md sync -o ./docs "https://domain.feishu.cn/wiki/settings/xxx"

  # 后续同步（使用已保存的配置）
  $ feishu2md sync -o ./docs

  # 或者在输出目录下直接运行
  $ cd ./docs && feishu2md sync
  ```

</details>

<details>
  <summary>Docker版本</summary>

  Docker 镜像：https://hub.docker.com/r/wwwsine/feishu2md

   Docker 命令：`docker run -it --rm -p 8080:8080 -e FEISHU_APP_ID=<your id> -e FEISHU_APP_SECRET=<your secret> -e GIN_MODE=release wwwsine/feishu2md`

   Docker Compose:

   ```yml
   # docker-compose.yml
   version: '3'
   services:
     feishu2md:
       image: wwwsine/feishu2md
       environment:
         FEISHU_APP_ID: <your id>
         FEISHU_APP_SECRET: <your secret>
         GIN_MODE: release
       ports:
         - "8080:8080"
   ```

   启动服务 `docker compose up -d`

   然后访问 https://127.0.0.1:8080 粘贴文档链接即可，文档链接可以通过 **分享 > 开启链接分享 > 复制链接** 获得。
</details>

## 感谢

- [chyroc/lark](https://github.com/chyroc/lark)
- [chyroc/lark_docs_md](https://github.com/chyroc/lark_docs_md)

## 开发指南

### 环境要求

- Go 1.21+

### 构建

```bash
# 构建 CLI
make build

# 构建 Web 服务
make server

# 构建全部
make all
```

### 测试

```bash
# 运行所有测试
make test

# 代码格式化
make format
```

注意：部分测试需要有效的飞书凭证（环境变量 `FEISHU_APP_ID` 和 `FEISHU_APP_SECRET`）。

### 跨平台编译

```bash
# 编译所有平台
make cross-build

# 仅编译 Linux
make cross-build-linux

# 仅编译 Windows
make cross-build-windows

# 仅编译 macOS
make cross-build-darwin
```

支持的平台：
| 平台 | 架构 |
|------|------|
| Linux | amd64, arm64 |
| Windows | amd64 |
| macOS | amd64, arm64 |

### Docker

```bash
# 构建镜像
make image

# 运行容器
make docker
```

### 项目结构

```
feishu2md/
├── cmd/              # CLI 入口和命令
│   ├── main.go       # 命令行入口
│   ├── download.go   # 下载命令
│   ├── sync.go       # 同步命令
│   └── config.go     # 配置命令
├── core/             # 核心业务逻辑
│   ├── client.go     # 飞书 API 客户端
│   ├── parser.go     # 文档解析器
│   ├── filter.go     # 目录过滤器
│   ├── cache.go      # 缓存管理
│   ├── config.go     # 全局配置
│   └── sync_config.go # 同步配置
├── utils/            # 工具函数
├── web/              # Web 服务
└── testdata/         # 测试数据
```

### 贡献

欢迎提交 PR！请确保：

1. 代码通过 `make format` 格式化
2. 所有测试通过 `make test`
3. 新功能包含相应的测试用例

