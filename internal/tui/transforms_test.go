package tui

import (
	"testing"

	"github.com/javinizer/javinizer-go/internal/worker"
	"github.com/stretchr/testify/assert"
)

func TestFilterVideoFiles(t *testing.T) {
	tests := []struct {
		name       string
		files      []FileItem
		extensions []string
		want       []FileItem
	}{
		{
			name: "valid extensions mp4 and mkv",
			files: []FileItem{
				{Path: "movie.mp4", Name: "movie.mp4"},
				{Path: "video.mkv", Name: "video.mkv"},
				{Path: "subtitle.srt", Name: "subtitle.srt"},
			},
			extensions: []string{".mp4", ".mkv"},
			want: []FileItem{
				{Path: "movie.mp4", Name: "movie.mp4"},
				{Path: "video.mkv", Name: "video.mkv"},
			},
		},
		{
			name: "empty extensions list",
			files: []FileItem{
				{Path: "movie.mp4", Name: "movie.mp4"},
			},
			extensions: []string{},
			want:       []FileItem{},
		},
		{
			name: "no matches",
			files: []FileItem{
				{Path: "subtitle.srt", Name: "subtitle.srt"},
				{Path: "image.jpg", Name: "image.jpg"},
			},
			extensions: []string{".mp4", ".mkv"},
			want:       []FileItem{},
		},
		{
			name: "mixed file types",
			files: []FileItem{
				{Path: "movie.mp4", Name: "movie.mp4"},
				{Path: "subtitle.srt", Name: "subtitle.srt"},
				{Path: "video.avi", Name: "video.avi"},
				{Path: "image.jpg", Name: "image.jpg"},
			},
			extensions: []string{".mp4", ".avi"},
			want: []FileItem{
				{Path: "movie.mp4", Name: "movie.mp4"},
				{Path: "video.avi", Name: "video.avi"},
			},
		},
		{
			name:       "empty input files",
			files:      []FileItem{},
			extensions: []string{".mp4", ".mkv"},
			want:       []FileItem{},
		},
		{
			name: "skip directories",
			files: []FileItem{
				{Path: "/videos/", Name: "videos", IsDir: true},
				{Path: "/videos/movie.mp4", Name: "movie.mp4", IsDir: false},
			},
			extensions: []string{".mp4"},
			want: []FileItem{
				{Path: "/videos/movie.mp4", Name: "movie.mp4", IsDir: false},
			},
		},
		{
			name: "case insensitive matching",
			files: []FileItem{
				{Path: "MOVIE.MP4", Name: "MOVIE.MP4"},
				{Path: "video.MKV", Name: "video.MKV"},
			},
			extensions: []string{".mp4", ".mkv"},
			want: []FileItem{
				{Path: "MOVIE.MP4", Name: "MOVIE.MP4"},
				{Path: "video.MKV", Name: "video.MKV"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalFiles := make([]FileItem, len(tt.files))
			copy(originalFiles, tt.files)

			got := FilterVideoFiles(tt.files, tt.extensions)

			assert.Equal(t, tt.want, got)

			// Verify immutability - original files unchanged
			assert.Equal(t, originalFiles, tt.files, "FilterVideoFiles should not modify input files")
		})
	}
}

func TestFormatFileStatus(t *testing.T) {
	tests := []struct {
		name string
		file FileItem
		want string
	}{
		{
			name: "matched file with ID",
			file: FileItem{
				Path:    "IPX-123.mp4",
				Matched: true,
				ID:      "IPX-123",
			},
			want: "matched [IPX-123]",
		},
		{
			name: "unmatched file",
			file: FileItem{
				Path:    "random-file.mp4",
				Matched: false,
				ID:      "",
			},
			want: "unmatched",
		},
		{
			name: "matched true but no ID",
			file: FileItem{
				Path:    "file.mp4",
				Matched: true,
				ID:      "",
			},
			want: "unmatched",
		},
		{
			name: "matched false with ID present",
			file: FileItem{
				Path:    "file.mp4",
				Matched: false,
				ID:      "IPX-999",
			},
			want: "unmatched",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatFileStatus(tt.file)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSortFilesByStatus(t *testing.T) {
	tests := []struct {
		name  string
		files []FileItem
		want  []FileItem
	}{
		{
			name: "mixed statuses - matched before unmatched",
			files: []FileItem{
				{Path: "c-unmatched.mp4", Matched: false},
				{Path: "a-matched.mp4", Matched: true, ID: "IPX-123"},
				{Path: "b-unmatched.mp4", Matched: false},
			},
			want: []FileItem{
				{Path: "a-matched.mp4", Matched: true, ID: "IPX-123"},
				{Path: "b-unmatched.mp4", Matched: false},
				{Path: "c-unmatched.mp4", Matched: false},
			},
		},
		{
			name: "all same status - alphabetical sort",
			files: []FileItem{
				{Path: "zebra.mp4", Matched: true, ID: "IPX-999"},
				{Path: "apple.mp4", Matched: true, ID: "IPX-123"},
				{Path: "banana.mp4", Matched: true, ID: "IPX-456"},
			},
			want: []FileItem{
				{Path: "apple.mp4", Matched: true, ID: "IPX-123"},
				{Path: "banana.mp4", Matched: true, ID: "IPX-456"},
				{Path: "zebra.mp4", Matched: true, ID: "IPX-999"},
			},
		},
		{
			name:  "empty files list",
			files: []FileItem{},
			want:  []FileItem{},
		},
		{
			name: "single file",
			files: []FileItem{
				{Path: "single.mp4", Matched: true, ID: "IPX-123"},
			},
			want: []FileItem{
				{Path: "single.mp4", Matched: true, ID: "IPX-123"},
			},
		},
		{
			name: "directories come last",
			files: []FileItem{
				{Path: "/videos/", Name: "videos", IsDir: true},
				{Path: "/videos/movie.mp4", Name: "movie.mp4", Matched: true, ID: "IPX-123"},
				{Path: "/docs/", Name: "docs", IsDir: true},
			},
			want: []FileItem{
				{Path: "/videos/movie.mp4", Name: "movie.mp4", Matched: true, ID: "IPX-123"},
				{Path: "/docs/", Name: "docs", IsDir: true},
				{Path: "/videos/", Name: "videos", IsDir: true},
			},
		},
		{
			name: "full sort priority test",
			files: []FileItem{
				{Path: "unmatched-2.mp4", Matched: false, IsDir: false},
				{Path: "/folder-a/", IsDir: true},
				{Path: "matched-2.mp4", Matched: true, ID: "IPX-456", IsDir: false},
				{Path: "unmatched-1.mp4", Matched: false, IsDir: false},
				{Path: "matched-1.mp4", Matched: true, ID: "IPX-123", IsDir: false},
				{Path: "/folder-b/", IsDir: true},
			},
			want: []FileItem{
				{Path: "matched-1.mp4", Matched: true, ID: "IPX-123", IsDir: false},
				{Path: "matched-2.mp4", Matched: true, ID: "IPX-456", IsDir: false},
				{Path: "unmatched-1.mp4", Matched: false, IsDir: false},
				{Path: "unmatched-2.mp4", Matched: false, IsDir: false},
				{Path: "/folder-a/", IsDir: true},
				{Path: "/folder-b/", IsDir: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalFiles := make([]FileItem, len(tt.files))
			copy(originalFiles, tt.files)

			got := SortFilesByStatus(tt.files)

			assert.Equal(t, tt.want, got)

			// Verify immutability - original files unchanged
			assert.Equal(t, originalFiles, tt.files, "SortFilesByStatus should not modify input files")
		})
	}
}

func TestFormatTaskStatus(t *testing.T) {
	tests := []struct {
		name   string
		status worker.TaskStatus
		want   string
	}{
		{
			name:   "running status",
			status: worker.TaskStatusRunning,
			want:   "RUN",
		},
		{
			name:   "success status",
			status: worker.TaskStatusSuccess,
			want:   "OK",
		},
		{
			name:   "failed status",
			status: worker.TaskStatusFailed,
			want:   "ERR",
		},
		{
			name:   "pending status",
			status: worker.TaskStatusPending,
			want:   "...",
		},
		{
			name:   "unknown status defaults to pending",
			status: worker.TaskStatus("unknown"), // Invalid status
			want:   "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTaskStatus(tt.status)
			assert.Equal(t, tt.want, got)
		})
	}
}
