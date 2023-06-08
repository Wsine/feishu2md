package main

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/88250/lute"
	"github.com/Wsine/feishu2md/core"
	"github.com/gin-gonic/gin"
)

func downloadHandler(c *gin.Context) {
	// get parameters
	feishu_docx_url, err := url.QueryUnescape(c.Query("url"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid encoded feishu/larksuite URL")
		return
	}

	// Validate the url
	reg := regexp.MustCompile("^https://[a-zA-Z0-9-]+.(feishu.cn|larksuite.com)/(docs|docx|wiki)/([a-zA-Z0-9]+)")
	matchResult := reg.FindStringSubmatch(feishu_docx_url)
	if matchResult == nil || len(matchResult) != 4 {
		c.String(http.StatusBadRequest, "Invalid feishu/larksuite URL pattern")
		return
	}

	config := core.NewConfig(
		os.Getenv("FEISHU_APP_ID"),
		os.Getenv("FEISHU_APP_SECRET"),
	)

	domain := matchResult[1]
	docType := matchResult[2]
	docToken := matchResult[3]

	ctx := context.Background()

	client := core.NewClient(
		config.Feishu.AppId, config.Feishu.AppSecret, domain,
	)

	parser := core.NewParser(ctx)
	markdown := ""

	// for a wiki page, we need to renew docType and docToken first
	if docType == "wiki" {
		node, err := client.GetWikiNodeInfo(ctx, docToken)
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal error: client.GetWikiNodeInfo")
			log.Panicf("error: %s", err)
			return
		}
		docType = node.ObjType
		docToken = node.ObjToken
	}

	docx, blocks, err := client.GetDocxContent(ctx, docToken)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal error: client.GetDocxContent")
		log.Panicf("error: %s", err)
		return
	}
	markdown = parser.ParseDocxContent(docx, blocks)

	zipBuffer := new(bytes.Buffer)
	writer := zip.NewWriter(zipBuffer)
	for _, imgToken := range parser.ImgTokens {
		localLink, rawImage, err := client.DownloadImageRaw(ctx, imgToken, config.Output.ImageDir)
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal error: client.DownloadImageRaw")
			log.Panicf("error: %s", err)
			return
		}
		markdown = strings.Replace(markdown, imgToken, localLink, 1)
		f, err := writer.Create(localLink)
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal error: zipWriter.Create")
			log.Panicf("error: %s", err)
			return
		}
		_, err = f.Write(rawImage)
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal error: zipWriter.Create.Write")
			log.Panicf("error: %s", err)
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
			c.String(http.StatusInternalServerError, "Internal error: zipWriter.Create")
			log.Panicf("error: %s", err)
			return
		}
		_, err = f.Write([]byte(result))
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal error: zipWriter.Create.Write")
			log.Panicf("error: %s", err)
			return
		}

		err = writer.Close()
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal error: zipWriter.Close")
			log.Panicf("error: %s", err)
			return
		}
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.zip"`, docToken))
		c.Data(http.StatusOK, "application/octet-stream", zipBuffer.Bytes())
	} else {
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.md"`, docToken))
		c.Data(http.StatusOK, "application/octet-stream", []byte(result))
	}
}
