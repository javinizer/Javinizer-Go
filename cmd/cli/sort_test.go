package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/database"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupSortTestDB creates a test database with proper structure for sort tests
func setupSortTestDB(t *testing.T) (configPath string, dbPath string) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath = filepath.Join(tmpDir, "data", "test.db")

	// Ensure database directory exists
	err := os.MkdirAll(filepath.Dir(dbPath), 0755)
	require.NoError(t, err)

	configPath, testCfg := createTestConfig(t,
		WithDatabaseDSN(dbPath),
		WithScraperPriority([]string{"r18dev", "dmm"}),
		WithDownloadCover(false),
	)

	// Initialize database with migrations
	db, err := database.New(testCfg)
	require.NoError(t, err)
	err = db.AutoMigrate()
	require.NoError(t, err)
	db.Close()

	return configPath, dbPath
}

// TestRunSort_Success tests successful file organization
func TestRunSort_Success(t *testing.T) {
	// Setup directories
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	destDir := filepath.Join(tmpDir, "dest")
	require.NoError(t, os.MkdirAll(sourceDir, 0755))
	require.NoError(t, os.MkdirAll(destDir, 0755))

	// Create test video file
	videoPath := createTestVideoFile(t, sourceDir, "IPX-535.mp4")

	// Setup config
	dbPath := filepath.Join(tmpDir, "test.db")
	configPath, testCfg := createTestConfig(t,
		WithDatabaseDSN(dbPath),
		WithScraperPriority([]string{"mock"}),
		WithOutputFolder("<ID> - <TITLE>"),
		WithOutputFile("<ID>"),
		WithNFOEnabled(false),
		WithDownloadCover(false),
	)
	testCfg.Output.MoveToFolder = true
	testCfg.Output.DownloadExtrafanart = false
	// Persist config mutations to file (runSort calls loadConfig which re-reads from file)
	require.NoError(t, config.Save(testCfg, configPath))

	// Pre-populate database with movie metadata
	dbConn, err := database.New(testCfg)
	require.NoError(t, err)
	err = dbConn.AutoMigrate()
	require.NoError(t, err)

	movie := createTestMovie("IPX-535", "Test Movie")
	movie.ContentID = "ipx00535"
	repo := database.NewMovieRepository(dbConn)
	err = repo.Upsert(movie)
	require.NoError(t, err)
	dbConn.Close()

	// Note: runSort creates its own scraper registry internally,
	// so we don't need to mock it for basic tests

	withTempConfigFile(t, configPath, func() {
		withTempConfig(t, testCfg, func() {
			deps := createTestDependencies(t, testCfg)
			defer deps.Close()

			stdout, _ := captureOutput(t, func() {
				cmd := &cobra.Command{}
				cmd.Flags().Bool("dry-run", false, "")
				cmd.Flags().Bool("recursive", true, "")
				cmd.Flags().String("dest", destDir, "")
				cmd.Flags().Bool("move", true, "")
				cmd.Flags().Bool("nfo", false, "")
				cmd.Flags().Bool("download", false, "")
				cmd.Flags().Bool("extrafanart", false, "")
				cmd.Flags().StringSlice("scrapers", nil, "")
				cmd.Flags().Bool("force-update", false, "")
				cmd.Flags().Bool("force-refresh", false, "")
				cmd.Flags().Bool("update", false, "")

				err := runSort(cmd, []string{sourceDir}, deps)
				require.NoError(t, err)
			})

			// Verify output shows success
			assert.Contains(t, stdout, "IPX-535")
			assert.Contains(t, stdout, "Found 1 video file(s)")
			assert.Contains(t, stdout, "Matched 1 file(s)")

			// Verify file was moved
			assertFileNotExists(t, videoPath)

			// Verify new location exists
			expectedPath := filepath.Join(destDir, "IPX-535 - Test Movie", "IPX-535.mp4")
			assertFileExists(t, expectedPath)
		})
	})
}

