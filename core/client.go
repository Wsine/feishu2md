package core

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/chyroc/lark/larkext"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chyroc/lark"
	"github.com/chyroc/lark_rate_limiter"
)

type Client struct {
	larkClient *lark.Lark
	SyncFile   *os.File
	SyncMap    map[string]string
}

func NewClient(appID, appSecret string, syncCfg ...*os.File) *Client {
	larkNew := lark.New(
		lark.WithAppCredential(appID, appSecret),
		lark.WithTimeout(60*time.Second),
		lark.WithApiMiddleware(lark_rate_limiter.Wait(5, 5)),
	)
	if len(syncCfg) > 0 {
		return &Client{
			larkClient: larkNew,
			SyncFile:   syncCfg[0],
			SyncMap:    make(map[string]string),
		}
	} else {
		return &Client{
			larkClient: larkNew,
			SyncFile:   nil,
			SyncMap:    make(map[string]string),
		}
	}
}

func (c *Client) InitSyncedFiles() {
	syncScanner := bufio.NewScanner(c.SyncFile)
	for syncScanner.Scan() {
		line := syncScanner.Text()
		parts := strings.Split(line, ",")
		// 8 只是为了限定当sync所需用token不存在，即文件有多余空行的情况
		if len(parts[0]) < 8 {
			continue
		}
		c.SyncMap[parts[0]] = parts[1]
	}

	if err := syncScanner.Err(); err != nil {
		fmt.Println("Error syncScanner file:", err)
	}

}

func (c *Client) SyncFileCheckAndWrite(ctx context.Context, docToken string) bool {
	if c.SyncFile == nil {
		fmt.Println("Current mode isn't sync mode")
		os.Exit(-1)
	} else {
		resp, _, err := c.larkClient.Drive.GetDriveFileMeta(ctx, &lark.GetDriveFileMetaReq{
			RequestDocs: []*lark.GetDriveFileMetaReqRequestDocs{{
				DocToken: docToken, DocType: "docx",
			}},
		})

		if err != nil {
			fmt.Println("GetFileMeta err ", err)
		}

		syncTime, ok := c.SyncMap[docToken]
		if (ok && syncTime < resp.Metas[0].LatestModifyTime) || (!ok) {
			writeN, err := c.SyncFile.WriteString(docToken + "," + resp.Metas[0].LatestModifyTime + "," + resp.Metas[0].Title + "\n")
			if err != nil {
				fmt.Printf("sync filer writeN %d err %d", writeN, err)
			}
			return true
		}

		return false
	}
	return true
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

func (c *Client) GetFolderName(ctx context.Context, folder_token string) (string, error) {
	curFolder := larkext.NewFolder(c.larkClient, folder_token)
	folderMeta, err := curFolder.Meta(ctx)

	if folderMeta != nil {
		return folderMeta.Name, nil
	}
	return "", err
}

func (c *Client) GetDriveFolderFileList(ctx context.Context, pageToken *string, folderToken *string) ([]*lark.GetDriveFileListRespFile, error) {
	resp, _, err := c.larkClient.Drive.GetDriveFileList(ctx, &lark.GetDriveFileListReq{
		PageSize:    nil,
		PageToken:   pageToken,
		FolderToken: folderToken,
	})
	if err != nil {
		return nil, err
	}
	files := resp.Files
	for resp.HasMore {
		resp, _, err = c.larkClient.Drive.GetDriveFileList(ctx, &lark.GetDriveFileListReq{
			PageSize:    nil,
			PageToken:   &resp.NextPageToken,
			FolderToken: folderToken,
		})
		if err != nil {
			return nil, err
		}
		files = append(files, resp.Files...)
	}
	return files, nil
}

func (c *Client) GetWikiName(ctx context.Context, spaceID string) (string, error) {
	resp, _, err := c.larkClient.Drive.GetWikiSpace(ctx, &lark.GetWikiSpaceReq{
		SpaceID: spaceID,
	})

	if err != nil {
		return "", err
	}

	return resp.Space.Name, nil
}

func (c *Client) GetWikiNodeList(ctx context.Context, spaceID string, parentNodeToken *string) ([]*lark.GetWikiNodeListRespItem, error) {
	resp, _, err := c.larkClient.Drive.GetWikiNodeList(ctx, &lark.GetWikiNodeListReq{
		SpaceID:         spaceID,
		PageSize:        nil,
		PageToken:       nil,
		ParentNodeToken: parentNodeToken,
	})

	if err != nil {
		return nil, err
	}

	nodes := resp.Items

	for resp.HasMore {
		resp, _, err := c.larkClient.Drive.GetWikiNodeList(ctx, &lark.GetWikiNodeListReq{
			SpaceID:         spaceID,
			PageSize:        nil,
			PageToken:       nil,
			ParentNodeToken: parentNodeToken,
		})

		if err != nil {
			return nil, err
		}

		nodes = append(nodes, resp.Items...)
	}

	return nodes, nil
}
