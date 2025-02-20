package mirror

import (
	"fmt"
	"net/url"
	"os"
	"wget/download"
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