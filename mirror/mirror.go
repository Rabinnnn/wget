package mirror

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

// A structure holding the parameters used during the mirroring process
type MirrorParams struct {
	URL           string
	OutputDir     string
	ConvertLinks  bool
	UseDynamic    bool
	RejectTypes   []string
	ExcludePaths  []string
	visited       sync.Map // Concurrent-safe map
	currentDepth  int
	maxDepth      int
	depthMutex    sync.Mutex // Protects currentDepth
	baseHost      string
	MaxConcurrent int
}

// GetMirrorParams parses the parameters passed for mirroring.
// It then populates the MirrorParams struct using the values.
func GetMirrorParams(urlStr, outputDir string, convertLinks bool, rejectTypes []string, excludePaths []string) *MirrorParams {
	baseURL, err := url.Parse(urlStr)
	if err != nil {
		fmt.Printf("Warning: Failed to parse URL: %v\n", err)
		return nil
	}

	return &MirrorParams{
		URL:           urlStr,
		OutputDir:     outputDir,
		ConvertLinks:  convertLinks,
		RejectTypes:   rejectTypes,
		ExcludePaths:  excludePaths,
		maxDepth:      5, // Maximum depth for nested links
		baseHost:      baseURL.Host,
		MaxConcurrent: 100000,
	}
}

