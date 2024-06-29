package main

import (
	"context"
	"fmt"
	"github.com/Wsine/feishu2md/core"
	"github.com/Wsine/feishu2md/utils"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var downloadFailureList = []string{}

const ApiLimitsPerSec = 5

func singleDownload(relPath string, url string, outputDir string, config *core.Config) {
	// If the output subdirectory for relPath does not exist, create it
	outputPath := filepath.Join(outputDir, relPath)
	subDir := filepath.Dir(outputPath)
	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		err := os.MkdirAll(subDir, 0755)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}
	}

	// Download the document
	err := downloadDocument(url, outputPath, false, config)
	if err != nil {
		fmt.Println("Error downloading document:", err)
		downloadFailureList = append(downloadFailureList, url)
	}
}

func logDownloadFailures() {
	// Log the URLs that failed to download in stderr
	if len(downloadFailureList) > 0 {
		fmt.Fprintln(os.Stderr, "The following URLs failed to download:")
		for _, url := range downloadFailureList {
			fmt.Fprintln(os.Stderr, url)
		}
		// Print the following message in Green background color
		_, _ = fmt.Fprintln(os.Stderr, "\033[42m\033[30mDon't worry, this is not a total failure.\033[0m")
		_, _ = fmt.Fprintln(os.Stderr, "\033[42m\033[30mSome of your documents may have been downloaded successfully.\033[0m")
		_, _ = fmt.Fprintln(os.Stderr, "\033[42m\033[30mYou can try to download the failed documents again.\033[0m")
	}
}

// Batch download all the documents in the pathMap to the output directory
// The pathMap is a map of {relativePath, url}
// This function downloads all the documents using the url to the relativePath in the output directory
func batchDownload(pathMap map[string]string, outputDir string, config *core.Config) error {
	utils.StopWhenErr = false

	var batchErr error = nil

	// API limit is 5 requests per second,
	// so we use a pool of 5 goroutines added to the pool every second
	readyOperators := make(chan struct{}, ApiLimitsPerSec)
	finishedOperators := make(chan struct{}, ApiLimitsPerSec)
	for i := 0; i < ApiLimitsPerSec; i++ {
		finishedOperators <- struct{}{}
	}
	downloadFinished := false
	// Set a timer to add 5 operators to the pool every 1.5 second (for safety)
	go func() {
		for {
			if downloadFinished {
				break
			}
			<-finishedOperators
			readyOperators <- struct{}{}
			<-time.After(1500 * time.Millisecond / ApiLimitsPerSec)
		}
	}()

	var wg sync.WaitGroup
	for relPath, url := range pathMap {
		wg.Add(1)
		go func(relPath, url string) {
			<-readyOperators
			singleDownload(relPath, url, outputDir, config)
			finishedOperators <- struct{}{}
			wg.Done()
		}(relPath, url)
	}

	wg.Wait()
	logDownloadFailures()

	if len(downloadFailureList) > 0 {
		batchErr = errors.New("Some documents failed to download")
	}
	return batchErr
}

func batchDownloadFolder(outputDir string, urlOrToken string, config *core.Config) error {
	baseFolderToken := urlOrToken
	// If this is batch download and a full directory url is provided, strip the last part (token)
	if strings.Contains(urlOrToken, "/") {
		segments := strings.Split(urlOrToken, "/")
		baseFolderToken = segments[len(segments)-1]
		baseFolderToken = strings.Split(baseFolderToken, "?")[0] // In case some user copy the url with query params
	}

	// Create client with context
	ctx := context.WithValue(context.Background(), "output", config.Output)

	client := core.NewClient(
		config.Feishu.AppId, config.Feishu.AppSecret,
	)
	pathMap, err := client.GetDriveStructure(ctx, baseFolderToken)

	if err != nil {
		return err
	}

	err = batchDownload(pathMap, outputDir, config)
	if err != nil {
		return err
	}

	return nil

}
