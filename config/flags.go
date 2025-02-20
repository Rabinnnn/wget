package config

import (
	
	"flag"
	"fmt"
	"os"
	//"wget/download"
)

// Flags struct holds all the configurable parameters for the download operation.
type Flags struct {
	OutputFile   string
	OutputDir    string
	RateLimit    string
	Background   bool
	InputFile    string
	Mirror       bool
	Reject       string
	Exclude      string
	RejectTypes []string
	ExcludePaths []string
	ConvertLinks bool
	UseDynamic   bool
	URLs         []string // Added to store URLs from the input file
}

// InitFlags initializes and parses command-line flags.
func InitFlags() *Flags {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flags := &Flags{}

	// Initialize flags with their default values and descriptions
	fs.StringVar(&flags.OutputFile, "O", "", "Save the file with a different name")
	fs.StringVar(&flags.OutputDir, "P", ".", "Save the file in a specific directory")
	fs.StringVar(&flags.RateLimit, "rate-limit", "", "Limit the download speed (e.g., 200k, 2M)")
	fs.BoolVar(&flags.Background, "B", false, "Download in the background")
	fs.StringVar(&flags.InputFile, "i", "", "File containing multiple URLs to download")
	fs.BoolVar(&flags.Mirror, "mirror", false, "Mirror a website")
	// fs.StringVar(&flags.Reject, "reject", "", "Reject specific file types (e.g., jpg,gif)")
	// fs.StringVar(&flags.Reject, "R", "", "Reject specific file types (e.g., jpg,gif)")
	// fs.StringVar(&flags.Exclude, "X", "", "Exclude specific directories (e.g., /js,/css)")
	// fs.StringVar(&flags.Exclude, "exclude", "", "Exclude specific directories (e.g., /js,/css)")
	
	// Mirror-related flags
	var rejectListShort, rejectListLong string
	fs.StringVar(&rejectListShort, "R", "", "Reject file types (comma-separated list)")
	fs.StringVar(&rejectListLong, "reject", "", "Reject file types (comma-separated list)")

	var excludeListShort, excludeListLong string
	fs.StringVar(&excludeListShort, "X", "", "Exclude directories (comma-separated list)")
	fs.StringVar(&excludeListLong, "exclude", "", "Exclude directories (comma-separated list)")

	fs.BoolVar(&flags.ConvertLinks, "convert-links", false, "Convert links for offline viewing")
	fs.BoolVar(&flags.UseDynamic, "dynamic", true, "Enable javascript rendering")

	// Parse flags, but skip the program name
	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Println(err)
		return nil
	}

	// If an input file is provided, read the URLs from it
	// if flags.InputFile != "" {
	// 	urls, err := download.ReadURLsFromFile(flags.InputFile)
	// 	if err != nil {
	// 		fmt.Printf("Error reading URLs from file: %v\n", err)
	// 		os.Exit(1)
	// 	}
	// 	flags.URLs = urls
	// }
	// Get URLs from remaining arguments
	args := fs.Args()
	if len(args) < 1 && flags.InputFile == "" {
		fmt.Println("no URL specified")
		return nil
	}

	// Store URLs
	flags.URLs = args

	return flags
}

