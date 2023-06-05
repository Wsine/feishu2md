package core

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Wsine/feishu2md/utils"
	"github.com/chyroc/lark"
	"github.com/elliotchance/orderedmap"
	strip "github.com/grokify/html-strip-tags-go"
	"github.com/olekukonko/tablewriter"
)

type Parser struct {
	ctx       context.Context
	ImgTokens []string
}

func NewParser(ctx context.Context) *Parser {
	return &Parser{ctx: ctx, ImgTokens: make([]string, 0)}
}

// =============================================================
// Parser utils
// =============================================================

var DocxCodeLang2MdStr = map[lark.DocxCodeLanguage]string{
	lark.DocxCodeLanguagePlainText:    "",
	lark.DocxCodeLanguageABAP:         "abap",
	lark.DocxCodeLanguageAda:          "ada",
	lark.DocxCodeLanguageApache:       "apache",
	lark.DocxCodeLanguageApex:         "apex",
	lark.DocxCodeLanguageAssembly:     "assembly",
	lark.DocxCodeLanguageBash:         "bash",
	lark.DocxCodeLanguageCSharp:       "csharp",
	lark.DocxCodeLanguageCPlusPlus:    "cpp",
	lark.DocxCodeLanguageC:            "c",
	lark.DocxCodeLanguageCOBOL:        "cobol",
	lark.DocxCodeLanguageCSS:          "css",
	lark.DocxCodeLanguageCoffeeScript: "coffeescript",
	lark.DocxCodeLanguageD:            "d",
	lark.DocxCodeLanguageDart:         "dart",
	lark.DocxCodeLanguageDelphi:       "delphi",
	lark.DocxCodeLanguageDjango:       "django",
	lark.DocxCodeLanguageDockerfile:   "dockerfile",
	lark.DocxCodeLanguageErlang:       "erlang",
	lark.DocxCodeLanguageFortran:      "fortran",
	lark.DocxCodeLanguageFoxPro:       "foxpro",
	lark.DocxCodeLanguageGo:           "go",
	lark.DocxCodeLanguageGroovy:       "groovy",
	lark.DocxCodeLanguageHTML:         "html",
	lark.DocxCodeLanguageHTMLBars:     "htmlbars",
	lark.DocxCodeLanguageHTTP:         "http",
	lark.DocxCodeLanguageHaskell:      "haskell",
	lark.DocxCodeLanguageJSON:         "json",
	lark.DocxCodeLanguageJava:         "java",
	lark.DocxCodeLanguageJavaScript:   "javascript",
	lark.DocxCodeLanguageJulia:        "julia",
	lark.DocxCodeLanguageKotlin:       "kotlin",
	lark.DocxCodeLanguageLateX:        "latex",
	lark.DocxCodeLanguageLisp:         "lisp",
	lark.DocxCodeLanguageLogo:         "logo",
	lark.DocxCodeLanguageLua:          "lua",
	lark.DocxCodeLanguageMATLAB:       "matlab",
	lark.DocxCodeLanguageMakefile:     "makefile",
	lark.DocxCodeLanguageMarkdown:     "markdown",
	lark.DocxCodeLanguageNginx:        "nginx",
	lark.DocxCodeLanguageObjective:    "objectivec",
	lark.DocxCodeLanguageOpenEdgeABL:  "openedge-abl",
	lark.DocxCodeLanguagePHP:          "php",
	lark.DocxCodeLanguagePerl:         "perl",
	lark.DocxCodeLanguagePostScript:   "postscript",
	lark.DocxCodeLanguagePower:        "powershell",
	lark.DocxCodeLanguageProlog:       "prolog",
	lark.DocxCodeLanguageProtoBuf:     "protobuf",
	lark.DocxCodeLanguagePython:       "python",
	lark.DocxCodeLanguageR:            "r",
	lark.DocxCodeLanguageRPG:          "rpg",
	lark.DocxCodeLanguageRuby:         "ruby",
	lark.DocxCodeLanguageRust:         "rust",
	lark.DocxCodeLanguageSAS:          "sas",
	lark.DocxCodeLanguageSCSS:         "scss",
	lark.DocxCodeLanguageSQL:          "sql",
	lark.DocxCodeLanguageScala:        "scala",
	lark.DocxCodeLanguageScheme:       "scheme",
	lark.DocxCodeLanguageScratch:      "scratch",
	lark.DocxCodeLanguageShell:        "shell",
	lark.DocxCodeLanguageSwift:        "swift",
	lark.DocxCodeLanguageThrift:       "thrift",
	lark.DocxCodeLanguageTypeScript:   "typescript",
	lark.DocxCodeLanguageVBScript:     "vbscript",
	lark.DocxCodeLanguageVisual:       "vbnet",
	lark.DocxCodeLanguageXML:          "xml",
	lark.DocxCodeLanguageYAML:         "yaml",
}