// ProcessUrl handles the URL passed for mirroring.
// It downloads the resources based on the specified parameters such as output name, directory, reject, and exclude.
// It handles the nested links recurssively.
func (m *MirrorParams) ProcessUrl(urlStr string, wg *sync.WaitGroup, sem chan struct{}) {
	defer wg.Done()          // mark when all goroutines have finished execution
	sem <- struct{}{}        // Acquire semaphore
	defer func() { <-sem }() // Ensure semaphore is released when the function completes.

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		fmt.Printf("failed to parse URL %s: %v\n", urlStr, err)
		return
	}

	cleanURL := *parsedURL
	cleanURL.Fragment = ""
	cleanURL.RawQuery = ""
	urlKey := cleanURL.String()

	// Use sync.Map for thread safety
	if _, exists := m.visited.Load(urlKey); exists {
		return
	}
	m.visited.Store(urlKey, true)

	// Protect `currentDepth` with a mutex
	m.depthMutex.Lock()
	if m.currentDepth > m.maxDepth {
		m.depthMutex.Unlock()
		return
	}
	m.currentDepth++
	m.depthMutex.Unlock()

	defer func() {
		m.depthMutex.Lock()
		m.currentDepth--
		m.depthMutex.Unlock()
	}()

	if parsedURL.Host != "" && parsedURL.Host != m.baseHost {
		fmt.Printf("Skipping external domain: %s\n", urlStr)
		return
	}

	if strings.Contains(parsedURL.Path, "/js/") {
		return
	}

	for _, excludePath := range m.ExcludePaths {
		normalizedExclude := strings.Trim(excludePath, "/")
		normalizedPath := strings.Trim(parsedURL.Path, "/")

		if strings.HasPrefix(normalizedPath, normalizedExclude) {
			fmt.Printf("Skipping excluded path: %s\n", urlStr)
			return
		}
	}

	filename := filepath.Base(parsedURL.Path)
	if filename == "" || filename == "/" {
		filename = "index.html"
	}

	shouldSaveFile := true

	for _, rejectedType := range m.RejectTypes {
		if strings.EqualFold(filename, rejectedType) {
			fmt.Printf("Skipping rejected file: %s\n", urlStr)
			shouldSaveFile = false
		}
	}

	ext := strings.ToLower(filepath.Ext(parsedURL.Path))
	if ext != "" {
		ext = strings.TrimPrefix(ext, ".")
		for _, rejectedType := range m.RejectTypes {
			if strings.EqualFold(ext, rejectedType) {
				fmt.Printf("Skipping rejected file type: %s\n", urlStr)
				shouldSaveFile = false
			}
		}
	}

	if shouldSaveFile {
		fmt.Printf("Downloading: %s\n", urlStr)
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		fmt.Printf("failed to create request: %v\n", err)
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("failed to download %s: %v\n", urlStr, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("failed to download %s: status code %d\n", urlStr, resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("failed to read response body: %v\n", err)
		return
	}

	outputPath := filepath.Join(m.OutputDir, m.convertToLocalPath(parsedURL))

	if strings.HasSuffix(outputPath, "/") || outputPath == m.OutputDir {
		outputPath = filepath.Join(outputPath, "index.html")
	}

	if info, err := os.Stat(outputPath); err == nil && info.IsDir() {
		outputPath = filepath.Join(outputPath, "index.html")
	}

	if shouldSaveFile {
		dir := filepath.Dir(outputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("failed to create directory %s: %v\n", dir, err)
			return
		}

		if err := os.WriteFile(outputPath, body, 0644); err != nil {
			fmt.Printf("failed to write file: %v\n", err)
			return
		}
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") {
		doc, err := html.Parse(bytes.NewReader(body))
		if err != nil {
			fmt.Printf("failed to parse HTML: %v\n", err)
			return
		}

		var processNode func(*html.Node)
		processNode = func(n *html.Node) {
			if n.Type == html.ElementNode {
				for i := 0; i < len(n.Attr); i++ {
					attr := n.Attr[i]
					switch attr.Key {
					case "href", "src":
						absURL, err := m.getAbsoluteURL(parsedURL, attr.Val)
						if err != nil {
							fmt.Printf("Warning: Failed to resolve URL %s: %v\n", attr.Val, err)
							continue
						}
						if strings.Contains(absURL.String(), "google-analytics.com") || strings.Contains(absURL.String(), "analytics.js") {
							continue
						}

						if absURL.Host == m.baseHost {
							if m.ConvertLinks {
								localPath := m.getRelativePath(parsedURL, absURL)
								n.Attr[i].Val = localPath
							} else {
								n.Attr[i].Val = absURL.String()
							}

							cleanAbsURL := *absURL
							cleanAbsURL.Fragment = ""
							cleanAbsURL.RawQuery = ""

							if _, exists := m.visited.Load(cleanAbsURL.String()); exists {
								continue
							}

							wg.Add(1)
							go m.ProcessUrl(absURL.String(), wg, sem)
						}
					case "style":
						urls := extractURLsFromCSS(attr.Val)
						for _, cssURL := range urls {
							absURL, err := m.getAbsoluteURL(parsedURL, cssURL)
							if err != nil {
								fmt.Printf("Warning: Failed to resolve URL %s: %v\n", cssURL, err)
								continue
							}
							if absURL.Host == m.baseHost {
								localPath := m.getRelativePath(parsedURL, absURL)
								if m.ConvertLinks {
									attr.Val = strings.ReplaceAll(attr.Val, fmt.Sprintf(`url('%s')`, cssURL), fmt.Sprintf(`url('%s')`, localPath))
									attr.Val = strings.ReplaceAll(attr.Val, fmt.Sprintf(`url("%s")`, cssURL), fmt.Sprintf(`url("%s")`, localPath))
									attr.Val = strings.ReplaceAll(attr.Val, fmt.Sprintf(`url(%s)`, cssURL), fmt.Sprintf(`url(%s')`, localPath))
									n.Attr[i] = attr
								}
								cleanAbsURL := *absURL
								cleanAbsURL.Fragment = ""
								cleanAbsURL.RawQuery = ""
								if _, exists := m.visited.Load(cleanAbsURL.String()); exists {
									continue
								}

								wg.Add(1)
								go m.ProcessUrl(absURL.String(), wg, sem)

							}
						}
					case "integrity":
						if i < len(n.Attr)-1 {
							n.Attr = append(n.Attr[:i], n.Attr[i+1:]...)
							i--
						} else {
							n.Attr = n.Attr[:i]
						}
					}
				}

				if n.Data == "style" && n.FirstChild != nil {
					cssContent := n.FirstChild.Data
					urls := extractURLsFromCSS(cssContent)
					for _, cssURL := range urls {
						absURL, err := m.getAbsoluteURL(parsedURL, cssURL)
						if err != nil {
							fmt.Printf("Warning: Failed to resolve URL %s: %v\n", cssURL, err)
							continue
						}

						if absURL.Host == m.baseHost {
							localPath := m.getRelativePath(parsedURL, absURL)
							if m.ConvertLinks {
								cssContent = strings.ReplaceAll(cssContent, fmt.Sprintf(`url('%s')`, cssURL), fmt.Sprintf(`url('%s')`, localPath))
								cssContent = strings.ReplaceAll(cssContent, fmt.Sprintf(`url("%s")`, cssURL), fmt.Sprintf(`url("%s")`, localPath))
								cssContent = strings.ReplaceAll(cssContent, fmt.Sprintf(`url(%s)`, cssURL), fmt.Sprintf(`url(%s')`, localPath))
								n.FirstChild.Data = cssContent
							}

							cleanAbsURL := *absURL
							cleanAbsURL.Fragment = ""
							cleanAbsURL.RawQuery = ""

							if _, exists := m.visited.Load(cleanAbsURL.String()); exists {
								continue
							}

							wg.Add(1)
							go m.ProcessUrl(absURL.String(), wg, sem)
						}
					}
				}
			}

			for c := n.FirstChild; c != nil; c = c.NextSibling {
				processNode(c)
			}
		}
		processNode(doc)

		if shouldSaveFile {
			var buf bytes.Buffer
			if err := html.Render(&buf, doc); err != nil {
				fmt.Printf("failed to render HTML: %v\n", err)
				<-sem
				return
			}

			if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
				fmt.Printf("failed to write updated HTML: %v\n", err)
				<-sem
				return
			}
		}
	} else if strings.Contains(contentType, "text/css") {
		cssContent := string(body)
		urls := extractURLsFromCSS(cssContent)
		for _, cssURL := range urls {
			absURL, err := m.getAbsoluteURL(parsedURL, cssURL)
			if err != nil {
				fmt.Printf("Warning: Failed to resolve URL %s: %v\n", cssURL, err)
				continue
			}

			if absURL.Host == m.baseHost {
				localPath := m.getRelativePath(parsedURL, absURL)
				if m.ConvertLinks {
					cssContent = strings.ReplaceAll(cssContent, fmt.Sprintf(`url('%s')`, cssURL), fmt.Sprintf(`url('%s')`, localPath))
					cssContent = strings.ReplaceAll(cssContent, fmt.Sprintf(`url("%s")`, cssURL), fmt.Sprintf(`url("%s")`, localPath))
					cssContent = strings.ReplaceAll(cssContent, fmt.Sprintf(`url(%s)`, cssURL), fmt.Sprintf(`url(%s')`, localPath))
				}

				cleanAbsURL := *absURL
				cleanAbsURL.Fragment = ""
				cleanAbsURL.RawQuery = ""

				if _, exists := m.visited.Load(cleanAbsURL.String()); exists {
					continue
				}

				wg.Add(1)
				go m.ProcessUrl(absURL.String(), wg, sem)
			}
		}

		if shouldSaveFile {
			if err := os.WriteFile(outputPath, []byte(cssContent), 0644); err != nil {
				fmt.Printf("failed to write updated CSS: %v\n", err)
				return
			}
		}
	}
}

