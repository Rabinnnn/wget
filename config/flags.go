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
}

// InitFlags initializes and parses command-line flags.
func InitFlags() *Flags {
    flags := &Flags{}
