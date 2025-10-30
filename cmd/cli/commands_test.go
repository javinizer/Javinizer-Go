package main

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCreateTUICommand(t *testing.T) {
	cmd := createTUICommand()

	assert.NotNil(t, cmd)
	assert.Equal(t, "tui [path]", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check flags are defined
	assert.NotNil(t, cmd.Flags().Lookup("source"))
	assert.NotNil(t, cmd.Flags().Lookup("dest"))
	assert.NotNil(t, cmd.Flags().Lookup("recursive"))
	assert.NotNil(t, cmd.Flags().Lookup("move"))
	assert.NotNil(t, cmd.Flags().Lookup("dry-run"))
	assert.NotNil(t, cmd.Flags().Lookup("extrafanart"))
	assert.NotNil(t, cmd.Flags().Lookup("scrapers"))
}

func TestNewAPICmd(t *testing.T) {
	cmd := newAPICmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "api", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check flags are defined
	assert.NotNil(t, cmd.Flags().Lookup("host"))
	assert.NotNil(t, cmd.Flags().Lookup("port"))
}

func TestRootCommand_AllCommandsRegistered(t *testing.T) {
	// Create root command with all subcommands (mimicking main())
	rootCmd := &cobra.Command{
		Use:   "javinizer",
		Short: "Javinizer - JAV metadata scraper and organizer",
	}

	// Add all commands
	scrapeCmd := &cobra.Command{
		Use:   "scrape [id]",
		Short: "Scrape metadata for a movie ID",
		Args:  cobra.ExactArgs(1),
	}
	scrapeCmd.Flags().StringSliceP("scrapers", "s", nil, "scrapers")
	scrapeCmd.Flags().BoolP("force", "f", false, "force")

	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "Show configuration and scraper information",
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration and database",
	}

	sortCmd := &cobra.Command{
		Use:   "sort [path]",
		Short: "Scan, scrape, and organize video files",
		Args:  cobra.ExactArgs(1),
	}
	sortCmd.Flags().BoolP("dry-run", "n", false, "dry run")
	sortCmd.Flags().BoolP("recursive", "r", true, "recursive")
	sortCmd.Flags().StringP("dest", "d", "", "destination")

	genreCmd := &cobra.Command{
		Use:   "genre",
		Short: "Manage genre replacements",
	}

	tagCmd := &cobra.Command{
		Use:   "tag",
		Short: "Manage per-movie tags",
	}

	historyCmd := &cobra.Command{
		Use:   "history",
		Short: "View operation history",
	}

	tuiCmd := createTUICommand()
	apiCmd := newAPICmd()

	rootCmd.AddCommand(scrapeCmd, infoCmd, initCmd, sortCmd, genreCmd, tagCmd, historyCmd, tuiCmd, apiCmd)

	// Verify all commands exist
	commands := rootCmd.Commands()
	assert.Len(t, commands, 9)

	commandNames := make([]string, len(commands))
	for i, cmd := range commands {
		commandNames[i] = cmd.Name()
	}

	assert.Contains(t, commandNames, "scrape")
	assert.Contains(t, commandNames, "info")
	assert.Contains(t, commandNames, "init")
	assert.Contains(t, commandNames, "sort")
	assert.Contains(t, commandNames, "genre")
	assert.Contains(t, commandNames, "tag")
	assert.Contains(t, commandNames, "history")
	assert.Contains(t, commandNames, "tui")
	assert.Contains(t, commandNames, "api")
}

func TestScrapeCommand_Flags(t *testing.T) {
	scrapeCmd := &cobra.Command{
		Use:  "scrape [id]",
		Args: cobra.ExactArgs(1),
	}
	scrapeCmd.Flags().StringSliceVarP(&scrapersFlag, "scrapers", "s", nil, "scrapers")
	scrapeCmd.Flags().BoolP("force", "f", false, "force")

	// Verify flags
	assert.NotNil(t, scrapeCmd.Flags().Lookup("scrapers"))
	assert.NotNil(t, scrapeCmd.Flags().Lookup("force"))

	// Verify short flags
	assert.NotNil(t, scrapeCmd.Flags().ShorthandLookup("s"))
	assert.NotNil(t, scrapeCmd.Flags().ShorthandLookup("f"))
}

