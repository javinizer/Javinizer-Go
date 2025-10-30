package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/javinizer/javinizer-go/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPercentage(t *testing.T) {
	tests := []struct {
		name     string
		part     int64
		total    int64
		expected float64
	}{
		{
			name:     "normal case 50%",
			part:     50,
			total:    100,
			expected: 50.0,
		},
		{
			name:     "normal case 25%",
			part:     25,
			total:    100,
			expected: 25.0,
		},
		{
			name:     "zero total returns zero",
			part:     10,
			total:    0,
			expected: 0.0,
		},
		{
			name:     "zero part returns zero",
			part:     0,
			total:    100,
			expected: 0.0,
		},
		{
			name:     "decimal result",
			part:     1,
			total:    3,
			expected: 33.33333333333333,
		},
		{
			name:     "100 percent",
			part:     100,
			total:    100,
			expected: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := percentage(tt.part, tt.total)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPrintMovie_BasicFields(t *testing.T) {
	// Create a test movie with basic fields
	releaseDate := time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC)
	movie := &models.Movie{
		ID:          "IPX-535",
		ContentID:   "ipx00535",
		Title:       "Test Movie Title",
		ReleaseDate: &releaseDate,
		Runtime:     120,
		Director:    "Test Director",
		Maker:       "Test Maker",
		Label:       "Test Label",
		Series:      "Test Series",
		RatingScore: 8.5,
		RatingVotes: 100,
		Description: "This is a test description",
	}

	// Capture stdout - printMovie writes to stdout
	// We just verify it doesn't panic
	assert.NotPanics(t, func() {
		printMovie(movie, nil)
	})
}

func TestPrintMovie_WithActresses(t *testing.T) {
	releaseDate := time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC)
	movie := &models.Movie{
		ID:          "IPX-535",
		Title:       "Test Movie",
		ReleaseDate: &releaseDate,
		Actresses: []models.Actress{
			{
				FirstName:    "Test",
				LastName:     "Actress",
				JapaneseName: "テスト女優",
			},
			{
				FirstName:    "Another",
				LastName:     "Actress",
				JapaneseName: "",
			},
		},
	}

	assert.NotPanics(t, func() {
		printMovie(movie, nil)
	})
}

func TestPrintMovie_WithGenres(t *testing.T) {
	releaseDate := time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC)
	movie := &models.Movie{
		ID:          "IPX-535",
		Title:       "Test Movie",
		ReleaseDate: &releaseDate,
		Genres: []models.Genre{
			{Name: "Drama"},
			{Name: "Romance"},
			{Name: "Action"},
		},
	}

	assert.NotPanics(t, func() {
		printMovie(movie, nil)
	})
}

func TestPrintMovie_WithTranslations(t *testing.T) {
	releaseDate := time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC)
	movie := &models.Movie{
		ID:          "IPX-535",
		Title:       "Test Movie",
		ReleaseDate: &releaseDate,
		Translations: []models.MovieTranslation{
			{
				Language:   "en",
				Title:      "English Title",
				SourceName: "r18dev",
			},
			{
				Language:   "ja",
				Title:      "Japanese Title",
				SourceName: "dmm",
			},
		},
		SourceName: "r18dev",
	}

	assert.NotPanics(t, func() {
		printMovie(movie, nil)
	})
}

func TestPrintMovie_WithMedia(t *testing.T) {
	releaseDate := time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC)
	movie := &models.Movie{
		ID:          "IPX-535",
		Title:       "Test Movie",
		ReleaseDate: &releaseDate,
		CoverURL:    "https://example.com/cover.jpg",
		PosterURL:   "https://example.com/poster.jpg",
		TrailerURL:  "https://example.com/trailer.mp4",
		Screenshots: []string{
			"https://example.com/screen1.jpg",
			"https://example.com/screen2.jpg",
		},
	}

	assert.NotPanics(t, func() {
		printMovie(movie, nil)
	})
}

func TestPrintMovie_WithScraperResults(t *testing.T) {
	releaseDate := time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC)
	movie := &models.Movie{
		ID:          "IPX-535",
		Title:       "Test Movie",
		ReleaseDate: &releaseDate,
	}

	results := []*models.ScraperResult{
		{
			Source:    "r18dev",
			SourceURL: "https://r18.dev/movies/IPX-535",
			Title:     "Test from R18",
		},
		{
			Source:    "dmm",
			SourceURL: "https://dmm.co.jp/digital/video/-/detail/=/cid=ipx00535",
			Title:     "Test from DMM",
		},
	}

	assert.NotPanics(t, func() {
		printMovie(movie, results)
	})
}

func TestPrintMovie_ManyActresses(t *testing.T) {
	releaseDate := time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC)
	movie := &models.Movie{
		ID:          "IPX-535",
		Title:       "Test Movie",
		ReleaseDate: &releaseDate,
		Actresses:   make([]models.Actress, 10),
	}

	// Fill with test actresses
	for i := 0; i < 10; i++ {
		movie.Actresses[i] = models.Actress{
			FirstName: "Actress",
			LastName:  string(rune('A' + i)),
		}
	}

	assert.NotPanics(t, func() {
		printMovie(movie, nil)
	})
}

func TestPrintMovie_ManyGenres(t *testing.T) {
	releaseDate := time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC)
	movie := &models.Movie{
		ID:          "IPX-535",
		Title:       "Test Movie",
		ReleaseDate: &releaseDate,
		Genres:      make([]models.Genre, 15),
	}

	// Fill with test genres
	for i := 0; i < 15; i++ {
		movie.Genres[i] = models.Genre{
			Name: "Genre" + string(rune('A'+i)),
		}
	}

	assert.NotPanics(t, func() {
		printMovie(movie, nil)
	})
}

