package download

import (
	"fmt"
	"io"
	"strings"
    "os"
    "golang.org/x/term"
	"time"
)

// ProgressWriter is a custom writer that tracks the progress of the download
// by updating download statistics like progress percentage, speed, and remaining time.
type ProgressWriter struct {
	writer      io.Writer    
	total       int64        
	downloaded  int64        
	lastPrinted time.Time    
	startTime   time.Time   
}

// NewProgressWriter creates a new ProgressWriter instance that tracks download progress.
// It initializes the writer and the total size of the file to be downloaded.
func NewProgressWriter(writer io.Writer, total int64) *ProgressWriter {
	return &ProgressWriter{
		writer:    writer,
		total:     total,
		startTime: time.Now(), 
	}
}

// getTerminalWidth gets the width of the terminal.
// Returns a fallback width of 80 if it can't determine the actual width.
func GetTerminalWidth() int {
	fd := int(os.Stdout.Fd())
	if width, _, err := term.GetSize(fd); err == nil {
		return width
	}
	return 50 // fallback width if we can't determine terminal width
}


// Write writes data to the underlying writer and tracks progress.
// It updates the amount of data downloaded and calls the `printProgress` function to display the progress.
// Write the data to the underlying writer (e.g., file or buffer).
func (p *ProgressWriter) Write(data []byte) (int, error) {
	
	n, err := p.writer.Write(data)
	if err != nil {
		return n, err 
	}

	
	p.downloaded += int64(n)

	
	p.printProgress()

	return n, nil
}

// printProgress prints the progress of the download to the console.
// It displays the downloaded data, total size, progress bar, download speed, and estimated remaining time.
func (p *ProgressWriter) printProgress() {
    // Limit the frequency of printing progress (only print every 500ms).
    if time.Since(p.lastPrinted) < time.Second/2 {
        return
    }

    p.lastPrinted = time.Now()

    if p.total == 0 {
        fmt.Println("Error: Total file size is zero.")
        return
    }

    totalKiB := float64(p.total) / 1024
    downloadedKiB := float64(p.downloaded) / 1024

    var percent float64
    var barWidth int
    terminalWidth:=GetTerminalWidth()
   
    // If the total size is unknown (Content-Length is -1), skip the percentage calculation.
    if p.total > 0 {
        percent = float64(p.downloaded) / float64(p.total) * 100
        barWidth = terminalWidth / 5 // bar width is a fifth of the terminal width
        if terminalWidth < 55 {
            barWidth = 3
        }
    } else {
        percent = -1 // Indicating no percentage calculation
        barWidth = 25 // Decrease width since we don't know the total size
    }

    // Calculate download speed in MiB/s by dividing the downloaded bytes by the elapsed time in seconds.
    elapsed := time.Since(p.startTime).Seconds()
    speed := float64(p.downloaded) / (1024 * 1024 * elapsed) // MiB/s

    // Create a progress bar (50 characters wide) based on the percentage completed.
    completed := int(float64(barWidth) * (float64(p.downloaded) / float64(p.total)))
    if completed < 0 {
        completed = 0 // Ensure the progress is non-negative
    }
    bar := strings.Repeat("=", completed)

    // Calculate the remaining time based on the current download speed and elapsed time.
    var remainingTime string
    if p.downloaded > 0 && p.total > 0 {
        bytesRemaining := p.total - p.downloaded
        timePerByte := elapsed / float64(p.downloaded)
        remainingSeconds := float64(bytesRemaining) * timePerByte

        if remainingSeconds < 1 {
            remainingTime = "0s"
        } else {
            remainingTime = fmt.Sprintf("%.1fs", remainingSeconds)
        }
    } else {
        remainingTime = "??s"
    }

    // If the total size is unknown, display a message instead of showing percentage
    if percent == -1 {
        fmt.Printf("\r %.2f KiB [%s] Downloading... %.2f MiB/s %s",
            downloadedKiB,
            bar,
            speed,
            remainingTime)
    } else {
        fmt.Printf("\r %.2f KiB / %.2f KiB [%s] %.2f%% %.2f MiB/s %s",
            downloadedKiB,
            totalKiB,
            bar,
            percent,
            speed,
            remainingTime)
    }

    if p.downloaded == p.total && p.total > 0 {
        fmt.Println()
    }
}
