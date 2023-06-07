package core_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/88250/lute"
	"github.com/Wsine/feishu2md/core"
	"github.com/Wsine/feishu2md/utils"
	"github.com/chyroc/lark"
	"github.com/stretchr/testify/assert"
)

func TestParseDocxContent(t *testing.T) {
	root := utils.RootDir()
	engine := lute.New(func(l *lute.Lute) {
		l.RenderOptions.AutoSpace = true
	})

	testdata := []string{
		"testdocx.1",
		"testdocx.2",
		"testdocx.3",
	}
	for _, td := range testdata {
		t.Run(td, func(t *testing.T) {
			jsonFile, err := os.Open(path.Join(root, "testdata", td+".json"))
			utils.CheckErr(err)
			defer jsonFile.Close()

			data := struct {
				Document *lark.DocxDocument `json:"document"`
				Blocks   []*lark.DocxBlock  `json:"blocks"`
			}{}
			byteValue, _ := ioutil.ReadAll(jsonFile)
			json.Unmarshal(byteValue, &data)

			parser := core.NewParser(context.Background())
			mdParsed := parser.ParseDocxContent(data.Document, data.Blocks)
			fmt.Println(mdParsed)
			mdParsed = engine.FormatStr("md", mdParsed)

			mdFile, err := ioutil.ReadFile(path.Join(root, "testdata", td+".md"))
			utils.CheckErr(err)
			mdExpected := string(mdFile)

			assert.Equal(t, mdExpected, mdParsed)
		})
	}
}
