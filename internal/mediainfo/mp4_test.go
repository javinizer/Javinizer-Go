package mediainfo

import "testing"

func TestMapMP4VideoCodec(t *testing.T) {
	tests := []struct {
		name     string
		fourcc   string
		expected string
	}{
		{"H.264 avc1", "avc1", "h264"},
		{"H.264 avc3", "avc3", "h264"},
		{"H.265 hvc1", "hvc1", "hevc"},
		{"H.265 hev1", "hev1", "hevc"},
		{"VP9", "vp09", "vp9"},
		{"VP8", "vp08", "vp8"},
		{"AV1", "av01", "av1"},
		{"MPEG4", "mp4v", "mpeg4"},
		{"Unknown", "xxxx", "xxxx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapMP4VideoCodec(tt.fourcc)
			if result != tt.expected {
				t.Errorf("mapMP4VideoCodec(%q) = %v, want %v", tt.fourcc, result, tt.expected)
			}
		})
	}
}

func TestMapMP4AudioCodec(t *testing.T) {
	tests := []struct {
		name     string
		fourcc   string
		expected string
	}{
		{"AAC", "mp4a", "aac"},
		{"MP3 dotted", ".mp3", "mp3"},
		{"MP3 spaced", "mp3 ", "mp3"},
		{"AC3", "ac-3", "ac3"},
		{"EAC3", "ec-3", "eac3"},
		{"Opus", "opus", "opus"},
		{"FLAC", "fLaC", "flac"},
		{"Unknown", "xxxx", "xxxx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapMP4AudioCodec(tt.fourcc)
			if result != tt.expected {
				t.Errorf("mapMP4AudioCodec(%q) = %v, want %v", tt.fourcc, result, tt.expected)
			}
		})
	}
}
