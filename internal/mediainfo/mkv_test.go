package mediainfo

import "testing"

func TestMapMKVVideoCodec(t *testing.T) {
	tests := []struct {
		name     string
		codecID  string
		expected string
	}{
		{"H.264 AVC", "V_MPEG4/ISO/AVC", "h264"},
		{"H.264 lowercase", "v_mpeg4/iso/avc", "h264"},
		{"H.264 H264", "V_H264", "h264"},
		{"H.265 HEVC", "V_MPEGH/ISO/HEVC", "hevc"},
		{"H.265 H265", "V_H265", "hevc"},
		{"VP9", "V_VP9", "vp9"},
		{"VP8", "V_VP8", "vp8"},
		{"AV1", "V_AV1", "av1"},
		{"MPEG4", "V_MPEG4/ISO/ASP", "mpeg4"},
		{"Theora", "V_THEORA", "theora"},
		{"Unknown with prefix", "V_UNKNOWN_CODEC", "UNKNOWN_CODEC"},
		{"Unknown without prefix", "UNKNOWN", "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapMKVVideoCodec(tt.codecID)
			if result != tt.expected {
				t.Errorf("mapMKVVideoCodec(%q) = %v, want %v", tt.codecID, result, tt.expected)
			}
		})
	}
}

func TestMapMKVAudioCodec(t *testing.T) {
	tests := []struct {
		name     string
		codecID  string
		expected string
	}{
		{"AAC", "A_AAC", "aac"},
		{"MP3", "A_MPEG/L3", "mp3"},
		{"MP3 direct", "A_MP3", "mp3"},
		{"AC3", "A_AC3", "ac3"},
		{"EAC3", "A_EAC3", "eac3"},
		{"EAC3 alt", "A_E-AC-3", "eac3"},
		{"DTS", "A_DTS", "dts"},
		{"Opus", "A_OPUS", "opus"},
		{"Vorbis", "A_VORBIS", "vorbis"},
		{"FLAC", "A_FLAC", "flac"},
		{"PCM", "A_PCM/INT/LIT", "pcm"},
		{"Unknown with prefix", "A_UNKNOWN_CODEC", "UNKNOWN_CODEC"},
		{"Unknown without prefix", "UNKNOWN", "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapMKVAudioCodec(tt.codecID)
			if result != tt.expected {
				t.Errorf("mapMKVAudioCodec(%q) = %v, want %v", tt.codecID, result, tt.expected)
			}
		})
	}
}
