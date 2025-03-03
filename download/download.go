package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
    "strings"
    "bufio"
    "net/url"
)

// Global output manager (singleton)
var outputManager *OutputManager
var outputManagerInit sync.Once

// GetOutputManager returns the singleton output manager instance
func GetOutputManager() *OutputManager {
	outputManagerInit.Do(func() {
		outputManager = NewOutputManager()
	})
	return outputManager
}

// ModifiedProgressWriter uses the output manager for tracking progress
type ModifiedProgressWriter struct {
	writer     io.Writer
	url        string
	total      int64
	downloaded int64
	manager    *OutputManager
}

// In download package
func ReadURLsFromFile(filename string) ([]string, error) {
	manager := GetOutputManager()
	
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
			manager.Log("Line %d: Empty URL, skipping", lineNumber)
			continue
		}

		// Validate URL
		parsedURL, err := url.Parse(urlText)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			manager.Log("Line %d: Invalid URL format '%s', skipping", lineNumber, urlText)
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
		manager.Log("Invalid URLs found in file:")
		for _, invalidURL := range invalidURLs {
			manager.Log("- %s", invalidURL)
		}
		manager.Log("Found %d valid URLs and %d invalid URLs",
			len(validURLs), len(invalidURLs))
	}

	if len(validURLs) == 0 {
		return nil, fmt.Errorf("no valid URLs found in file %s", filename)
	}

	return validURLs, nil
}

// NewModifiedProgressWriter creates a new progress writer that integrates with the output manager
func NewModifiedProgressWriter(writer io.Writer, url string, total int64) *ModifiedProgressWriter {
	manager := GetOutputManager()
	manager.RegisterDownload(url, total)
	
	return &ModifiedProgressWriter{
		writer:     writer,
		url:        url,
		total:      total,
		downloaded: 0,
		manager:    manager,
	}
}

// Write writes data and updates progress through the output manager
func (p *ModifiedProgressWriter) Write(data []byte) (int, error) {
	n, err := p.writer.Write(data)
	if err != nil {
		return n, err
	}
	
	p.downloaded += int64(n)
	p.manager.UpdateProgress(p.url, p.downloaded)
	
	return n, nil
}

// DownloadFile with output manager integration
func DownloadFile(fileURL, outputFile, outputDir, rateLimit string, background bool) error {
	manager := GetOutputManager()
	manager.Log("Starting download of %s", fileURL)
	
	// Make an HTTP GET request to the file URL
	resp, err := http.Get(fileURL)
	if err != nil {
		manager.Log("Error requesting %s: %v", fileURL, err)
		return err
	}
	defer resp.Body.Close()
	
	// Check if the server returned a successful HTTP status
	if resp.StatusCode != http.StatusOK {
		manager.Log("Status error for %s: %s", fileURL, resp.Status)
		return fmt.Errorf("status: %s", resp.Status)
	}
	manager.Log("Response received for %s: %s", fileURL, resp.Status)
	
	// Get the content length of the file
	contentLength := resp.ContentLength
	manager.Log("Content size for %s: %d bytes (%.2f MB)", 
		fileURL, contentLength, float64(contentLength)/(1024*1024))
	
	// If the output file name is not provided, use the base name of the URL as the file name
	fileName := outputFile
	if fileName == "" {
		fileName = filepath.Base(fileURL)
	}
	
	// Set the full file path where the file will be saved
	filePath := filepath.Join(outputDir, fileName)
	manager.Log("Saving %s to: %s", fileURL, filePath)
	
	// Ensure the output directory exists
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		manager.Log("Error creating directory for %s: %v", fileURL, err)
		return err
	}
	
	// Create the output file
	file, err := os.Create(filePath)
	if err != nil {
		manager.Log("Error creating file for %s: %v", fileURL, err)
		return err
	}
	defer file.Close()
	
	// Set up the writer with rate limiting if needed
	var writer io.Writer = file
	if rateLimit != "" {
		limit, err := ParseRateLimit(rateLimit)
		if err != nil {
			manager.Log("Error parsing rate limit for %s: %v", fileURL, err)
			return err
		}
		writer = NewRateLimitedWriter(file, limit)
	}
	
	// Use our modified progress writer
	progressWriter := NewModifiedProgressWriter(writer, fileURL, contentLength)
	_, err = io.Copy(progressWriter, resp.Body)
	
	if err != nil {
		manager.Log("Error downloading %s: %v", fileURL, err)
		return err
	}
	
	manager.Log("Completed download of %s", fileURL)
	manager.CompleteDownload(fileURL)
	
	return nil
}

// Modified DownloadMultipleFiles with output manager integration
func DownloadMultipleFiles(urls []string, outputDir, rateLimit string, background bool) {
	manager := GetOutputManager()
	manager.Log("Starting download of %d files", len(urls))
	
	var wg sync.WaitGroup
	for _, u := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			err := DownloadFile(url, "", outputDir, rateLimit, background)
			if err != nil {
				manager.Log("Error downloading %s: %v", url, err)
			}
		}(u)
	}
	
	wg.Wait()
	manager.Log("All downloads completed")
}

// ParseRateLimit is a placeholder for the actual function in utils
func ParseRateLimit(rateLimit string) (int64, error) {
	// You should import and use the actual function from utils
	// This is just a placeholder for the example
	return 1024 * 1024, nil // 1 MB/s default
}