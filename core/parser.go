package core

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/chyroc/lark"
)

type Parser struct {
	ctx       context.Context
	ImgTokens []string
}

func NewParser(ctx context.Context) *Parser {
	return &Parser{ctx: ctx, ImgTokens: make([]string, 0)}
}

// =============================================================
// Parse the old version of document (docs)
// =============================================================

func (p *Parser) ParseDocContent(docs *lark.DocContent) string {
	buf := new(strings.Builder)
	buf.WriteString(p.ParseDocParagraph(docs.Title, true))
	buf.WriteString("\n")
	buf.WriteString(p.ParseDocBody(docs.Body))
	return buf.String()
}

func (p *Parser) ParseDocParagraph(para *lark.DocParagraph, isTitle bool) string {
	buf := new(strings.Builder)
	if isTitle {
		buf.WriteString("# ")
		buf.WriteString(para.Elements[0].TextRun.Text)
		buf.WriteString("\n")
	} else {
		postWrite := ""
		if style := para.Style; style != nil {
			if style.HeadingLevel > 0 {
				buf.WriteString(strings.Repeat("#", int(style.HeadingLevel)))
				buf.WriteString(" ")
			} else if list := style.List; list != nil {
				switch list.Type {
				case "number":
					buf.WriteString(strconv.Itoa(list.Number) + ".")
				case "bullet":
					buf.WriteString("-")
				case "checkBox":
					buf.WriteString("- [ ]")
				case "checkedBox":
					buf.WriteString("- [x]")
				}
				buf.WriteString(" ")
			} else if style.Quote {
				buf.WriteString("> ")
			} else {
				switch style.Align {
				case "right":
				case "center":
					buf.WriteString(fmt.Sprintf("<div style=\"text-align: %s\">", style.Align))
					postWrite += "</div>"
				default:
				}
			}
		}
		for _, e := range para.Elements {
			buf.WriteString(p.ParseDocParagraphElement(e))
		}
		buf.WriteString(postWrite)
		buf.WriteString("\n")
	}
	return buf.String()
}

func (p *Parser) ParseDocBody(body *lark.DocBody) string {
	buf := new(strings.Builder)
	for _, b := range body.Blocks {
		buf.WriteString(p.ParseDocBlock(b))
		buf.WriteString("\n")
	}
	return buf.String()
}

func (p *Parser) ParseDocBlock(b *lark.DocBlock) string {
	switch b.Type {
	case lark.DocBlockTypeParagraph:
		return p.ParseDocParagraph(b.Paragraph, false)
	case lark.DocBlockTypeGallery:
		return p.ParseDocGallery(b.Gallery)
	case lark.DocBlockTypeCode:
		return p.ParseDocCode(b.Code)
	default:
		return ""
	}
}

func (p *Parser) ParseDocParagraphElement(e *lark.DocParagraphElement) string {
	switch e.Type {
	case lark.DocParagraphElementTypeTextRun:
		return p.ParseDocTextRun(e.TextRun)
	case lark.DocParagraphElementTypeDocsLink:
		return p.ParseDocDocsLink(e.DocsLink)
	case lark.DocParagraphElementTypeEquation:
		return p.ParseDocEquation(e.Equation)
	default:
		return ""
	}
}

func (p *Parser) ParseDocTextRun(tr *lark.DocTextRun) string {
	buf := new(strings.Builder)
	postWrite := ""
	if style := tr.Style; style != nil {
		if style.Bold {
			buf.WriteString("**")
			postWrite = "**"
		} else if style.Italic {
			buf.WriteString("*")
			postWrite = "*"
		} else if style.StrikeThrough {
			buf.WriteString("~~")
			postWrite = "~~"
		} else if style.Underline {
			buf.WriteString("<u>")
			postWrite = "</u>"
		} else if style.CodeInline {
			buf.WriteString("`")
			postWrite = "`"
		} else if link := style.Link; link != nil {
			buf.WriteString("[")
			postWrite = fmt.Sprintf("](%s)", link.URL)
		}
	}
	buf.WriteString(tr.Text)
	buf.WriteString(postWrite)
	return buf.String()
}

func (p *Parser) ParseDocDocsLink(l *lark.DocDocsLink) string {
	buf := new(strings.Builder)
	buf.WriteString(fmt.Sprintf("[](%s)", l.URL))
	return buf.String()
}

func (p *Parser) ParseDocEquation(eq *lark.DocEquation) string {
	buf := new(strings.Builder)
	buf.WriteString("$$" + eq.Equation + "$$")
	return buf.String()
}

func (p *Parser) ParseDocGallery(g *lark.DocGallery) string {
	buf := new(strings.Builder)
	for _, img := range g.ImageList {
		buf.WriteString(p.ParseDocImageItem(img))
	}
	return buf.String()
}

func (p *Parser) ParseDocImageItem(img *lark.DocImageItem) string {
	buf := new(strings.Builder)
	buf.WriteString(fmt.Sprintf("![](%s)", img.FileToken))
	buf.WriteString("\n")
	p.ImgTokens = append(p.ImgTokens, img.FileToken)
	return buf.String()
}

func (p *Parser) ParseDocCode(c *lark.DocCode) string {
	buf := new(strings.Builder)
	buf.WriteString("```")
	buf.WriteString(c.Language)
	buf.WriteString("\n")
	buf.WriteString(p.ParseDocBody(c.Body))
	buf.WriteString("```")
	buf.WriteString("\n")
	return buf.String()
}

// =============================================================
// Parse the new version of document (docx)
// =============================================================

func (p *Parser) ParseDocxContent(doc *lark.DocxDocument, blocks []*lark.DocxBlock) string {
	buf := new(strings.Builder)
	// buf.WriteString(p.ParseDocxDocument(doc))
	// buf.WriteString("\n")
	for _, v := range blocks {
		buf.WriteString(p.ParseDocxBlock(v))
		buf.WriteString("\n")
	}
	return buf.String()
}

