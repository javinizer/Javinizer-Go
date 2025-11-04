package imageutil

import (
	"testing"
)

func TestConstructAwsimgsrcPosterURL(t *testing.T) {
	tests := []struct {
		name        string
		coverURL    string
		expectedURL string
	}{
		{
			name:        "digital video format",
			coverURL:    "https://pics.dmm.co.jp/digital/video/sone00860/sone00860pl.jpg",
			expectedURL: "https://awsimgsrc.dmm.com/dig/video/sone00860/sone00860ps.jpg",
		},
		{
			name:        "mono movie format",
			coverURL:    "https://pics.dmm.co.jp/mono/movie/adult/118abw001/118abw001pl.jpg",
			expectedURL: "https://awsimgsrc.dmm.com/dig/mono/movie/118abw001/118abw001ps.jpg",
		},
		{
			name:        "awsimgsrc already - pl.jpg",
			coverURL:    "https://awsimgsrc.dmm.com/dig/video/ipx00535/ipx00535pl.jpg",
			expectedURL: "https://awsimgsrc.dmm.com/dig/video/ipx00535/ipx00535ps.jpg",
		},
		{
			name:        "awsimgsrc mono format - pl.jpg",
			coverURL:    "https://awsimgsrc.dmm.com/dig/mono/movie/mdb087/mdb087pl.jpg",
			expectedURL: "https://awsimgsrc.dmm.com/dig/mono/movie/mdb087/mdb087ps.jpg",
		},
		{
			name:        "empty URL",
			coverURL:    "",
			expectedURL: "",
		},
		{
			name:        "invalid URL format",
			coverURL:    "https://example.com/image.jpg",
			expectedURL: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := constructAwsimgsrcPosterURL(tt.coverURL)
			if result != tt.expectedURL {
				t.Errorf("constructAwsimgsrcPosterURL() = %v, want %v", result, tt.expectedURL)
			}
		})
	}
}

func TestGetOptimalPosterURL(t *testing.T) {
	tests := []struct {
		name            string
		coverURL        string
		expectedCrop    bool
		expectedContain string // Check if result contains this substring
	}{
		{
			name:            "empty cover URL",
			coverURL:        "",
			expectedCrop:    false, // Backend handles all cropping now
			expectedContain: "",
		},
		{
			name:            "invalid cover URL format",
			coverURL:        "https://example.com/image.jpg",
			expectedCrop:    false, // Backend handles all cropping now
			expectedContain: "example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posterURL, shouldCrop := GetOptimalPosterURL(tt.coverURL, nil)

			if shouldCrop != tt.expectedCrop {
				t.Errorf("GetOptimalPosterURL() shouldCrop = %v, want %v", shouldCrop, tt.expectedCrop)
			}

			if tt.expectedContain != "" && posterURL != tt.coverURL {
				t.Errorf("GetOptimalPosterURL() posterURL = %v, want %v", posterURL, tt.coverURL)
			}
		})
	}
}
