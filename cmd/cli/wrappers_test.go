package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/history"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Wrapper Function Tests
// ============================================================================

func TestRunWithDeps_Success(t *testing.T) {
	configPath, _ := setupGenreTestDB(t)

	withTempConfigFile(t, configPath, func() {
		// Load config first
		err := loadConfig()
		require.NoError(t, err)

		cmd := &cobra.Command{}
		callCount := 0

		// Test function that receives dependencies
		testFunc := func(cmd *cobra.Command, args []string, deps *Dependencies) error {
			callCount++
			// Verify deps are passed correctly
			assert.NotNil(t, deps)
			assert.NotNil(t, deps.Config)
			assert.NotNil(t, deps.DB)
			assert.NotNil(t, deps.ScraperRegistry)
			return nil
		}

		wrappedFunc := runWithDeps(testFunc)
		err = wrappedFunc(cmd, []string{})
		require.NoError(t, err)

		// Verify function was called
		assert.Equal(t, 1, callCount)
	})
}

func TestRunWithDeps_FunctionError(t *testing.T) {
	configPath, _ := setupGenreTestDB(t)

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		require.NoError(t, err)

		cmd := &cobra.Command{}

		// Test function that returns error
		testFunc := func(cmd *cobra.Command, args []string, deps *Dependencies) error {
			return fmt.Errorf("test error from function")
		}

		wrappedFunc := runWithDeps(testFunc)
		err = wrappedFunc(cmd, []string{})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "test error from function")
	})
}

func TestRunWithDeps_ConfigLoadError(t *testing.T) {
	// Set invalid config file path
	withTempConfigFile(t, "/nonexistent/path/config.yaml", func() {
		cmd := &cobra.Command{}

		testFunc := func(cmd *cobra.Command, args []string, deps *Dependencies) error {
			t.Fatal("Should not reach here")
			return nil
		}

		wrappedFunc := runWithDeps(testFunc)
		err := wrappedFunc(cmd, []string{})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load config")
	})
}

func TestRunWithDeps_DependencyInitError(t *testing.T) {
	// Create config with invalid database path (e.g., unwritable location)
	configPath, testCfg := createTestConfig(t, WithDatabaseDSN("/dev/null/invalid/test.db"))

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		require.NoError(t, err)

		// Override global config with invalid DB path
		cfg = testCfg

		cmd := &cobra.Command{}

		testFunc := func(cmd *cobra.Command, args []string, deps *Dependencies) error {
			t.Fatal("Should not reach here")
			return nil
		}

		wrappedFunc := runWithDeps(testFunc)
		err = wrappedFunc(cmd, []string{})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to initialize dependencies")
	})
}

func TestRunWithDeps_ArgsPassedThrough(t *testing.T) {
	configPath, _ := setupGenreTestDB(t)

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		require.NoError(t, err)

		cmd := &cobra.Command{}
		expectedArgs := []string{"arg1", "arg2", "arg3"}
		var receivedArgs []string

		testFunc := func(cmd *cobra.Command, args []string, deps *Dependencies) error {
			receivedArgs = args
			return nil
		}

		wrappedFunc := runWithDeps(testFunc)
		err = wrappedFunc(cmd, expectedArgs)
		require.NoError(t, err)

		assert.Equal(t, expectedArgs, receivedArgs)
	})
}

func TestRunWithConfig_Success(t *testing.T) {
	configPath, _ := setupGenreTestDB(t)

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		require.NoError(t, err)

		cmd := &cobra.Command{}
		callCount := 0

		testFunc := func(cmd *cobra.Command, args []string, cfg *config.Config) error {
			callCount++
			// Verify config is passed correctly
			assert.NotNil(t, cfg)
			assert.NotEmpty(t, cfg.Database.DSN)
			return nil
		}

		wrappedFunc := runWithConfig(testFunc)
		err = wrappedFunc(cmd, []string{})
		require.NoError(t, err)

		assert.Equal(t, 1, callCount)
	})
}

func TestRunWithConfig_FunctionError(t *testing.T) {
	configPath, _ := setupGenreTestDB(t)

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		require.NoError(t, err)

		cmd := &cobra.Command{}

		testFunc := func(cmd *cobra.Command, args []string, cfg *config.Config) error {
			return fmt.Errorf("config function error")
		}

		wrappedFunc := runWithConfig(testFunc)
		err = wrappedFunc(cmd, []string{})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "config function error")
	})
}

