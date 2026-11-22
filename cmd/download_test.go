package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Wsine/feishu2md/core"
	"github.com/chyroc/lark"
)

type imageDownloadTransport func(*http.Request) (*http.Response, error)

func (f imageDownloadTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestDownloadDocumentImageLinks(t *testing.T) {
	root := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	oldConfig, oldOpts, oldTransport := dlConfig, dlOpts, http.DefaultTransport
	t.Cleanup(func() {
		dlConfig, dlOpts, http.DefaultTransport = oldConfig, oldOpts, oldTransport
		if err := os.Chdir(oldWD); err != nil {
			t.Error(err)
		}
	})

	tests := []struct {
		name, outputDir, imageDir string
		wiki, skip                bool
	}{
		{name: "current directory", outputDir: ".", imageDir: "static"},
		{name: "nested wiki", outputDir: filepath.Join("Wiki", "Chapter"), imageDir: "static", wiki: true},
		{name: "absolute output", outputDir: filepath.Join(root, "absolute"), imageDir: "static"},
		{name: "custom image directory", outputDir: "custom", imageDir: filepath.Join("assets", "pictures")},
		{name: "parent image directory", outputDir: "shared", imageDir: filepath.Join("..", "images")},
		{name: "skip image download", outputDir: "skip", imageDir: "static", skip: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imageRequests := 0
			// 所有 HTTP 请求都在内存中响应，未知请求直接失败，不访问真实飞书服务。
			http.DefaultTransport = imageDownloadTransport(func(r *http.Request) (*http.Response, error) {
				var data interface{}
				header := http.Header{"Content-Type": {"application/json"}}
				switch r.URL.Path {
				case "/open-apis/auth/v3/tenant_access_token/internal":
					data = map[string]interface{}{"code": 0, "tenant_access_token": "test-token", "expire": 7200}
				case "/open-apis/wiki/v2/spaces/get_node":
					data = map[string]interface{}{"data": map[string]interface{}{"node": map[string]string{"obj_type": "docx", "obj_token": "document"}}}
				case "/open-apis/docx/v1/documents/document":
					data = map[string]interface{}{"data": map[string]interface{}{"document": &lark.DocxDocument{DocumentID: "document", Title: "Sample"}}}
				case "/open-apis/docx/v1/documents/document/blocks":
					data = map[string]interface{}{"data": map[string]interface{}{"items": []*lark.DocxBlock{
						{BlockID: "document", BlockType: lark.DocxBlockTypePage, Children: []string{"image"}, Page: &lark.DocxBlockText{Elements: []*lark.DocxTextElement{{TextRun: &lark.DocxTextElementTextRun{Content: "Sample"}}}}},
						{BlockID: "image", BlockType: lark.DocxBlockTypeImage, Image: &lark.DocxBlockImage{Token: "asset"}},
					}}}
				case "/open-apis/drive/v1/medias/asset/download":
					imageRequests++
					header.Set("Content-Type", "image/png")
					header.Set("Content-Disposition", `attachment; filename="asset.png"`)
					return &http.Response{StatusCode: http.StatusOK, Header: header, Body: io.NopCloser(strings.NewReader("image-data"))}, nil
				default:
					return nil, fmt.Errorf("unexpected request: %s", r.URL.Path)
				}
				body, err := json.Marshal(data)
				if err != nil {
					return nil, err
				}
				return &http.Response{StatusCode: http.StatusOK, Header: header, Body: io.NopCloser(strings.NewReader(string(body)))}, nil
			})
			dlConfig = *core.NewConfig("test-app", "test-secret")
			dlConfig.Output.ImageDir = tt.imageDir
			dlConfig.Output.SkipImgDownload = tt.skip
			dlOpts = DownloadOpts{}
			url := "https://example.feishu.cn/docx/document"
			if tt.wiki {
				url = "https://example.feishu.cn/wiki/node"
			}
			if err := downloadDocument(context.Background(), core.NewClient("test-app", "test-secret"), url, &DownloadOpts{outputDir: tt.outputDir}); err != nil {
				t.Fatal(err)
			}
			markdown, err := os.ReadFile(filepath.Join(tt.outputDir, "document.md"))
			if err != nil {
				t.Fatal(err)
			}
			if tt.skip {
				if imageRequests != 0 || !strings.Contains(string(markdown), "![](asset)") {
					t.Fatalf("skipped image download changed: requests=%d markdown=%s", imageRequests, markdown)
				}
				return
			}
			wantLink := filepath.ToSlash(filepath.Join(tt.imageDir, "asset.png"))
			if !strings.Contains(string(markdown), "![]("+wantLink+")") {
				t.Errorf("image link should be relative to document: want %q in %s", wantLink, markdown)
			}
			image, err := os.ReadFile(filepath.Join(tt.outputDir, filepath.FromSlash(wantLink)))
			if err != nil || string(image) != "image-data" || imageRequests != 1 {
				t.Fatalf("image file mismatch: requests=%d data=%q err=%v", imageRequests, image, err)
			}
		})
	}
}
