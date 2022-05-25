# Feishu2Md

这是一个下载飞书文档为 Markdown 文件的工具，使用 Go 语言实现。

## 如何使用

借助 Go 语言跨平台的特性，已预先编译好了 x86 平台的可执行文件，可以在 [Release](https://github.com/Wsine/feishu2md/releases) 中下载，并将相应平台的 feishu2md 可执行文件放置在 PATH 路径中即可。

查阅帮助文档：

```bash
$ feishu2md -h
NAME:
   feishu2md - download feishu doc as markdown file

USAGE:
   feishu2md [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config    generate config file (default: false)
   --help, -h  show help (default: false)
```

生成配置文件：

通过 `feishu2md --config` 命令即可生成该工具的配置文件。生成的配置文件路径为：

- Windows: %AppData%/feishu2md/config.json

- Linux: $XDG_CONFIG_HOME/feishu2md/config.json

- Mac: $XDG_CONFIG_HOME/feishu2md/config.json

如无配置 XDG_CONFIG_HOME 环境变量，则默认为 ~/.config 目录。

生成的配置文件需要填写 APP ID 和 APP SECRET 信息，请参考 [飞书官方文档](https://open.feishu.cn/document/ukTMukTMukTM/ukDNz4SO0MjL5QzM/get-) 获取。

Image Dir 为存放文档中图片的文件夹名称。

```json
{
 "feishu": {
  "app_id": "",
  "app_secret": ""
 },
 "output": {
  "image_dir": "static"
 }
}
```

下载为 Markdown：

通过 `feishu2md <your feishu doc url>` 直接下载，文档链接可以通过 **分享 > 开启链接分享 > 复制链接** 获得。

示例：

```bash
$ feishu2md https://domain.feishu.cn/docs/doctoken
```

## 感谢

- [chyroc/lark](https://github.com/chyroc/lark)
- [chyroc/lark_docs_md](https://github.com/chyroc/lark_docs_md)
