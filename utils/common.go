package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func CheckErr(e error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		fmt.Fprintf(
			os.Stderr,
			"\n%s\n\n%s\n\n",
			strings.Repeat("=", 20),
			"Report the following if it is a bug",
		)
		panic(e)
	}
}

func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "  ")
	return string(s)
}