func renderMarkdownTable(data [][]string) string {
	builder := &strings.Builder{}
	table := tablewriter.NewWriter(builder)
	table.SetCenterSeparator("|")
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(false)
	table.SetAutoMergeCells(false)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetHeader(data[0])
	table.AppendBulk(data[1:])
	table.Render()
	return builder.String()
}

// =============================================================
// Parse the new version of document (docx)
// =============================================================

func (p *Parser) ParseDocxContent(doc *lark.DocxDocument, blocks []*lark.DocxBlock) string {
	// block map
	// - Table cell block needs block map to collect children blocks
	// - ParseDocxContent needs block map to avoid duplicate rendering
	blockMap := orderedmap.NewOrderedMap()
	for _, block := range blocks {
		blockMap.Set(block.BlockID, block)
	}

	buf := new(strings.Builder)
	// buf.WriteString(p.ParseDocxDocument(doc))
	// buf.WriteString("\n")
	for _, v := range blocks {
		buf.WriteString(p.ParseDocxBlock(v, blockMap))
		buf.WriteString("\n")
	}
	return buf.String()
}

func (p *Parser) ParseDocxDocument(doc *lark.DocxDocument) string {
	return doc.Title
}

func (p *Parser) ParseDocxBlock(b *lark.DocxBlock, blockMap *orderedmap.OrderedMap) string {
	if _, ok := blockMap.Get(b.BlockID); blockMap != nil && !ok {
		// ignore rendered children block
		return ""
	}

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
		// calculate indent level
		indentLevel := 1
		parent := blockMap.GetOrDefault(b.ParentID, nil)
		for {
			if parent == nil || parent.(*lark.DocxBlock).BlockType != lark.DocxBlockTypeBullet {
				break
			}
			indentLevel += 1
			parent = blockMap.GetOrDefault(parent.(*lark.DocxBlock).ParentID, nil)
		}
		buf.WriteString(strings.Repeat("  ", indentLevel-1))
		buf.WriteString("- ")
		buf.WriteString(p.ParseDocxBlockText(b.Bullet))
	case lark.DocxBlockTypeOrdered:
		// calculate indent level
		indentLevel := 1
		parent := blockMap.GetOrDefault(b.ParentID, nil)
		for {
			if parent == nil || parent.(*lark.DocxBlock).BlockType != lark.DocxBlockTypeOrdered {
				break
			}
			indentLevel += 1
			parent = blockMap.GetOrDefault(parent.(*lark.DocxBlock).ParentID, nil)
		}
		buf.WriteString(strings.Repeat("  ", indentLevel-1))
		buf.WriteString("1. ")
		buf.WriteString(p.ParseDocxBlockText(b.Ordered))
	case lark.DocxBlockTypeCode:
		buf.WriteString("```" + DocxCodeLang2MdStr[b.Code.Style.Language] + "\n")
		buf.WriteString(strings.TrimSpace(p.ParseDocxBlockText(b.Code)))
		buf.WriteString("\n```")
	case lark.DocxBlockTypeQuote:
		buf.WriteString("> ")
		buf.WriteString(p.ParseDocxBlockText(b.Quote))
	case lark.DocxBlockTypeEquation:
		buf.WriteString("$$\n")
		buf.WriteString(p.ParseDocxBlockText(b.Equation))
		buf.WriteString("\n$$")
	case lark.DocxBlockTypeTodo:
		if b.Todo.Style.Done {
			buf.WriteString("- [x] ")
		} else {
			buf.WriteString("- [ ] ")
		}
		buf.WriteString(p.ParseDocxBlockText(b.Todo))
	case lark.DocxBlockTypeImage:
		buf.WriteString(p.ParseDocxBlockImage(b.Image))
	case lark.DocxBlockTypeTableCell:
		buf.WriteString(p.ParseDocxBlockTableCell(b.BlockID, blockMap))
	case lark.DocxBlockTypeTable:
		buf.WriteString(p.ParseDocxBlockTable(b.ParentID, b.Table, blockMap))
	case lark.DocxBlockTypeQuoteContainer:
		buf.WriteString(p.ParseDocxBlockQuoteContainer(b.BlockID, b.QuoteContainer, blockMap))
	default:
		return ""
	}
	return buf.String()
}

