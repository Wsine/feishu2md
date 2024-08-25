package utils

import (
	"testing"
)

func TestUnescapeURL(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
		want   string
	}{
		{
			name:   "url unescape success",
			rawURL: "https%3A%2F%2Fsspai.com%2Fpost%2F58509",
			want:   "https://sspai.com/post/58509",
		},
		{
			name:   "url unescape failed, keep it",
			rawURL: "https$3A$2F$2Fsspai.com$2Fpost$2F58509",
			want:   "https$3A$2F$2Fsspai.com$2Fpost$2F58509",
		},
		{
			name:   "url not need to unescape, keep it",
			rawURL: "https://sample.feishu.cn/docs/doccnByZP6puODElAYySJkPIfUb",
			want:   "https://sample.feishu.cn/docs/doccnByZP6puODElAYySJkPIfUb",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UnescapeURL(tt.rawURL); got != tt.want {
				t.Errorf("URL = %v\nGot = %v\nExpected = %v", tt.rawURL, got, tt.want)
			}
		})
	}
}

func TestValidateDownloadURL(t *testing.T) {
	tests := []struct {
		name  string
		url   string
		noErr bool
	}{
		{
			name:  "validate feishu url success",
			url:   "https://sample.feishu.cn/docx/doccnByZP6puODElAYySJkPIfUb",
			noErr: true,
		},
		{
			name:  "validate larksuite url success",
			url:   "https://sample.larksuite.com/wiki/doccnByZP6puODElAYySJkPIfUb",
			noErr: true,
		},
		{
			name:  "validate larksuite url success",
			url:   "https://sample.sg.larksuite.com/wiki/doccnByZP6puODElAYySJkPIfUb",
			noErr: true,
		},
		{
			name:  "validate feishu url success",
			url:   "https://sample.f.mioffice.cn/docx/doccnByZP6puODElAYySJkPIfUb",
			noErr: true,
		},
		{
			name:  "validate arbitrary url failed",
			url:   "https://google.com",
			noErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, got := ValidateDocumentURL(tt.url); (got == nil) != tt.noErr {
				t.Errorf("ValidateDownloadURL(%v)", tt.url)
			}
		})
	}
}

func TestValidWikiURL(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		prefix string
		token  string
		noErr  bool
	}{
		{
			name:   "valid wiki setting success",
			url:    "",
			prefix: "",
			token:  "",
			noErr:  false,
		},
		{
			name:   "validate docs url failed",
			url:    "https://sample.sg.larksuite.com/wiki/doccnByZP6puODElAYySJkPIfUb",
			prefix: "",
			token:  "",
			noErr:  false,
		},
		{
			name:   "validate feishu url failed",
			url:    "https://sample.feishu.cn/docx/doccnByZP6puODElAYySJkPIfUb",
			prefix: "",
			token:  "",
			noErr:  false,
		},
		{
			name:   "validate larksuite wiki settings success",
			url:    "https://sample.sg.larksuite.com/wiki/settings/doccnByZP6puODElAYySJkPIfUb",
			prefix: "https://sample.sg.larksuite.com",
			token:  "doccnByZP6puODElAYySJkPIfUb",
			noErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if prefix, token, got := ValidateWikiURL(tt.url); (got == nil) != tt.noErr || prefix != tt.prefix || token != tt.token {
				t.Errorf("ValidateWikiURL(%v) = %v, %v; want prefix = %v, want token = %v", tt.url, prefix, token, tt.prefix, tt.token)
			}
		})
	}
}