func TestLoadConfig_InvalidFile(t *testing.T) {
	// Save original config file
	originalCfgFile := cfgFile

	// Create temp dir for test
	tmpDir := t.TempDir()

	// Set config file to non-existent path
	cfgFile = filepath.Join(tmpDir, "nonexistent.yaml")

	// Reset after test
	defer func() {
		cfgFile = originalCfgFile
		cfg = nil
	}()

	err := loadConfig()

	// loadConfig uses LoadOrCreate which creates a default config if missing
	// So we expect no error
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestLoadConfig_ValidFile(t *testing.T) {
	// Save original config file
	originalCfgFile := cfgFile

	// Create temp dir for test
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	validConfig := `
server:
  host: "localhost"
  port: 8080

database:
  type: "sqlite"
  dsn: "data/test.db"

scrapers:
  user_agent: "Javinizer/Test"
  priority:
    - "r18dev"
    - "dmm"
  proxy:
    enabled: false
  r18dev:
    enabled: true
  dmm:
    enabled: true
    scrape_actress: true

output:
  folder_format: "<ID> - <TITLE>"
  file_format: "<ID>"
  download_cover: true
  download_extrafanart: false
  download_proxy:
    enabled: false

logging:
  level: "info"
  format: "text"
  output: "stdout"
`

	require.NoError(t, os.WriteFile(configPath, []byte(validConfig), 0644))

	// Set config file
	cfgFile = configPath

	// Reset after test
	defer func() {
		cfgFile = originalCfgFile
		cfg = nil
	}()

	err := loadConfig()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "data/test.db", cfg.Database.DSN)
	assert.Contains(t, cfg.Scrapers.Priority, "r18dev")
	assert.Contains(t, cfg.Scrapers.Priority, "dmm")
}

func TestLoadConfig_WithProxyEnabled(t *testing.T) {
	originalCfgFile := cfgFile
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configWithProxy := `
server:
  host: "localhost"
  port: 8080

database:
  type: "sqlite"
  dsn: "data/test.db"

scrapers:
  priority:
    - "r18dev"
  proxy:
    enabled: true
    url: "http://proxy.example.com:8080"
    username: "user"
    password: "pass"
  r18dev:
    enabled: true
  dmm:
    enabled: false

output:
  download_proxy:
    enabled: false

logging:
  level: "info"
  format: "text"
  output: "stdout"
`

	require.NoError(t, os.WriteFile(configPath, []byte(configWithProxy), 0644))

	cfgFile = configPath

	defer func() {
		cfgFile = originalCfgFile
		cfg = nil
	}()

	err := loadConfig()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.True(t, cfg.Scrapers.Proxy.Enabled)
	assert.Equal(t, "http://proxy.example.com:8080", cfg.Scrapers.Proxy.URL)
}

func TestLoadConfig_ProxyEnabledButEmptyURL(t *testing.T) {
	originalCfgFile := cfgFile
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configWithBadProxy := `
server:
  host: "localhost"
  port: 8080

database:
  type: "sqlite"
  dsn: "data/test.db"

scrapers:
  priority:
    - "r18dev"
  proxy:
    enabled: true
    url: ""
  r18dev:
    enabled: true
  dmm:
    enabled: false

output:
  download_proxy:
    enabled: false

logging:
  level: "info"
  format: "text"
  output: "stdout"
`

	require.NoError(t, os.WriteFile(configPath, []byte(configWithBadProxy), 0644))

	cfgFile = configPath

	defer func() {
		cfgFile = originalCfgFile
		cfg = nil
	}()

	err := loadConfig()

	// loadConfig should disable proxy if URL is empty
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.False(t, cfg.Scrapers.Proxy.Enabled) // Should be disabled
}

func TestLoadConfig_WithVerboseFlag(t *testing.T) {
	originalCfgFile := cfgFile
	originalVerboseFlag := verboseFlag

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	validConfig := `
server:
  host: "localhost"
  port: 8080

database:
  type: "sqlite"
  dsn: "data/test.db"

scrapers:
  priority:
    - "r18dev"
  proxy:
    enabled: false
  r18dev:
    enabled: true
  dmm:
    enabled: false

output:
  download_proxy:
    enabled: false

logging:
  level: "info"
  format: "text"
  output: "stdout"
`

	require.NoError(t, os.WriteFile(configPath, []byte(validConfig), 0644))

	cfgFile = configPath
	verboseFlag = true

	defer func() {
		cfgFile = originalCfgFile
		verboseFlag = originalVerboseFlag
		cfg = nil
	}()

	err := loadConfig()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	// Logger should be initialized with debug level due to verbose flag
	// We can't easily verify this without exposing logger state, but we verify no error
}

func TestLoadConfig_MalformedYAML(t *testing.T) {
	originalCfgFile := cfgFile

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	malformedConfig := `
server:
  host: "localhost"
  port: "not-a-number"  # Invalid - port should be int
`

	require.NoError(t, os.WriteFile(configPath, []byte(malformedConfig), 0644))

	cfgFile = configPath

	defer func() {
		cfgFile = originalCfgFile
		cfg = nil
	}()

	err := loadConfig()

	// Should get error due to malformed YAML
	assert.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "config")
}
