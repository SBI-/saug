package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Thread struct {
	Tid        string
	URL        string
	Pages      int
	FolderName string
}

type Overview struct {
	NumberOfPages struct {
		Value int `xml:"value,attr"`
	} `xml:"number-of-pages"`
	Title string `xml:"title"`
}

const (
	baseURL        = "https://forum.mods.de/bb/xml/thread.php?TID=%s"
	pageURLPattern = baseURL + "&page=%d"
	imgPattern     = `\[img\]([^[]*)\[/img\]`
	timeout        = 20 * time.Second
)

func validateThreads(context context.Context, threadIDs []string) ([]Thread, error) {
	var validThreads []Thread
	for _, tid := range threadIDs {
		url := fmt.Sprintf(baseURL, tid)
		req, err := http.NewRequestWithContext(context, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating request: %w", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error fetching thread: %w", err)
		}

		defer resp.Body.Close()

		var overview Overview
		if err := xml.NewDecoder(resp.Body).Decode(&overview); err != nil {
			return nil, fmt.Errorf("error parsing XML: %w", err)
		}
		if overview.NumberOfPages.Value > 0 {
			validThreads = append(validThreads, Thread{
				FolderName: fmt.Sprintf("%s (%s)", overview.Title, tid),
				Tid:        tid,
				URL:        url,
				Pages:      overview.NumberOfPages.Value,
			})
		} else {
			fmt.Printf(fmt.Sprintf("Invalid thread id: %s. Skipping.\n", tid))
		}

	}
	return validThreads, nil
}

func makeFolders(ids []string) error {
	for _, name := range ids {
		if _, err := os.Stat(name); os.IsNotExist(err) {
			if err := os.MkdirAll(name, os.ModePerm); err != nil {
				return fmt.Errorf("error creating directory %s: %w", name, err)
			}
		}
	}
	return nil
}

func getPages(context context.Context, thread Thread) ([]string, error) {
	var pages []string
	for page := 1; page <= thread.Pages; page++ {
		url := fmt.Sprintf(pageURLPattern, thread.Tid, page)
		req, err := http.NewRequestWithContext(context, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating request: %w", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error fetching page: %w", err)
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading page content: %w", err)
		}
		pages = append(pages, string(body))
	}
	return pages, nil
}

func extractURLs(pages []string) []string {
	var urls []string
	re := regexp.MustCompile(imgPattern)
	for _, page := range pages {
		matches := re.FindAllStringSubmatch(page, -1)
		for _, match := range matches {
			if len(match) > 1 {
				url := strings.Replace(match[1], "/thumb/", "/img/", -1)
				url = strings.Replace(url, "\n", "", -1)
				url = strings.Replace(url, "\r", "", -1)
				urls = append(urls, url)
			}
		}
	}
	return urls
}

func filterURLs(urls []string) []string {
	var filtered []string
	for _, url := range urls {
		if strings.Contains(url, "abload.de/") {
			filtered = append(filtered, url)
		}
	}
	return filtered
}

func downloadURLs(context context.Context, thread Thread, urls []string, wait *sync.WaitGroup) {
	defer wait.Done()
	for _, url := range urls {
		fileName := strings.Split(filepath.Base(url), "?")[0]
		location := filepath.Join(thread.FolderName, fileName)

		req, err := http.NewRequestWithContext(context, "GET", url, nil)
		if err != nil {
			log.Printf("error creating request: %v", err)
			continue
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("error downloading URL %s: %s, url, err\n", url, err)
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			file, err := os.Create(location)
			if err != nil {
				log.Printf("error creating file %s: %v", location, err)
			}

			defer file.Close()

			_, err = io.Copy(file, resp.Body)
			if err != nil {
				log.Printf("error saving file %s: %v", location, err)
			}
		}
	}
}

func run(threadIDs []string) error {
	context, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	threads, err := validateThreads(context, threadIDs)
	if err != nil {
		return fmt.Errorf("error validating threads: %w", err)
	}

	var folderNames []string
	for _, thread := range threads {
		folderNames = append(folderNames, thread.FolderName)
	}

	if err := makeFolders(folderNames); err != nil {
		return fmt.Errorf("error creating folders: %w", err)
	}

	var wait sync.WaitGroup
	for _, thread := range threads {
		pages, err := getPages(context, thread)
		if err != nil {
			return fmt.Errorf("error getting pages for thread %s: %w", thread.Tid, err)
		}

		urls := extractURLs(pages)
		abloadURLs := filterURLs(urls)

		wait.Add(1)
		go downloadURLs(context, thread, abloadURLs, &wait)
	}
	wait.Wait()

	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide thread IDs as arguments.")
		return
	}
	if err := run(os.Args[1:]); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