// TestRunSort_MultipleFiles tests organizing multiple files
func TestRunSort_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	destDir := filepath.Join(tmpDir, "dest")
	require.NoError(t, os.MkdirAll(sourceDir, 0755))
	require.NoError(t, os.MkdirAll(destDir, 0755))

	// Create multiple test video files
	createTestVideoFile(t, sourceDir, "IPX-535.mp4")
	createTestVideoFile(t, sourceDir, "ABC-123.mkv")

	dbPath := filepath.Join(tmpDir, "test.db")
	configPath, testCfg := createTestConfig(t,
		WithDatabaseDSN(dbPath),
		WithScraperPriority([]string{"mock"}),
		WithOutputFolder("<ID>"),
		WithOutputFile("<ID>"),
		WithNFOEnabled(false),
		WithDownloadCover(false),
	)
	testCfg.Output.MoveToFolder = true

	// Pre-populate database with both movies
	dbConn, err := database.New(testCfg)
	require.NoError(t, err)
	err = dbConn.AutoMigrate()
	require.NoError(t, err)

	repo := database.NewMovieRepository(dbConn)
	movie1 := createTestMovie("IPX-535", "Movie One")
	movie2 := createTestMovie("ABC-123", "Movie Two")
	require.NoError(t, repo.Upsert(movie1))
	require.NoError(t, repo.Upsert(movie2))
	dbConn.Close()

	withTempConfigFile(t, configPath, func() {
		withTempConfig(t, testCfg, func() {
			deps := createTestDependencies(t, testCfg)
			defer deps.Close()

			stdout, _ := captureOutput(t, func() {
				cmd := &cobra.Command{}
				cmd.Flags().Bool("dry-run", false, "")
				cmd.Flags().Bool("recursive", true, "")
				cmd.Flags().String("dest", destDir, "")
				cmd.Flags().Bool("move", true, "")
				cmd.Flags().Bool("nfo", false, "")
				cmd.Flags().Bool("download", false, "")
				cmd.Flags().Bool("extrafanart", false, "")
				cmd.Flags().StringSlice("scrapers", nil, "")
				cmd.Flags().Bool("force-update", false, "")
				cmd.Flags().Bool("force-refresh", false, "")
				cmd.Flags().Bool("update", false, "")

				err := runSort(cmd, []string{sourceDir}, deps)
				require.NoError(t, err)
			})

			assert.Contains(t, stdout, "Found 2 video file(s)")
			assert.Contains(t, stdout, "Matched 2 file(s)")

			// Verify both files organized
			assertFileExists(t, filepath.Join(destDir, "IPX-535", "IPX-535.mp4"))
			assertFileExists(t, filepath.Join(destDir, "ABC-123", "ABC-123.mkv"))
		})
	})
}

// TestRunSort_DryRun tests dry-run mode doesn't move files
func TestRunSort_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	destDir := filepath.Join(tmpDir, "dest")
	require.NoError(t, os.MkdirAll(sourceDir, 0755))
	require.NoError(t, os.MkdirAll(destDir, 0755))

	videoPath := createTestVideoFile(t, sourceDir, "IPX-535.mp4")

	dbPath := filepath.Join(tmpDir, "test.db")
	configPath, testCfg := createTestConfig(t,
		WithDatabaseDSN(dbPath),
		WithOutputFolder("<ID>"),
		WithOutputFile("<ID>"),
		WithNFOEnabled(false),
		WithDownloadCover(false),
	)
	testCfg.Output.MoveToFolder = true

	dbConn, err := database.New(testCfg)
	require.NoError(t, err)
	err = dbConn.AutoMigrate()
	require.NoError(t, err)

	repo := database.NewMovieRepository(dbConn)
	movie := createTestMovie("IPX-535", "Test Movie")
	require.NoError(t, repo.Upsert(movie))
	dbConn.Close()

	withTempConfigFile(t, configPath, func() {
		withTempConfig(t, testCfg, func() {
			deps := createTestDependencies(t, testCfg)
			defer deps.Close()

			stdout, _ := captureOutput(t, func() {
				cmd := &cobra.Command{}
				cmd.Flags().Bool("dry-run", false, "")
				cmd.Flags().Set("dry-run", "true")
				cmd.Flags().Bool("recursive", true, "")
				cmd.Flags().String("dest", destDir, "")
				cmd.Flags().Bool("move", false, "")
				cmd.Flags().Bool("nfo", false, "")
				cmd.Flags().Bool("download", false, "")
				cmd.Flags().Bool("extrafanart", false, "")
				cmd.Flags().StringSlice("scrapers", nil, "")
				cmd.Flags().Bool("force-update", false, "")
				cmd.Flags().Bool("force-refresh", false, "")
				cmd.Flags().Bool("update", false, "")

				err := runSort(cmd, []string{sourceDir}, deps)
				require.NoError(t, err)
			})

			// Verify dry-run output
			assert.Contains(t, stdout, "DRY RUN")
			assert.Contains(t, stdout, "Run without --dry-run to apply changes")
			assert.Contains(t, stdout, "Would organize")

			// Verify file was NOT moved
			assertFileExists(t, videoPath)

			// Verify destination doesn't exist
			expectedPath := filepath.Join(destDir, "IPX-535", "IPX-535.mp4")
			assertFileNotExists(t, expectedPath)
		})
	})
}

