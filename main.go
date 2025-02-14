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