func TestRunWithConfig_ConfigLoadError(t *testing.T) {
	withTempConfigFile(t, "/nonexistent/path/config.yaml", func() {
		cmd := &cobra.Command{}

		testFunc := func(cmd *cobra.Command, args []string, cfg *config.Config) error {
			t.Fatal("Should not reach here")
			return nil
		}

		wrappedFunc := runWithConfig(testFunc)
		err := wrappedFunc(cmd, []string{})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load config")
	})
}

func TestRunWithConfig_ArgsPassedThrough(t *testing.T) {
	configPath, _ := setupGenreTestDB(t)

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		require.NoError(t, err)

		cmd := &cobra.Command{}
		expectedArgs := []string{"arg1", "arg2"}
		var receivedArgs []string

		testFunc := func(cmd *cobra.Command, args []string, cfg *config.Config) error {
			receivedArgs = args
			return nil
		}

		wrappedFunc := runWithConfig(testFunc)
		err = wrappedFunc(cmd, expectedArgs)
		require.NoError(t, err)

		assert.Equal(t, expectedArgs, receivedArgs)
	})
}

// ============================================================================
// History Command Tests
// ============================================================================

func TestRunHistoryList_Empty(t *testing.T) {
	configPath, _ := setupGenreTestDB(t)

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		require.NoError(t, err)

		deps := createTestDependencies(t, cfg)
		defer deps.Close()

		cmd := &cobra.Command{}
		cmd.Flags().Int("limit", 10, "")
		cmd.Flags().String("operation", "", "")
		cmd.Flags().String("status", "", "")

		stdout, _ := captureOutput(t, func() {
			err = runHistoryList(cmd, []string{}, deps)
			require.NoError(t, err)
		})

		// Verify empty message
		assert.Contains(t, stdout, "No history records found")
	})
}

func TestRunHistoryList_MultipleOperations(t *testing.T) {
	configPath, _ := setupGenreTestDB(t)

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		require.NoError(t, err)

		deps := createTestDependencies(t, cfg)
		defer deps.Close()

		// Add history records
		logger := history.NewLogger(deps.DB)
		require.NoError(t, logger.LogScrape("IPX-001", "http://example.com", nil, nil))
		require.NoError(t, logger.LogOrganize("IPX-001", "/src/file.mp4", "/dest/file.mp4", false, nil))
		require.NoError(t, logger.LogDownload("IPX-001", "http://example.com/cover.jpg", "/dest/cover.jpg", "cover", nil))
		require.NoError(t, logger.LogNFO("IPX-001", "/dest/IPX-001.nfo", nil))

		cmd := &cobra.Command{}
		cmd.Flags().Int("limit", 10, "")
		cmd.Flags().String("operation", "", "")
		cmd.Flags().String("status", "", "")

		stdout, _ := captureOutput(t, func() {
			err = runHistoryList(cmd, []string{}, deps)
			require.NoError(t, err)
		})

		// Verify output contains all operations
		assert.Contains(t, stdout, "IPX-001")
		assert.Contains(t, stdout, "scrape")
		assert.Contains(t, stdout, "organize")
		assert.Contains(t, stdout, "download")
		assert.Contains(t, stdout, "nfo")
		assert.Contains(t, stdout, "=== Operation History ===")
	})
}

func TestRunHistoryList_WithLimit(t *testing.T) {
	configPath, _ := setupGenreTestDB(t)

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		require.NoError(t, err)

		deps := createTestDependencies(t, cfg)
		defer deps.Close()

		// Add multiple history records
		logger := history.NewLogger(deps.DB)
		for i := 1; i <= 5; i++ {
			movieID := fmt.Sprintf("IPX-%03d", i)
			require.NoError(t, logger.LogScrape(movieID, "http://example.com", nil, nil))
		}

		cmd := &cobra.Command{}
		cmd.Flags().Int("limit", 3, "")
		cmd.Flags().String("operation", "", "")
		cmd.Flags().String("status", "", "")

		stdout, _ := captureOutput(t, func() {
			err = runHistoryList(cmd, []string{}, deps)
			require.NoError(t, err)
		})

		// Verify output contains records
		assert.Contains(t, stdout, "Showing 3 record(s)")
	})
}

