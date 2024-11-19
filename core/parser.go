package core

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Wsine/feishu2md/utils"
	"github.com/chyroc/lark"
	"github.com/olekukonko/tablewriter"
)

type Parser struct {
	useHTMLTags bool
	ImgTokens   []string
	blockMap    map[string]*lark.DocxBlock
}

func NewParser(config OutputConfig) *Parser {
	return &Parser{
		useHTMLTags: config.UseHTMLTags,
		ImgTokens:   make([]string, 0),
		blockMap:    make(map[string]*lark.DocxBlock),
	}
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
	for _, block := range blocks {
		p.blockMap[block.BlockID] = block
	}

	entryBlock := p.blockMap[doc.DocumentID]
	return p.ParseDocxBlock(entryBlock, 0)
}

func (p *Parser) ParseDocxBlock(b *lark.DocxBlock, indentLevel int) string {
	buf := new(strings.Builder)
	buf.WriteString(strings.Repeat("\t", indentLevel))
	switch b.BlockType {
	case lark.DocxBlockTypePage:
		buf.WriteString(p.ParseDocxBlockPage(b))
	case lark.DocxBlockTypeText:
		buf.WriteString(p.ParseDocxBlockText(b.Text))
	case lark.DocxBlockTypeCallout:
		buf.WriteString(p.ParseDocxBlockCallout(b))
	case lark.DocxBlockTypeHeading1:
		buf.WriteString(p.ParseDocxBlockHeading(b, 1))
	case lark.DocxBlockTypeHeading2:
		buf.WriteString(p.ParseDocxBlockHeading(b, 2))
	case lark.DocxBlockTypeHeading3:
		buf.WriteString(p.ParseDocxBlockHeading(b, 3))
	case lark.DocxBlockTypeHeading4:
		buf.WriteString(p.ParseDocxBlockHeading(b, 4))
	case lark.DocxBlockTypeHeading5:
		buf.WriteString(p.ParseDocxBlockHeading(b, 5))
	case lark.DocxBlockTypeHeading6:
		buf.WriteString(p.ParseDocxBlockHeading(b, 6))
	case lark.DocxBlockTypeHeading7:
		buf.WriteString(p.ParseDocxBlockHeading(b, 7))
	case lark.DocxBlockTypeHeading8:
		buf.WriteString(p.ParseDocxBlockHeading(b, 8))
	case lark.DocxBlockTypeHeading9:
		buf.WriteString(p.ParseDocxBlockHeading(b, 9))
	case lark.DocxBlockTypeBullet:
		buf.WriteString(p.ParseDocxBlockBullet(b, indentLevel))
	case lark.DocxBlockTypeOrdered:
		buf.WriteString(p.ParseDocxBlockOrdered(b, indentLevel))
	case lark.DocxBlockTypeCode:
		buf.WriteString("```" + DocxCodeLang2MdStr[b.Code.Style.Language] + "\n")
		buf.WriteString(strings.TrimSpace(p.ParseDocxBlockText(b.Code)))
		buf.WriteString("\n```\n")
	case lark.DocxBlockTypeQuote:
		buf.WriteString("> ")
		buf.WriteString(p.ParseDocxBlockText(b.Quote))
	case lark.DocxBlockTypeEquation:
		buf.WriteString("$$\n")
		buf.WriteString(p.ParseDocxBlockText(b.Equation))
		buf.WriteString("\n$$\n")
	case lark.DocxBlockTypeTodo:
		if b.Todo.Style.Done {
			buf.WriteString("- [x] ")
		} else {
			buf.WriteString("- [ ] ")
		}
		buf.WriteString(p.ParseDocxBlockText(b.Todo))
	case lark.DocxBlockTypeDivider:
		buf.WriteString("---\n")
	case lark.DocxBlockTypeImage:
		buf.WriteString(p.ParseDocxBlockImage(b.Image))
	case lark.DocxBlockTypeTableCell:
		buf.WriteString(p.ParseDocxBlockTableCell(b))
	case lark.DocxBlockTypeTable:
		buf.WriteString(p.ParseDocxBlockTable(b.Table))
	case lark.DocxBlockTypeQuoteContainer:
		buf.WriteString(p.ParseDocxBlockQuoteContainer(b))
	case lark.DocxBlockTypeGrid:
		buf.WriteString(p.ParseDocxBlockGrid(b, indentLevel))
	default:
	}
	return buf.String()
}

