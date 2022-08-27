package utils

import (
	"testing"
)

func TestUnescapeURL(t *testing.T) {
	type args struct {
		rawURL string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "url unescape success",
			args: args{
				rawURL: "https%3A%2F%2Fsspai.com%2Fpost%2F58509",
			},
			want: "https://sspai.com/post/58509",
		},
		{
			name: "url unescape failed, keep it",
			args: args{
				rawURL: "https$3A$2F$2Fsspai.com$2Fpost$2F58509",
			},
			want: "https$3A$2F$2Fsspai.com$2Fpost$2F58509",
		},
		{
			name: "url not need to unescape, keep it",
			args: args{
				rawURL: "https://sample.feishu.cn/docs/doccnByZP6puODElAYySJkPIfUb",
			},
			want: "https://sample.feishu.cn/docs/doccnByZP6puODElAYySJkPIfUb",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UnescapeURL(tt.args.rawURL); got != tt.want {
				t.Errorf("unescapeURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