func TestRunHistoryList_FilterByOperation(t *testing.T) {
	configPath, _ := setupGenreTestDB(t)

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		require.NoError(t, err)

		deps := createTestDependencies(t, cfg)
		defer deps.Close()

		// Add different operation types
		logger := history.NewLogger(deps.DB)
		require.NoError(t, logger.LogScrape("IPX-001", "http://example.com", nil, nil))
		require.NoError(t, logger.LogOrganize("IPX-002", "/src", "/dest", false, nil))
		require.NoError(t, logger.LogDownload("IPX-003", "http://example.com/cover.jpg", "/path", "cover", nil))

		cmd := &cobra.Command{}
		cmd.Flags().Int("limit", 10, "")
		cmd.Flags().String("operation", "scrape", "")
		cmd.Flags().String("status", "", "")

		stdout, _ := captureOutput(t, func() {
			err = runHistoryList(cmd, []string{}, deps)
			require.NoError(t, err)
		})

		// Verify only scrape operations shown
		assert.Contains(t, stdout, "IPX-001")
		assert.Contains(t, stdout, "scrape")
		// Other operations should not be in output
		assert.NotContains(t, stdout, "organize")
		assert.NotContains(t, stdout, "download")
	})
}

func TestRunHistoryList_FilterByStatus(t *testing.T) {
	configPath, _ := setupGenreTestDB(t)

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		require.NoError(t, err)

		deps := createTestDependencies(t, cfg)
		defer deps.Close()

		// Add records with different statuses
		logger := history.NewLogger(deps.DB)
		logger.LogScrape("IPX-001", "http://example.com", nil, nil)                         // success
		logger.LogScrape("IPX-002", "http://example.com", nil, fmt.Errorf("scrape failed")) // failed

		cmd := &cobra.Command{}
		cmd.Flags().Int("limit", 10, "")
		cmd.Flags().String("operation", "", "")
		cmd.Flags().String("status", "success", "")

		stdout, _ := captureOutput(t, func() {
			err = runHistoryList(cmd, []string{}, deps)
			require.NoError(t, err)
		})

		// Verify only success status shown
		assert.Contains(t, stdout, "IPX-001")
		assert.NotContains(t, stdout, "IPX-002")
	})
}

func TestRunHistoryList_DryRunFlag(t *testing.T) {
	configPath, _ := setupGenreTestDB(t)

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		require.NoError(t, err)

		deps := createTestDependencies(t, cfg)
		defer deps.Close()

		// Add dry-run record
		logger := history.NewLogger(deps.DB)
		logger.LogOrganize("IPX-001", "/src", "/dest", true, nil) // dry-run = true

		cmd := &cobra.Command{}
		cmd.Flags().Int("limit", 10, "")
		cmd.Flags().String("operation", "", "")
		cmd.Flags().String("status", "", "")

		stdout, _ := captureOutput(t, func() {
			err = runHistoryList(cmd, []string{}, deps)
			require.NoError(t, err)
		})

		// Verify dry-run indicator shown
		assert.Contains(t, stdout, "IPX-001")
		assert.Contains(t, stdout, "✓") // Dry run checkmark
	})
}

func TestRunHistoryStats_Empty(t *testing.T) {
	configPath, _ := setupGenreTestDB(t)

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		require.NoError(t, err)

		deps := createTestDependencies(t, cfg)
		defer deps.Close()

		cmd := &cobra.Command{}

		stdout, _ := captureOutput(t, func() {
			err = runHistoryStats(cmd, []string{}, deps)
			require.NoError(t, err)
		})

		// Verify stats output structure
		assert.Contains(t, stdout, "=== History Statistics ===")
		assert.Contains(t, stdout, "Total Operations: 0")
		assert.Contains(t, stdout, "By Status:")
		assert.Contains(t, stdout, "By Operation:")
	})
}

