package mirror

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// A structure holding the parameters used during the mirroring process
type MirrorParams struct{
    URL          string
	OutputDir    string
	ConvertLinks bool
	UseDynamic   bool
	RejectTypes  []string
	ExcludePaths []string
	visited      map[string]bool
	currentDepth int
	maxDepth     int
	baseHost     string
}


func GetMirrorParams(urlStr, outputDir string, rejectTypes []string, excludePaths []string, convertLinks bool) *MirrorParams{
    baseURL, err := url.Parse(urlStr)
    if err != nil {
		fmt.Printf("Warning: Failed to parse URL: %v\n", err)
		return nil
	}

    return &MirrorParams{
		URL:          urlStr,
		OutputDir:    outputDir,
		ConvertLinks: convertLinks,
		RejectTypes:  rejectTypes,
		ExcludePaths: excludePaths,
		visited:      make(map[string]bool),
		maxDepth:     5, // Maximum depth for nested links
		baseHost:     baseURL.Host,
	}
}



// ProcessUrl downloads and processes a single URL
func (m *MirrorParams) ProcessUrl(urlStr string) error {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("failed to parse URL %s: %v", urlStr, err)
	}

	// Remove fragments and query parameters for visited check
	cleanURL := *parsedURL
	cleanURL.Fragment = ""
	cleanURL.RawQuery = ""
	urlKey := cleanURL.String()

	if m.visited[urlKey] {
		return nil
	}
	m.visited[urlKey] = true

	if m.currentDepth > m.maxDepth {
		return nil
	}

	// Only process URLs from the same domain
	if parsedURL.Host != "" && parsedURL.Host != m.baseHost {
		fmt.Printf("Skipping external domain: %s\n", urlStr)
		return nil
	}


	filename := filepath.Base(parsedURL.Path)
	if filename == "" || filename == "/" {
		filename = "index.html"
	}

	// Flag to track if this file should be saved
	shouldSaveFile := true


	fmt.Printf("Downloading: %s\n", urlStr)

	// Download the URL
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download %s: %v", urlStr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s: status code %d", urlStr, resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// Prepare output path for all cases
	outputPath := filepath.Join(m.OutputDir, m.convertToLocalPath(parsedURL))

	// If path ends with a slash, append index.html
	if strings.HasSuffix(outputPath, "/") || outputPath == m.OutputDir {
		outputPath = filepath.Join(outputPath, "index.html")
	}

	// Ensure the file doesn't exist as a directory
	if info, err := os.Stat(outputPath); err == nil && info.IsDir() {
		outputPath = filepath.Join(outputPath, "index.html")
	}

	// Save file if not rejected
	if shouldSaveFile {
		// Create directory if it doesn't exist
		dir := filepath.Dir(outputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}

		if err := os.WriteFile(outputPath, body, 0644); err != nil {
			return fmt.Errorf("failed to write file: %v", err)
		}
	}


	return nil
}



// convertToLocalPath transforms an absolute path into a relative path
func (m *MirrorParams) convertToLocalPath(u *url.URL) string {
	// Get the path without query parameters and fragments
	cleanPath := u.Path

	// Split the path into components
	parts := strings.Split(strings.TrimPrefix(cleanPath, "/"), "/")

	// Handle dynamic paths
	if len(parts) > 0 {
		// Convert paths with extensions to files
		if hasFileExtension(parts[len(parts)-1]) {
			return filepath.Join(u.Host, cleanPath)
		}

		// Handle paths that look like API endpoints or dynamic routes
		// Examples: /api/v1/users, /users/123, /repo/branch/path
		if hasNumericID(cleanPath) || hasDynamicParts(parts) {
			// Store under a 'pages' directory to separate dynamic content
			return filepath.Join(u.Host, "pages", cleanPath, "index.html")
		}
	}

	// Default handling for other paths
	path := filepath.Join(u.Host, strings.TrimPrefix(cleanPath, "/"))

	// If path is empty or ends with a slash, append index.html
	if path == u.Host || strings.HasSuffix(path, "/") || !hasFileExtension(path) {
		path = filepath.Join(path, "index.html")
	}

	return path
}


// hasFileExtension checks if a path has a file extension
func hasFileExtension(path string) bool {
	ext := filepath.Ext(path)
	return ext != "" && !strings.Contains(ext, "/")
}

// containsNumericID checks if a path contains numeric IDs
// Examples: /users/123, /posts/456/comments
func hasNumericID(path string) bool {
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if isNumeric(part) {
			return true
		}
	}
	return false
}

// isNumeric checks if a string is numeric
func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// containsDynamicSegments checks if path parts look like dynamic segments
// Examples: /blob/master/file.txt, /tree/main/src, /api/v1/users
func hasDynamicParts(parts []string) bool {
	// Common dynamic path patterns
	dynamicParts := []string{
		"api",
		"v1", "v2", "v3", // API versions
		"blob", "tree", // Repository patterns
		"branch", "tag",
		"commit", "pull",
		"issues", "wiki",
		"raw", "edit",
	}

	for _, part := range parts {
		for _, dynamicPart := range dynamicParts {
			if strings.EqualFold(part, dynamicPart) {
				return true
			}
		}
	}

	return false
}