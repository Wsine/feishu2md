package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

var StopWhenErr = true

func CheckErr(e error) error {
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		fmt.Fprintf(
			os.Stderr,
			"\n%s\n\n%s\n\n",
			strings.Repeat("=", 20),
			"Report the following if it is a bug",
		)
		if StopWhenErr {
			panic(e)
		}
	}
	return e
}

func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "  ")
	return string(s)
}

func SanitizeFileName(title string) string {
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		title = strings.ReplaceAll(title, char, "_")
	}
	return title
}
