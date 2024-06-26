package core

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/chyroc/lark"
)

type Client struct {
	larkClient *lark.Lark
}

func NewClient(appID, appSecret string) *Client {
	return &Client{
		larkClient: lark.New(
			lark.WithAppCredential(appID, appSecret),
			lark.WithTimeout(60*time.Second),
		),
	}
}

func (c *Client) DownloadImage(ctx context.Context, imgToken, outDir string) (string, error) {
	resp, _, err := c.larkClient.Drive.DownloadDriveMedia(ctx, &lark.DownloadDriveMediaReq{
		FileToken: imgToken,
	})
	if err != nil {
		return imgToken, err
	}
	fileext := filepath.Ext(resp.Filename)
	filename := fmt.Sprintf("%s/%s%s", outDir, imgToken, fileext)
	err = os.MkdirAll(filepath.Dir(filename), 0o755)
	if err != nil {
		return imgToken, err
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0o666)
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

func (c *Client) DownloadImageRaw(ctx context.Context, imgToken, imgDir string) (string, []byte, error) {
	resp, _, err := c.larkClient.Drive.DownloadDriveMedia(ctx, &lark.DownloadDriveMediaReq{
		FileToken: imgToken,
	})
	if err != nil {
		return imgToken, nil, err
	}
	fileext := filepath.Ext(resp.Filename)
	filename := fmt.Sprintf("%s/%s%s", imgDir, imgToken, fileext)
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.File)
	return filename, buf.Bytes(), nil
}

func (c *Client) GetDocxContent(ctx context.Context, docToken string) (*lark.DocxDocument, []*lark.DocxBlock, error) {
	resp, _, err := c.larkClient.Drive.GetDocxDocument(ctx, &lark.GetDocxDocumentReq{
		DocumentID: docToken,
	})
	if err != nil {
		return nil, nil, err
	}
	docx := &lark.DocxDocument{
		DocumentID: resp.Document.DocumentID,
		RevisionID: resp.Document.RevisionID,
		Title:      resp.Document.Title,
	}
	var blocks []*lark.DocxBlock
	var pageToken *string
	for {
		resp2, _, err := c.larkClient.Drive.GetDocxBlockListOfDocument(ctx, &lark.GetDocxBlockListOfDocumentReq{
			DocumentID: docx.DocumentID,
			PageToken:  pageToken,
		})
		if err != nil {
			return docx, nil, err
		}
		blocks = append(blocks, resp2.Items...)
		pageToken = &resp2.PageToken
		if !resp2.HasMore {
			break
		}
	}
	return docx, blocks, nil
}

func (c *Client) GetWikiNodeInfo(ctx context.Context, token string) (*lark.GetWikiNodeRespNode, error) {
	resp, _, err := c.larkClient.Drive.GetWikiNode(ctx, &lark.GetWikiNodeReq{
		Token: token,
	})
	if err != nil {
		return nil, err
	}
	return resp.Node, nil
}

func (c *Client) GetDriveFolderFileList(ctx context.Context, pageToken *string, folderToken *string) (*lark.GetDriveFileListResp, error) {
	resp, _, err := c.larkClient.Drive.GetDriveFileList(ctx, &lark.GetDriveFileListReq{
		PageSize:    nil,
		PageToken:   pageToken,
		FolderToken: folderToken,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetDriveStructureRecursion(ctx context.Context, folderToken *string, currentPath string, pathToURL map[string]string) error {
	resp, err := c.GetDriveFolderFileList(ctx, nil, folderToken)
	if err != nil {
		return err
	}
	files := resp.Files
	for resp.HasMore {
		resp, err = c.GetDriveFolderFileList(ctx, &resp.NextPageToken, folderToken)
		if err != nil {
			return err
		}
		files = append(files, resp.Files...)
	}

	for _, file := range files {
		path := currentPath + "/" + file.Name
		if file.Type == "folder" {
			err = c.GetDriveStructureRecursion(ctx, &file.Token, path, pathToURL)
			if err != nil {
				return err
			}
		} else {
			pathToURL[path] = file.URL
		}
	}

	return nil
}

func (c *Client) GetDriveStructure(ctx context.Context, baseFolderToken *string) (map[string]string, error) {
	structure := map[string]string{}
	err := c.GetDriveStructureRecursion(ctx, baseFolderToken, ".", structure)
	return structure, err
}
