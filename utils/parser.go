package utils

import (
    "fmt"
    "strings"
)

// ParseRateLimit takes a rate limit string (e.g., "10k", "5m") and converts it 
// to an integer value in bytes (e.g., 10240 for "10k", 5242880 for "5m").
func ParseRateLimit(rateLimit string) (int64, error) {
    rateLimit = strings.ToLower(rateLimit)  // Normalize to lowercase to handle both "K" and "k"
    var multiplier int64 = 1  // Default multiplier for no suffix (1)

    // Check for rate limit suffixes like 'k' for kilobytes and 'm' for megabytes.
    switch {
    case strings.HasSuffix(rateLimit, "k"):
        multiplier = 1024  // 1k = 1024 bytes
        rateLimit = strings.TrimSuffix(rateLimit, "k")  // Remove the 'k' suffix
    case strings.HasSuffix(rateLimit, "m"):
        multiplier = 1024 * 1024  // 1m = 1024 * 1024 bytes
        rateLimit = strings.TrimSuffix(rateLimit, "m")  // Remove the 'm' suffix
    }
    
    // Parse the remaining part of the rate limit string as an integer.
    rate, err := ParseInt(rateLimit)
    if err != nil {
        return 0, err  
    }

    
    return rate * multiplier, nil
}

// ParseInt converts a string to an integer, returning an error if parsing fails.
func ParseInt(s string) (int64, error) {
    var result int64
    _, err := fmt.Sscanf(s, "%d", &result)  
    return result, err  
}
