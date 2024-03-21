package utils

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

func UnescapeURL(rawURL string) string {
	if u, err := url.QueryUnescape(rawURL); err == nil {
		return u
	}
	return rawURL
}

func ValidateDownloadURL(url string) (string, string, string, error) {
	hosts := []string{"feishu.cn", "larksuite.com", "larkoffice.com"}

	reg := regexp.MustCompile("^https://([\\w-]+.)?(" + strings.Join(hosts, "|") + ")/(docs|docx|wiki)/([a-zA-Z0-9]+)")
	matchResult := reg.FindStringSubmatch(url)
	if matchResult == nil || len(matchResult) != 5 {
		return "", "", "", errors.Errorf("Invalid document URL format")
	}

	return matchResult[2], matchResult[3], matchResult[4], nil
}
