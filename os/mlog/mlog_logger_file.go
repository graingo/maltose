package mlog

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/graingo/maltose/errors/merror"
)

// fileWriter is a writer that writes to files based on date patterns or fixed file names.
type fileWriter struct {
	basePath    string        // Base directory path
	filePattern string        // File name pattern
	autoClean   int           // Auto cleanup days
	mu          sync.Mutex    // Mutex for concurrency safety
	file        *os.File      // Current open file
	currentPath string        // Current file path
	lastCheck   time.Time     // Last file check time
	lastCleanup time.Time     // Last cleanup check time
	stopChan    chan struct{} // Stop channel for cleanup goroutine
	isDateMode  bool          // Whether using date pattern mode
}

// newFileWriter creates a new DateWriter.
func newFileWriter(basePath string, filePattern string, autoClean int) (*fileWriter, error) {
	// Ensure directory exists
	dir := logDir(basePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, merror.Wrapf(err, "failed to create log directory: %s", dir)
	}

	w := &fileWriter{
		basePath:    basePath,
		filePattern: filePattern,
		autoClean:   autoClean,
		lastCheck:   time.Now(),
		lastCleanup: time.Now(),
		stopChan:    make(chan struct{}),
		isDateMode:  isDatePattern(filePattern),
	}

	// Start cleanup goroutine if auto cleanup is enabled and using date mode
	if autoClean > 0 && w.isDateMode {
		go w.cleanupRoutine()
	}

	return w, nil
}

// Write implements the io.Writer interface.
func (w *fileWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Lazy initialization: open the file on first write.
	if w.file == nil {
		if err = w.checkAndRotate(); err != nil {
			return 0, err
		}
	}

	// Check if we need to switch to a new file based on current date
	// Only check periodically and only if using date mode
	if w.isDateMode && time.Since(w.lastCheck) >= time.Minute {
		if err := w.checkAndRotate(); err != nil {
			return 0, err
		}
		w.lastCheck = time.Now()
	}

	return w.file.Write(p)
}

// Close closes the current file and stops the cleanup goroutine.
func (w *fileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Stop cleanup goroutine if running
	close(w.stopChan)

	// Close file
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// checkAndRotate checks if the file needs to be rotated based on the current date.
func (w *fileWriter) checkAndRotate() error {
	// Generate file path based on current date or fixed pattern
	var filePath string
	if w.isDateMode {
		filePath = w.formatFilePath(time.Now())
	} else {
		filePath = filepath.Join(logDir(w.basePath), w.filePattern)
	}

	// If the path is the same, no need to rotate
	if filePath == w.currentPath && w.file != nil {
		return nil
	}

	// Close current file if open
	if w.file != nil {
		w.file.Close()
		w.file = nil
	}

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return merror.Wrapf(err, "failed to create log directory: %s", dir)
	}

	// Open new file
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return merror.Wrapf(err, "failed to open log file: %s", filePath)
	}

	// Update state
	w.file = file
	w.currentPath = filePath

	return nil
}

// formatFilePath formats the file path based on the current date.
func (w *fileWriter) formatFilePath(t time.Time) string {
	// Replace date placeholders
	pattern := w.filePattern
	pattern = strings.ReplaceAll(pattern, "{Y}", t.Format("2006"))
	pattern = strings.ReplaceAll(pattern, "{y}", t.Format("06"))
	pattern = strings.ReplaceAll(pattern, "{m}", t.Format("01"))
	pattern = strings.ReplaceAll(pattern, "{d}", t.Format("02"))
	pattern = strings.ReplaceAll(pattern, "{H}", t.Format("15"))
	pattern = strings.ReplaceAll(pattern, "{i}", t.Format("04"))
	pattern = strings.ReplaceAll(pattern, "{s}", t.Format("05"))

	// Combine with base path
	return filepath.Join(logDir(w.basePath), pattern)
}

// cleanupRoutine periodically cleans up old log files.
func (w *fileWriter) cleanupRoutine() {
	ticker := time.NewTicker(time.Hour) // Check every hour
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.cleanup()
		case <-w.stopChan:
			return
		}
	}
}

// cleanup removes old log files based on autoClean setting.
func (w *fileWriter) cleanup() {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Skip if auto cleanup is not enabled or not in date mode
	if w.autoClean <= 0 || !w.isDateMode {
		return
	}

	// Convert date pattern to regex pattern for matching
	regexPattern := convertDatePatternToRegex(w.filePattern)
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		// Log error but continue
		fmt.Printf("Invalid pattern for cleanup: %s, error: %v\n", w.filePattern, err)
		return
	}

	// Get all files in directory
	dir := logDir(w.basePath)
	files, err := os.ReadDir(dir)
	if err != nil {
		// Log error but continue
		fmt.Printf("Failed to read directory for cleanup: %s, error: %v\n", dir, err)
		return
	}

	now := time.Now()
	maxAge := time.Duration(w.autoClean) * 24 * time.Hour // Convert to days

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Check if file matches pattern
		if !re.MatchString(file.Name()) {
			continue
		}

		// Check file age
		info, err := file.Info()
		if err != nil {
			continue
		}

		if now.Sub(info.ModTime()) > maxAge {
			filePath := filepath.Join(dir, file.Name())
			// Skip current file
			if filePath == w.currentPath {
				continue
			}
			if err := os.Remove(filePath); err != nil {
				// Log error but continue
				fmt.Printf("Failed to remove file %s: %v\n", filePath, err)
			}
		}
	}

	w.lastCleanup = now
}

// convertDatePatternToRegex converts a date pattern to a regex pattern.
func convertDatePatternToRegex(pattern string) string {
	pattern = regexp.QuoteMeta(pattern)
	pattern = strings.ReplaceAll(pattern, "\\{Y\\}", "\\d{4}")
	pattern = strings.ReplaceAll(pattern, "\\{y\\}", "\\d{2}")
	pattern = strings.ReplaceAll(pattern, "\\{m\\}", "\\d{2}")
	pattern = strings.ReplaceAll(pattern, "\\{d\\}", "\\d{2}")
	pattern = strings.ReplaceAll(pattern, "\\{H\\}", "\\d{2}")
	pattern = strings.ReplaceAll(pattern, "\\{i\\}", "\\d{2}")
	pattern = strings.ReplaceAll(pattern, "\\{s\\}", "\\d{2}")
	return "^" + pattern + "$"
}

// isDatePattern checks if a file pattern contains date placeholders.
func isDatePattern(pattern string) bool {
	return strings.Contains(pattern, "{") && strings.Contains(pattern, "}")
}

// logDir determines the directory for the log files.
// If basePath is a simple name like "logs" without any path separators or extension,
// it's treated as a directory. Otherwise, we extract the directory part.
// This helps avoid `filepath.Dir("logs")` returning ".", which would place logs in the current directory.
func logDir(basePath string) string {
	if !strings.ContainsAny(basePath, `/\`) && filepath.Ext(basePath) == "" {
		return basePath
	}
	return filepath.Dir(basePath)
}
