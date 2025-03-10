package download

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"wget/utils"
)

// DownloadFile downloads a file from the provided URL, saves it to the specified output directory and file, and applies a rate limit if provided.
// DownloadFile downloads a file from the provided URL, saves it to the specified output directory and file, and applies a rate limit if provided.
func DownloadFile(fileURL, outputFile, outputDir, rateLimit string, background bool) error {
	startTime := time.Now()
	fmt.Printf("start at %s\n", startTime.Format("2006-01-02 15:04:05"))

	// Make an HTTP GET request to the file URL.
	resp, err := http.Get(fileURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if the server returned a successful HTTP status.
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status: %s", resp.Status)
	}
	fmt.Printf("sending request, awaiting response... status %s\n", resp.Status)

	// Get the content length of the file.
	contentLength := resp.ContentLength
	fmt.Printf("content size: %d [~%.2fMB]\n", contentLength, float64(contentLength)/(1024*1024))

	// If the output file name is not provided, use the base name of the URL as the file name.
	fileName := outputFile
	if fileName == "" {
		fileName = filepath.Base(fileURL)
	}

	// Set the full file path where the file will be saved.
	filePath := filepath.Join(outputDir, fileName)
	fmt.Printf("saving file to: %s\n", filePath)

	// Ensure the output directory exists (create if it doesn't).
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return err
	}

	// Create the output file in the specified location.
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Set up the writer. If rate limit is specified, apply rate limiting to the writer.
	var writer io.Writer = file
	if rateLimit != "" {
		limit, err := utils.ParseRateLimit(rateLimit)
		if err != nil {
			return err
		}
		writer = NewRateLimitedWriter(file, limit)
	}

	// Only use progress writer if not in background mode
	if !background {
		// Set up a writer that will track download progress.
		progressWriter := NewProgressWriter(writer, contentLength)
		_, err = io.Copy(progressWriter, resp.Body)
	} else {
		// In background mode, just copy the data without progress tracking
		_, err = io.Copy(writer, resp.Body)
	}

	if err != nil {
		return err
	}

	fmt.Printf("\nDownloaded [%s]\n", fileURL)
	fmt.Printf("finished at %s\n", time.Now().Format("2006-01-02 15:04:05"))
	return nil
}

// DownloadMultipleFiles initiates downloading multiple files concurrently using goroutines.
// A wait group is used to synchronize the completion of multiple downloads.
// Loop through all provided URLs and download them concurrently.
// Increment the wait group counter for each download.
// Start a new goroutine for each download.
// Ensure the counter is decremented when the download completes.
// DownloadMultipleFiles downloads multiple files in parallel from the provided URLs
func DownloadMultipleFiles(urls []string, outputDir, rateLimit string, background bool) {
    var wg sync.WaitGroup
    for _, u := range urls {
        wg.Add(1)
        go func(url string) {
            defer wg.Done()
            err := DownloadFile(url, "", outputDir, rateLimit, background)
            if err != nil {
                fmt.Printf("Error downloading %s: %v\n", url, err)
            }
        }(u)
    }
    // Wait for all downloads to complete.
    wg.Wait()
    fmt.Println("Download finished.")
}

// Helper function to read URLs from a file
// In download package
func ReadURLsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename) // Open the file containing URLs
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", filename, err)
	}
	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			err = closeErr // Return the error of closing the file
		}
	}()

	var validURLs []string
	var invalidURLs []string
	scanner := bufio.NewScanner(file) // Scanner to read the file line by line
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		urlText := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if urlText == "" {
			fmt.Printf("Line %d: Empty URL, skipping\n", lineNumber)
			continue
		}

		// Validate URL
		parsedURL, err := url.Parse(urlText)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			fmt.Printf("Line %d: Invalid URL format '%s', skipping\n", lineNumber, urlText)
			invalidURLs = append(invalidURLs, fmt.Sprintf("Line %d: %s", lineNumber, urlText))
			continue
		}

		// URL is valid format
		validURLs = append(validURLs, urlText)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning file: %v", err)
	}

	// Report invalid URLs if any were found
	if len(invalidURLs) > 0 {
		fmt.Println("\nInvalid URLs found in file:")
		for _, invalidURL := range invalidURLs {
			fmt.Println("- " + invalidURL)
		}
		fmt.Printf("\nFound %d valid URLs and %d invalid URLs\n",
			len(validURLs), len(invalidURLs))
	}

	if len(validURLs) == 0 {
		return nil, fmt.Errorf("no valid URLs found in file %s", filename)
	}

	return validURLs, nil
}
