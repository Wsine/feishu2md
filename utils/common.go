package utils

import (
  "fmt"
  "strings"
)

func CheckErr(e error) {
  if e != nil {
    fmt.Println(e)
    fmt.Printf(
      "\n%s\n\n%s\n\n",
      strings.Repeat("=", 20),
      "Report the following if it is a bug",
    )
    panic(e)
  }
}