func TestRunHistoryStats_MultipleOperations(t *testing.T) {
	configPath, _ := setupGenreTestDB(t)

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		require.NoError(t, err)

		deps := createTestDependencies(t, cfg)
		defer deps.Close()

		// Add various operations
		logger := history.NewLogger(deps.DB)
		logger.LogScrape("IPX-001", "http://example.com", nil, nil)                          // success
		logger.LogScrape("IPX-002", "http://example.com", nil, nil)                          // success
		logger.LogScrape("IPX-003", "http://example.com", nil, fmt.Errorf("failed"))         // failed
		logger.LogOrganize("IPX-001", "/src", "/dest", false, nil)                           // success
		logger.LogDownload("IPX-001", "http://example.com/cover.jpg", "/path", "cover", nil) // success
		logger.LogNFO("IPX-001", "/path", nil)                                               // success

		cmd := &cobra.Command{}

		stdout, _ := captureOutput(t, func() {
			err = runHistoryStats(cmd, []string{}, deps)
			require.NoError(t, err)
		})

		// Verify stats output
		assert.Contains(t, stdout, "Total Operations: 6")
		assert.Contains(t, stdout, "Success:  5")
		assert.Contains(t, stdout, "Failed:   1")
		assert.Contains(t, stdout, "Scrape:   3")
		assert.Contains(t, stdout, "Organize: 1")
		assert.Contains(t, stdout, "Download: 1")
		assert.Contains(t, stdout, "NFO:      1")
	})
}

func TestRunHistoryStats_Percentages(t *testing.T) {
	configPath, _ := setupGenreTestDB(t)

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		require.NoError(t, err)

		deps := createTestDependencies(t, cfg)
		defer deps.Close()

		// Add operations with known distribution
		logger := history.NewLogger(deps.DB)
		for i := 0; i < 8; i++ {
			logger.LogScrape(fmt.Sprintf("IPX-%03d", i), "http://example.com", nil, nil) // 8 success
		}
		for i := 0; i < 2; i++ {
			logger.LogScrape(fmt.Sprintf("IPX-%03d", i+10), "http://example.com", nil, fmt.Errorf("failed")) // 2 failed
		}

		cmd := &cobra.Command{}

		stdout, _ := captureOutput(t, func() {
			err = runHistoryStats(cmd, []string{}, deps)
			require.NoError(t, err)
		})

		// Verify percentages are shown (8/10 = 80%, 2/10 = 20%)
		assert.Contains(t, stdout, "80.0%") // Success percentage
		assert.Contains(t, stdout, "20.0%") // Failed percentage
	})
}

func TestRunHistoryMovie_Success(t *testing.T) {
	configPath, _ := setupGenreTestDB(t)

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		require.NoError(t, err)

		deps := createTestDependencies(t, cfg)
		defer deps.Close()

		// Add history for specific movie
		logger := history.NewLogger(deps.DB)
		require.NoError(t, logger.LogScrape("IPX-001", "http://example.com", nil, nil))
		require.NoError(t, logger.LogOrganize("IPX-001", "/src", "/dest", false, nil))
		require.NoError(t, logger.LogDownload("IPX-001", "http://example.com/cover.jpg", "/path", "cover", nil))

		// Add history for different movie (should not appear)
		require.NoError(t, logger.LogScrape("IPX-002", "http://example.com", nil, nil))

		cmd := &cobra.Command{}

		stdout, _ := captureOutput(t, func() {
			err = runHistoryMovie(cmd, []string{"IPX-001"}, deps)
			require.NoError(t, err)
		})

		// Verify output contains operations for IPX-001 only
		assert.Contains(t, stdout, "=== History for IPX-001 ===")
		assert.Contains(t, stdout, "scrape")
		assert.Contains(t, stdout, "organize")
		assert.Contains(t, stdout, "download")
		assert.Contains(t, stdout, "Total: 3 operation(s)")
		assert.NotContains(t, stdout, "IPX-002")
	})
}

func TestRunHistoryMovie_NotFound(t *testing.T) {
	configPath, _ := setupGenreTestDB(t)

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		require.NoError(t, err)

		deps := createTestDependencies(t, cfg)
		defer deps.Close()

		// Add history for different movie
		logger := history.NewLogger(deps.DB)
		require.NoError(t, logger.LogScrape("IPX-001", "http://example.com", nil, nil))

		cmd := &cobra.Command{}

		stdout, _ := captureOutput(t, func() {
			err = runHistoryMovie(cmd, []string{"IPX-999"}, deps)
			require.NoError(t, err)
		})

		// Verify "not found" message
		assert.Contains(t, stdout, "No history found for movie: IPX-999")
	})
}