// TestRunSort_CopyMode tests copying files instead of moving
func TestRunSort_CopyMode(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	destDir := filepath.Join(tmpDir, "dest")
	require.NoError(t, os.MkdirAll(sourceDir, 0755))
	require.NoError(t, os.MkdirAll(destDir, 0755))

	videoPath := createTestVideoFile(t, sourceDir, "IPX-535.mp4")

	dbPath := filepath.Join(tmpDir, "test.db")
	configPath, testCfg := createTestConfig(t,
		WithDatabaseDSN(dbPath),
		WithOutputFolder("<ID>"),
		WithOutputFile("<ID>"),
		WithNFOEnabled(false),
		WithDownloadCover(false),
	)
	testCfg.Output.MoveToFolder = true

	dbConn, err := database.New(testCfg)
	require.NoError(t, err)
	err = dbConn.AutoMigrate()
	require.NoError(t, err)

	repo := database.NewMovieRepository(dbConn)
	movie := createTestMovie("IPX-535", "Test Movie")
	require.NoError(t, repo.Upsert(movie))
	dbConn.Close()

	withTempConfigFile(t, configPath, func() {
		withTempConfig(t, testCfg, func() {
			deps := createTestDependencies(t, testCfg)
			defer deps.Close()

			stdout, _ := captureOutput(t, func() {
				cmd := &cobra.Command{}
				cmd.Flags().Bool("dry-run", false, "")
				cmd.Flags().Bool("recursive", true, "")
				cmd.Flags().String("dest", destDir, "")
				cmd.Flags().Bool("move", false, "")
				cmd.Flags().Set("move", "false")
				cmd.Flags().Bool("nfo", false, "")
				cmd.Flags().Bool("download", false, "")
				cmd.Flags().Bool("extrafanart", false, "")
				cmd.Flags().StringSlice("scrapers", nil, "")
				cmd.Flags().Bool("force-update", false, "")
				cmd.Flags().Bool("force-refresh", false, "")
				cmd.Flags().Bool("update", false, "")

				err := runSort(cmd, []string{sourceDir}, deps)
				require.NoError(t, err)
			})

			assert.Contains(t, stdout, "COPY")
			assert.Contains(t, stdout, "Organized")

			// Verify original file still exists (copy mode)
			assertFileExists(t, videoPath)

			// Verify copy was created
			expectedPath := filepath.Join(destDir, "IPX-535", "IPX-535.mp4")
			assertFileExists(t, expectedPath)
		})
	})
}

