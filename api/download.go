package handler

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/88250/lute"
	"github.com/Wsine/feishu2md/core"
)

func Handler(w http.ResponseWriter, r *http.Request) {
  // get parameters
  user_access_token := r.URL.Query().Get("code")
  feishu_docx_url, err := url.QueryUnescape(r.URL.Query().Get("state"))
  if err != nil {
    fmt.Fprintf(w, "<h1>Invalid encoded feishu/larksuite URL</h1>")
		return
  }

  // Validate the url
  reg := regexp.MustCompile("^https://[a-zA-Z0-9-]+.(feishu.cn|larksuite.com)/(docs|docx|wiki)/([a-zA-Z0-9]+)")
	matchResult := reg.FindStringSubmatch(feishu_docx_url)
	if matchResult == nil || len(matchResult) != 4 {
    fmt.Fprintf(w, "<h1>Invalid feishu/larksuite URL pattern</h1>")
		return
	}

  config := core.NewConfig(
    os.Getenv("FEISHU_APP_ID"),
    user_access_token,
  )

	domain := matchResult[1]
	docType := matchResult[2]
	docToken := matchResult[3]
	fmt.Println("Captured document token:", docToken)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "OutputConfig", config.Output)

	client := core.NewClient(
		config.Feishu.AppId, config.Feishu.AppSecret, domain,
	)

	parser := core.NewParser(ctx)
	title := ""
	markdown := ""

	// for a wiki page, we need to renew docType and docToken first
	if docType == "wiki" {
		node, err := client.GetWikiNodeInfo(ctx, docToken)
		if err != nil {
      fmt.Fprintf(w, "<h1>Internal error: client.GetWikiNodeInfo</h1>")
			return
		}
		docType = node.ObjType
		docToken = node.ObjToken
	}

	if docType == "docx" {
		docx, blocks, err := client.GetDocxContent(ctx, docToken)
		if err != nil {
      fmt.Fprintf(w, "<h1>Internal error: client.GetDocxContent</h1>")
			return
		}
		markdown = parser.ParseDocxContent(docx, blocks)
		title = docx.Title
	} else {
		doc, err := client.GetDocContent(ctx, docToken)
		if err != nil {
      fmt.Fprintf(w, "<h1>Internal error: client.GetDocContent</h1>")
			return
		}
		markdown = parser.ParseDocContent(doc)
		for _, element := range doc.Title.Elements {
			title += element.TextRun.Text
		}
	}

  zipBuffer := new(bytes.Buffer)
  writer := zip.NewWriter(zipBuffer)
	for _, imgToken := range parser.ImgTokens {
		localLink, rawImage, err := client.DownloadImageRaw(ctx, imgToken)
		if err != nil {
      fmt.Fprintf(w, "<h1>Internal error: client.DownloadImageRaw</h1>")
			return
		}
		markdown = strings.Replace(markdown, imgToken, localLink, 1)
    f, err := writer.Create(localLink)
    if err != nil {
      fmt.Fprintf(w, "<h1>Internal error: zipWriter.Create</h1>")
			return
    }
    _, err = f.Write(rawImage)
    if err != nil {
      fmt.Fprintf(w, "<h1>Internal error: zipWriter.Create.Write</h1>")
			return
    }
	}

	engine := lute.New(func(l *lute.Lute) {
		l.RenderOptions.AutoSpace = true
	})
	result := engine.FormatStr("md", markdown)

  // Set response
  if len(parser.ImgTokens) > 0 {
    mdName := fmt.Sprintf("%s.md", docToken)
    f, err := writer.Create(mdName)
    if err != nil {
      fmt.Fprintf(w, "<h1>Internal error: zipWriter.Create</h1>")
			return
    }
    _, err = f.Write([]byte(result))
    if err != nil {
      fmt.Fprintf(w, "<h1>Internal error: zipWriter.Create.Write</h1>")
			return
    }

    err = writer.Close()
    if err != nil {
      fmt.Fprintf(w, "<h1>Internal error: zipWriter.Close</h1>")
			return
    }
    w.Header().Set("Content-Type", "application/zip")
    w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.zip"`, docToken))
    w.Write(zipBuffer.Bytes())
  } else {
    w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.md"`, docToken))
    w.Header().Set("Content-Type", "text/markdown")
    fmt.Fprintf(w, result)
  }
}