func TestRunHistoryMovie_WithPaths(t *testing.T) {
	configPath, _ := setupGenreTestDB(t)

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		require.NoError(t, err)

		deps := createTestDependencies(t, cfg)
		defer deps.Close()

		// Add history with paths
		logger := history.NewLogger(deps.DB)
		require.NoError(t, logger.LogOrganize("IPX-001", "/source/file.mp4", "/destination/file.mp4", false, nil))

		cmd := &cobra.Command{}

		stdout, _ := captureOutput(t, func() {
			err = runHistoryMovie(cmd, []string{"IPX-001"}, deps)
			require.NoError(t, err)
		})

		// Verify paths are shown
		assert.Contains(t, stdout, "From: /source/file.mp4")
		assert.Contains(t, stdout, "To:   /destination/file.mp4")
	})
}

func TestRunHistoryMovie_WithError(t *testing.T) {
	configPath, _ := setupGenreTestDB(t)

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		require.NoError(t, err)

		deps := createTestDependencies(t, cfg)
		defer deps.Close()

		// Add failed operation
		logger := history.NewLogger(deps.DB)
		logger.LogScrape("IPX-001", "http://example.com", nil, fmt.Errorf("network timeout"))

		cmd := &cobra.Command{}

		stdout, _ := captureOutput(t, func() {
			err = runHistoryMovie(cmd, []string{"IPX-001"}, deps)
			require.NoError(t, err)
		})

		// Verify error is shown
		assert.Contains(t, stdout, "Error: network timeout")
		assert.Contains(t, stdout, "❌") // Failed icon
	})
}

func TestRunHistoryClean_NoRecords(t *testing.T) {
	configPath, _ := setupGenreTestDB(t)

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		require.NoError(t, err)

		deps := createTestDependencies(t, cfg)
		defer deps.Close()

		cmd := &cobra.Command{}
		cmd.Flags().Int("days", 30, "")

		stdout, _ := captureOutput(t, func() {
			err = runHistoryClean(cmd, []string{}, deps)
			require.NoError(t, err)
		})

		// Verify no cleanup message
		assert.Contains(t, stdout, "No records older than 30 days found")
	})
}

func TestRunHistoryClean_WithRecords(t *testing.T) {
	configPath, _ := setupGenreTestDB(t)

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		require.NoError(t, err)

		deps := createTestDependencies(t, cfg)
		defer deps.Close()

		// Add some recent records
		logger := history.NewLogger(deps.DB)
		require.NoError(t, logger.LogScrape("IPX-001", "http://example.com", nil, nil))
		require.NoError(t, logger.LogScrape("IPX-002", "http://example.com", nil, nil))

		cmd := &cobra.Command{}
		cmd.Flags().Int("days", 30, "")

		// Verify the command runs without error
		// Note: Testing actual deletion behavior with aged records requires complex
		// database manipulation that doesn't work reliably in tests. The cleanup logic
		// itself is tested via the repository layer tests.
		captureOutput(t, func() {
			err = runHistoryClean(cmd, []string{}, deps)
			require.NoError(t, err)
		})
	})
}

func TestRunHistoryClean_CustomDays(t *testing.T) {
	configPath, _ := setupGenreTestDB(t)

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		require.NoError(t, err)

		deps := createTestDependencies(t, cfg)
		defer deps.Close()

		// Add a recent record
		logger := history.NewLogger(deps.DB)
		require.NoError(t, logger.LogScrape("IPX-001", "http://example.com", nil, nil))

		// Clean records older than 7 days (custom threshold)
		cmd := &cobra.Command{}
		cmd.Flags().Int("days", 7, "")

		stdout, _ := captureOutput(t, func() {
			err = runHistoryClean(cmd, []string{}, deps)
			require.NoError(t, err)
		})

		// Verify output mentions the custom 7-day threshold
		assert.Contains(t, stdout, "7 days")
	})
}

// ============================================================================
// NewDependencies Function Tests
// ============================================================================