// TestRunSort_WithNFO tests NFO file generation
func TestRunSort_WithNFO(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	destDir := filepath.Join(tmpDir, "dest")
	require.NoError(t, os.MkdirAll(sourceDir, 0755))
	require.NoError(t, os.MkdirAll(destDir, 0755))

	createTestVideoFile(t, sourceDir, "IPX-535.mp4")

	dbPath := filepath.Join(tmpDir, "test.db")
	configPath, testCfg := createTestConfig(t,
		WithDatabaseDSN(dbPath),
		WithOutputFolder("<ID>"),
		WithOutputFile("<ID>"),
		WithNFOEnabled(true),
		WithDownloadCover(false),
	)
	testCfg.Output.MoveToFolder = true
	testCfg.Metadata.NFO.Enabled = true
	testCfg.Metadata.NFO.PerFile = false

	dbConn, err := database.New(testCfg)
	require.NoError(t, err)
	err = dbConn.AutoMigrate()
	require.NoError(t, err)

	repo := database.NewMovieRepository(dbConn)
	movie := createTestMovie("IPX-535", "Test Movie")
	require.NoError(t, repo.Upsert(movie))
	dbConn.Close()

	withTempConfigFile(t, configPath, func() {
		withTempConfig(t, testCfg, func() {
			deps := createTestDependencies(t, testCfg)
			defer deps.Close()

			stdout, _ := captureOutput(t, func() {
				cmd := &cobra.Command{}
				cmd.Flags().Bool("dry-run", false, "")
				cmd.Flags().Bool("recursive", true, "")
				cmd.Flags().String("dest", destDir, "")
				cmd.Flags().Bool("move", true, "")
				cmd.Flags().Bool("nfo", true, "")
				cmd.Flags().Set("nfo", "true")
				cmd.Flags().Bool("download", false, "")
				cmd.Flags().Bool("extrafanart", false, "")
				cmd.Flags().StringSlice("scrapers", nil, "")
				cmd.Flags().Bool("force-update", false, "")
				cmd.Flags().Bool("force-refresh", false, "")
				cmd.Flags().Bool("update", false, "")

				err := runSort(cmd, []string{sourceDir}, deps)
				require.NoError(t, err)
			})

			assert.Contains(t, stdout, "Generating NFO files")
			assert.Contains(t, stdout, "IPX-535.nfo")

			// Verify NFO file was created
			nfoPath := filepath.Join(destDir, "IPX-535", "IPX-535.nfo")
			assertFileExists(t, nfoPath)

			// Verify NFO contains XML content
			content, err := os.ReadFile(nfoPath)
			require.NoError(t, err)
			assert.Contains(t, string(content), "<?xml")
			assert.Contains(t, string(content), "Test Movie")
		})
	})
}

// TestRunSort_NestedDirectories tests recursive directory scanning
func TestRunSort_NestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	subDir := filepath.Join(sourceDir, "subdir")
	destDir := filepath.Join(tmpDir, "dest")
	require.NoError(t, os.MkdirAll(subDir, 0755))
	require.NoError(t, os.MkdirAll(destDir, 0755))

	// Create files in nested structure
	createTestVideoFile(t, sourceDir, "IPX-535.mp4")
	createTestVideoFile(t, subDir, "ABC-123.mkv")

	dbPath := filepath.Join(tmpDir, "test.db")
	configPath, testCfg := createTestConfig(t,
		WithDatabaseDSN(dbPath),
		WithOutputFolder("<ID>"),
		WithOutputFile("<ID>"),
		WithNFOEnabled(false),
		WithDownloadCover(false),
	)
	testCfg.Output.MoveToFolder = true

	dbConn, err := database.New(testCfg)
	require.NoError(t, err)
	err = dbConn.AutoMigrate()
	require.NoError(t, err)

	repo := database.NewMovieRepository(dbConn)
	require.NoError(t, repo.Upsert(createTestMovie("IPX-535", "Movie One")))
	require.NoError(t, repo.Upsert(createTestMovie("ABC-123", "Movie Two")))
	dbConn.Close()

	withTempConfigFile(t, configPath, func() {
		withTempConfig(t, testCfg, func() {
			deps := createTestDependencies(t, testCfg)
			defer deps.Close()

			stdout, _ := captureOutput(t, func() {
				cmd := &cobra.Command{}
				cmd.Flags().Bool("dry-run", false, "")
				cmd.Flags().Bool("recursive", true, "")
				cmd.Flags().Set("recursive", "true")
				cmd.Flags().String("dest", destDir, "")
				cmd.Flags().Bool("move", true, "")
				cmd.Flags().Bool("nfo", false, "")
				cmd.Flags().Bool("download", false, "")
				cmd.Flags().Bool("extrafanart", false, "")
				cmd.Flags().StringSlice("scrapers", nil, "")
				cmd.Flags().Bool("force-update", false, "")
				cmd.Flags().Bool("force-refresh", false, "")
				cmd.Flags().Bool("update", false, "")

				err := runSort(cmd, []string{sourceDir}, deps)
				require.NoError(t, err)
			})

			// Should find files in nested directories
			assert.Contains(t, stdout, "Found 2 video file(s)")

			// Verify both files organized
			assertFileExists(t, filepath.Join(destDir, "IPX-535", "IPX-535.mp4"))
			assertFileExists(t, filepath.Join(destDir, "ABC-123", "ABC-123.mkv"))
		})
	})
}

