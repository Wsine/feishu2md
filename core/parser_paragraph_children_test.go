package core_test

import (
	"strings"
	"testing"

	"github.com/88250/lute"
	"github.com/Wsine/feishu2md/core"
	"github.com/chyroc/lark"
)

func TestParseDocxParagraphChildren(t *testing.T) {
	text := func(value string) *lark.DocxBlockText {
		return &lark.DocxBlockText{Elements: []*lark.DocxTextElement{{TextRun: &lark.DocxTextElementTextRun{Content: value}}}}
	}
	paragraph := func(id, value string, children ...string) *lark.DocxBlock {
		return &lark.DocxBlock{BlockID: id, BlockType: lark.DocxBlockTypeText, Text: text(value), Children: children}
	}
	image := func(id string) *lark.DocxBlock {
		return &lark.DocxBlock{BlockID: id, BlockType: lark.DocxBlockTypeImage, Image: &lark.DocxBlockImage{Token: id}}
	}
	tests := []struct {
		name     string
		children []string
		blocks   []*lark.DocxBlock
		body     string
		tokens   []string
	}{
		{"leaf paragraph", []string{"parent"}, []*lark.DocxBlock{paragraph("parent", "Parent")}, "Parent\n\n", nil},
		{"nested and root images", []string{"parent", "root"}, []*lark.DocxBlock{paragraph("parent", "Parent", "nested"), image("nested"), image("root")}, "Parent\n\n![](nested)\n\n![](root)\n\n", []string{"nested", "root"}},
		{"recursive paragraphs", []string{"parent"}, []*lark.DocxBlock{paragraph("parent", "Parent", "child"), paragraph("child", "Child", "nested"), image("nested")}, "Parent\n\nChild\n\n![](nested)\n\n", []string{"nested"}},
		{"child order", []string{"parent"}, []*lark.DocxBlock{paragraph("parent", "Parent", "second", "first"), image("first"), image("second")}, "Parent\n\n![](second)\n\n![](first)\n\n", []string{"second", "first"}},
		{"child text", []string{"parent"}, []*lark.DocxBlock{paragraph("parent", "Parent", "child"), paragraph("child", "Child")}, "Parent\n\nChild\n\n", nil},
		{"paragraph inside list", []string{"list"}, []*lark.DocxBlock{{BlockID: "list", BlockType: lark.DocxBlockTypeBullet, Bullet: text("Item"), Children: []string{"parent"}}, paragraph("parent", "Parent", "nested"), image("nested")}, "- Item\n\tParent\n\n\t![](nested)\n\n", []string{"nested"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := &lark.DocxBlock{BlockID: "page", BlockType: lark.DocxBlockTypePage, Page: text("Document"), Children: tt.children}
			parser := core.NewParser(core.NewConfig("", "").Output)
			markdown := parser.ParseDocxContent(&lark.DocxDocument{DocumentID: "page"}, append(tt.blocks, page))
			if want := "# Document\n\n" + tt.body; markdown != want {
				t.Errorf("got %q, want %q", markdown, want)
			}
			if strings.Join(parser.ImgTokens, ",") != strings.Join(tt.tokens, ",") {
				t.Errorf("download tokens: got %v, want %v", parser.ImgTokens, tt.tokens)
			}
			// 验证格式化后仍渲染成图片，避免段落缩进将图片语法误变成代码。
			engine := lute.New()
			rendered := engine.MarkdownStr("", engine.FormatStr("md", markdown))
			if strings.Count(rendered, "<img ") != len(tt.tokens) {
				t.Errorf("expected %d rendered images: %s", len(tt.tokens), rendered)
			}
			if tt.name == "paragraph inside list" {
				if imageAt, listEnd := strings.Index(rendered, "<img "), strings.Index(rendered, "</ul>"); imageAt < 0 || listEnd < imageAt {
					t.Errorf("image escaped its list: %s", rendered)
				}
			}
		})
	}
}
