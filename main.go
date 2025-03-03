package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"wget/config"
	"wget/download"
	"wget/mirror"
)

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("error getting home directory: %v", err)
		}
		return filepath.Join(homeDir, path[2:]), nil
	}
	return path, nil
}

func main() {
    // Initialize flags and parse command-line arguments
    flags := config.InitFlags()
   // flag.Parse()
    
    // If background download flag is set, redirect output to a log file
    if flags.Background {
        logFile, err := os.Create("wget-log") // Create a log file
        if err != nil {
            fmt.Println("Error creating log file:", err)
            return
        }
        defer func() {
            closeErr := logFile.Close()
            if closeErr != nil {
                fmt.Println("Error closing log file:", closeErr)
            }
        }()
        os.Stdout = logFile // Redirect stdout to log file
        os.Stderr = logFile // Redirect stderr to log file
        fmt.Println("Output will be written to 'wget-log'.")
    }
    
    
        // If input file is provided, read URLs and initiate downloading multiple files
        if flags.InputFile != "" {
            urls, err := download.ReadURLsFromFile(flags.InputFile) // Correct call
            if err != nil {
                fmt.Println("Error reading URLs from file:", err)
                os.Exit(1)
            }
            download.DownloadMultipleFiles(urls, flags.OutputDir, flags.RateLimit, flags.Background)
            if err != nil {
                fmt.Println("Error downloading multiple files:", err)
            }
            return
        }
    // If mirror flag is set, mirror the website specified by the URL argument
    if flags.Mirror {

        if len(flags.URLs) != 1 {
            fmt.Println("Mirror mode requires exactly one URL")
            os.Exit(1)
        }
        
        // Set output directory
		outputDir := "mirrors"
		if flags.OutputDir != "" {
			if expanded, err := expandPath(flags.OutputDir); err != nil {
                fmt.Printf("error: %v\n", err)
				os.Exit(1) 
			} else {
				outputDir = expanded
			}
		}

		// Create mirror options
		MirrorParams := mirror.GetMirrorParams(flags.URLs[0], outputDir, flags.ConvertLinks, flags.RejectTypes, flags.ExcludePaths)
		if MirrorParams == nil {
            fmt.Printf("failed to create mirror options\n")
			os.Exit(1)
		}

		// Start mirroring
		fmt.Printf("Starting mirror of %s\n", flags.URLs[0])
		fmt.Printf("Output directory: %s\n", outputDir)

		if err := MirrorParams.Mirror(); err != nil {
            fmt.Printf("mirroring failed: %v\n", err)
			os.Exit(1) 
		}

		return
    }
    // If no flags match, download a single file from the provided URL argument
    if len(flags.URLs) == 0 {
        fmt.Println("URL is required for file download")
        return 
    }
    fileURL := flags.URLs[0]
   
    if err := download.DownloadFile(fileURL, flags.OutputFile, flags.OutputDir, flags.RateLimit , flags.Background); err != nil {
        fmt.Printf("download failed: %v\n", err)
        return 
    }
}
