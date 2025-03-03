package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
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
	flag.BoolVar(&flags.Background, "background", false, "Run download in background mode without showing progress")
	
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

	
	args := fs.Args()
	if len(args) < 1 && flags.InputFile == "" {
		fmt.Println("no URL specified")
		return nil
	}

	// Store URLs
	flags.URLs = args

		// Process reject lists (combine short and long options)
		rejectTypes := []string{}
		if rejectListShort != "" {
			rejectTypes = append(rejectTypes, strings.Split(rejectListShort, ",")...)
		}
		if rejectListLong != "" {
			rejectTypes = append(rejectTypes, strings.Split(rejectListLong, ",")...)
		}
		for i := range rejectTypes {
			rejectTypes[i] = strings.TrimSpace(rejectTypes[i])
		}
		flags.RejectTypes = rejectTypes

		// Process exclude lists (combine short and long options)
		excludePaths := []string{}
		if excludeListShort != "" {
			excludePaths = append(excludePaths, strings.Split(excludeListShort, ",")...)
		}
		if excludeListLong != "" {
			excludePaths = append(excludePaths, strings.Split(excludeListLong, ",")...)
		}
		for i := range excludePaths {
			excludePaths[i] = strings.TrimSpace(excludePaths[i])
		}
		flags.ExcludePaths = excludePaths


	return flags
}

