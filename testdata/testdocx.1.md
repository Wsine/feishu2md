# 一日一技：飞书文档转换为 Markdown

随着少数派逐渐 All in 飞书，我们少数派作者们也逐渐迁移到飞书文档进行写稿。飞书文档提供了 Web 平台的富文本编辑器，配合「少数派助手」这个服务，可以将稿件一键发布到少数派平台，着实是非常方便。

不少的少数派作者都有自己的博客平台，而大部分的博客平台都是使用 Markdown 作为输入从而生成 HTML 发布到网络中的。但是，飞书只支持 Markdown 语法的编辑，却不支持导出为 Markdown 文件下载，这打断了我们一直以来已经完善的发布博客流程。

本文就提供一种将飞书文档转换为 Markdown 文件的方法，来弥补这个 Gap。

关联阅读：

- [《内容团队协作的最佳形式：少数派编辑部如何用飞书》](https://sspai.com/post/58509)
- [《如何使用「少数派助手」从飞书文档发布文章》](https://sspai.com/post/68135)

## 现有的方法痛点

飞书支持的导出格式为 Word 和 PDF 两种格式。如需编辑，我们就只能选择 Word 格式，然后使用文档格式转换的瑞士军刀 pandoc 从 Word 文档转换为 Markdown 文件。参考命令：`pandoc test.docx -o test.md` 。但是，如今这种方法已经不可靠了，如果尝试将本文转换，则会得到下图的格式。

![](boxcnbK20aJ9pePyziodIvjXTce)

从图中的效果可以看出，文档中多了很多冗余的换行，列表格式消失不见，图片丢失等问题。究其原因，是因为导出的 Word 文档没有使用 Word 内建的富文本样式，而全部使用的自定义样式。至于图片问题，转换后的 Markdown 文档中的图片格式是 `![Generated](media/image1.png){width="5.90625in" height="2.8020833333333335in"}` 。可以通过将 Word 文档的 docx 后缀改为 zip，然后从压缩包中整体提取 word/media 文件夹来修复图片的问题。但其它的格式问题，依然是一个头疼的问题。

另一方面，在没有 pandoc 转换工具的情况下，如果要获得 Markdown 文件，我理解的最便捷的方法如下：

1. 全文复制飞书文档的富文本内容
2. 全文粘贴到本地的 markdown 编辑器中
3. （可选）逐个下载文档中的图片并替换 markdown 文件中的图片

当完成第 2 步的时候，其实文档看起来已经完整了，但是仔细观察会发现文档中的图片是飞书的临时链接，且只有 24 小时的有效时间。因此，为了有效地保留图片，需要进行第 3 步手动下载图片替换。当一篇文档中的图片非常多的时候，手动下载替换是一个非常枯燥的事情。

如果是使用图床的作者，可以在第 2 步的文档后直接使用图床上传工具（如：PicGo）进行图片上传快速替换，甚至 Typora 编辑器中就自带了这个功能。但是，由于图片链接是临时链接，没有文件后缀（.jpg/.png/.gif），当上传到图床后也丢失了这个信息，虽然不影响图床的回传，但是后面如果需要替换图床将会是一个灾难。

## 使用 Feishu2Md 工具

在进行了大量的搜索后，我其实也没有找到现有的转换工具能够转换飞书文档为 Markdown 文件下载的。但是，十分幸运，我碰巧找到了 [chyroc](https://github.com/chyroc) 使用飞书的 Open API 实现的飞书文档解析器 [lark_docs_md](https://github.com/chyroc/lark_docs_md) 。因此，我决定基于这个库开发一个下载工具，也就是小标题的 Feishu2Md 工具。

Feishu2Md 已开源并发布在 Github 中： [https://github.com/Wsine/feishu2md](https://github.com/Wsine/feishu2md)

<strong>下载 feishu2md </strong>-<strong> </strong>得益于 golang 本身的多平台编译特性，我已经为 Windows/Linux/Mac 都预编译了该工具的可执行文件，可以直接从 [Github Release](https://github.com/Wsine/feishu2md/releases) 中下载，从压缩包中提取自己平台的 feishu2md 二进制可执行文件即可，建议放置在 PATH 路径中。

<strong>生成配置文件</strong> - feishu2md 需要使用飞书的 Open API 提取飞书文档，因此需要配置相应的 App ID 和 App Secret 进行 API 的调用。首先，进入飞书的 [开发者后台](https://open.feishu.cn/app) 然后创建一个企业自建应用，信息可以任意填，发布但不必等待审核通过。然后在创建的应用页面中，找到「凭证与基础信息」，即可找到 App ID 和 App Secret 信息。

![](boxcnh7JKLbFaWhHKHveYzGMNZg)

执行 `feishu2md --config` 命令会生成该工具的配置文件。生成的配置文件路径为：

- Windows: %AppData%/feishu2md/config.json
- Linux: $XDG_CONFIG_HOME/feishu2md/config.json
- Mac: $XDG_CONFIG_HOME/feishu2md/config.json

如无配置 XDG_CONFIG_HOME 环境变量，则默认为 ~/.config 目录

将 App ID 和 App Secret 填入配置文件 config.json 中的相应位置。另外，image_dir 配置项为存放文档中图片的文件夹名称。

<strong>下载飞书文档</strong> - 通过 `feishu2md <你的飞书文档链接>` 直接下载，文档链接可以通过 分享 > 开启链接分享 > 复制链接 获得。

![](boxcnqt9YDTirkKlTATlQI025Ig)

调用示例：

```
feishu2md [一日一技：飞书文档转换为 Markdown](https://oaztcemx3k.feishu.cn/docs/doccnrOvzeQ8BSnfsXj8jwJHC3c#)
```

![](boxcnAb2MgMQoUMDLLf3ySogueh)

格式转换可能会有一些细微的渲染差异，毕竟 markdown 本身的标准也有很多套，建议手动检查一下。而最头疼的图片问题，该工具也已经帮忙整体处理好了。然后就可以愉快地用以前的工作流发布博客了。

## 开发感言

由于 lark_docs_md 是使用 golang 实现的，因此这也是我首次使用 golang 进行开发。对于开发小工具，整体的开发体验非常良好，而且还能编译得到二进制以及享受多平台编译的好处。工具可能还有一些不是很完善的地方，如有问题可以提 issue，我有时间会进行修复的。

最后，欢迎试用，欢迎 PR ~
