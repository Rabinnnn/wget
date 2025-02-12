package download

import (
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "sync"
    "time"
    "wget/utils"
)

func DownloadFile(fileURL, outputFile, outputDir, rateLimit string) error {
    startTime := time.Now().Format("2006-01-02 15:04:05")
    fmt.Printf("Start at %s\n", startTime)

    resp, err := http.Get(fileURL)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("status: %s", resp.Status)
    }
    fmt.Printf("Sending request, awaiting response... status %s\n", resp.Status)

    contentLength := resp.ContentLength
    fmt.Printf("Content size: %d [~%.2fMB]\n", contentLength, float64(contentLength)/(1024*1024))

    fileName := outputFile
    if fileName == "" {
        fileName = filepath.Base(fileURL)
    }
    filePath := filepath.Join(outputDir, fileName)
    fmt.Printf("Saving file to: %s\n", filePath)

    if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
        return err
    }

    file, err := os.Create(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    var writer io.Writer = file
    if rateLimit != "" {
        limit, err := utils.ParseRateLimit(rateLimit)
        if err != nil {
            return err
        }
        writer = NewRateLimitedWriter(file, limit)
    }

    progressWriter := NewProgressWriter(writer, contentLength)
    _, err = io.Copy(progressWriter, resp.Body)
    if err != nil {
        return err
    }

    fmt.Printf("\nDownloaded [%s]\n", fileURL)
    fmt.Printf("Finished at %s\n", time.Now().Format("2006-01-02 15:04:05"))
    return nil
}

func DownloadMultipleFiles(urls []string, outputDir, rateLimit string) {
    var wg sync.WaitGroup
    for _, u := range urls {
        wg.Add(1)
        go func(url string) {
            defer wg.Done()
            err := DownloadFile(url, "", outputDir, rateLimit)
            if err != nil {
                fmt.Printf("Error downloading %s: %v\n", url, err)
            }
        }(u)
    }
    wg.Wait()
    fmt.Println("Download finished.")
}