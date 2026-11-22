package core_test

import (
	"fmt"
	"html"
	"strings"
	"testing"

	"github.com/88250/lute"
	"github.com/Wsine/feishu2md/core"
	"github.com/chyroc/lark"
)

func TestParseDocxNestedCodeFences(t *testing.T) {
	text := func(content string) *lark.DocxBlockText {
		return &lark.DocxBlockText{Elements: []*lark.DocxTextElement{
			{TextRun: &lark.DocxTextElementTextRun{Content: content}},
		}}
	}
	tests := []struct {
		name    string
		depth   int
		content string
	}{
		{"top-level HTML", 0, "<a href=\"https://example.com\">\n  example\n</a>"},
		{"one list level", 1, "<a href=\"https://example.com\">\n  example\n</a>"},
		{"two list levels", 2, "<a href=\"https://example.com\">\n  example\n</a>"},
		{"top-level blank line", 0, "start\n\n  middle\nend"},
		{"nested blank line", 1, "start\n\n  middle\nend"},
		{"deeply nested blank line", 2, "start\n\n  middle\nend"},
		{"empty nested code", 1, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := &lark.DocxBlock{BlockID: "code", BlockType: lark.DocxBlockTypeCode, Code: text(tt.content)}
			code.Code.Style = &lark.DocxTextStyle{Language: lark.DocxCodeLanguageHTML}
			page := &lark.DocxBlock{BlockID: "page", BlockType: lark.DocxBlockTypePage, Page: text("Code")}
			blocks := []*lark.DocxBlock{page, code}
			parent := page
			for depth := 0; depth < tt.depth; depth++ {
				item := &lark.DocxBlock{BlockID: fmt.Sprintf("item%d", depth), ParentID: parent.BlockID, BlockType: lark.DocxBlockTypeBullet, Bullet: text("Item")}
				parent.Children = []string{item.BlockID}
				blocks = append(blocks, item)
				parent = item
			}
			parent.Children = []string{code.BlockID}
			code.ParentID = parent.BlockID
			parser := core.NewParser(core.NewConfig("", "").Output)
			markdown := parser.ParseDocxContent(&lark.DocxDocument{DocumentID: "page"}, blocks)
			indent := strings.Repeat("\t", tt.depth)
			wantCode := indent + "```html\n" + indent + strings.ReplaceAll(tt.content, "\n", "\n"+indent) + "\n" + indent + "```\n"
			if !strings.Contains(markdown, wantCode) {
				t.Errorf("inconsistent code block indentation: want %q in %q", wantCode, markdown)
			}

			// 通过实际 Markdown 渲染检查，确保代码不变为空块或逃出所属列表。
			engine := lute.New()
			engine.RenderOptions.AutoSpace = true
			engine.SetCodeSyntaxHighlight(false)
			formatted := engine.FormatStr("md", markdown)
			rendered := engine.MarkdownStr("", formatted)
			wantHTML := "<pre><code class=\"language-html\">" + tt.content + "\n</code></pre>"
			if strings.Count(rendered, "<pre><code") != 1 || !strings.Contains(html.UnescapeString(rendered), wantHTML) {
				t.Errorf("expected one complete code block %q, got %s", wantHTML, rendered)
			}
			if tt.depth > 0 {
				codeAt := strings.Index(rendered, "<pre><code")
				listEnd := strings.Index(rendered, "</ul>")
				if codeAt < 0 || listEnd < codeAt {
					t.Errorf("code escaped its list: %s", rendered)
				}
			}
		})
	}
}