// TestRunSort_TemplateFormatting tests various template formats
func TestRunSort_TemplateFormatting(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	destDir := filepath.Join(tmpDir, "dest")
	require.NoError(t, os.MkdirAll(sourceDir, 0755))
	require.NoError(t, os.MkdirAll(destDir, 0755))

	createTestVideoFile(t, sourceDir, "IPX-535.mp4")

	dbPath := filepath.Join(tmpDir, "test.db")
	configPath, testCfg := createTestConfig(t,
		WithDatabaseDSN(dbPath),
		WithOutputFolder("<ID> [<MAKER>] - <TITLE> (<YEAR>)"),
		WithOutputFile("<ID>"),
		WithNFOEnabled(false),
		WithDownloadCover(false),
	)
	testCfg.Output.MoveToFolder = true

	dbConn, err := database.New(testCfg)
	require.NoError(t, err)
	err = dbConn.AutoMigrate()
	require.NoError(t, err)

	repo := database.NewMovieRepository(dbConn)
	movie := createTestMovie("IPX-535", "Complex Template Test")
	releaseDate := time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC)
	movie.ReleaseDate = &releaseDate
	movie.Maker = "Premium Studio"
	require.NoError(t, repo.Upsert(movie))
	dbConn.Close()

	withTempConfigFile(t, configPath, func() {
		withTempConfig(t, testCfg, func() {
			deps := createTestDependencies(t, testCfg)
			defer deps.Close()

			stdout, _ := captureOutput(t, func() {
				cmd := &cobra.Command{}
				cmd.Flags().Bool("dry-run", false, "")
				cmd.Flags().Bool("recursive", true, "")
				cmd.Flags().String("dest", destDir, "")
				cmd.Flags().Bool("move", true, "")
				cmd.Flags().Bool("nfo", false, "")
				cmd.Flags().Bool("download", false, "")
				cmd.Flags().Bool("extrafanart", false, "")
				cmd.Flags().StringSlice("scrapers", nil, "")
				cmd.Flags().Bool("force-update", false, "")
				cmd.Flags().Bool("force-refresh", false, "")
				cmd.Flags().Bool("update", false, "")

				err := runSort(cmd, []string{sourceDir}, deps)
				require.NoError(t, err)
			})

			assert.Contains(t, stdout, "Organized")

			// Verify complex template was applied correctly
			expectedPath := filepath.Join(destDir, "IPX-535 [Premium Studio] - Complex Template Test (2023)", "IPX-535.mp4")
			assertFileExists(t, expectedPath)
		})
	})
}

// TestRunSort_NoFiles tests handling of empty directory
func TestRunSort_NoFiles(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	require.NoError(t, os.MkdirAll(sourceDir, 0755))

	dbPath := filepath.Join(tmpDir, "test.db")
	configPath, testCfg := createTestConfig(t,
		WithDatabaseDSN(dbPath),
		WithOutputFolder("<ID>"),
		WithOutputFile("<ID>"),
	)

	withTempConfigFile(t, configPath, func() {
		withTempConfig(t, testCfg, func() {
			deps := createTestDependencies(t, testCfg)
			defer deps.Close()

			stdout, _ := captureOutput(t, func() {
				cmd := &cobra.Command{}
				cmd.Flags().Bool("dry-run", false, "")
				cmd.Flags().Bool("recursive", true, "")
				cmd.Flags().String("dest", "", "")
				cmd.Flags().Bool("move", false, "")
				cmd.Flags().Bool("nfo", false, "")
				cmd.Flags().Bool("download", false, "")
				cmd.Flags().Bool("extrafanart", false, "")
				cmd.Flags().StringSlice("scrapers", nil, "")
				cmd.Flags().Bool("force-update", false, "")
				cmd.Flags().Bool("force-refresh", false, "")
				cmd.Flags().Bool("update", false, "")

				err := runSort(cmd, []string{sourceDir}, deps)
				require.NoError(t, err)
			})

			assert.Contains(t, stdout, "No files to process")
		})
	})
}

