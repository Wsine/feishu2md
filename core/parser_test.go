package core_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

func TestParseDocxContent(t *testing.T) {
	root := utils.RootDir()
	jsonFile, err := os.Open(path.Join(root, "data", "testdata.2.json"))
	utils.CheckErr(err)
	defer jsonFile.Close()
	data := struct {
		Document *lark.DocxDocument `json:"document"`
		Blocks   []*lark.DocxBlock  `json:"blocks"`
	}{}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &data)

	title := data.Document.Title
	if title != "一日一技：飞书文档转换为 Markdown" {
		t.Errorf("The parsed title is not correct.")
	}

	parser := core.NewParser(context.Background())
	fmt.Println(parser.ParseDocxContent(data.Document, data.Blocks))
}

func TestParseDocContent_table(t *testing.T) {
	type test struct {
		name      string
		inputJSON string
		expectMD  string
	}

	tests := []test{
		{
			name:      "parse code block",
			inputJSON: "docs_code_block.json",
			expectMD:  "docs_code_block.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := utils.RootDir()
			jsonFile, err := os.Open(path.Join(root, "data", tt.inputJSON))
			utils.CheckErr(err)
			defer jsonFile.Close()
			mdFile, err := os.Open(path.Join(root, "data", tt.expectMD))
			utils.CheckErr(err)
			defer mdFile.Close()

			docs := new(lark.DocContent)
			err = json.NewDecoder(jsonFile).Decode(docs)
			utils.CheckErr(err)
			parser := core.NewParser(context.Background())
			got := parser.ParseDocContent(docs)

			var expect bytes.Buffer
			_ , err = io.Copy(&expect, mdFile)
			utils.CheckErr(err)
			if want:= expect.String();  got != want {
				t.Errorf("got %v, want %v", got, want)
			}
		})
	}
}
