package download

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

// ProgressWriter is a custom writer that tracks the progress of the download
// by updating download statistics like progress percentage, speed, and remaining time.
type ProgressWriter struct {
	writer      io.Writer
	total       int64
	downloaded  int64
	lastPrinted time.Time
	startTime   time.Time
	lastWidth   int // Store the last known terminal width
}

// NewProgressWriter creates a new ProgressWriter instance that tracks download progress.
// It initializes the writer and the total size of the file to be downloaded.
func NewProgressWriter(writer io.Writer, total int64) *ProgressWriter {
	return &ProgressWriter{
		writer:     writer,
		total:      total,
		startTime:  time.Now(),
		lastWidth:  GetTerminalWidth(), // Initialize with current terminal width
		lastPrinted: time.Now().Add(-1 * time.Second), // Ensure first update prints immediately
	}
}

// GetTerminalWidth gets the width of the terminal.
// Returns a fallback width of 50 if it can't determine the actual width.
func GetTerminalWidth() int {
	fd := int(os.Stdout.Fd())
	if width, _, err := term.GetSize(fd); err == nil {
		return width
	}
	return 50 // fallback width if we can't determine terminal width
}

// Write writes data to the underlying writer and tracks progress.
// It updates the amount of data downloaded and calls the `printProgress` function to display the progress.
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
	if time.Since(p.lastPrinted) < time.Second/5 && p.downloaded < p.total {
		return
	}

	// Check if terminal width has changed since last update
	currentWidth := GetTerminalWidth()
	terminalResized := currentWidth != p.lastWidth
	p.lastWidth = currentWidth // Update the stored width

	p.lastPrinted = time.Now()

	if p.total == 0 {
		fmt.Println("Error: Total file size is zero.")
		return
	}

	totalKiB := float64(p.total) / 1024
	downloadedKiB := float64(p.downloaded) / 1024

	var percent float64
	var barWidth int
	terminalWidth := currentWidth
	
	// Clear the line completely on resize to prevent artifacts
	if terminalResized {
		fmt.Print("\r\033[K")
	}

	// If the total size is unknown (Content-Length is -1), skip the percentage calculation.
	if p.total > 0 {
		percent = float64(p.downloaded) / float64(p.total) * 100
		barWidth = terminalWidth / 5 // bar width is a fifth of the terminal width
		
		// Ensure minimum bar width
		if barWidth < 3 {
			barWidth = 3
		}
	} else {
		percent = -1 // Indicating no percentage calculation
		barWidth = 25 // Default width since we don't know the total size
		
		// Adjust for very small terminals
		if barWidth > terminalWidth/3 {
			barWidth = terminalWidth / 3
		}
	}

	// Calculate download speed in MiB/s by dividing the downloaded bytes by the elapsed time in seconds.
	elapsed := time.Since(p.startTime).Seconds()
	speed := float64(p.downloaded) / (1024 * 1024 * elapsed) // MiB/s

	// Create a progress bar based on the percentage completed.
	completed := int(float64(barWidth) * (float64(p.downloaded) / float64(p.total)))
	if completed < 0 {
		completed = 0 // Ensure the progress is non-negative
	}
	if completed > barWidth {
		completed = barWidth // Ensure progress doesn't exceed bar width
	}
	
	bar := strings.Repeat("=", completed)
	
	// If bar is not complete, add a > character to show progress direction
	if completed < barWidth && completed > 0 {
		bar = bar[:len(bar)-1] + ">" + strings.Repeat(" ", barWidth-completed)
	} else {
		bar = bar + strings.Repeat(" ", barWidth-completed)
	}

	// Calculate the remaining time based on the current download speed and elapsed time.
	var remainingTime string
	if p.downloaded > 0 && p.total > 0 {
		bytesRemaining := p.total - p.downloaded
		timePerByte := elapsed / float64(p.downloaded)
		remainingSeconds := float64(bytesRemaining) * timePerByte

		if remainingSeconds < 1 {
			remainingTime = "0s"
		} else if remainingSeconds < 60 {
			remainingTime = fmt.Sprintf("%.1fs", remainingSeconds)
		} else if remainingSeconds < 3600 {
			remainingTime = fmt.Sprintf("%.1fm", remainingSeconds/60)
		} else {
			remainingTime = fmt.Sprintf("%.1fh", remainingSeconds/3600)
		}
	} else {
		remainingTime = "??s"
	}

	// If the total size is unknown, display a message instead of showing percentage
	if percent == -1 {
		fmt.Printf("\r\033[K %.2f KiB [%s] Downloading... %.2f MiB/s %s",
			downloadedKiB,
			bar,
			speed,
			remainingTime)
	} else {
		if terminalWidth < 55 {
			// For narrow terminals, use two lines
			// First clear the current line and print the first line of information
			fmt.Printf("\r\033[K %.2f KiB / %.2f KiB\n", downloadedKiB, totalKiB)
			// Then clear the next line and print the second line of information
			fmt.Printf("\r\033[K [%s] %.2f%% %.2f MiB/s %s", 
				bar, percent, speed, remainingTime)
			// Move cursor back up to be ready for the next update
			if p.downloaded != p.total {
				fmt.Print("\033[1A")
			}
		} else {
			// For wider terminals, everything on one line
			fmt.Printf("\r\033[K %.2f KiB / %.2f KiB [%s] %.2f%% %.2f MiB/s %s",
				downloadedKiB, totalKiB, bar, percent, speed, remainingTime)
		}
	}

	// Print a newline when download is complete
	if p.downloaded == p.total && p.total > 0 {
		if terminalWidth < 55 {
			fmt.Println() // Extra newline for the two-line display
		}
		fmt.Println()
	}
}