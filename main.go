package main

import (
  "context"
  "fmt"
  "os"

  "github.com/chyroc/lark"
  "github.com/chyroc/lark/larkext"
  "github.com/chyroc/lark_docs_md"
  "github.com/joho/godotenv"
)

func checkErr(e error) {
  if e != nil {
    panic(e)
  }
}

func main() {
  godotenv.Load()
  appId := os.Getenv("FEISHU_APP_ID")
  appSecret := os.Getenv("FEISHU_APP_SECRET")
  client := lark.New(
    lark.WithAppCredential(appId, appSecret),
  )

  docToken := os.Getenv("FEISHU_DOC_TOKEN")
  doc, err := larkext.NewDoc(client, docToken).Content(context.Background())
  checkErr(err)

  result := lark_docs_md.DocMarkdown(context.Background(), doc, &lark_docs_md.FormatOpt{
    LarkClient: client,
    StaticDir: "static",
    FilePrefix: "static",
  })
  fmt.Println(result)

  f, err := os.Create("test.md")
  checkErr(err)
  defer f.Close()
  nBytes, err := f.WriteString(result)
  checkErr(err)
  fmt.Printf("Wrote %d bytes\n", nBytes)
  f.Sync()
}

