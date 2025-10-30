package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/javinizer/javinizer-go/internal/matcher"
	"github.com/javinizer/javinizer-go/internal/models"
	"github.com/javinizer/javinizer-go/internal/scanner"
	"github.com/stretchr/testify/assert"
)

// Test helpers that would typically be in integration tests
// but are simple enough to unit test

func TestPrintMovie_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		movie *models.Movie
	}{
		{
			name: "empty movie",
			movie: &models.Movie{
				ID: "TEST-001",
			},
		},
		{
			name: "only ID and Title",
			movie: &models.Movie{
				ID:    "TEST-002",
				Title: "Test Title",
			},
		},
		{
			name: "with very long description",
			movie: &models.Movie{
				ID:          "TEST-003",
				Title:       "Test",
				Description: "This is a very long description that should be wrapped properly when displayed. " + "It contains multiple sentences and should handle wrapping gracefully. " + "The wrapping logic should handle this without issues and maintain readability. " + "Additional text to make this even longer and test the wrapping boundaries. " + "More text to ensure we're testing multi-line wrapping properly.",
			},
		},
		{
			name: "with same cover and poster URL",
			movie: &models.Movie{
				ID:        "TEST-004",
				CoverURL:  "https://example.com/cover.jpg",
				PosterURL: "https://example.com/cover.jpg", // Same as cover
			},
		},
		{
			name: "with no source name but has translations",
			movie: &models.Movie{
				ID:    "TEST-005",
				Title: "Test",
				Translations: []models.MovieTranslation{
					{
						Language:   "en",
						SourceName: "source1",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				printMovie(tt.movie, nil)
			})
		})
	}
}

func TestLoadConfig_EdgeCases(t *testing.T) {
	t.Run("with empty config file path", func(t *testing.T) {
		originalCfgFile := cfgFile
		tmpDir := t.TempDir()

		// Create a config file with minimal content
		configPath := filepath.Join(tmpDir, "minimal.yaml")
		minimalConfig := `
server:
  host: ""
  port: 0
database:
  dsn: ":memory:"
scrapers:
  priority: []
  proxy:
    enabled: false
  r18dev:
    enabled: false
  dmm:
    enabled: false
output:
  download_proxy:
    enabled: false
logging:
  level: "error"
  output: "stdout"
`
		err := os.WriteFile(configPath, []byte(minimalConfig), 0644)
		assert.NoError(t, err)

		cfgFile = configPath
		defer func() {
			cfgFile = originalCfgFile
			cfg = nil
		}()

		err = loadConfig()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
	})

	t.Run("with download proxy enabled but empty URL", func(t *testing.T) {
		originalCfgFile := cfgFile
		tmpDir := t.TempDir()

		configPath := filepath.Join(tmpDir, "proxy.yaml")
		proxyConfig := `
server:
  host: "localhost"
  port: 8080
database:
  dsn: "data/test.db"
scrapers:
  priority: ["r18dev"]
  proxy:
    enabled: false
  r18dev:
    enabled: true
  dmm:
    enabled: false
output:
  download_proxy:
    enabled: true
    url: ""
logging:
  level: "info"
  output: "stdout"
`
		err := os.WriteFile(configPath, []byte(proxyConfig), 0644)
		assert.NoError(t, err)

		cfgFile = configPath
		defer func() {
			cfgFile = originalCfgFile
			cfg = nil
		}()

		err = loadConfig()
		// Should succeed and disable proxy
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.False(t, cfg.Output.DownloadProxy.Enabled)
	})

	t.Run("with proxy URL containing credentials", func(t *testing.T) {
		originalCfgFile := cfgFile
		tmpDir := t.TempDir()

		configPath := filepath.Join(tmpDir, "proxy-creds.yaml")
		proxyConfig := `
server:
  host: "localhost"
  port: 8080
database:
  dsn: "data/test.db"
scrapers:
  priority: ["r18dev"]
  proxy:
    enabled: true
    url: "http://user:pass@proxy.example.com:8080"
  r18dev:
    enabled: true
  dmm:
    enabled: false
output:
  download_proxy:
    enabled: false
logging:
  level: "info"
  output: "stdout"
`
		err := os.WriteFile(configPath, []byte(proxyConfig), 0644)
		assert.NoError(t, err)

		cfgFile = configPath
		defer func() {
			cfgFile = originalCfgFile
			cfg = nil
		}()

		err = loadConfig()
		// Should succeed and sanitize URL in logs
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.True(t, cfg.Scrapers.Proxy.Enabled)
	})
}

func TestBuildFileTree_EdgeCases(t *testing.T) {
	t.Run("with absolute base path", func(t *testing.T) {
		basePath := "/absolute/path/to/videos"
		files := []scanner.FileInfo{
			{
				Path: "/absolute/path/to/videos/movie.mp4",
				Name: "movie.mp4",
				Size: 1024,
			},
		}
		matchMap := make(map[string]matcher.MatchResult)

		result := buildFileTree(basePath, files, matchMap)
		assert.NotEmpty(t, result)
	})

	t.Run("with very deep nesting", func(t *testing.T) {
		basePath := "/test"
		files := []scanner.FileInfo{
			{
				Path: "/test/a/b/c/d/e/f/movie.mp4",
				Name: "movie.mp4",
				Size: 1024,
			},
		}
		matchMap := make(map[string]matcher.MatchResult)

		result := buildFileTree(basePath, files, matchMap)
		assert.NotEmpty(t, result)

		// Should have multiple directory levels
		dirCount := 0
		for _, item := range result {
			if item.IsDir {
				dirCount++
			}
		}
		assert.Greater(t, dirCount, 3) // At least a few directory levels
	})

	t.Run("with mixed matched and unmatched files", func(t *testing.T) {
		basePath := "/test"
		files := []scanner.FileInfo{
			{
				Path: "/test/matched.mp4",
				Name: "matched.mp4",
				Size: 1024,
			},
			{
				Path: "/test/unmatched.mp4",
				Name: "unmatched.mp4",
				Size: 2048,
			},
		}
		matchMap := map[string]matcher.MatchResult{
			"/test/matched.mp4": {
				ID: "IPX-535",
			},
		}

		result := buildFileTree(basePath, files, matchMap)
		assert.NotEmpty(t, result)

		matchedCount := 0
		unmatchedCount := 0
		for _, item := range result {
			if !item.IsDir {
				if item.Matched {
					matchedCount++
				} else {
					unmatchedCount++
				}
			}
		}

		assert.Equal(t, 1, matchedCount)
		assert.Equal(t, 1, unmatchedCount)
	})
}

func TestPercentage_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		part     int64
		total    int64
		expected float64
	}{
		{
			name:     "very large numbers",
			part:     1000000,
			total:    10000000,
			expected: 10.0,
		},
		{
			name:     "part equals total",
			part:     50,
			total:    50,
			expected: 100.0,
		},
		{
			name:     "part greater than total (should still calculate)",
			part:     150,
			total:    100,
			expected: 150.0,
		},
		{
			name:     "both zero",
			part:     0,
			total:    0,
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := percentage(tt.part, tt.total)
			assert.Equal(t, tt.expected, result)
		})
	}
}