func TestNewDependencies_Success(t *testing.T) {
	tmpDir := t.TempDir()
	_, testCfg := createTestConfig(t, WithDatabaseDSN(tmpDir+"/data/test.db"))

	// Ensure config is valid
	require.NotNil(t, testCfg)

	deps, err := NewDependencies(testCfg)
	require.NoError(t, err)
	defer deps.Close()

	// Verify dependencies are initialized
	assert.NotNil(t, deps.Config)
	assert.NotNil(t, deps.DB)
	assert.NotNil(t, deps.ScraperRegistry)

	// Verify database directory was created
	dbDir := tmpDir + "/data"
	_, statErr := os.Stat(dbDir)
	assert.NoError(t, statErr, "database directory should exist")

	// Verify scrapers are registered
	scrapers := deps.ScraperRegistry.GetAll()
	assert.GreaterOrEqual(t, len(scrapers), 2, "should have at least 2 scrapers (r18dev, dmm)")
}

func TestNewDependencies_NilConfig(t *testing.T) {
	deps, err := NewDependencies(nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "config cannot be nil")
	assert.Nil(t, deps)
}

func TestNewDependencies_DatabaseCreated(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/data/javinizer.db"
	_, testCfg := createTestConfig(t, WithDatabaseDSN(dbPath))

	deps, err := NewDependencies(testCfg)
	require.NoError(t, err)
	defer deps.Close()

	// Verify database file was created
	_, statErr := os.Stat(dbPath)
	assert.NoError(t, statErr, "database file should exist")
}

func TestNewDependencies_ScraperRegistryPopulated(t *testing.T) {
	tmpDir := t.TempDir()
	_, testCfg := createTestConfig(t, WithDatabaseDSN(tmpDir+"/test.db"))

	// Enable both scrapers in config
	testCfg.Scrapers.R18Dev.Enabled = true
	testCfg.Scrapers.DMM.Enabled = true

	deps, err := NewDependencies(testCfg)
	require.NoError(t, err)
	defer deps.Close()

	// Verify scrapers are registered
	scrapers := deps.ScraperRegistry.GetAll()
	assert.Len(t, scrapers, 2, "should have 2 scrapers")

	// Verify scraper names
	scraperNames := make(map[string]bool)
	for _, scraper := range scrapers {
		scraperNames[scraper.Name()] = true
	}
	assert.True(t, scraperNames["r18dev"], "r18dev scraper should be registered")
	assert.True(t, scraperNames["dmm"], "dmm scraper should be registered")
}

func TestNewDependencies_DatabaseMigrations(t *testing.T) {
	tmpDir := t.TempDir()
	_, testCfg := createTestConfig(t, WithDatabaseDSN(tmpDir+"/test.db"))

	deps, err := NewDependencies(testCfg)
	require.NoError(t, err)
	defer deps.Close()

	// Verify tables exist by querying them
	var count int64

	// Check movies table
	err = deps.DB.Table("movies").Count(&count).Error
	assert.NoError(t, err, "movies table should exist")

	// Check actresses table
	err = deps.DB.Table("actresses").Count(&count).Error
	assert.NoError(t, err, "actresses table should exist")

	// Check genres table
	err = deps.DB.Table("genres").Count(&count).Error
	assert.NoError(t, err, "genres table should exist")

	// Check genre_replacements table
	err = deps.DB.Table("genre_replacements").Count(&count).Error
	assert.NoError(t, err, "genre_replacements table should exist")

	// Check history table
	err = deps.DB.Table("history").Count(&count).Error
	assert.NoError(t, err, "history table should exist")
}

func TestNewDependencies_Close(t *testing.T) {
	tmpDir := t.TempDir()
	_, testCfg := createTestConfig(t, WithDatabaseDSN(tmpDir+"/test.db"))

	deps, err := NewDependencies(testCfg)
	require.NoError(t, err)

	// Close dependencies
	err = deps.Close()
	assert.NoError(t, err)

	// Verify database connection is closed (attempting query should fail)
	var count int64
	err = deps.DB.Table("movies").Count(&count).Error
	assert.Error(t, err, "database should be closed")
}

// ============================================================================
// Helper Function Tests
// ============================================================================

func TestPercentage_Normal(t *testing.T) {
	result := percentage(25, 100)
	assert.Equal(t, 25.0, result)
}

func TestPercentage_ZeroTotal(t *testing.T) {
	result := percentage(10, 0)
	assert.Equal(t, 0.0, result)
}

func TestPercentage_ZeroPart(t *testing.T) {
	result := percentage(0, 100)
	assert.Equal(t, 0.0, result)
}

func TestPercentage_Decimal(t *testing.T) {
	result := percentage(1, 3)
	assert.InDelta(t, 33.333333, result, 0.00001)
}
