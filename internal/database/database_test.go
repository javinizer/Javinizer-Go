package database

import (
	"testing"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/logger"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedLevel logger.LogLevel
	}{
		{"lowercase silent", "silent", logger.Silent},
		{"lowercase info", "info", logger.Info},
		{"lowercase warn", "warn", logger.Warn},
		{"lowercase error", "error", logger.Error},
		{"empty string defaults to silent", "", logger.Silent},
		{"mixed case Info", "Info", logger.Info},
		{"uppercase WARN", "WARN", logger.Warn},
		{"with leading whitespace", "  info", logger.Info},
		{"with trailing whitespace", "warn  ", logger.Warn},
		{"with both whitespace", "  error  ", logger.Error},
		{"invalid value defaults to silent", "invalid", logger.Silent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the actual parseLogLevel function used in production
			result := parseLogLevel(tt.input)

			if result != tt.expectedLevel {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.input, result, tt.expectedLevel)
			}
		})
	}
}

func TestDBClose(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "sqlite",
			DSN:  ":memory:",
		},
		Logging: config.LoggingConfig{
			Level: "error",
		},
	}

	db, err := New(cfg)
	require.NoError(t, err)

	// Close the database
	err = db.Close()
	require.NoError(t, err)

	// Try to use closed database (should work since it's just a connection close)
	// We can't easily test that it's actually closed without errors,
	// but the test validates that Close() doesn't panic
}

func TestDBAutoMigrate(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "sqlite",
			DSN:  ":memory:",
		},
		Logging: config.LoggingConfig{
			Level: "error",
		},
	}

	db, err := New(cfg)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Test AutoMigrate creates all tables
	err = db.AutoMigrate()
	require.NoError(t, err)

	// Verify tables exist by checking if we can create records
	repo := NewMovieRepository(db)
	movie := &models.Movie{
		ContentID: "test-migrate",
		ID:        "TEST-MIGRATE",
		Title:     "Migration Test",
	}
	err = repo.Create(movie)
	require.NoError(t, err)

	found, err := repo.FindByID("TEST-MIGRATE")
	require.NoError(t, err)
	assert.Equal(t, "Migration Test", found.Title)
}

func TestParseLogLevelEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"all lowercase info", "INFO"},
		{"with tabs", "\tinfo\t"},
		{"with newlines", "\nwarn\n"},
		{"mixed case mixed", "ErRoR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			assert.NotEqual(t, logger.Silent, result, "Should not default to silent for valid inputs")
		})
	}
}
