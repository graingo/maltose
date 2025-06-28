package mlog

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/graingo/maltose/errors/merror"
)

// RotationConfig holds all the configuration for log file rotation.
type rotationConfig struct {
	// MaxSize is the maximum size in megabytes of the log file before it gets rotated.
	// It is only applicable for 'size' rotation type.
	MaxSize int `mconv:"max_size"` // (MB)
	// MaxBackups is the maximum number of old log files to retain.
	// It is only applicable for 'size' rotation type.
	MaxBackups int `mconv:"max_backups"` // (files)
	// MaxAge is the maximum number of days to retain old log files.
	// It is applicable for both 'size' and 'date' rotation types.
	MaxAge int `mconv:"max_age"` // (days)
}

// fileWriter is a writer that writes to files based on date patterns or fixed file names.
type fileWriter struct {
	pathPattern  string        // Full path pattern for the log file
	isDateMode   bool          // Whether using date pattern mode
	mu           sync.Mutex    // Mutex for concurrency safety
	file         *os.File      // Current open file
	currentPath  string        // Current file path
	lastCheck    time.Time     // Last file check time
	lastCleanup  time.Time     // Last cleanup check time
	stopChan     chan struct{} // Stop channel for cleanup goroutine
	cfg          *rotationConfig
	cleanupRegex *regexp.Regexp
}

// newFileWriter creates a new fileWriter based on the provided rotation config.
func newFileWriter(filepath string, cfg *rotationConfig) (*fileWriter, error) {
	if filepath == "" {
		return nil, merror.New("filepath for log rotation cannot be empty")
	}

	isDateMode := isDatePattern(filepath)

	// Ensure directory exists
	dir := logDir(filepath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, merror.Wrapf(err, "failed to create log directory: %s", dir)
	}

	w := &fileWriter{
		pathPattern: filepath,
		cfg:         cfg,
		lastCheck:   time.Now(),
		lastCleanup: time.Now(),
		stopChan:    make(chan struct{}),
		isDateMode:  isDateMode,
	}

	// Start cleanup goroutine if needed
	if cfg.MaxAge > 0 || (!isDateMode && cfg.MaxBackups > 0) {
		if w.isDateMode {
			regexPattern := convertDatePatternToRegex(w.pathPattern)
			re, err := regexp.Compile(regexPattern)
			if err != nil {
				return nil, merror.Wrapf(err, "invalid file pattern for cleanup regex compilation: %s", w.pathPattern)
			}
			w.cleanupRegex = re
		}
		go w.cleanupRoutine()
	}

	return w, nil
}

// Write implements the io.Writer interface.
func (w *fileWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Lazy initialization or periodic rotation
	if err = w.checkAndRotate(); err != nil {
		return 0, err
	}

	// Rotate by size if not in date mode
	if !w.isDateMode && w.cfg.MaxSize > 0 {
		if stat, err := w.file.Stat(); err == nil {
			// Rotate if size exceeds max size
			if stat.Size() >= int64(w.cfg.MaxSize)*1024*1024 {
				if err := w.rotate(); err != nil {
					return 0, err
				}
			}
		}
	}

	return w.file.Write(p)
}

// rotate performs a size-based rotation.
func (w *fileWriter) rotate() error {
	// Close existing file
	if err := w.file.Close(); err != nil {
		return err
	}
	w.file = nil

	// Rename current log file to a backup name
	backupPath := w.backupFilePath()
	if err := os.Rename(w.currentPath, backupPath); err != nil {
		return merror.Wrapf(err, "failed to rename log file for rotation: %s", w.currentPath)
	}

	// Re-open the original file, which will be new and empty
	return w.checkAndRotate()
}

