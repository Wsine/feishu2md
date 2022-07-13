package core

import (
	"context"

	"github.com/Wsine/feishu2md/utils"
	"github.com/larksuite/oapi-sdk-go/core"
	doc "github.com/larksuite/oapi-sdk-go/service/doc/v2"
)

type DocClient struct {
  docService *doc.Service
}

func NewDocClient(appID, appSecret, domain string) *DocClient {
  c := new(DocClient)

  settings := core.NewInternalAppSettings(
    core.SetAppCredentials(appID, appSecret),
  )
  conf := core.NewConfig(
    core.Domain("https://open." + domain), settings,
  )
  c.docService = doc.NewService(conf)

  return c
}

func (c *DocClient) GetContent(ctx context.Context, docToken string) string {
  coreCtx := core.WrapContext(ctx)
  reqCall := c.docService.Docs.Content(coreCtx)
  reqCall.SetDocToken(docToken)
  result, err := reqCall.Do()
  utils.CheckErr(err)
  return result.Content
}
