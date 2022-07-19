package core_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/Wsine/feishu2md/core"
	"github.com/Wsine/feishu2md/utils"
	"github.com/chyroc/lark"
)

func TestParseDocContent(t *testing.T) {
  root := utils.RootDir()
  jsonFile, err := os.Open(path.Join(root, "data", "testdata.1.json"))
  utils.CheckErr(err)
  defer jsonFile.Close()
  var docs lark.DocContent
  byteValue, _ := ioutil.ReadAll(jsonFile)
  json.Unmarshal(byteValue, &docs)

  title := docs.Title.Elements[0].TextRun.Text
  if title != "一日一技：飞书文档转换为 Markdown" {
    t.Errorf("The parsed title is not correct.")
  }

  parser := core.NewParser(context.Background())
  fmt.Println(parser.ParseDocContent(&docs))
}

