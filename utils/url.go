package utils

import (
	"net/url"
	"regexp"

	"github.com/pkg/errors"
)

func UnescapeURL(rawURL string) string {
	if u, err := url.QueryUnescape(rawURL); err == nil {
		return u
	}
	return rawURL
}

func ValidateDocumentURL(url string) (string, string, error) {
	reg := regexp.MustCompile("^https://[\\w-.]+/(docs|docx|wiki)/([a-zA-Z0-9]+)")
	matchResult := reg.FindStringSubmatch(url)
	if matchResult == nil || len(matchResult) != 3 {
		return "", "", errors.Errorf("Invalid feishu/larksuite document URL pattern")
	}
	docType := matchResult[1]
	docToken := matchResult[2]
	return docType, docToken, nil
}

func ValidateFolderURL(url string) (string, error) {
  reg := regexp.MustCompile("^https://[\\w-.]+/drive/folder/([a-zA-Z0-9]+)")
  matchResult := reg.FindStringSubmatch(url)
  if matchResult == nil || len(matchResult) != 2 {
    return "", errors.Errorf("Invalid feishu/larksuite folder URL pattern")
  }
  folderToken := matchResult[1]
  return folderToken, nil
}
