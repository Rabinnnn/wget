package main

import (
    "flag"
    "fmt"
    "os"
    "bufio"
    "wget/config"
    "wget/download"
    "wget/mirror"
)

func main() {
    // Initialize flags and parse command-line arguments
    flags := config.InitFlags()
    flag.Parse()
    // If background download flag is set, redirect output to a log file
    if flags.Background {
        logFile, err := os.Create("wget-log") // Create a log file
        if err != nil {
            fmt.Println("Error creating log file:", err)
            return
        }
        defer logFile.Close()
        os.Stdout = logFile // Redirect stdout to log file
        os.Stderr = logFile // Redirect stderr to log file
        fmt.Println("Output will be written to 'wget-log'.")
    }
    // If input file is provided, read URLs and initiate downloading multiple files
    if flags.InputFile != "" {
        urls, err := readURLsFromFile(flags.InputFile) // Read URLs from the file
        if err != nil {
            fmt.Println("Error reading URLs from file:", err)
            return
        }
        download.DownloadMultipleFiles(urls, flags.OutputDir, flags.RateLimit) // Download multiple files
        return
    }
    // If mirror flag is set, mirror the website specified by the URL argument
    if flags.Mirror {
        if len(flag.Args()) == 0 {
            fmt.Println("Error: URL is required for mirroring") // URL is required for mirroring
            return
        }
        websiteURL := flag.Args()[0]
        err := mirror.MirrorWebsite(websiteURL, flags.RateLimit) // Start mirroring the website
        if err != nil {
            fmt.Println("Error mirroring website:", err)
        }
        return
    }
    // If no flags match, download a single file from the provided URL argument
    if len(flag.Args()) == 0 {
        fmt.Println("Error: URL is required") // URL is required for file download
        return
    }

    fileURL := flag.Args()[0]
    err := download.DownloadFile(fileURL, flags.OutputFile, flags.OutputDir, flags.RateLimit) // Download single file
    if err != nil {
        fmt.Println("Error downloading file:", err)
    }
