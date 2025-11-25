package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/javinizer/javinizer-go/internal/worker"
)

// FilterVideoFiles returns only video files matching the provided extensions.
// This is a pure function: it creates a new slice without modifying the input.
//
// Parameters:
//   - files: slice of FileItem to filter
//   - extensions: list of valid video extensions (e.g., ".mp4", ".mkv", ".avi")
//
// Returns:
//   - new slice containing only files matching the extensions
func FilterVideoFiles(files []FileItem, extensions []string) []FileItem {
	if len(extensions) == 0 {
		return []FileItem{}
	}

	filtered := make([]FileItem, 0, len(files))

	for _, file := range files {
		// Skip directories
		if file.IsDir {
			continue
		}

		// Check if file matches any extension
		for _, ext := range extensions {
			if strings.HasSuffix(strings.ToLower(file.Path), strings.ToLower(ext)) {
				filtered = append(filtered, file)
				break
			}
		}
	}

	return filtered
}

// FormatFileStatus returns a human-readable status string for a file based on its matched state.
// This is a pure function with no side effects.
//
// Status format:
//   - "matched [ID]" - file has been matched to a JAV ID
//   - "unmatched" - file has not been matched yet
//
// Note: FileItem struct currently tracks Matched and ID fields only.
// Processing/Progress state is tracked separately in the TUI Model.
//
// Parameters:
//   - file: FileItem to format status for
//
// Returns:
//   - formatted status string
func FormatFileStatus(file FileItem) string {
	if file.Matched && file.ID != "" {
		return fmt.Sprintf("matched [%s]", file.ID)
	}
	return "unmatched"
}

// SortFilesByStatus sorts files by matched state and then alphabetically by path.
// This is a pure function: it creates a new sorted slice without modifying the input.
//
// Sort order:
//  1. Matched files (sorted alphabetically by path)
//  2. Unmatched files (sorted alphabetically by path)
//  3. Directories (sorted alphabetically by path)
//
// Parameters:
//   - files: slice of FileItem to sort
//
// Returns:
//   - new sorted slice (original unchanged)
func SortFilesByStatus(files []FileItem) []FileItem {
	// Create new slice to preserve immutability
	sorted := make([]FileItem, len(files))
	copy(sorted, files)

	sort.Slice(sorted, func(i, j int) bool {
		// Priority 1: Directories come last
		if sorted[i].IsDir && !sorted[j].IsDir {
			return false
		}
		if !sorted[i].IsDir && sorted[j].IsDir {
			return true
		}

		// Priority 2: Matched files come before unmatched files
		if sorted[i].Matched && !sorted[j].Matched {
			return true
		}
		if !sorted[i].Matched && sorted[j].Matched {
			return false
		}

		// Priority 3: Within same status group, sort alphabetically by path
		return sorted[i].Path < sorted[j].Path
	})

	return sorted
}

// FormatTaskStatus returns a human-readable status badge for a worker task.
// This is a pure function extracted from the TaskList component.
//
// Status format:
//   - worker.TaskStatusRunning -> "RUN"
//   - worker.TaskStatusSuccess -> "OK"
//   - worker.TaskStatusFailed -> "ERR"
//   - worker.TaskStatusPending -> "..."
//
// Parameters:
//   - status: worker.TaskStatus enum value
//
// Returns:
//   - formatted status badge string (without styling)
func FormatTaskStatus(status worker.TaskStatus) string {
	switch status {
	case worker.TaskStatusRunning:
		return "RUN"
	case worker.TaskStatusSuccess:
		return "OK"
	case worker.TaskStatusFailed:
		return "ERR"
	case worker.TaskStatusPending:
		return "..."
	default:
		return "..."
	}
}
