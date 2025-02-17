package download

import (
    "io"
    "time"
)

type RateLimitedWriter struct {
    writer    io.Writer
    bandwidth int64
}

func NewRateLimitedWriter(writer io.Writer, bandwidth int64) *RateLimitedWriter {
    return &RateLimitedWriter{writer: writer, bandwidth: bandwidth}
}

func (r *RateLimitedWriter) Write(p []byte) (int, error) {
    start := time.Now()
    n, err := r.writer.Write(p)
    if err != nil {
        return n, err
    }
    
    elapsed := time.Since(start)
    expectedTime := time.Duration(n) * time.Second / time.Duration(r.bandwidth)
    if elapsed < expectedTime {
        time.Sleep(expectedTime - elapsed)
    }
    return n, nil
}