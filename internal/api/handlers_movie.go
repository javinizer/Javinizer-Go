package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/javinizer/javinizer-go/internal/logging"
	"github.com/javinizer/javinizer-go/internal/models"
)

// scrapeMovie godoc
// @Summary Scrape movie metadata
// @Description Scrape metadata from configured sources and cache in database
// @Tags movies
// @Accept json
// @Produce json
// @Param request body ScrapeRequest true "Movie ID to scrape"
// @Success 200 {object} ScrapeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/scrape [post]
func scrapeMovie(deps *ServerDependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ScrapeRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, ErrorResponse{Error: err.Error()})
			return
		}

		// Check if already in database
		existing, err := deps.MovieRepo.FindByID(req.ID)
		if err == nil && existing != nil {
			c.JSON(200, ScrapeResponse{
				Cached: true,
				Movie:  existing,
			})
			return
		}

		// Scrape from sources in priority order - use getters for thread-safe access
		results := []*models.ScraperResult{}
		errors := []string{}

		for _, scraper := range deps.GetRegistry().GetByPriority(deps.GetConfig().Scrapers.Priority) {
			result, err := scraper.Search(req.ID)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", scraper.Name(), err))
				continue
			}
			results = append(results, result)
		}

		if len(results) == 0 {
			c.JSON(404, ErrorResponse{
				Error:  "Movie not found",
				Errors: errors,
			})
			return
		}

		// Aggregate results - use getter for thread-safe access
		movie, err := deps.GetAggregator().Aggregate(results)
		if err != nil {
			c.JSON(500, ErrorResponse{Error: err.Error()})
			return
		}

		movie.OriginalFileName = req.ID

		// Save to database (upsert: create or update)
		if err := deps.MovieRepo.Upsert(movie); err != nil {
			logging.Errorf("Failed to save movie to database: %v", err)
		}

		c.JSON(200, ScrapeResponse{
			Cached:      false,
			Movie:       movie,
			SourcesUsed: len(results),
			Errors:      errors,
		})
	}
}

// getMovie godoc
// @Summary Get movie by ID
// @Description Retrieve movie metadata from cache by ID
// @Tags movies
// @Produce json
// @Param id path string true "Movie ID" example:"IPX-535"
// @Success 200 {object} MovieResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/movie/{id} [get]
func getMovie(deps *ServerDependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		movie, err := deps.MovieRepo.FindByID(id)
		if err != nil {
			c.JSON(404, ErrorResponse{Error: "Movie not found"})
			return
		}

		c.JSON(200, MovieResponse{Movie: movie})
	}
}

// listMovies godoc
// @Summary List cached movies
// @Description Get a list of cached movies from the database
// @Tags movies
// @Produce json
// @Success 200 {object} MoviesResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/movies [get]
func listMovies(deps *ServerDependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := 20
		offset := 0

		movies, err := deps.MovieRepo.List(limit, offset)
		if err != nil {
			c.JSON(500, ErrorResponse{Error: err.Error()})
			return
		}

		c.JSON(200, MoviesResponse{
			Movies: movies,
			Count:  len(movies),
		})
	}
}