// TestRunSort_NoMatches tests handling files with no JAV IDs
func TestRunSort_NoMatches(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	require.NoError(t, os.MkdirAll(sourceDir, 0755))

	// Create file without JAV ID pattern
	createTestVideoFile(t, sourceDir, "random_movie.mp4")

	dbPath := filepath.Join(tmpDir, "test.db")
	configPath, testCfg := createTestConfig(t,
		WithDatabaseDSN(dbPath),
		WithOutputFolder("<ID>"),
		WithOutputFile("<ID>"),
	)

	withTempConfigFile(t, configPath, func() {
		withTempConfig(t, testCfg, func() {
			deps := createTestDependencies(t, testCfg)
			defer deps.Close()

			stdout, _ := captureOutput(t, func() {
				cmd := &cobra.Command{}
				cmd.Flags().Bool("dry-run", false, "")
				cmd.Flags().Bool("recursive", true, "")
				cmd.Flags().String("dest", "", "")
				cmd.Flags().Bool("move", false, "")
				cmd.Flags().Bool("nfo", false, "")
				cmd.Flags().Bool("download", false, "")
				cmd.Flags().Bool("extrafanart", false, "")
				cmd.Flags().StringSlice("scrapers", nil, "")
				cmd.Flags().Bool("force-update", false, "")
				cmd.Flags().Bool("force-refresh", false, "")
				cmd.Flags().Bool("update", false, "")

				err := runSort(cmd, []string{sourceDir}, deps)
				require.NoError(t, err)
			})

			assert.Contains(t, stdout, "Found 1 video file(s)")
			assert.Contains(t, stdout, "No JAV IDs found in filenames")
		})
	})
}

// TestRunSort_NoMetadata tests handling files not in database
func TestRunSort_NoMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	destDir := filepath.Join(tmpDir, "dest")
	require.NoError(t, os.MkdirAll(sourceDir, 0755))
	require.NoError(t, os.MkdirAll(destDir, 0755))

	createTestVideoFile(t, sourceDir, "IPX-535.mp4")

	dbPath := filepath.Join(tmpDir, "test.db")
	configPath, testCfg := createTestConfig(t,
		WithDatabaseDSN(dbPath),
		WithScraperPriority([]string{"mock"}),
		WithOutputFolder("<ID>"),
		WithOutputFile("<ID>"),
		WithNFOEnabled(false),
		WithDownloadCover(false),
	)
	testCfg.Output.MoveToFolder = true

	// Note: Real scrapers will fail for non-existent IDs naturally
	// We test that the command handles scraper failures gracefully

	withTempConfigFile(t, configPath, func() {
		withTempConfig(t, testCfg, func() {
			deps := createTestDependencies(t, testCfg)
			defer deps.Close()

			stdout, _ := captureOutput(t, func() {
				cmd := &cobra.Command{}
				cmd.Flags().Bool("dry-run", false, "")
				cmd.Flags().Bool("recursive", true, "")
				cmd.Flags().String("dest", destDir, "")
				cmd.Flags().Bool("move", false, "")
				cmd.Flags().Bool("nfo", false, "")
				cmd.Flags().Bool("download", false, "")
				cmd.Flags().Bool("extrafanart", false, "")
				cmd.Flags().StringSlice("scrapers", nil, "")
				cmd.Flags().Bool("force-update", false, "")
				cmd.Flags().Bool("force-refresh", false, "")
				cmd.Flags().Bool("update", false, "")

				err := runSort(cmd, []string{sourceDir}, deps)
				require.NoError(t, err)
			})

			// Should report no metadata found
			assert.Contains(t, stdout, "Scraping metadata")
			assert.Contains(t, stdout, "Failed: 1")
			assert.Contains(t, stdout, "No metadata found")
		})
	})
}
