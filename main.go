package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/88250/lute"
	"github.com/chyroc/lark"
	"github.com/chyroc/lark/larkext"
	"github.com/chyroc/lark_docs_md"
	"github.com/urfave/cli/v2"
)

func generateConfig() error {
  configFilePath, err := getConfigFilePath()
  checkErr(err)
  if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
    if err := os.MkdirAll(filepath.Dir(configFilePath), os.ModePerm); err != nil {
      return err
    }
    config := Config{
      Feishu: Feishu{AppId: "", AppSecret: ""},
      Output: Output{ImageDir: "static"},
    }
    configJson, err := json.MarshalIndent(config, "", " ")
    checkErr(err)
    file, err := os.Create(configFilePath)
    checkErr(err)
    defer file.Close()
    _, err = file.WriteString(string(configJson))
    checkErr(err)
    fmt.Printf("Generated config file on %s\n", configFilePath)
  } else {
    fmt.Printf("Config file exists on %s\n", configFilePath)
  }
  return nil
}

func handleUrl(url string) error {
  configFilePath, err := getConfigFilePath()
  checkErr(err)
  configFile, err := os.Open(configFilePath)
  checkErr(err)
  defer configFile.Close()

  var config Config
  byteValue, err := ioutil.ReadAll(configFile)
  checkErr(err)
  json.Unmarshal(byteValue, &config)
  if config.Feishu.AppId == "" || config.Feishu.AppSecret == "" {
    return fmt.Errorf("Please fill in the app id and secret on %s first\n", configFilePath)
  }

  client := lark.New(
    lark.WithAppCredential(config.Feishu.AppId, config.Feishu.AppSecret),
  )

  reg := regexp.MustCompile("^https://[a-zA-Z0-9]+.(?:feishu.cn|larksuite.com)/docs/([a-zA-Z0-9]+)")
  matchResult := reg.FindStringSubmatch(url)
  if matchResult == nil || len(matchResult) != 2 {
    return fmt.Errorf("Invalid feishu/larksuite URL containing docToken\n")
  }
  docToken := matchResult[1]
  fmt.Println("Captured doc token:", docToken)
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

  fmt.Printf("Downloaded markdown file to %s\n", mdName)
  return nil
}

func main() {
  app := &cli.App{
    Name: "feishu2md",
    Version: "v0.1.2",
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
