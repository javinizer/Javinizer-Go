package worker

import (
	"context"
	"fmt"

	"github.com/javinizer/javinizer-go/internal/aggregator"
	"github.com/javinizer/javinizer-go/internal/database"
	"github.com/javinizer/javinizer-go/internal/downloader"
	"github.com/javinizer/javinizer-go/internal/matcher"
	"github.com/javinizer/javinizer-go/internal/models"
	"github.com/javinizer/javinizer-go/internal/nfo"
	"github.com/javinizer/javinizer-go/internal/organizer"
)

// ScrapeTask scrapes metadata for a JAV ID
type ScrapeTask struct {
	BaseTask
	javID       string
	registry    *models.ScraperRegistry
	aggregator  *aggregator.Aggregator
	movieRepo   *database.MovieRepository
	progressTracker *ProgressTracker
}

// NewScrapeTask creates a new scrape task
func NewScrapeTask(
	javID string,
	registry *models.ScraperRegistry,
	agg *aggregator.Aggregator,
	movieRepo *database.MovieRepository,
	progressTracker *ProgressTracker,
) *ScrapeTask {
	return &ScrapeTask{
		BaseTask: BaseTask{
			id:          javID,
			taskType:    TaskTypeScrape,
			description: fmt.Sprintf("Scraping metadata for %s", javID),
		},
		javID:           javID,
		registry:        registry,
		aggregator:      agg,
		movieRepo:       movieRepo,
		progressTracker: progressTracker,
	}
}

func (t *ScrapeTask) Execute(ctx context.Context) error {
	// Check cache first
	if _, err := t.movieRepo.FindByID(t.javID); err == nil {
		t.progressTracker.Update(t.id, 1.0, "Found in cache", 0)
		return nil
	}

	// Scrape from sources
	t.progressTracker.Update(t.id, 0.2, "Querying scrapers...", 0)

	results := make([]*models.ScraperResult, 0)
	scrapers := t.registry.GetByPriority([]string{"r18dev", "dmm"})

	for i, scraper := range scrapers {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		progress := 0.2 + (float64(i) / float64(len(scrapers)) * 0.5)
		t.progressTracker.Update(t.id, progress, fmt.Sprintf("Scraping from %s...", scraper.Name()), 0)

		result, err := scraper.Search(t.javID)
		if err != nil {
			continue
		}
		results = append(results, result)
	}

	if len(results) == 0 {
		return fmt.Errorf("no results found from any scraper")
	}

	t.progressTracker.Update(t.id, 0.8, "Aggregating metadata...", 0)

	// Aggregate results
	movie, err := t.aggregator.Aggregate(results)
	if err != nil {
		return fmt.Errorf("failed to aggregate: %w", err)
	}

	// Save to database
	t.progressTracker.Update(t.id, 0.9, "Saving to database...", 0)
	if err := t.movieRepo.Upsert(movie); err != nil {
		return fmt.Errorf("failed to save to database: %w", err)
	}

	t.progressTracker.Update(t.id, 1.0, "Completed", 0)
	return nil
}

// DownloadTask downloads media for a movie
type DownloadTask struct {
	BaseTask
	movie           *models.Movie
	targetDir       string
	downloader      *downloader.Downloader
	progressTracker *ProgressTracker
}

// NewDownloadTask creates a new download task
func NewDownloadTask(
	movie *models.Movie,
	targetDir string,
	dl *downloader.Downloader,
	progressTracker *ProgressTracker,
) *DownloadTask {
	return &DownloadTask{
		BaseTask: BaseTask{
			id:          fmt.Sprintf("download-%s", movie.ID),
			taskType:    TaskTypeDownload,
			description: fmt.Sprintf("Downloading media for %s", movie.ID),
		},
		movie:           movie,
		targetDir:       targetDir,
		downloader:      dl,
		progressTracker: progressTracker,
	}
}

