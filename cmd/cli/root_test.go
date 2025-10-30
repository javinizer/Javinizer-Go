package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRootCommand_Execute tests the root command without actually running subcommands
func TestRootCommand_Execute(t *testing.T) {
	// Create a minimal root command like main() does
	rootCmd := &cobra.Command{
		Use:   "javinizer",
		Short: "Javinizer - JAV metadata scraper and organizer",
	}

	// Add persistent flags
	var testCfgFile string
	var testVerboseFlag bool
	rootCmd.PersistentFlags().StringVar(&testCfgFile, "config", "configs/config.yaml", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&testVerboseFlag, "verbose", "v", false, "enable debug logging")

	// Capture output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	// Test help command
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Javinizer")
	assert.Contains(t, output, "JAV metadata scraper")
}

// TestAllCommands_Help tests that help works for all commands
func TestAllCommands_Help(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "scrape help",
			args:     []string{"scrape", "--help"},
			expected: "Scrape metadata",
		},
		{
			name:     "info help",
			args:     []string{"info", "--help"},
			expected: "Show configuration",
		},
		{
			name:     "init help",
			args:     []string{"init", "--help"},
			expected: "Initialize configuration",
		},
		{
			name:     "sort help",
			args:     []string{"sort", "--help"},
			expected: "scrape",
		},
		{
			name:     "genre help",
			args:     []string{"genre", "--help"},
			expected: "Manage genre",
		},
		{
			name:     "tag help",
			args:     []string{"tag", "--help"},
			expected: "custom tags",
		},
		{
			name:     "history help",
			args:     []string{"history", "--help"},
			expected: "history",
		},
		{
			name:     "tui help",
			args:     []string{"tui", "--help"},
			expected: "interactive",
		},
		{
			name:     "api help",
			args:     []string{"api", "--help"},
			expected: "API server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd := buildTestRootCommand()

			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()
			assert.NoError(t, err)

			output := buf.String()
			assert.Contains(t, strings.ToLower(output), strings.ToLower(tt.expected))
		})
	}
}

