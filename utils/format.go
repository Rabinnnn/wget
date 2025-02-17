package utils

import "fmt"

// FormatBytes takes a byte value (in bytes) and converts it into a human-readable string 
// with appropriate units (B, KB, MB, GB, etc.).
// It uses binary prefixes where 1 KB = 1024 bytes, 1 MB = 1024 KB, etc.
func FormatBytes(bytes int64) string {
    const unit = 1024 
    
    if bytes < unit {
        return fmt.Sprintf("%d B", bytes)
    }
    
    // Initialize variables for the division factor and exponent.
    div, exp := int64(unit), 0
    
    // Loop to determine which unit (KB, MB, GB, etc.) is appropriate.
    // This divides the byte value by 1024 until it reaches a value less than 1024.
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit  // Scale the divisor by 1024 each time (KB, MB, GB, etc.)
        exp++        // Increase exponent to represent the next unit (K, M, G, etc.)
    }
    
    // Format the final result with two decimal places, and use the appropriate unit 
    // based on the exponent (K for kilo, M for mega, etc.).
    return fmt.Sprintf("%.2f %cB", 
        float64(bytes)/float64(div),  // Calculate the value in the current unit
        "KMGTPE"[exp])                // Use the correct character from "KMGTPE" (for KB, MB, etc.)
}
