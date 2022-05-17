package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/88250/lute"
	"github.com/chyroc/lark"
	"github.com/chyroc/lark/larkext"
	"github.com/chyroc/lark_docs_md"
	"github.com/urfave/cli/v2"
)

var configName string = "feishu.json"

func checkErr(e error) {
  if e != nil {
    panic(e)
  }
}

func generateConfig() error {
  if _, err := os.Stat(configName); errors.Is(err, os.ErrNotExist) {
    config := Config{
      Feishu: Feishu{AppId: "", AppSecret: ""},
      Output: Output{ImageDir: "static"},
    }
    file, err := json.MarshalIndent(config, "", " ")
    checkErr(err)
    err = ioutil.WriteFile(configName, file, 0644)
    checkErr(err)
  }
  return nil
}

func handleUrl(url string) error {
  configFile, err := os.Open(configName)
  checkErr(err)
  defer configFile.Close()

  var config Config
  byteValue, err := ioutil.ReadAll(configFile)
  checkErr(err)
  json.Unmarshal(byteValue, &config)
  fmt.Printf("%+v\n", config)

  client := lark.New(
    lark.WithAppCredential(config.Feishu.AppId, config.Feishu.AppSecret),
  )

  docToken := url[strings.LastIndex(url, "/") + 1 : ]
  doc, err := larkext.NewDoc(client, docToken).Content(context.Background())
  checkErr(err)

  result := lark_docs_md.DocMarkdown(context.Background(), doc, &lark_docs_md.FormatOpt{
    LarkClient: client,
    StaticDir: config.Output.ImageDir,
    FilePrefix: config.Output.ImageDir,
  })
  engine := lute.New(func (l *lute.Lute) {
    l.RenderOptions.AutoSpace = true
  })
  result = engine.FormatStr("md", result)

  mdName := fmt.Sprintf("%s.md", docToken)
  mdFile, err := os.Create(mdName)
  checkErr(err)
  defer mdFile.Close()
  nBytes, err := mdFile.WriteString(result)
  checkErr(err)
  fmt.Printf("Wrote %d bytes\n", nBytes)
  mdFile.Sync()
  return nil
}

func main() {
  app := &cli.App{
    Name: "feishu2md",
    Usage: "download feishu doc as markdown file",
    Flags: []cli.Flag{
      &cli.BoolFlag{
        Name: "config",
        Usage: "generate config file",
      },
    },
    Action: func(ctx *cli.Context) error {
      if ctx.Bool("config") {
        err := generateConfig()
        checkErr(err)
      } else if ctx.NArg() == 0 {
        cli.ShowAppHelp(ctx)
      } else {
        url := ctx.Args().Get(0)
        err := handleUrl(url)
        checkErr(err)
      }
      return nil
    },
  }

  err := app.Run(os.Args)
  checkErr(err)
}