func (p *Parser) ParseDocxDocument(doc *lark.DocxDocument) string {
	return doc.Title
}

func (p *Parser) ParseDocxBlock(b *lark.DocxBlock) string {
	buf := new(strings.Builder)
	switch b.BlockType {
	case lark.DocxBlockTypePage:
		buf.WriteString("# ")
		buf.WriteString(p.ParseDocxBlockText(b.Page))
	case lark.DocxBlockTypeText:
		return p.ParseDocxBlockText(b.Text)
	case lark.DocxBlockTypeHeading1:
		buf.WriteString("# ")
		buf.WriteString(p.ParseDocxBlockText(b.Heading1))
	case lark.DocxBlockTypeHeading2:
		buf.WriteString("## ")
		buf.WriteString(p.ParseDocxBlockText(b.Heading2))
	case lark.DocxBlockTypeHeading3:
		buf.WriteString("### ")
		buf.WriteString(p.ParseDocxBlockText(b.Heading3))
	case lark.DocxBlockTypeHeading4:
		buf.WriteString("#### ")
		buf.WriteString(p.ParseDocxBlockText(b.Heading4))
	case lark.DocxBlockTypeHeading5:
		buf.WriteString("##### ")
		buf.WriteString(p.ParseDocxBlockText(b.Heading5))
	case lark.DocxBlockTypeHeading6:
		buf.WriteString("###### ")
		buf.WriteString(p.ParseDocxBlockText(b.Heading6))
	case lark.DocxBlockTypeHeading7:
		buf.WriteString("####### ")
		buf.WriteString(p.ParseDocxBlockText(b.Heading7))
	case lark.DocxBlockTypeHeading8:
		buf.WriteString("######## ")
		buf.WriteString(p.ParseDocxBlockText(b.Heading8))
	case lark.DocxBlockTypeHeading9:
		buf.WriteString("######### ")
		buf.WriteString(p.ParseDocxBlockText(b.Heading9))
	case lark.DocxBlockTypeBullet:
		buf.WriteString("- ")
		buf.WriteString(p.ParseDocxBlockText(b.Bullet))
	case lark.DocxBlockTypeOrdered:
		buf.WriteString("1. ")
		buf.WriteString(p.ParseDocxBlockText(b.Ordered))
	case lark.DocxBlockTypeCode:
		buf.WriteString("```\n")
		buf.WriteString(p.ParseDocxBlockText(b.Code))
		buf.WriteString("\n```")
	case lark.DocxBlockTypeQuote:
		buf.WriteString("> ")
		buf.WriteString(p.ParseDocxBlockText(b.Quote))
	case lark.DocxBlockTypeEquation:
		buf.WriteString("$$\n")
		buf.WriteString(p.ParseDocxBlockText(b.Equation))
		buf.WriteString("\n$$")
	case lark.DocxBlockTypeTodo:
		if b.Ordered.Style.Done {
			buf.WriteString("- [x] ")
		} else {
			buf.WriteString("- [ ] ")
		}
		buf.WriteString(p.ParseDocxBlockText(b.Todo))
	case lark.DocxBlockTypeImage:
		buf.WriteString(p.ParseDocxBlockImage(b.Image))
	default:
		return ""
	}
	return buf.String()
}

func (p *Parser) ParseDocxBlockText(b *lark.DocxBlockText) string {
	buf := new(strings.Builder)
	for _, e := range b.Elements {
		buf.WriteString(p.ParseDocxTextElement(e))
	}
	buf.WriteString("\n")
	return buf.String()
}

func (p *Parser) ParseDocxTextElement(e *lark.DocxTextElement) string {
	buf := new(strings.Builder)
	if e.TextRun != nil {
		buf.WriteString(p.ParseDocxTextElementTextRun(e.TextRun))
	}
	if e.MentionUser != nil {
		buf.WriteString(e.MentionUser.UserID)
	}
	if e.MentionDoc != nil {
		buf.WriteString(fmt.Sprintf("[%s](%s)", e.MentionDoc.Title, e.MentionDoc.URL))
	}
	if e.Equation != nil {
		buf.WriteString("%%" + e.Equation.Content + "%%")
	}
	return buf.String()
}

func (p *Parser) ParseDocxTextElementTextRun(tr *lark.DocxTextElementTextRun) string {
	buf := new(strings.Builder)
	postWrite := ""
	if style := tr.TextElementStyle; style != nil {
		if style.Bold {
			buf.WriteString("**")
			postWrite = "**"
		} else if style.Italic {
			buf.WriteString("*")
			postWrite = "*"
		} else if style.Strikethrough {
			buf.WriteString("~~")
			postWrite = "~~"
		} else if style.Underline {
			buf.WriteString("<u>")
			postWrite = "</u>"
		} else if style.InlineCode {
			buf.WriteString("`")
			postWrite = "`"
		} else if link := style.Link; link != nil {
			buf.WriteString("[")
			postWrite = fmt.Sprintf("](%s)", link.URL)
		}
	}
	buf.WriteString(tr.Content)
	buf.WriteString(postWrite)
	return buf.String()
}

func (p *Parser) ParseDocxBlockImage(img *lark.DocxBlockImage) string {
	buf := new(strings.Builder)
	buf.WriteString(fmt.Sprintf("![](%s)", img.Token))
	buf.WriteString("\n")
	p.ImgTokens = append(p.ImgTokens, img.Token)
	return buf.String()
}

func (p *Parser) ParseDocxWhatever(body *lark.DocBody) string {
	buf := new(strings.Builder)

	return buf.String()
}
