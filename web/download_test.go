package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/chyroc/lark"
	"github.com/gin-gonic/gin"
)

type paragraphImageTransport func(*http.Request) (*http.Response, error)

func (f paragraphImageTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestDownloadZIPIncludesParagraphImages(t *testing.T) {
	oldTransport, oldMode := http.DefaultTransport, gin.Mode()
	t.Cleanup(func() { http.DefaultTransport = oldTransport; gin.SetMode(oldMode) })
	gin.SetMode(gin.TestMode)
	t.Setenv("FEISHU_APP_ID", "paragraph-test-app")
	t.Setenv("FEISHU_APP_SECRET", "paragraph-test-secret")
	requests := make(map[string]int)
	// 模拟 SDK 的 HTTP 边界；不读取真实凭据，也不向飞书发送请求。
	http.DefaultTransport = paragraphImageTransport(func(r *http.Request) (*http.Response, error) {
		header := http.Header{"Content-Type": {"application/json"}}
		var data interface{}
		switch r.URL.Path {
		case "/open-apis/auth/v3/tenant_access_token/internal":
			data = map[string]interface{}{"code": 0, "tenant_access_token": "test-token", "expire": 7200}
		case "/open-apis/docx/v1/documents/document":
			data = map[string]interface{}{"data": map[string]interface{}{"document": &lark.DocxDocument{DocumentID: "document", Title: "Images"}}}
		case "/open-apis/docx/v1/documents/document/blocks":
			text := func(s string) *lark.DocxBlockText {
				return &lark.DocxBlockText{Elements: []*lark.DocxTextElement{{TextRun: &lark.DocxTextElementTextRun{Content: s}}}}
			}
			data = map[string]interface{}{"data": map[string]interface{}{"items": []*lark.DocxBlock{
				{BlockID: "document", BlockType: lark.DocxBlockTypePage, Page: text("Images"), Children: []string{"parent", "root"}},
				{BlockID: "parent", BlockType: lark.DocxBlockTypeText, Text: text("Paragraph"), Children: []string{"nested"}},
				{BlockID: "nested", BlockType: lark.DocxBlockTypeImage, Image: &lark.DocxBlockImage{Token: "nested"}},
				{BlockID: "root", BlockType: lark.DocxBlockTypeImage, Image: &lark.DocxBlockImage{Token: "root"}},
			}}}
		case "/open-apis/drive/v1/medias/nested/download", "/open-apis/drive/v1/medias/root/download":
			token := strings.Split(r.URL.Path, "/")[5]
			requests[token]++
			header.Set("Content-Type", "image/png")
			header.Set("Content-Disposition", `attachment; filename="image.png"`)
			return &http.Response{StatusCode: 200, Header: header, Body: io.NopCloser(strings.NewReader(token + "-data"))}, nil
		default:
			return nil, fmt.Errorf("unexpected request: %s", r.URL.Path)
		}
		body, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		return &http.Response{StatusCode: 200, Header: header, Body: io.NopCloser(bytes.NewReader(body))}, nil
	})
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest("GET", "/download?url="+url.QueryEscape("https://example.feishu.cn/docx/document"), nil)
	downloadHandler(ctx)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d: %s", recorder.Code, recorder.Body)
	}
	reader, err := zip.NewReader(bytes.NewReader(recorder.Body.Bytes()), int64(recorder.Body.Len()))
	if err != nil {
		t.Fatal(err)
	}
	files := make(map[string]string)
	for _, entry := range reader.File {
		file, err := entry.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			t.Fatal(err)
		}
		files[entry.Name] = string(data)
	}
	if len(files) != 3 {
		t.Errorf("want Markdown and two images, got %v", files)
	}
	for _, token := range []string{"nested", "root"} {
		path := "static/" + token + ".png"
		if requests[token] != 1 || files[path] != token+"-data" {
			t.Errorf("missing or corrupt image %s: requests=%d, files=%v", token, requests[token], files)
		}
		if !strings.Contains(files["document.md"], "![]("+path+")") {
			t.Errorf("missing Markdown link %s: %s", path, files["document.md"])
		}
	}
}