func (m *MirrorParams) ProcessUrlWrapper(urlStr string) error {
	var wg sync.WaitGroup
	sem := make(chan struct{}, m.MaxConcurrent) // Limit concurrency

	wg.Add(1)
	go m.ProcessUrl(urlStr, &wg, sem)

	wg.Wait()
	return nil
}

// convertToLocalPath transforms a URL to local file path
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

func (m *MirrorParams) Mirror() error {
	// Create output directory
	if err := os.MkdirAll(m.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	fmt.Printf("Starting mirror of %s\n", m.URL)
	fmt.Printf("Output directory: %s\n", m.OutputDir)

	return m.ProcessUrlWrapper(m.URL)
}

// getAbsoluteURL transforms relative URL to Absolute URL
func (m *MirrorParams) getAbsoluteURL(base *url.URL, ref string) (*url.URL, error) {
	refURL, err := url.Parse(ref)
	if err != nil {
		return nil, err
	}
	return base.ResolveReference(refURL), nil
}

// getRelativePath converts a URL to a relative link path for use in HTML
func (m *MirrorParams) getRelativePath(base, ref *url.URL) string {
	// If the reference URL is absolute (starts with a protocol), keep it as is
	if ref.Scheme != "" || ref.Host != "" {
		if ref.Host != base.Host {
			// External link, keep it as is
			return ref.String()
		}
	}

	// Get the local path for the reference URL
	localPath := m.convertToLocalPath(ref)

	// Get the local path for the base URL
	basePath := m.convertToLocalPath(base)
	baseDir := filepath.Dir(basePath)

	// Calculate the relative path from base to ref
	rel, err := filepath.Rel(baseDir, localPath)
	if err != nil {
		// Fallback to absolute path if relative path calculation fails
		return "/" + localPath
	}

	// Convert Windows backslashes to forward slashes for URLs
	return strings.ReplaceAll(rel, "\\", "/")
}

func extractURLsFromCSS(css string) []string {
	var urls []string
	// Match url(...) patterns in CSS
	urlPattern := regexp.MustCompile(`url\(['"]?([^'"\)]+)['"]?\)`)
	matches := urlPattern.FindAllStringSubmatch(css, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			urls = append(urls, match[1])
		}
	}
	return urls
}