// buildTestRootCommand builds a full command tree for testing
func buildTestRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "javinizer",
		Short: "Javinizer - JAV metadata scraper and organizer",
		Long:  `A metadata scraper and file organizer for Japanese Adult Videos (JAV)`,
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "configs/config.yaml", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "enable debug logging")

	// Scrape command
	scrapeCmd := &cobra.Command{
		Use:   "scrape [id]",
		Short: "Scrape metadata for a movie ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Stub for testing
		},
	}
	scrapeCmd.Flags().StringSliceVarP(&scrapersFlag, "scrapers", "s", nil, "Comma-separated list of scrapers")
	scrapeCmd.Flags().BoolP("force", "f", false, "Force refresh metadata")

	// Info command
	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "Show configuration and scraper information",
		Run: func(cmd *cobra.Command, args []string) {
			// Stub for testing
		},
	}

	// Init command
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration and database",
		Run: func(cmd *cobra.Command, args []string) {
			// Stub for testing
		},
	}

	// Sort command
	sortCmd := &cobra.Command{
		Use:   "sort [path]",
		Short: "Scan, scrape, and organize video files",
		Long:  `Scans a directory for video files, scrapes metadata, generates NFO files, downloads media, and organizes files`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Stub for testing
		},
	}
	sortCmd.Flags().BoolP("dry-run", "n", false, "Preview operations")
	sortCmd.Flags().BoolP("recursive", "r", true, "Scan recursively")
	sortCmd.Flags().StringP("dest", "d", "", "Destination directory")

	// Genre command
	genreCmd := &cobra.Command{
		Use:   "genre",
		Short: "Manage genre replacements",
		Long:  `Manage genre name replacements for customizing genre names from scrapers`,
	}

	genreAddCmd := &cobra.Command{
		Use:   "add <original> <replacement>",
		Short: "Add a genre replacement",
		Args:  cobra.ExactArgs(2),
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	genreListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all genre replacements",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	genreRemoveCmd := &cobra.Command{
		Use:   "remove <original>",
		Short: "Remove a genre replacement",
		Args:  cobra.ExactArgs(1),
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	genreCmd.AddCommand(genreAddCmd, genreListCmd, genreRemoveCmd)

	// Tag command
	tagCmd := &cobra.Command{
		Use:   "tag",
		Short: "Manage per-movie tags",
		Long:  `Manage custom tags for individual movies (stored in database, appear in NFO files)`,
	}

	tagAddCmd := &cobra.Command{
		Use:   "add <movie_id> <tag> [tag2]...",
		Short: "Add tag(s) to a movie",
		Args:  cobra.MinimumNArgs(2),
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	tagListCmd := &cobra.Command{
		Use:   "list [movie_id]",
		Short: "List tags for a movie or all tag mappings",
		Args:  cobra.MaximumNArgs(1),
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	tagRemoveCmd := &cobra.Command{
		Use:   "remove <movie_id> [tag]",
		Short: "Remove tag(s) from a movie",
		Args:  cobra.RangeArgs(1, 2),
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	tagSearchCmd := &cobra.Command{
		Use:   "search <tag>",
		Short: "Find all movies with a specific tag",
		Args:  cobra.ExactArgs(1),
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	tagAllTagsCmd := &cobra.Command{
		Use:   "tags",
		Short: "List all unique tags in database",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	tagCmd.AddCommand(tagAddCmd, tagListCmd, tagRemoveCmd, tagSearchCmd, tagAllTagsCmd)

	// History command
	historyCmd := &cobra.Command{
		Use:   "history",
		Short: "View operation history",
		Long:  `View and manage the history of scrape, organize, download, and NFO operations`,
	}

	historyListCmd := &cobra.Command{
		Use:   "list",
		Short: "List recent operations",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	historyListCmd.Flags().IntP("limit", "n", 20, "Number of records")
	historyListCmd.Flags().StringP("operation", "o", "", "Filter by operation")
	historyListCmd.Flags().StringP("status", "s", "", "Filter by status")

	historyStatsCmd := &cobra.Command{
		Use:   "stats",
		Short: "Show operation statistics",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	historyMovieCmd := &cobra.Command{
		Use:   "movie <id>",
		Short: "Show history for a specific movie",
		Args:  cobra.ExactArgs(1),
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	historyCleanCmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean up old history records",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	historyCleanCmd.Flags().IntP("days", "d", 30, "Delete records older than days")

	historyCmd.AddCommand(historyListCmd, historyStatsCmd, historyMovieCmd, historyCleanCmd)

	// TUI and API commands
	tuiCmd := createTUICommand()
	apiCmd := newAPICmd()

	rootCmd.AddCommand(scrapeCmd, infoCmd, initCmd, sortCmd, genreCmd, tagCmd, historyCmd, tuiCmd, apiCmd)

	return rootCmd
}

// TestLoadConfig_CompleteFlow tests the full config loading flow
func TestLoadConfig_CompleteFlow(t *testing.T) {
	originalCfgFile := cfgFile
	originalVerboseFlag := verboseFlag

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "complete.yaml")

	completeConfig := `
server:
  host: "0.0.0.0"
  port: 9090

database:
  type: "sqlite"
  dsn: "data/test-complete.db"

scrapers:
  user_agent: "TestAgent/1.0"
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

metadata:
  ignore_genres: ["genre1", "genre2"]
  required_fields: ["title", "id"]

output:
  folder_format: "<ID> - <TITLE>"
  file_format: "<ID>"
  move_to_folder: true
  download_cover: true
  download_extrafanart: true
  download_proxy:
    enabled: false

logging:
  level: "debug"
  format: "json"
  output: "stdout"

performance:
  max_workers: 3
  buffer_size: 50
`

	require.NoError(t, os.WriteFile(configPath, []byte(completeConfig), 0644))

	cfgFile = configPath
	verboseFlag = false

	defer func() {
		cfgFile = originalCfgFile
		verboseFlag = originalVerboseFlag
		cfg = nil
	}()

	err := loadConfig()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify all config sections loaded
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "data/test-complete.db", cfg.Database.DSN)
	assert.Equal(t, "TestAgent/1.0", cfg.Scrapers.UserAgent)
	assert.Contains(t, cfg.Scrapers.Priority, "r18dev")
	assert.Contains(t, cfg.Scrapers.Priority, "dmm")
	assert.True(t, cfg.Scrapers.R18Dev.Enabled)
	assert.True(t, cfg.Scrapers.DMM.Enabled)
	assert.True(t, cfg.Scrapers.DMM.ScrapeActress)
	assert.Contains(t, cfg.Metadata.IgnoreGenres, "genre1")
	assert.Contains(t, cfg.Metadata.RequiredFields, "title")
	assert.Equal(t, "<ID> - <TITLE>", cfg.Output.FolderFormat)
	assert.Equal(t, "<ID>", cfg.Output.FileFormat)
	assert.True(t, cfg.Output.MoveToFolder)
	assert.True(t, cfg.Output.DownloadCover)
	assert.True(t, cfg.Output.DownloadExtrafanart)
	assert.False(t, cfg.Output.DownloadProxy.Enabled)
	assert.Equal(t, 3, cfg.Performance.MaxWorkers)
	assert.Equal(t, 50, cfg.Performance.BufferSize)
}
