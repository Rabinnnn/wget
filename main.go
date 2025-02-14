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
