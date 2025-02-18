package config
import "flag"
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
    ConvertLinks bool
    UseDynamic bool   
}

// InitFlags initializes and parses command-line flags.
func InitFlags() *Flags {
    flags := &Flags{}
    
    // Initialize flags with their default values and descriptions
    flag.StringVar(&flags.OutputFile, "O", "", "Save the file with a different name")
    flag.StringVar(&flags.OutputDir, "P", ".", "Save the file in a specific directory")
    flag.StringVar(&flags.RateLimit, "rate-limit", "", "Limit the download speed (e.g., 200k, 2M)")
    flag.BoolVar(&flags.Background, "B", false, "Download in the background")
    flag.StringVar(&flags.InputFile, "i", "", "File containing multiple URLs to download")
    flag.BoolVar(&flags.Mirror, "mirror", false, "Mirror a website")
    flag.StringVar(&flags.Reject, "reject", "", "Reject specific file types (e.g., jpg,gif)") 
    flag.StringVar(&flags.Reject, "R", "", "Reject specific file types (e.g., jpg,gif)") 
    flag.StringVar(&flags.Exclude, "X", "", "Exclude specific directories (e.g., /js,/css)")
    flag.StringVar(&flags.Exclude, "exclude", "", "Exclude specific directories (e.g., /js,/css)")
    flag.BoolVar(&flags.ConvertLinks, "convert-links", false, "Convert links for offline viewing")
    flag.BoolVar(&flags.UseDynamic, "dynamic", true, "Enable javascript rendering")

    

    
    return flags
}
