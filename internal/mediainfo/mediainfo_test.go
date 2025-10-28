package mediainfo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectContainer(t *testing.T) {
	tests := []struct {
		name     string
		header   []byte
		expected string
	}{
		{
			name: "MP4 container",
			header: []byte{
				0x00, 0x00, 0x00, 0x20, 'f', 't', 'y', 'p',
				'i', 's', 'o', '5', 0x00, 0x00, 0x00, 0x00,
			},
			expected: "mp4",
		},
		{
			name: "MKV container",
			header: []byte{
				0x1A, 0x45, 0xDF, 0xA3, 0x01, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x1F, 0x42, 0x86, 0x81, 0x01,
			},
			expected: "mkv",
		},
		{
			name: "FLV container",
			header: []byte{
				'F', 'L', 'V', 0x01, 0x05, 0x00, 0x00, 0x00,
				0x09, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			expected: "flv",
		},
		{
			name: "AVI container",
			header: []byte{
				'R', 'I', 'F', 'F', 0x00, 0x00, 0x00, 0x00,
				'A', 'V', 'I', ' ', 'L', 'I', 'S', 'T',
			},
			expected: "avi",
		},
		{
			name: "Unknown container",
			header: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectContainer(tt.header)
			if result != tt.expected {
				t.Errorf("detectContainer() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestVideoInfo_GetResolution(t *testing.T) {
	tests := []struct {
		name     string
		height   int
		expected string
	}{
		{"4K", 2160, "4K"},
		{"1080p", 1080, "1080p"},
		{"720p", 720, "720p"},
		{"480p", 480, "480p"},
		{"SD", 360, "SD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &VideoInfo{Height: tt.height}
			if got := v.GetResolution(); got != tt.expected {
				t.Errorf("GetResolution() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestVideoInfo_GetAudioChannelDescription(t *testing.T) {
	tests := []struct {
		name     string
		channels int
		expected string
	}{
		{"Mono", 1, "Mono"},
		{"Stereo", 2, "Stereo"},
		{"5.1", 6, "5.1"},
		{"7.1", 8, "7.1"},
		{"Custom", 4, "4 channels"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &VideoInfo{AudioChannels: tt.channels}
			if got := v.GetAudioChannelDescription(); got != tt.expected {
				t.Errorf("GetAudioChannelDescription() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAnalyze_InvalidFile(t *testing.T) {
	_, err := Analyze("/nonexistent/file.mp4")
	if err == nil {
		t.Error("Analyze() should return error for nonexistent file")
	}
}

func TestAnalyze_TooSmallFile(t *testing.T) {
	// Create a temporary file that's too small to be a valid video
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "small.mp4")

	if err := os.WriteFile(tmpFile, []byte("too small"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := Analyze(tmpFile)
	if err == nil {
		t.Error("Analyze() should return error for too small file")
	}
}
