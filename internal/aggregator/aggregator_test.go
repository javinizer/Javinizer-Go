package aggregator

import (
	"testing"

	"github.com/javinizer/javinizer-go/internal/config"
)

func TestIsRegexPattern(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected bool
	}{
		{"Starts with caret", "^Featured", true},
		{"Ends with dollar", "mosaic$", true},
		{"Contains dot star", ".*mosaic.*", true},
		{"Contains dot plus", ".+test", true},
		{"Contains backslash", "\\d+", true},
		{"Contains brackets", "[0-9]", true},
		{"Contains parentheses", "(test)", true},
		{"Contains pipe", "test|demo", true},
		{"Contains question mark", "test?", true},
		{"Contains asterisk", "test*", true},
		{"Contains plus", "test+", true},
		{"Plain string", "Featured Actress", false},
		{"Plain string with space", "Big Tits", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRegexPattern(tt.pattern)
			if result != tt.expected {
				t.Errorf("isRegexPattern(%q) = %v, want %v", tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestCompileGenreRegexes(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		wantLen  int
	}{
		{
			name:     "Only regex patterns",
			patterns: []string{"^Featured", ".*mosaic.*", "test$"},
			wantLen:  3,
		},
		{
			name:     "Mixed regex and plain",
			patterns: []string{"^Featured", "Plain Text", ".*mosaic.*"},
			wantLen:  2,
		},
		{
			name:     "Only plain strings",
			patterns: []string{"Featured Actress", "Big Tits"},
			wantLen:  0,
		},
		{
			name:     "Empty list",
			patterns: []string{},
			wantLen:  0,
		},
		{
			name:     "Invalid regex",
			patterns: []string{"^(unclosed", "valid$"},
			wantLen:  1, // Only valid one compiles
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Metadata: config.MetadataConfig{
					IgnoreGenres: tt.patterns,
				},
			}

			agg := New(cfg)

			if len(agg.ignoreGenreRegexes) != tt.wantLen {
				t.Errorf("compileGenreRegexes() compiled %d patterns, want %d", 
					len(agg.ignoreGenreRegexes), tt.wantLen)
			}
		})
	}
}

func TestIsGenreIgnored(t *testing.T) {
	tests := []struct {
		name          string
		ignoreGenres  []string
		genreToTest   string
		shouldIgnore  bool
	}{
		{
			name:         "Exact match",
			ignoreGenres: []string{"Featured Actress", "Sample"},
			genreToTest:  "Featured Actress",
			shouldIgnore: true,
		},
		{
			name:         "Regex prefix match",
			ignoreGenres: []string{"^Featured"},
			genreToTest:  "Featured Actress",
			shouldIgnore: true,
		},
		{
			name:         "Regex suffix match with space",
			ignoreGenres: []string{"mosaic$"},
			genreToTest:  "HD mosaic",
			shouldIgnore: true, // Ends with "mosaic"
		},
		{
			name:         "Regex suffix match success",
			ignoreGenres: []string{"mosaic$"},
			genreToTest:  "mosaic",
			shouldIgnore: true,
		},
		{
			name:         "Regex contains match",
			ignoreGenres: []string{".*mosaic.*"},
			genreToTest:  "HD mosaic available",
			shouldIgnore: true,
		},
		{
			name:         "Multiple patterns first matches",
			ignoreGenres: []string{"^Featured", ".*mosaic.*", "Sample"},
			genreToTest:  "Featured Actress",
			shouldIgnore: true,
		},
		{
			name:         "Multiple patterns second matches",
			ignoreGenres: []string{"^Featured", ".*mosaic.*", "Sample"},
			genreToTest:  "HD mosaic",
			shouldIgnore: true,
		},
		{
			name:         "Multiple patterns third matches",
			ignoreGenres: []string{"^Featured", ".*mosaic.*", "Sample"},
			genreToTest:  "Sample",
			shouldIgnore: true,
		},
		{
			name:         "No match",
			ignoreGenres: []string{"^Featured", ".*mosaic.*"},
			genreToTest:  "Beautiful Girl",
			shouldIgnore: false,
		},
		{
			name:         "Case sensitive exact",
			ignoreGenres: []string{"Sample"},
			genreToTest:  "sample",
			shouldIgnore: false,
		},
		{
			name:         "Case sensitive regex",
			ignoreGenres: []string{"^featured"},
			genreToTest:  "Featured Actress",
			shouldIgnore: false,
		},
		{
			name:         "Case insensitive regex",
			ignoreGenres: []string{"(?i)^featured"},
			genreToTest:  "Featured Actress",
			shouldIgnore: true,
		},
		{
			name:         "Complex regex",
			ignoreGenres: []string{"^(HD|4K|VR)"},
			genreToTest:  "HD",
			shouldIgnore: true,
		},
		{
			name:         "Complex regex no match",
			ignoreGenres: []string{"^(HD|4K|VR)"},
			genreToTest:  "Beautiful Girl",
			shouldIgnore: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Metadata: config.MetadataConfig{
					IgnoreGenres: tt.ignoreGenres,
				},
			}

			agg := New(cfg)
			result := agg.isGenreIgnored(tt.genreToTest)

			if result != tt.shouldIgnore {
				t.Errorf("isGenreIgnored(%q) = %v, want %v", 
					tt.genreToTest, result, tt.shouldIgnore)
			}
		})
	}
}

func TestGenreFilteringIntegration(t *testing.T) {
	// This test verifies that regex patterns work end-to-end in genre filtering
	cfg := &config.Config{
		Scrapers: config.ScrapersConfig{
			Priority: []string{"test"},
		},
		Metadata: config.MetadataConfig{
			IgnoreGenres: []string{
				"^Featured",      // Regex: starts with "Featured"
				".*mosaic.*",     // Regex: contains "mosaic"
				"Sample",         // Exact: exactly "Sample"
				"^(HD|4K)",       // Regex: starts with HD or 4K
			},
		},
	}

	agg := New(cfg)

	// Verify regex compilation
	if len(agg.ignoreGenreRegexes) != 3 {
		t.Errorf("Expected 3 compiled regex patterns, got %d", len(agg.ignoreGenreRegexes))
	}

	// Test genres that should be filtered
	shouldFilter := []string{
		"Featured Actress",  // Matches ^Featured
		"HD mosaic",         // Matches .*mosaic.*
		"Sample",            // Exact match
		"HD",                // Matches ^(HD|4K)
		"4K",                // Matches ^(HD|4K)
		"mosaic version",    // Matches .*mosaic.*
	}

	for _, genre := range shouldFilter {
		if !agg.isGenreIgnored(genre) {
			t.Errorf("Genre %q should be filtered but wasn't", genre)
		}
	}

	// Test genres that should NOT be filtered
	shouldKeep := []string{
		"Beautiful Girl",
		"Blowjob",
		"Creampie",
		"featured actress", // Case sensitive
		"High Definition",  // Not "HD"
	}

	for _, genre := range shouldKeep {
		if agg.isGenreIgnored(genre) {
			t.Errorf("Genre %q should be kept but was filtered", genre)
		}
	}
}
