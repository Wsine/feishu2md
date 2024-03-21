package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Wsine/feishu2md/core"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type web struct {
  app.Compo

  Url string
}

func (w *web) Render() app.UI {
  return app.Div().Body(
    app.H1().Text("feishu2md wasm"),
    app.Input().Type("text").Placeholder("Enter your Feishu URL").OnChange(w.ValueTo(&w.Url)),
    app.Button().Text("Download").OnClick(w.convert),
  )
}

func (w *web) convert(ctx app.Context, e app.Event) {
  fmt.Println(w.Url)
	client := core.NewClient(
		"cli_a267b1fe00b8d00b", "2GQrMtnfY5VzkoHJakLjCdmeoiEPIda6", "feishu.cn",
	)
	docx, blocks, err := client.GetDocxContent(ctx, "doxcnXhd93zqoLnmVPGIPTy7AFe")
  fmt.Println(docx, blocks, err)
}

func main() {
  app.Route("/", &web{})
  app.RunWhenOnBrowser()

  err := app.GenerateStaticWebsite("./dist", &app.Handler{
    Name:       "feishu2md",
    Description: "feishu2md wasm",
    Resources:   app.GitHubPages("test"),
  })
  if err != nil {
    log.Fatal(err)
  }

  http.Handle("/", http.FileServer(http.Dir("./dist")))
  log.Print("Server started at http://localhost:8080")
  err = http.ListenAndServe(":8080", nil)
  if err != nil {
    log.Fatal(err)
  }
}