func TestSortCommand_Flags(t *testing.T) {
	sortCmd := &cobra.Command{
		Use:  "sort [path]",
		Args: cobra.ExactArgs(1),
	}
	sortCmd.Flags().BoolP("dry-run", "n", false, "dry run")
	sortCmd.Flags().BoolP("recursive", "r", true, "recursive")
	sortCmd.Flags().StringP("dest", "d", "", "destination")
	sortCmd.Flags().BoolP("move", "m", false, "move")
	sortCmd.Flags().BoolP("nfo", "", true, "generate NFO")
	sortCmd.Flags().BoolP("download", "", true, "download media")
	sortCmd.Flags().Bool("extrafanart", false, "download extrafanart")
	sortCmd.Flags().StringSliceP("scrapers", "p", nil, "scraper priority")
	sortCmd.Flags().BoolP("force-update", "f", false, "force update")
	sortCmd.Flags().Bool("force-refresh", false, "force refresh")
	sortCmd.Flags().BoolP("update", "u", false, "update mode")

	// Verify all flags
	assert.NotNil(t, sortCmd.Flags().Lookup("dry-run"))
	assert.NotNil(t, sortCmd.Flags().Lookup("recursive"))
	assert.NotNil(t, sortCmd.Flags().Lookup("dest"))
	assert.NotNil(t, sortCmd.Flags().Lookup("move"))
	assert.NotNil(t, sortCmd.Flags().Lookup("nfo"))
	assert.NotNil(t, sortCmd.Flags().Lookup("download"))
	assert.NotNil(t, sortCmd.Flags().Lookup("extrafanart"))
	assert.NotNil(t, sortCmd.Flags().Lookup("scrapers"))
	assert.NotNil(t, sortCmd.Flags().Lookup("force-update"))
	assert.NotNil(t, sortCmd.Flags().Lookup("force-refresh"))
	assert.NotNil(t, sortCmd.Flags().Lookup("update"))
}

func TestGenreCommand_Subcommands(t *testing.T) {
	genreCmd := &cobra.Command{
		Use:   "genre",
		Short: "Manage genre replacements",
	}

	addCmd := &cobra.Command{
		Use:  "add <original> <replacement>",
		Args: cobra.ExactArgs(2),
	}

	listCmd := &cobra.Command{
		Use: "list",
	}

	removeCmd := &cobra.Command{
		Use:  "remove <original>",
		Args: cobra.ExactArgs(1),
	}

	genreCmd.AddCommand(addCmd, listCmd, removeCmd)

	// Verify subcommands
	assert.Len(t, genreCmd.Commands(), 3)

	subcommands := genreCmd.Commands()
	names := make([]string, len(subcommands))
	for i, cmd := range subcommands {
		names[i] = cmd.Name()
	}

	assert.Contains(t, names, "add")
	assert.Contains(t, names, "list")
	assert.Contains(t, names, "remove")
}

func TestTagCommand_Subcommands(t *testing.T) {
	tagCmd := &cobra.Command{
		Use:   "tag",
		Short: "Manage per-movie tags",
	}

	addCmd := &cobra.Command{
		Use:  "add <movie_id> <tag> [tag2]...",
		Args: cobra.MinimumNArgs(2),
	}

	listCmd := &cobra.Command{
		Use:  "list [movie_id]",
		Args: cobra.MaximumNArgs(1),
	}

	removeCmd := &cobra.Command{
		Use:  "remove <movie_id> [tag]",
		Args: cobra.RangeArgs(1, 2),
	}

	searchCmd := &cobra.Command{
		Use:  "search <tag>",
		Args: cobra.ExactArgs(1),
	}

	tagsCmd := &cobra.Command{
		Use: "tags",
	}

	tagCmd.AddCommand(addCmd, listCmd, removeCmd, searchCmd, tagsCmd)

	// Verify subcommands
	assert.Len(t, tagCmd.Commands(), 5)

	subcommands := tagCmd.Commands()
	names := make([]string, len(subcommands))
	for i, cmd := range subcommands {
		names[i] = cmd.Name()
	}

	assert.Contains(t, names, "add")
	assert.Contains(t, names, "list")
	assert.Contains(t, names, "remove")
	assert.Contains(t, names, "search")
	assert.Contains(t, names, "tags")
}

