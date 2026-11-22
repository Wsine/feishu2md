package core_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/88250/lute"
	"github.com/Wsine/feishu2md/core"
	"github.com/chyroc/lark"
)

func TestParseDocxCompositeTextStyles(t *testing.T) {
	link := &lark.DocxTextElementStyleLink{URL: "https%3A%2F%2Fexample.com%2Fasset"}
	tests := []struct {
		name  string
		style *lark.DocxTextElementStyle
		md    string
		html  string
	}{
		{"nil style", nil, "骨架资产", "骨架资产"},
		{"empty style", &lark.DocxTextElementStyle{}, "骨架资产", "骨架资产"},
		{"bold", &lark.DocxTextElementStyle{Bold: true}, "**骨架资产**", "<strong>骨架资产</strong>"},
		{"italic", &lark.DocxTextElementStyle{Italic: true}, "_骨架资产_", "<em>骨架资产</em>"},
		{"strikethrough", &lark.DocxTextElementStyle{Strikethrough: true}, "~~骨架资产~~", "<del>骨架资产</del>"},
		{"underline", &lark.DocxTextElementStyle{Underline: true}, "<u>骨架资产</u>", "<u>骨架资产</u>"},
		{"inline code", &lark.DocxTextElementStyle{InlineCode: true}, "`骨架资产`", "`骨架资产`"},
		{"link", &lark.DocxTextElementStyle{Link: link}, "[骨架资产](https://example.com/asset)", "[骨架资产](https://example.com/asset)"},
		{"underline and link", &lark.DocxTextElementStyle{Underline: true, Link: link}, "<u>[骨架资产](https://example.com/asset)</u>", "<u>[骨架资产](https://example.com/asset)</u>"},
		{"bold and italic", &lark.DocxTextElementStyle{Bold: true, Italic: true}, "**_骨架资产_**", "<strong><em>骨架资产</em></strong>"},
		{"italic and strikethrough", &lark.DocxTextElementStyle{Italic: true, Strikethrough: true}, "_~~骨架资产~~_", "<em><del>骨架资产</del></em>"},
		{"bold and link", &lark.DocxTextElementStyle{Bold: true, Link: link}, "**[骨架资产](https://example.com/asset)**", "<strong>[骨架资产](https://example.com/asset)</strong>"},
		{"linked inline code", &lark.DocxTextElementStyle{InlineCode: true, Link: link}, "[`骨架资产`](https://example.com/asset)", "[`骨架资产`](https://example.com/asset)"},
		{"all styles", &lark.DocxTextElementStyle{Bold: true, Italic: true, Strikethrough: true, Underline: true, InlineCode: true, Link: link}, "**_~~<u>[`骨架资产`](https://example.com/asset)</u>~~_**", "<strong><em><del><u>[`骨架资产`](https://example.com/asset)</u></del></em></strong>"},
	}
	for _, html := range []bool{false, true} {
		parser := core.NewParser(core.OutputConfig{UseHTMLTags: html})
		for _, tt := range tests {
			t.Run(fmt.Sprintf("html=%t/%s", html, tt.name), func(t *testing.T) {
				got := parser.ParseDocxTextElementTextRun(&lark.DocxTextElementTextRun{Content: "骨架资产", TextElementStyle: tt.style})
				want := tt.md
				if html {
					want = tt.html
				}
				if got != want {
					t.Fatalf("got %q, want %q", got, want)
				}
			})
		}
	}
}

func TestParseDocxCompositeStylesRenderAsMarkup(t *testing.T) {
	for _, html := range []bool{false, true} {
		t.Run(fmt.Sprintf("html=%t", html), func(t *testing.T) {
			parser := core.NewParser(core.OutputConfig{UseHTMLTags: html})
			md := parser.ParseDocxTextElementTextRun(&lark.DocxTextElementTextRun{
				Content: "asset",
				TextElementStyle: &lark.DocxTextElementStyle{
					Bold: true, Italic: true, Strikethrough: true, Underline: true, InlineCode: true,
					Link: &lark.DocxTextElementStyleLink{URL: "https://example.com/asset"},
				},
			})
			// 验证实际 Markdown 渲染结果，防止链接被包进代码而变成不可点击的文本。
			rendered := lute.New().MarkdownStr("", md)
			for _, markup := range []string{"<strong>", "<em>", "<del>", "<u>", `<a href="https://example.com/asset">`, "<code>asset</code>"} {
				if !strings.Contains(rendered, markup) {
					t.Errorf("rendered markup missing %q: %s", markup, rendered)
				}
			}
		})
	}
}
