package download

import (
    "fmt"
    "io"
    "time"
    "wget/utils"
)

type ProgressWriter struct {
    writer      io.Writer
    total       int64
    downloaded  int64
    lastPrinted time.Time
}

func NewProgressWriter(writer io.Writer, total int64) *ProgressWriter {
    return &ProgressWriter{writer: writer, total: total}
}

func (p *ProgressWriter) Write(data []byte) (int, error) {
    n, err := p.writer.Write(data)
    if err != nil {
        return n, err
    }
    p.downloaded += int64(n)
    p.printProgress()
    return n, nil
}

func (p *ProgressWriter) printProgress() {
    if time.Since(p.lastPrinted) < time.Second {
        return
    }
    p.lastPrinted = time.Now()
    percent := float64(p.downloaded) / float64(p.total) * 100
    fmt.Printf("\rDownloading... %.2f%% (%s/%s)", 
        percent, 
        utils.FormatBytes(p.downloaded), 
        utils.FormatBytes(p.total))
}