package main

import (
	"fmt"
	"github.com/Wsine/feishu2md/core"
	"github.com/Wsine/feishu2md/utils"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

type BatchDownloadOpts struct {
	baseDir      string
	outputDir    string
	larkSpaceURL string
}

const ApiLimitsPerSec = 5

var batchDownloadOpts = BatchDownloadOpts{}

// Parse the content of a .url file and return the URL
// The content of a .url file is like this:
// [InternetShortcut]
// URL=https://doesnotexists.larksuite.com/docx/doccnL4J5Z6QJ5Z6QJ5Z6QJ5Z6Q
// Object=doccnL4J5Z6QJ5Z6QJ5Z6QJ5Z6Q
// The URL is the link to the document
func parseURL(content string, larkSpaceURL string) (string, error) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "URL=") {
			URL := strings.TrimPrefix(line, "URL=")
			// Use regex to capture the prefix of the URL:
			// Pattern: https://{USELESS}.com/{URL}
			pattern := fmt.Sprintf("https://.*?\\.com/(.*)")
			URL = regexp.MustCompile(pattern).FindStringSubmatch(URL)[1]

			URL = fmt.Sprintf("https://%s/%s", larkSpaceURL, URL)
			return URL, nil
		}
	}
	return "", errors.New("URL not found in the content")
}

// Batch download all the documents in the pathMap to the output directory
// The pathMap is a map of {relativePath, url}
// This function downloads all the documents using the url to the relativePath in the output directory
func batchDownload(pathMap map[string]string, outputDir string) error {
	// Load config
	configPath, err := core.GetConfigFilePath()
	utils.CheckErr(err)
	config, err := core.ReadConfigFromFile(configPath)
	utils.CheckErr(err)

	var batchErr error = nil

	downloadFunc := func(relPath, url string) {
		// If the output subdirectory for relPath does not exist, create it
		outputPath := filepath.Join(outputDir, relPath)
		subDir := filepath.Dir(outputPath)
		if _, err := os.Stat(subDir); os.IsNotExist(err) {
			err := os.MkdirAll(subDir, 0755)
			if err != nil {
				fmt.Println("Error creating directory:", err)
				batchErr = err
				return
			}
		}

		// Download the document
		err := downloadDocument(url, outputPath, false, config)
		if err != nil {
			fmt.Println("Error downloading document:", err)
			batchErr = err
			return
		}

		fmt.Printf("Downloaded markdown file to %s\n", filepath.Join(outputDir, relPath))
	}

	// API limit is 5 requests per second,
	// so we use a pool of 5 goroutines added to the pool every second
	operators := make(chan struct{}, 0)
	downloadFinished := false
	// Set a timer to add 5 operators to the pool every 1.5 second (for safety)
	go func() {
		for {
			if downloadFinished {
				break
			}
			for i := 0; i < ApiLimitsPerSec; i++ {
				operators <- struct{}{}
			}
			<-time.After(1500 * time.Millisecond)
		}
	}()

	var wg sync.WaitGroup
	for relPath, url := range pathMap {
		wg.Add(1)
		go func(relPath, url string) {
			<-operators
			downloadFunc(relPath, url)
			wg.Done()
		}(relPath, url)
	}

	wg.Wait()
	return batchErr
}

// `baseDir` is the base directory for the all of the Feishu document direcotry
// you downloaded
//
// In Feishu, you can download a document as a directory, but the directory only
// contains a bunch of .url files, which are actually the links to the documents
// This function will fetch all files within the base directory with .url extension
// and download them all to the output directory with the same hierarchy structure
//
// For example, if you have a directory structure like this:
// baseDir
// ├── docFolder1
// │   ├── doc1.url
// │   └── doc2.url
// └── docFolder2
//
//	├── doc3.url
//	└── doc4.url
//
// `batchDownload` will download all the documents to the output directory with the
// same structure:
// outputDir
// ├── docFolder1
// │   ├── doc1.md
// │   └── doc2.md
// └── docFolder2
//
//	├── doc3.md
//	└── doc4.md
func handleBatchDownloadCommand(baseDir string, outputDir string, larkSpaceURL string, opts *BatchDownloadOpts) error {
	// Validate the base directory
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return errors.Errorf("Base directory does not exist: %s", baseDir)
	}

	pathMap := make(map[string]string)

	// Find all files under `baseDir` with .url extension
	err := filepath.Walk(baseDir,
		func(path string, info os.FileInfo, err error) error {
			// DEBUG: print path
			fmt.Println(path)
			if err != nil {
				return err
			}
			if filepath.Ext(path) == ".url" {
				// Read the content of the .url file
				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				// Parse the URL from the content
				url, err := parseURL(string(content), larkSpaceURL)
				if err != nil {
					return err
				}
				// Save pair {relativePath, url} to the map
				relPath, err := filepath.Rel(baseDir, path)
				relPath = strings.TrimSuffix(relPath, ".url") + ".md"
				if err != nil {
					return err
				}
				pathMap[relPath] = url
			}
			return nil
		},
	)

	if err != nil {
		return err
	}

	err = batchDownload(pathMap, outputDir)
	if err != nil {
		return err
	}

	return nil

}
