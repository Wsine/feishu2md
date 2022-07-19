package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/chyroc/lark"
)

type Client struct {
  larkCilent *lark.Lark
}

func NewClient(appID, appSecret, domain string) *Client {
  return &Client{
    larkCilent: lark.New(
      lark.WithAppCredential(appID, appSecret),
      lark.WithOpenBaseURL("https://open." + domain),
    ),
  }
}

func (c *Client) GetDocContent(ctx context.Context, docToken string) (*lark.DocContent, error) {
  resp, _, err := c.larkCilent.Drive.GetDriveDocContent(ctx, &lark.GetDriveDocContentReq{
    DocToken: docToken,
  })
  if err != nil {
    return nil, err
  }
  doc := &lark.DocContent{}
  err = json.Unmarshal([]byte(resp.Content), doc)
  if err != nil {
    return doc, err
  }
  return doc, nil
}

func (c *Client) DownloadImage(ctx context.Context, imgToken string) (string, error) {
  resp, _, err := c.larkCilent.Drive.DownloadDriveMedia(ctx, &lark.DownloadDriveMediaReq{
    FileToken: imgToken,
  })
  if err != nil {
    return imgToken, err
  }
  fileext := filepath.Ext(resp.Filename)
  filename := fmt.Sprintf("%s/%s%s", "static", imgToken, fileext)
  err = os.MkdirAll(filepath.Dir(filename), 0o755)
  file, err := os.OpenFile(filename, os.O_CREATE | os.O_WRONLY, 0o666)
  if err != nil {
    return imgToken, err
  }
  defer file.Close()
  _, err = io.Copy(file, resp.File)
  if err != nil {
    return imgToken, err
  }
  return filename, nil
}