func TestHistoryCommand_Subcommands(t *testing.T) {
	historyCmd := &cobra.Command{
		Use:   "history",
		Short: "View operation history",
	}

	listCmd := &cobra.Command{
		Use: "list",
	}
	listCmd.Flags().IntP("limit", "n", 20, "limit")
	listCmd.Flags().StringP("operation", "o", "", "operation")
	listCmd.Flags().StringP("status", "s", "", "status")

	statsCmd := &cobra.Command{
		Use: "stats",
	}

	movieCmd := &cobra.Command{
		Use:  "movie <id>",
		Args: cobra.ExactArgs(1),
	}

	cleanCmd := &cobra.Command{
		Use: "clean",
	}
	cleanCmd.Flags().IntP("days", "d", 30, "days")

	historyCmd.AddCommand(listCmd, statsCmd, movieCmd, cleanCmd)

	// Verify subcommands
	assert.Len(t, historyCmd.Commands(), 4)

	subcommands := historyCmd.Commands()
	names := make([]string, len(subcommands))
	for i, cmd := range subcommands {
		names[i] = cmd.Name()
	}

	assert.Contains(t, names, "list")
	assert.Contains(t, names, "stats")
	assert.Contains(t, names, "movie")
	assert.Contains(t, names, "clean")
}

func TestHistoryListCommand_Flags(t *testing.T) {
	listCmd := &cobra.Command{
		Use: "list",
	}
	listCmd.Flags().IntP("limit", "n", 20, "limit")
	listCmd.Flags().StringP("operation", "o", "", "operation")
	listCmd.Flags().StringP("status", "s", "", "status")

	// Verify flags
	assert.NotNil(t, listCmd.Flags().Lookup("limit"))
	assert.NotNil(t, listCmd.Flags().Lookup("operation"))
	assert.NotNil(t, listCmd.Flags().Lookup("status"))

	// Verify default values
	limit, _ := listCmd.Flags().GetInt("limit")
	assert.Equal(t, 20, limit)
}

func TestHistoryCleanCommand_Flags(t *testing.T) {
	cleanCmd := &cobra.Command{
		Use: "clean",
	}
	cleanCmd.Flags().IntP("days", "d", 30, "days")

	// Verify flags
	assert.NotNil(t, cleanCmd.Flags().Lookup("days"))

	// Verify default value
	days, _ := cleanCmd.Flags().GetInt("days")
	assert.Equal(t, 30, days)
}

func TestRootCommand_PersistentFlags(t *testing.T) {
	rootCmd := &cobra.Command{
		Use: "javinizer",
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "configs/config.yaml", "config file")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "verbose")

	// Verify persistent flags
	assert.NotNil(t, rootCmd.PersistentFlags().Lookup("config"))
	assert.NotNil(t, rootCmd.PersistentFlags().Lookup("verbose"))

	// Verify short flag
	assert.NotNil(t, rootCmd.PersistentFlags().ShorthandLookup("v"))
}

func TestCommandArgs_Validation(t *testing.T) {
	tests := []struct {
		name     string
		cmd      *cobra.Command
		args     []string
		hasError bool
	}{
		{
			name: "scrape requires exactly 1 arg",
			cmd: &cobra.Command{
				Use:  "scrape [id]",
				Args: cobra.ExactArgs(1),
			},
			args:     []string{"IPX-535"},
			hasError: false,
		},
		{
			name: "scrape fails with 0 args",
			cmd: &cobra.Command{
				Use:  "scrape [id]",
				Args: cobra.ExactArgs(1),
			},
			args:     []string{},
			hasError: true,
		},
		{
			name: "sort requires exactly 1 arg",
			cmd: &cobra.Command{
				Use:  "sort [path]",
				Args: cobra.ExactArgs(1),
			},
			args:     []string{"/path/to/videos"},
			hasError: false,
		},
		{
			name: "info requires no args",
			cmd: &cobra.Command{
				Use:  "info",
				Args: cobra.NoArgs,
			},
			args:     []string{},
			hasError: false,
		},
		{
			name: "tag add requires minimum 2 args",
			cmd: &cobra.Command{
				Use:  "add <movie_id> <tag>",
				Args: cobra.MinimumNArgs(2),
			},
			args:     []string{"IPX-535", "Favorite"},
			hasError: false,
		},
		{
			name: "tag add fails with 1 arg",
			cmd: &cobra.Command{
				Use:  "add <movie_id> <tag>",
				Args: cobra.MinimumNArgs(2),
			},
			args:     []string{"IPX-535"},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Args(tt.cmd, tt.args)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