func (t *DownloadTask) Execute(ctx context.Context) error {
	t.progressTracker.Update(t.id, 0.1, "Starting downloads...", 0)

	results, err := t.downloader.DownloadAll(t.movie, t.targetDir)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	downloaded := 0
	for _, r := range results {
		if r.Downloaded {
			downloaded++
		}
	}

	t.progressTracker.Update(t.id, 1.0, fmt.Sprintf("Downloaded %d files", downloaded), 0)
	return nil
}

// OrganizeTask organizes a video file
type OrganizeTask struct {
	BaseTask
	match           matcher.MatchResult
	movie           *models.Movie
	destPath        string
	moveFiles       bool
	organizer       *organizer.Organizer
	progressTracker *ProgressTracker
}

// NewOrganizeTask creates a new organize task
func NewOrganizeTask(
	match matcher.MatchResult,
	movie *models.Movie,
	destPath string,
	moveFiles bool,
	org *organizer.Organizer,
	progressTracker *ProgressTracker,
) *OrganizeTask {
	operation := "copy"
	if moveFiles {
		operation = "move"
	}

	return &OrganizeTask{
		BaseTask: BaseTask{
			id:          fmt.Sprintf("organize-%s", match.File.Name),
			taskType:    TaskTypeOrganize,
			description: fmt.Sprintf("Organizing %s (%s)", match.File.Name, operation),
		},
		match:           match,
		movie:           movie,
		destPath:        destPath,
		moveFiles:       moveFiles,
		organizer:       org,
		progressTracker: progressTracker,
	}
}

func (t *OrganizeTask) Execute(ctx context.Context) error {
	t.progressTracker.Update(t.id, 0.2, "Planning organization...", 0)

	// Plan the organization
	plan, err := t.organizer.Plan(t.match, t.movie, t.destPath)
	if err != nil {
		return fmt.Errorf("failed to plan: %w", err)
	}

	// Validate plan
	t.progressTracker.Update(t.id, 0.4, "Validating plan...", 0)
	if issues := organizer.ValidatePlan(plan); len(issues) > 0 {
		return fmt.Errorf("validation failed: %v", issues)
	}

	// Execute plan
	t.progressTracker.Update(t.id, 0.6, "Executing plan...", 0)
	var result *organizer.OrganizeResult
	var execErr error

	if t.moveFiles {
		// Execute moves the file (the default behavior)
		result, execErr = t.organizer.Execute(plan, false)
	} else {
		// Copy copies instead of moving
		result, execErr = t.organizer.Copy(plan, false)
	}

	if execErr != nil {
		return fmt.Errorf("failed to organize: %w", execErr)
	}

	if result.Error != nil {
		return fmt.Errorf("organize error: %w", result.Error)
	}

	t.progressTracker.Update(t.id, 1.0, "Organized successfully", 0)
	return nil
}

// NFOTask generates an NFO file
type NFOTask struct {
	BaseTask
	movie           *models.Movie
	targetDir       string
	generator       *nfo.Generator
	progressTracker *ProgressTracker
}

// NewNFOTask creates a new NFO generation task
func NewNFOTask(
	movie *models.Movie,
	targetDir string,
	gen *nfo.Generator,
	progressTracker *ProgressTracker,
) *NFOTask {
	return &NFOTask{
		BaseTask: BaseTask{
			id:          fmt.Sprintf("nfo-%s", movie.ID),
			taskType:    TaskTypeNFO,
			description: fmt.Sprintf("Generating NFO for %s", movie.ID),
		},
		movie:           movie,
		targetDir:       targetDir,
		generator:       gen,
		progressTracker: progressTracker,
	}
}

func (t *NFOTask) Execute(ctx context.Context) error {
	t.progressTracker.Update(t.id, 0.5, "Generating NFO...", 0)

	if err := t.generator.Generate(t.movie, t.targetDir); err != nil {
		return fmt.Errorf("failed to generate NFO: %w", err)
	}

	t.progressTracker.Update(t.id, 1.0, "NFO generated", 0)
	return nil
}
