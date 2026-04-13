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
	"github.com/chyroc/lark_rate_limiter"
)

// wrapErr enriches lark API errors with response metadata (request ID, status code)
// so that 403/other failures are debuggable without guessing.
func wrapErr(err error, resp *lark.Response, apiName string) error {
	if err == nil {
		return nil
	}
	if resp != nil {
		return fmt.Errorf("[%s] %w (request_id=%s, status_code=%d)", apiName, err, resp.RequestID, resp.StatusCode)
	}
	return fmt.Errorf("[%s] %w", apiName, err)
}

type Client struct {
	larkClient      *lark.Lark
	userAccessToken string // stores user access token
}

func NewClient(appID, appSecret string) *Client {
	return &Client{
		larkClient: lark.New(
			lark.WithAppCredential(appID, appSecret),
			lark.WithTimeout(60*time.Second),
			lark.WithApiMiddleware(lark_rate_limiter.Wait(4, 4)),
		),
	}
}

// NewClientWithUserToken creates a client that stores user access token for user-identity operations.
// Note: The lark SDK requires user token to be passed per-method via lark.WithUserAccessToken().
// This constructor stores the token in Client struct for use with method-level options.
func NewClientWithUserToken(appID, appSecret, userToken string) *Client {
	return &Client{
		larkClient: lark.New(
			lark.WithAppCredential(appID, appSecret),
			lark.WithTimeout(60*time.Second),
			lark.WithApiMiddleware(lark_rate_limiter.Wait(4, 4)),
		),
		userAccessToken: userToken,
	}
}

func (c *Client) userTokenOpts() []lark.MethodOptionFunc {
	if c.userAccessToken != "" {
		return []lark.MethodOptionFunc{lark.WithUserAccessToken(c.userAccessToken)}
	}
	return nil
}

func (c *Client) DownloadImage(ctx context.Context, imgToken, outDir string) (string, error) {
	opts := c.userTokenOpts()
	resp, response, err := c.larkClient.Drive.DownloadDriveMedia(ctx, &lark.DownloadDriveMediaReq{
		FileToken: imgToken,
	}, opts...)
	if err != nil {
		return imgToken, wrapErr(err, response, "DownloadImage")
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
	opts := c.userTokenOpts()
	resp, response, err := c.larkClient.Drive.DownloadDriveMedia(ctx, &lark.DownloadDriveMediaReq{
		FileToken: imgToken,
	}, opts...)
	if err != nil {
		return imgToken, nil, wrapErr(err, response, "DownloadImageRaw")
	}
	fileext := filepath.Ext(resp.Filename)
	filename := fmt.Sprintf("%s/%s%s", imgDir, imgToken, fileext)
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.File)
	return filename, buf.Bytes(), nil
}

func (c *Client) GetDocxContent(ctx context.Context, docToken string) (*lark.DocxDocument, []*lark.DocxBlock, error) {
	opts := c.userTokenOpts()
	resp, response, err := c.larkClient.Drive.GetDocxDocument(ctx, &lark.GetDocxDocumentReq{
		DocumentID: docToken,
	}, opts...)
	if err != nil {
		return nil, nil, wrapErr(err, response, "GetDocxDocument")
	}
	docx := &lark.DocxDocument{
		DocumentID: resp.Document.DocumentID,
		RevisionID: resp.Document.RevisionID,
		Title:      resp.Document.Title,
	}
	var blocks []*lark.DocxBlock
	var pageToken *string
	for {
		resp2, response2, err := c.larkClient.Drive.GetDocxBlockListOfDocument(ctx, &lark.GetDocxBlockListOfDocumentReq{
			DocumentID: docx.DocumentID,
			PageToken:  pageToken,
		}, opts...)
		if err != nil {
			return docx, nil, wrapErr(err, response2, "GetDocxBlockListOfDocument")
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
	opts := c.userTokenOpts()
	resp, response, err := c.larkClient.Drive.GetWikiNode(ctx, &lark.GetWikiNodeReq{
		Token: token,
	}, opts...)
	if err != nil {
		return nil, wrapErr(err, response, "GetWikiNode")
	}
	return resp.Node, nil
}

func (c *Client) GetDriveFolderFileList(ctx context.Context, pageToken *string, folderToken *string) ([]*lark.GetDriveFileListRespFile, error) {
	opts := c.userTokenOpts()
	resp, response, err := c.larkClient.Drive.GetDriveFileList(ctx, &lark.GetDriveFileListReq{
		PageSize:    nil,
		PageToken:   pageToken,
		FolderToken: folderToken,
	}, opts...)
	if err != nil {
		return nil, wrapErr(err, response, "GetDriveFileList")
	}
	files := resp.Files
	for resp.HasMore {
		resp, response, err = c.larkClient.Drive.GetDriveFileList(ctx, &lark.GetDriveFileListReq{
			PageSize:    nil,
			PageToken:   &resp.NextPageToken,
			FolderToken: folderToken,
		}, opts...)
		if err != nil {
			return nil, wrapErr(err, response, "GetDriveFileList")
		}
		files = append(files, resp.Files...)
	}
	return files, nil
}

func (c *Client) GetWikiName(ctx context.Context, spaceID string) (string, error) {
	opts := c.userTokenOpts()
	resp, response, err := c.larkClient.Drive.GetWikiSpace(ctx, &lark.GetWikiSpaceReq{
		SpaceID: spaceID,
	}, opts...)

	if err != nil {
		return "", wrapErr(err, response, "GetWikiSpace")
	}

	return resp.Space.Name, nil
}

func (c *Client) GetWikiNodeList(ctx context.Context, spaceID string, parentNodeToken *string) ([]*lark.GetWikiNodeListRespItem, error) {
	opts := c.userTokenOpts()
	resp, response, err := c.larkClient.Drive.GetWikiNodeList(ctx, &lark.GetWikiNodeListReq{
		SpaceID:         spaceID,
		PageSize:        nil,
		PageToken:       nil,
		ParentNodeToken: parentNodeToken,
	}, opts...)

	if err != nil {
		return nil, wrapErr(err, response, "GetWikiNodeList")
	}

	nodes := resp.Items
	previousPageToken := ""

	for resp.HasMore && previousPageToken != resp.PageToken {
		previousPageToken = resp.PageToken
		resp, response, err := c.larkClient.Drive.GetWikiNodeList(ctx, &lark.GetWikiNodeListReq{
			SpaceID:         spaceID,
			PageSize:        nil,
			PageToken:       &resp.PageToken,
			ParentNodeToken: parentNodeToken,
		}, opts...)

		if err != nil {
			return nil, wrapErr(err, response, "GetWikiNodeList")
		}

		nodes = append(nodes, resp.Items...)
	}

	return nodes, nil
}