func (p *Parser) ParseDocxBlockPage(b *lark.DocxBlock) string {
	buf := new(strings.Builder)

	buf.WriteString("# ")
	buf.WriteString(p.ParseDocxBlockText(b.Page))
	buf.WriteString("\n")

	for _, childId := range b.Children {
		childBlock := p.blockMap[childId]
		buf.WriteString(p.ParseDocxBlock(childBlock, 0))
		buf.WriteString("\n")
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

func (p *Parser) ParseDocxBlockCallout(b *lark.DocxBlock) string {
	buf := new(strings.Builder)

	buf.WriteString(">[!TIP] \n")

	for _, childId := range b.Children {
		childBlock := p.blockMap[childId]
		buf.WriteString(p.ParseDocxBlock(childBlock, 0))
	}

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
		buf.WriteString(
			fmt.Sprintf("[%s](%s)", e.MentionDoc.Title, utils.UnescapeURL(e.MentionDoc.URL)))
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
			if p.useHTMLTags {
				buf.WriteString("<strong>")
				postWrite = "</strong>"
			} else {
				buf.WriteString("**")
				postWrite = "**"
			}
		} else if style.Italic {
			if p.useHTMLTags {
				buf.WriteString("<em>")
				postWrite = "</em>"
			} else {
				buf.WriteString("_")
				postWrite = "_"
			}
		} else if style.Strikethrough {
			if p.useHTMLTags {
				buf.WriteString("<del>")
				postWrite = "</del>"
			} else {
				buf.WriteString("~~")
				postWrite = "~~"
			}
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

func (p *Parser) ParseDocxBlockHeading(b *lark.DocxBlock, headingLevel int) string {
	buf := new(strings.Builder)

	buf.WriteString(strings.Repeat("#", headingLevel))
	buf.WriteString(" ")

	headingText := reflect.ValueOf(b).Elem().FieldByName(fmt.Sprintf("Heading%d", headingLevel))
	buf.WriteString(p.ParseDocxBlockText(headingText.Interface().(*lark.DocxBlockText)))

	for _, childId := range b.Children {
		childBlock := p.blockMap[childId]
		buf.WriteString(p.ParseDocxBlock(childBlock, 0))
	}

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

func (p *Parser) ParseDocxBlockBullet(b *lark.DocxBlock, indentLevel int) string {
	buf := new(strings.Builder)

	buf.WriteString("- ")
	buf.WriteString(p.ParseDocxBlockText(b.Bullet))

	for _, childId := range b.Children {
		childBlock := p.blockMap[childId]
		buf.WriteString(p.ParseDocxBlock(childBlock, indentLevel+1))
	}

	return buf.String()
}

func (p *Parser) ParseDocxBlockOrdered(b *lark.DocxBlock, indentLevel int) string {
	buf := new(strings.Builder)

	// calculate order and indent level
	parent := p.blockMap[b.ParentID]
	order := 1
	for idx, child := range parent.Children {
		if child == b.BlockID {
			for i := idx - 1; i >= 0; i-- {
				if p.blockMap[parent.Children[i]].BlockType == lark.DocxBlockTypeOrdered {
					order += 1
				} else {
					break
				}
			}
			break
		}
	}

	buf.WriteString(fmt.Sprintf("%d. ", order))
	buf.WriteString(p.ParseDocxBlockText(b.Ordered))

	for _, childId := range b.Children {
		childBlock := p.blockMap[childId]
		buf.WriteString(p.ParseDocxBlock(childBlock, indentLevel+1))
	}

	return buf.String()
}

func (p *Parser) ParseDocxBlockTableCell(b *lark.DocxBlock) string {
	buf := new(strings.Builder)

	for _, child := range b.Children {
		block := p.blockMap[child]
		content := p.ParseDocxBlock(block, 0)
		buf.WriteString(content + "<br/>")
	}

	return buf.String()
}

func (p *Parser) ParseDocxBlockTable(t *lark.DocxBlockTable) string {
	// - First row as header
	// - Ignore cell merging
	// - First row as header
	// - Ignore cell merging
	var rows [][]string
	mergeInfoMap := map[int64]map[int64]*lark.DocxBlockTablePropertyMergeInfo{}

	// 构建单元格合并信息的映射
	if t.Property.MergeInfo != nil {
		for i, merge := range t.Property.MergeInfo {
			rowIndex := int64(i) / t.Property.ColumnSize
			colIndex := int64(i) % t.Property.ColumnSize
			if _, exists := mergeInfoMap[int64(rowIndex)]; !exists {
				mergeInfoMap[int64(rowIndex)] = map[int64]*lark.DocxBlockTablePropertyMergeInfo{}
			}
			mergeInfoMap[rowIndex][colIndex] = merge
		}
	}
	for i, blockId := range t.Cells {
		block := p.blockMap[blockId]
		cellContent := p.ParseDocxBlock(block, 0)
		cellContent = strings.ReplaceAll(cellContent, "\n", "")
		rowIndex := int64(i) / t.Property.ColumnSize
		colIndex := int64(i) % t.Property.ColumnSize
		// 初始化行
		for len(rows) <= int(rowIndex) {
			rows = append(rows, []string{})
		}
		for len(rows[rowIndex]) <= int(colIndex) {
			rows[rowIndex] = append(rows[rowIndex], "")
		}
		// 设置单元格内容
		rows[rowIndex][colIndex] = cellContent
	}

	// 渲染为 HTML 表格
	buf := new(strings.Builder)
	buf.WriteString("<table>\n")

	// 跟踪已经处理过的合并单元格
	processedCells := map[string]bool{}

	// 构建 HTML 表格内容
	for rowIndex, row := range rows {
		buf.WriteString("<tr>\n")
		for colIndex, cellContent := range row {
			cellKey := fmt.Sprintf("%d-%d", rowIndex, colIndex)

			// 跳过已处理的单元格
			if processedCells[cellKey] {
				continue
			}

			mergeInfo := mergeInfoMap[int64(rowIndex)][int64(colIndex)]
			if mergeInfo != nil {
				// 合并单元格
				buf.WriteString(fmt.Sprintf(
					`<td rowspan="%d" colspan="%d">%s</td>`,
					mergeInfo.RowSpan, mergeInfo.ColSpan, cellContent,
				))
				// 标记合并范围内的所有单元格为已处理
				for r := rowIndex; r < rowIndex+int(mergeInfo.RowSpan); r++ {
					for c := colIndex; c < colIndex+int(mergeInfo.ColSpan); c++ {
						processedCells[fmt.Sprintf("%d-%d", r, c)] = true
					}
				}
			} else {
				// 普通单元格
				buf.WriteString(fmt.Sprintf("<td>%s</td>", cellContent))
			}
		}
		buf.WriteString("</tr>\n")
	}
	buf.WriteString("</table>\n")

	return buf.String()
}

func (p *Parser) ParseDocxBlockQuoteContainer(b *lark.DocxBlock) string {
	buf := new(strings.Builder)

	for _, child := range b.Children {
		block := p.blockMap[child]
		buf.WriteString("> ")
		buf.WriteString(p.ParseDocxBlock(block, 0))
	}

	return buf.String()
}

func (p *Parser) ParseDocxBlockGrid(b *lark.DocxBlock, indentLevel int) string {
	buf := new(strings.Builder)

	for _, child := range b.Children {
		columnBlock := p.blockMap[child]
		for _, child := range columnBlock.Children {
			block := p.blockMap[child]
			buf.WriteString(p.ParseDocxBlock(block, indentLevel))
		}
	}

	return buf.String()
}

