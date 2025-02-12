package utils

import (
    "fmt"
    "strings"
)

func ParseRateLimit(rateLimit string) (int64, error) {
    rateLimit = strings.ToLower(rateLimit)
    var multiplier int64 = 1
    
    switch {
    case strings.HasSuffix(rateLimit, "k"):
        multiplier = 1024
        rateLimit = strings.TrimSuffix(rateLimit, "k")
    case strings.HasSuffix(rateLimit, "m"):
        multiplier = 1024 * 1024
        rateLimit = strings.TrimSuffix(rateLimit, "m")
    }
    
    rate, err := ParseInt(rateLimit)
    if err != nil {
        return 0, err
    }
    return rate * multiplier, nil
}

func ParseInt(s string) (int64, error) {
    var result int64
    _, err := fmt.Sscanf(s, "%d", &result)
    return result, err
}