// backupFilePath generates a backup file path with a timestamp.
// e.g., /path/to/app.2023-10-27T10-00-00.000.log
func (w *fileWriter) backupFilePath() string {
	dir := filepath.Dir(w.currentPath)
	filename := filepath.Base(w.currentPath)
	ext := filepath.Ext(filename)
	prefix := filename[:len(filename)-len(ext)]
	timestamp := time.Now().Format("20060102150405000")

	return filepath.Join(dir, fmt.Sprintf("%s.%s%s", prefix, timestamp, ext))
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
		filePath = w.pathPattern
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
	pattern := w.pathPattern
	pattern = strings.ReplaceAll(pattern, "{Y}", t.Format("2006"))
	pattern = strings.ReplaceAll(pattern, "{y}", t.Format("06"))
	pattern = strings.ReplaceAll(pattern, "{m}", t.Format("01"))
	pattern = strings.ReplaceAll(pattern, "{d}", t.Format("02"))
	pattern = strings.ReplaceAll(pattern, "{H}", t.Format("15"))
	pattern = strings.ReplaceAll(pattern, "{i}", t.Format("04"))
	pattern = strings.ReplaceAll(pattern, "{s}", t.Format("05"))

	return pattern
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

	// For size mode, cleanup is based on MaxAge or MaxBackups.
	// For date mode, cleanup is based on cfg.MaxAge.
	if w.isDateMode {
		if w.cfg.MaxAge <= 0 {
			return
		}
	} else {
		if w.cfg.MaxAge <= 0 && w.cfg.MaxBackups <= 0 {
			return
		}
	}

	dir := logDir(w.pathPattern)
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("Failed to read directory for cleanup: %s, error: %v\n", dir, err)
		return
	}

	if w.isDateMode {
		w.cleanupDateMode(files, dir)
	} else {
		w.cleanupSizeMode(files, dir)
	}

	w.lastCleanup = time.Now()
}

func (w *fileWriter) cleanupDateMode(files []os.DirEntry, dir string) {
	if w.cleanupRegex == nil || w.cfg.MaxAge <= 0 {
		return
	}
	maxAge := time.Duration(w.cfg.MaxAge) * 24 * time.Hour
	now := time.Now()

	for _, file := range files {
		if file.IsDir() || !w.cleanupRegex.MatchString(file.Name()) {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		if now.Sub(info.ModTime()) > maxAge {
			filePath := filepath.Join(dir, file.Name())
			if filePath == w.currentPath {
				continue
			}
			os.Remove(filePath)
		}
	}
}

type backupFile struct {
	path    string
	modTime time.Time
}

func (w *fileWriter) cleanupSizeMode(files []os.DirEntry, dir string) {
	if w.cfg.MaxAge <= 0 && w.cfg.MaxBackups <= 0 {
		return
	}
	var backupFiles []backupFile
	filePattern := filepath.Base(w.pathPattern)
	prefix := filePattern[:len(filePattern)-len(filepath.Ext(filePattern))]
	ext := filepath.Ext(filePattern)

	for _, file := range files {
		if file.IsDir() || !strings.HasPrefix(file.Name(), prefix) || !strings.HasSuffix(file.Name(), ext) {
			continue
		}
		// Skip the main log file
		if file.Name() == filePattern {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}
		backupFiles = append(backupFiles, backupFile{
			path:    filepath.Join(dir, file.Name()),
			modTime: info.ModTime(),
		})
	}

	// Sort by mod time, oldest first
	sort.Slice(backupFiles, func(i, j int) bool {
		return backupFiles[i].modTime.Before(backupFiles[j].modTime)
	})

	// Cleanup by max age
	if w.cfg.MaxAge > 0 {
		maxAgeDuration := time.Duration(w.cfg.MaxAge) * 24 * time.Hour
		now := time.Now()
		var filesToKeep []backupFile
		for _, f := range backupFiles {
			if now.Sub(f.modTime) > maxAgeDuration {
				os.Remove(f.path)
			} else {
				filesToKeep = append(filesToKeep, f)
			}
		}
		backupFiles = filesToKeep
	}

	// Cleanup by max backups
	if w.cfg.MaxBackups > 0 && len(backupFiles) > w.cfg.MaxBackups {
		filesToRemove := backupFiles[:len(backupFiles)-w.cfg.MaxBackups]
		for _, f := range filesToRemove {
			os.Remove(f.path)
		}
	}
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
