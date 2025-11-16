package scanner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/spf13/afero"
)

var (
	ErrMaxFilesExceeded = errors.New("maximum file limit exceeded")
	ErrScanTimeout      = errors.New("scan operation timed out")
)

const (
	// MaxSkippedFiles is the maximum number of skipped file paths to store
	// to prevent unbounded memory growth when scanning directories with millions of non-video files
	MaxSkippedFiles = 1000
)

// Scanner finds video files based on configuration
type Scanner struct {
	fs     afero.Fs
	config *config.MatchingConfig
}

// NewScanner creates a new file scanner
func NewScanner(fs afero.Fs, cfg *config.MatchingConfig) *Scanner {
	return &Scanner{
		fs:     fs,
		config: cfg,
	}
}

// FileInfo represents a discovered video file
type FileInfo struct {
	Path      string    // Full absolute path
	Name      string    // Filename without path
	Extension string    // File extension (e.g., ".mp4")
	Size      int64     // File size in bytes
	ModTime   time.Time // Last modified time
	Dir       string    // Directory containing the file
}

// ScanResult contains the results of a directory scan
type ScanResult struct {
	Files        []FileInfo // Matched video files
	Skipped      []string   // Sample of skipped files (capped at MaxSkippedFiles)
	SkippedCount int        // Total count of skipped files
	Errors       []error    // Errors encountered during scan
	LimitReached bool       // Whether max file limit was reached
	TimedOut     bool       // Whether scan timed out
	TotalScanned int        // Total number of files scanned before limit/timeout
}

// Scan recursively scans a directory for video files (no limits)
func (s *Scanner) Scan(rootPath string) (*ScanResult, error) {
	// Call ScanWithLimits with no limits (context.Background(), maxFiles = 0)
	return s.ScanWithLimits(context.Background(), rootPath, 0)
}

