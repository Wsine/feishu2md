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
	type Args struct {
		url       string
		allowHost string
	}
	tests := []struct {
		name  string
		args  Args
		noErr bool
	}{
		{
			name: "validate feishu url success",
			args: Args{
				url:       "https://sample.feishu.cn/docx/doccnByZP6puODElAYySJkPIfUb",
				allowHost: "",
			},
			noErr: true,
		},
		{
			name: "validate larksuite url success",
			args: Args{
				url:       "https://sample.larksuite.com/wiki/doccnByZP6puODElAYySJkPIfUb",
				allowHost: "",
			},
			noErr: true,
		},
		{
			name: "validate feishu url success with allow host",
			args: Args{
				url:       "https://f.mioffice.cn/docx/doccnByZP6puODElAYySJkPIfUb",
				allowHost: "f.mioffice.cn",
			},
			noErr: true,
		},
		{
			name: "validate arbitrary url failed",
			args: Args{
				url:       "https://google.com",
				allowHost: "",
			},
			noErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, allowHost := tt.args.url, tt.args.allowHost
			if _, _, _, got := ValidateDownloadURL(url, allowHost); (got == nil) != tt.noErr {
				t.Errorf("ValidateDownloadURL(%v, %v)", url, allowHost)
			}
		})
	}
}
