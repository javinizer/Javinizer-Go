package tui

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/javinizer/javinizer-go/internal/aggregator"
	"github.com/javinizer/javinizer-go/internal/database"
	"github.com/javinizer/javinizer-go/internal/downloader"
	"github.com/javinizer/javinizer-go/internal/matcher"
	"github.com/javinizer/javinizer-go/internal/models"
	"github.com/javinizer/javinizer-go/internal/nfo"
	"github.com/javinizer/javinizer-go/internal/organizer"
	"github.com/javinizer/javinizer-go/internal/worker"
)

// ProcessingCoordinator coordinates task execution for the TUI
type ProcessingCoordinator struct {
	pool            *worker.Pool
	progressTracker *worker.ProgressTracker
	movieRepo       *database.MovieRepository
	registry        *models.ScraperRegistry
	aggregator      *aggregator.Aggregator
	downloader      *downloader.Downloader
	organizer       *organizer.Organizer
	nfoGenerator    *nfo.Generator
	destPath        string
	moveFiles       bool
	scrapeEnabled   bool
	downloadEnabled bool
	organizeEnabled bool
	nfoEnabled      bool
}

// NewProcessingCoordinator creates a new processing coordinator
func NewProcessingCoordinator(
	pool *worker.Pool,
	progressTracker *worker.ProgressTracker,
	movieRepo *database.MovieRepository,
	registry *models.ScraperRegistry,
	agg *aggregator.Aggregator,
	dl *downloader.Downloader,
	org *organizer.Organizer,
	nfoGen *nfo.Generator,
	destPath string,
	moveFiles bool,
) *ProcessingCoordinator {
	return &ProcessingCoordinator{
		pool:            pool,
		progressTracker: progressTracker,
		movieRepo:       movieRepo,
		registry:        registry,
		aggregator:      agg,
		downloader:      dl,
		organizer:       org,
		nfoGenerator:    nfoGen,
		destPath:        destPath,
		moveFiles:       moveFiles,
		scrapeEnabled:   true,
		downloadEnabled: true,
		organizeEnabled: true,
		nfoEnabled:      true,
	}
}

// SetOptions configures which operations to perform
func (pc *ProcessingCoordinator) SetOptions(scrape, download, organize, nfo bool) {
	pc.scrapeEnabled = scrape
	pc.downloadEnabled = download
	pc.organizeEnabled = organize
	pc.nfoEnabled = nfo
}

// ProcessFiles processes the selected files with matched JAV IDs
func (pc *ProcessingCoordinator) ProcessFiles(
	ctx context.Context,
	files []FileItem,
	matches map[string]matcher.MatchResult,
) error {
	for _, file := range files {
		if !file.Selected || !file.Matched {
			continue
		}

		match, found := matches[file.Path]
		if !found {
			continue
		}

		// Submit scrape task
		var movie *models.Movie
		if pc.scrapeEnabled {
			scrapeTask := worker.NewScrapeTask(
				match.ID,
				pc.registry,
				pc.aggregator,
				pc.movieRepo,
				pc.progressTracker,
			)
			if err := pc.pool.Submit(scrapeTask); err != nil {
				return fmt.Errorf("failed to submit scrape task for %s: %w", match.ID, err)
			}

			// Try to get movie from repo for subsequent tasks
			movie, _ = pc.movieRepo.FindByID(match.ID)
		}

		// Submit download task (if we have movie metadata)
		if pc.downloadEnabled && movie != nil {
			downloadDir := filepath.Join(pc.destPath, match.ID)
			downloadTask := worker.NewDownloadTask(
				movie,
				downloadDir,
				pc.downloader,
				pc.progressTracker,
			)
			if err := pc.pool.Submit(downloadTask); err != nil {
				return fmt.Errorf("failed to submit download task for %s: %w", match.ID, err)
			}
		}

		// Submit organize task
		if pc.organizeEnabled {
			organizeTask := worker.NewOrganizeTask(
				match,
				movie,
				pc.destPath,
				pc.moveFiles,
				pc.organizer,
				pc.progressTracker,
			)
			if err := pc.pool.Submit(organizeTask); err != nil {
				return fmt.Errorf("failed to submit organize task for %s: %w", match.ID, err)
			}
		}

		// Submit NFO task (if we have movie metadata)
		if pc.nfoEnabled && movie != nil {
			nfoDir := filepath.Join(pc.destPath, match.ID)
			nfoTask := worker.NewNFOTask(
				movie,
				nfoDir,
				pc.nfoGenerator,
				pc.progressTracker,
			)
			if err := pc.pool.Submit(nfoTask); err != nil {
				return fmt.Errorf("failed to submit NFO task for %s: %w", match.ID, err)
			}
		}
	}

	return nil
}

// Wait waits for all tasks to complete
func (pc *ProcessingCoordinator) Wait() error {
	return pc.pool.Wait()
}

// Stop stops the worker pool
func (pc *ProcessingCoordinator) Stop() {
	pc.pool.Stop()
}