// ScanWithLimits recursively scans a directory for video files with timeout and file count limits
// maxFiles = 0 means no limit
func (s *Scanner) ScanWithLimits(ctx context.Context, rootPath string, maxFiles int) (*ScanResult, error) {
	result := &ScanResult{
		Files:   make([]FileInfo, 0),
		Skipped: make([]string, 0),
		Errors:  make([]error, 0),
	}

	// Ensure path is absolute
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	// Verify path exists
	if _, err := s.fs.Stat(absPath); err != nil {
		return nil, err
	}

	// Walk the directory tree
	fileCount := 0
	err = filepath.WalkDir(absPath, func(path string, d os.DirEntry, err error) error {
		// Check context for timeout/cancellation (check every 100 files for performance)
		fileCount++
		if fileCount%100 == 0 {
			select {
			case <-ctx.Done():
				result.TimedOut = true
				return filepath.SkipAll // Stop walking
			default:
			}
		}

		if err != nil {
			result.Errors = append(result.Errors, err)
			return nil // Continue scanning
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		result.TotalScanned++

		// Check if file matches criteria
		if s.shouldIncludeFile(path, d) {
			info, err := d.Info()
			if err != nil {
				result.Errors = append(result.Errors, err)
				return nil
			}

			// Skip symlinks to prevent metadata leakage from protected files
			if info.Mode()&os.ModeSymlink != 0 {
				result.SkippedCount++
				if len(result.Skipped) < MaxSkippedFiles {
					result.Skipped = append(result.Skipped, path)
				}
				return nil
			}

			fileInfo := FileInfo{
				Path:      path,
				Name:      d.Name(),
				Extension: filepath.Ext(path),
				Size:      info.Size(),
				ModTime:   info.ModTime(),
				Dir:       filepath.Dir(path),
			}

			result.Files = append(result.Files, fileInfo)

			// Check if we've reached the file limit
			if maxFiles > 0 && len(result.Files) >= maxFiles {
				result.LimitReached = true
				return filepath.SkipAll // Stop walking
			}
		} else {
			// Track skipped files count, but only store first MaxSkippedFiles paths to prevent memory issues
			result.SkippedCount++
			if len(result.Skipped) < MaxSkippedFiles {
				result.Skipped = append(result.Skipped, path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// ScanSingle scans a single file or directory (non-recursive)
func (s *Scanner) ScanSingle(path string) (*ScanResult, error) {
	result := &ScanResult{
		Files:   make([]FileInfo, 0),
		Skipped: make([]string, 0),
		Errors:  make([]error, 0),
	}

	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	// Check if it's a file or directory
	info, err := s.fs.Stat(absPath)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		// Scan directory (non-recursive)
		entries, err := afero.ReadDir(s.fs, absPath)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			fullPath := filepath.Join(absPath, entry.Name())
			if s.shouldIncludeFile(fullPath, nil) {
				// Skip symlinks to prevent metadata leakage from protected files
				if entry.Mode()&os.ModeSymlink != 0 {
					result.SkippedCount++
					if len(result.Skipped) < MaxSkippedFiles {
						result.Skipped = append(result.Skipped, fullPath)
					}
					continue
				}

				fileInfo := FileInfo{
					Path:      fullPath,
					Name:      entry.Name(),
					Extension: filepath.Ext(fullPath),
					Size:      entry.Size(),
					ModTime:   entry.ModTime(),
					Dir:       absPath,
				}

				result.Files = append(result.Files, fileInfo)
			} else {
				result.SkippedCount++
				if len(result.Skipped) < MaxSkippedFiles {
					result.Skipped = append(result.Skipped, fullPath)
				}
			}
		}
	} else {
		// Single file
		if s.shouldIncludeFile(absPath, nil) {
			// Skip symlinks to prevent metadata leakage from protected files
			if info.Mode()&os.ModeSymlink != 0 {
				result.SkippedCount++
				if len(result.Skipped) < MaxSkippedFiles {
					result.Skipped = append(result.Skipped, absPath)
				}
			} else {
				fileInfo := FileInfo{
					Path:      absPath,
					Name:      info.Name(),
					Extension: filepath.Ext(absPath),
					Size:      info.Size(),
					ModTime:   info.ModTime(),
					Dir:       filepath.Dir(absPath),
				}

				result.Files = append(result.Files, fileInfo)
			}
		} else {
			result.SkippedCount++
			if len(result.Skipped) < MaxSkippedFiles {
				result.Skipped = append(result.Skipped, absPath)
			}
		}
	}

	return result, nil
}

// shouldIncludeFile checks if a file should be included based on configuration
func (s *Scanner) shouldIncludeFile(path string, entry os.DirEntry) bool {
	// Check extension
	ext := strings.ToLower(filepath.Ext(path))
	hasValidExt := false
	for _, validExt := range s.config.Extensions {
		if ext == strings.ToLower(validExt) {
			hasValidExt = true
			break
		}
	}
	if !hasValidExt {
		return false
	}

	// Check exclude patterns (glob patterns)
	basename := filepath.Base(path)
	for _, pattern := range s.config.ExcludePatterns {
		matched, err := filepath.Match(pattern, basename)
		if err == nil && matched {
			return false
		}
	}

	// Check minimum file size
	if s.config.MinSizeMB > 0 {
		var size int64
		if entry != nil {
			info, err := entry.Info()
			if err == nil {
				size = info.Size()
			}
		} else {
			info, err := s.fs.Stat(path)
			if err == nil {
				size = info.Size()
			}
		}

		minBytes := int64(s.config.MinSizeMB) * 1024 * 1024
		if size < minBytes {
			return false
		}
	}

	return true
}

// Filter filters a list of files based on configuration
func (s *Scanner) Filter(files []string) []FileInfo {
	result := make([]FileInfo, 0)

	for _, path := range files {
		if !s.shouldIncludeFile(path, nil) {
			continue
		}

		// Get file info
		info, err := s.fs.Stat(path)
		if err != nil {
			continue
		}

		if info.IsDir() {
			continue
		}

		// Skip symlinks to prevent metadata leakage from protected files
		if info.Mode()&os.ModeSymlink != 0 {
			continue
		}

		fileInfo := FileInfo{
			Path:      path,
			Name:      info.Name(),
			Extension: filepath.Ext(path),
			Size:      info.Size(),
			ModTime:   info.ModTime(),
			Dir:       filepath.Dir(path),
		}

		result = append(result, fileInfo)
	}

	return result
}
