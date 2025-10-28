package api

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/logging"
	"github.com/javinizer/javinizer-go/internal/models"
)

// Mutex to serialize config updates (prevents concurrent read-modify-write races)
var configMutex sync.Mutex

// healthCheck godoc
// @Summary Health check
// @Description Check API health and list enabled scrapers
// @Tags system
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func healthCheck(registry *models.ScraperRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		scrapers := []string{}
		for _, s := range registry.GetEnabled() {
			scrapers = append(scrapers, s.Name())
		}
		c.JSON(200, HealthResponse{
			Status:   "ok",
			Scrapers: scrapers,
		})
	}
}

// getConfig godoc
// @Summary Get configuration
// @Description Retrieve the current server configuration
// @Tags system
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/config [get]
func getConfig(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, cfg)
	}
}

// getAvailableScrapers godoc
// @Summary Get available scrapers
// @Description Get list of all available scrapers with their display names and enabled status
// @Tags system
// @Produce json
// @Success 200 {object} AvailableScrapersResponse
// @Router /api/v1/scrapers [get]
func getAvailableScrapers(registry *models.ScraperRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		scrapers := []ScraperInfo{}

		// Get all registered scrapers
		for _, scraper := range registry.GetAll() {
			name := scraper.Name()

			// Map internal names to display names
			displayName := name
			var options []ScraperOption

			switch name {
			case "r18dev":
				displayName = "R18.dev"
				// R18Dev has no additional options
				options = []ScraperOption{}
			case "dmm":
				displayName = "DMM/Fanza"
				// DMM scraper options
				options = []ScraperOption{
					{
						Key:         "scrape_actress",
						Label:       "Scrape Actress Information",
						Description: "Enable detailed actress data scraping from DMM (may be slower)",
						Type:        "boolean",
					},
				}
			}

			scrapers = append(scrapers, ScraperInfo{
				Name:        name,
				DisplayName: displayName,
				Enabled:     scraper.IsEnabled(),
				Options:     options,
			})
		}

		c.JSON(200, AvailableScrapersResponse{
			Scrapers: scrapers,
		})
	}
}

// updateConfig godoc
// @Summary Update configuration
// @Description Update and save the server configuration
// @Tags system
// @Accept json
// @Produce json
// @Param config body config.Config true "Full configuration object"
// @Success 200 {object} map[string]interface{} "message: Configuration saved successfully"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/config [put]
func updateConfig(cfg *config.Config, cfgFile string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Serialize updates to prevent concurrent read-modify-write races
		configMutex.Lock()
		defer configMutex.Unlock()

		// Parse incoming config
		var newConfig config.Config
		if err := c.ShouldBindJSON(&newConfig); err != nil {
			c.JSON(400, ErrorResponse{Error: "Invalid configuration format"})
			return
		}

		// Save to YAML file (empty arrays are preserved, not removed)
		if err := config.Save(&newConfig, cfgFile); err != nil {
			logging.Errorf("Failed to save config: %v", err)
			c.JSON(500, ErrorResponse{Error: "Failed to save configuration"})
			return
		}

		// Update the in-memory config
		*cfg = newConfig

		logging.Info("Configuration updated successfully")
		c.JSON(200, gin.H{
			"message": "Configuration saved successfully",
		})
	}
}