func (p *Parser) ParseDocxBlockText(b *lark.DocxBlockText) string {
	buf := new(strings.Builder)
	numElem := len(b.Elements)
	for _, e := range b.Elements {
		inline := numElem > 1
		buf.WriteString(p.ParseDocxTextElement(e, inline))
	}
	buf.WriteString("\n")
	return buf.String()
}

func (p *Parser) ParseDocxTextElement(e *lark.DocxTextElement, inline bool) string {
	buf := new(strings.Builder)
	if e.TextRun != nil {
		buf.WriteString(p.ParseDocxTextElementTextRun(e.TextRun))
	}
	if e.MentionUser != nil {
		buf.WriteString(e.MentionUser.UserID)
	}
	if e.MentionDoc != nil {
		buf.WriteString(fmt.Sprintf("[%s](%s)", e.MentionDoc.Title, utils.UnescapeURL(e.MentionDoc.URL)))
	}
	if e.Equation != nil {
		symbol := "$$"
		if inline {
			symbol = "$"
		}
		buf.WriteString(symbol + strings.TrimSuffix(e.Equation.Content, "\n") + symbol)
	}
	return buf.String()
}

func (p *Parser) ParseDocxTextElementTextRun(tr *lark.DocxTextElementTextRun) string {
	buf := new(strings.Builder)
	postWrite := ""
	if style := tr.TextElementStyle; style != nil {
		if style.Bold {
			buf.WriteString("<strong>")
			postWrite = "</strong>"
		} else if style.Italic {
			buf.WriteString("<em>")
			postWrite = "</em>"
		} else if style.Strikethrough {
			buf.WriteString("<del>")
			postWrite = "</del>"
		} else if style.Underline {
			buf.WriteString("<u>")
			postWrite = "</u>"
		} else if style.InlineCode {
			buf.WriteString("`")
			postWrite = "`"
		} else if link := style.Link; link != nil {
			buf.WriteString("[")
			postWrite = fmt.Sprintf("](%s)", utils.UnescapeURL(link.URL))
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

func (p *Parser) ParseDocxBlockTableCell(blockId string, blockMap *orderedmap.OrderedMap) string {
	var contents string
	for _, key := range blockMap.Keys() {
		value, ok := blockMap.Get(key)
		if !ok {
			continue
		}
		block := value.(*lark.DocxBlock)
		if block.ParentID != blockId {
			continue
		}

		content := p.ParseDocxBlock(block, blockMap)
		if content == "" {
			continue
		}
		contents += content
		// remove table cell children block from map
		blockMap.Delete(block.BlockID)
	}
	contents = strings.Join(strings.Fields(strings.ReplaceAll(strings.TrimSpace(strip.StripTags(contents)), "\n", "<br/>")), " ")
	return contents
}

func (p *Parser) ParseDocxBlockTable(documentId string, t *lark.DocxBlockTable, blockMap *orderedmap.OrderedMap) string {
	// - First row as header
	// - Ignore cell merging
	var rows [][]string
	for i, blockId := range t.Cells {
		block, ok := blockMap.Get(blockId)
		if !ok {
			log.Printf("got invalid block cell '%s', document: %s\n", blockId, documentId)
			continue
		}

		content := p.ParseDocxBlock(block.(*lark.DocxBlock), blockMap)
		rowIndex := int64(i) / t.Property.ColumnSize
		if len(rows) < int(rowIndex)+1 {
			rows = append(rows, []string{})
		}
		rows[rowIndex] = append(rows[rowIndex], content)
		// remove table cell block from map
		blockMap.Delete(blockId)
	}

	buf := new(strings.Builder)
	buf.WriteString("\n")
	buf.WriteString(renderMarkdownTable(rows))
	buf.WriteString("\n")
	return buf.String()
}

func (p *Parser) ParseDocxBlockQuoteContainer(blockId string, q *lark.DocxBlocQuoteContainer, blockMap *orderedmap.OrderedMap) string {
	contents := "> "
	for _, key := range blockMap.Keys() {
		value, ok := blockMap.Get(key)
		if !ok {
			continue
		}
		block := value.(*lark.DocxBlock)
		if block.ParentID != blockId {
			continue
		}

		content := p.ParseDocxBlock(block, blockMap)
		if content == "" {
			continue
		}
		contents += content
		// remove quote container children block from map
		blockMap.Delete(block.BlockID)
	}
	contents = strings.Join(strings.Fields(strings.ReplaceAll(strings.TrimSpace(strip.StripTags(contents)), "\n", "<br/>")), " ")
	return contents
}
