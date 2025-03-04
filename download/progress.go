package download

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
)

// OutputManager handles all console output for multiple downloads
type OutputManager struct {
	mu             sync.Mutex
	progressBars   map[string]*ProgressInfo
	terminalHeight int
	terminalWidth  int
	logMessages    []string
	maxLogLines    int
}

// ProgressInfo stores information about a download's progress
type ProgressInfo struct {
	URL         string
	Total       int64
	Downloaded  int64
	Speed       float64
	StartTime   time.Time
	LastUpdated time.Time
}

// NewOutputManager creates a new output manager
func NewOutputManager() *OutputManager {
	height, width := getTerminalSize()
	return &OutputManager{
		progressBars:   make(map[string]*ProgressInfo),
		terminalHeight: height,
		terminalWidth:  width,
		maxLogLines:    10, // Maximum number of log messages to keep
	}
}

// getTerminalSize gets the current terminal dimensions
func getTerminalSize() (int, int) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 24, 80 // Fallback values
	}
	return height, width
}

// RegisterDownload adds a new download to be tracked
func (om *OutputManager) RegisterDownload(url string, total int64) {
	om.mu.Lock()
	defer om.mu.Unlock()

	om.progressBars[url] = &ProgressInfo{
		URL:       url,
		Total:     total,
		StartTime: time.Now(),
	}
	
	// Update display immediately
	om.updateDisplay()
}

// UpdateProgress updates the progress of a download
func (om *OutputManager) UpdateProgress(url string, downloaded int64) {
	om.mu.Lock()
	defer om.mu.Unlock()

	if info, exists := om.progressBars[url]; exists {
		info.Downloaded = downloaded
		
		// Calculate speed
		elapsed := time.Since(info.StartTime).Seconds()
		if elapsed > 0 {
			info.Speed = float64(downloaded) / (1024 * 1024 * elapsed) // MiB/s
		}
		
		info.LastUpdated = time.Now()
		
		// Update display if enough time has passed since last update
		if len(om.progressBars) == 1 || time.Since(info.LastUpdated) > time.Millisecond*200 {
			om.updateDisplay()
		}
	}
}

// CompleteDownload marks a download as complete and removes its progress bar
func (om *OutputManager) CompleteDownload(url string) {
	om.mu.Lock()
	defer om.mu.Unlock()

	delete(om.progressBars, url)
	om.updateDisplay()
}

// Log adds a log message to be displayed
func (om *OutputManager) Log(format string, args ...interface{}) {
	om.mu.Lock()
	defer om.mu.Unlock()

	message := fmt.Sprintf(format, args...)
	om.logMessages = append(om.logMessages, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), message))
	
	// Keep only the last N log messages
	if len(om.logMessages) > om.maxLogLines {
		om.logMessages = om.logMessages[len(om.logMessages)-om.maxLogLines:]
	}
	
	om.updateDisplay()
}

// updateDisplay redraws the entire console output
func (om *OutputManager) updateDisplay() {
	// Get current terminal size
	om.terminalHeight, om.terminalWidth = getTerminalSize()
	
	// Clear screen and move cursor to top-left
	fmt.Print("\033[2J\033[H")
	
	// Calculate available space
	logArea := len(om.logMessages)
	if logArea > om.maxLogLines {
		logArea = om.maxLogLines
	}
	
	progressArea := len(om.progressBars)
	if progressArea == 0 {
		progressArea = 1 // Always leave at least one line for progress
	}
	
	// Make sure we have enough room
	totalNeeded := logArea + progressArea + 1 // +1 for the divider
	if totalNeeded > om.terminalHeight {
		// If not enough room, reduce log area
		logArea = om.terminalHeight - progressArea - 1
		if logArea < 0 {
			logArea = 0
		}
	}
	
	// Print log messages
	startIdx := 0
	if len(om.logMessages) > logArea {
		startIdx = len(om.logMessages) - logArea
	}
	
	for i := startIdx; i < len(om.logMessages); i++ {
		fmt.Println(om.logMessages[i])
	}
	
	// Print divider if we have both logs and progress bars
	if logArea > 0 && len(om.progressBars) > 0 {
		fmt.Println(strings.Repeat("-", om.terminalWidth))
	}
	
	// Print progress bars
	for url, info := range om.progressBars {
		om.printProgressBar(url, info)
	}
}

// printProgressBar prints a single progress bar
func (om *OutputManager) printProgressBar(url string, info *ProgressInfo) {
	// Calculate values
	downloadedKiB := float64(info.Downloaded) / 1024
	totalKiB := float64(info.Total) / 1024
	
	// Create short URL for display
	displayURL := url
	if len(url) > 30 {
		displayURL = "..." + url[len(url)-27:]
	}
	
	// Calculate progress percentage and bar
	var percent float64
	barWidth := om.terminalWidth / 4
	if barWidth < 10 {
		barWidth = 10
	}
	
	if info.Total > 0 {
		percent = float64(info.Downloaded) / float64(info.Total) * 100
	} else {
		percent = -1
	}
	
	// Create progress bar
	completed := 0
	if info.Total > 0 {
		completed = int(float64(barWidth) * (float64(info.Downloaded) / float64(info.Total)))
		if completed > barWidth {
			completed = barWidth
		}
	}
	
	bar := strings.Repeat("=", completed)
	if completed < barWidth && completed > 0 {
		bar = bar[:len(bar)-1] + ">" + strings.Repeat(" ", barWidth-completed)
	} else {
		bar = bar + strings.Repeat(" ", barWidth-completed)
	}
	
	// Calculate remaining time
	var remainingTime string
	if info.Downloaded > 0 && info.Total > 0 {
		bytesRemaining := info.Total - info.Downloaded
		timePerByte := time.Since(info.StartTime).Seconds() / float64(info.Downloaded)
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
	
	// Print progress information
	if percent == -1 {
		fmt.Printf("%s: %.2f KiB [%s] %.2f MiB/s %s\n",
			displayURL,
			downloadedKiB,
			bar,
			info.Speed,
			remainingTime)
	} else {
		fmt.Printf(" %.2f/%.2f KiB [%s] %.1f%% %.2f MiB/s %s\n",
			//displayURL,
			downloadedKiB,
			totalKiB,
			bar,
			percent,
			info.Speed,
			remainingTime)
	}
}