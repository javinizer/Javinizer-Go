package api

import (
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/models"
)

// getAvailableScrapers godoc
// @Summary Get available scrapers
// @Description Get list of all available scrapers with their display names, enabled status, and configuration options. Scrapers are ordered by priority from config.
// @Tags system
// @Produce json
// @Success 200 {object} AvailableScrapersResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/scrapers [get]
func getAvailableScrapers(deps *ServerDependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		scrapers := []ScraperInfo{}
		cfg := deps.GetConfig()
		profileChoices := proxyProfileChoices(cfg)

		// Use getter to get current registry (respects config reloads)
		registry := deps.GetRegistry()
		registered := registry.GetAll()
		scraperByName := make(map[string]models.Scraper, len(registered))
		for _, scraper := range registered {
			scraperByName[scraper.Name()] = scraper
		}

		// Build deterministic order:
		// 1) config scrapers.priority order
		// 2) any remaining registered scrapers (sorted by name)
		orderedNames := make([]string, 0, len(scraperByName))
		seen := make(map[string]bool, len(scraperByName))
		if cfg != nil {
			for _, name := range cfg.Scrapers.Priority {
				if _, ok := scraperByName[name]; !ok || seen[name] {
					continue
				}
				orderedNames = append(orderedNames, name)
				seen[name] = true
			}
		}
		remainingNames := make([]string, 0, len(scraperByName))
		for name := range scraperByName {
			if !seen[name] {
				remainingNames = append(remainingNames, name)
			}
		}
		sort.Strings(remainingNames)
		orderedNames = append(orderedNames, remainingNames...)

		for _, name := range orderedNames {
			scraper := scraperByName[name]
			// Map internal names to display names
			displayName := name
			var options []ScraperOption

			switch name {
			case "r18dev":
				displayName = "R18.dev"
				options = []ScraperOption{
					{
						Key:         "language",
						Label:       "Language",
						Description: "Language for metadata fields from R18.dev",
						Type:        "select",
						Choices: []ScraperChoice{
							{Value: "en", Label: "English"},
							{Value: "ja", Label: "Japanese"},
						},
					},
				}
				options = append(options, scraperFakeUserAgentOptions()...)
				options = append(options, scraperProxyOptions(profileChoices)...)
				options = append(options, scraperDownloadProxyOptions(profileChoices)...)
			case "dmm":
				displayName = "DMM/Fanza"
				// DMM scraper options
				minTimeout := 5
				maxTimeout := 120
				options = []ScraperOption{
					{
						Key:         "scrape_actress",
						Label:       "Scrape Actress Information",
						Description: "Extract actress names and IDs from DMM. Disable for faster scraping if you only need actress data from other sources.",
						Type:        "boolean",
					},
					{
						Key:         "enable_browser",
						Label:       "Enable browser mode",
						Description: "Use browser automation for video.dmm.co.jp (required for JavaScript-rendered content)",
						Type:        "boolean",
					},
					{
						Key:         "browser_timeout",
						Label:       "Browser timeout",
						Description: "Maximum time to wait for browser operations",
						Type:        "number",
						Min:         &minTimeout,
						Max:         &maxTimeout,
						Unit:        "seconds",
					},
				}
				options = append(options, scraperFakeUserAgentOptions()...)
				options = append(options, scraperProxyOptions(profileChoices)...)
				options = append(options, scraperDownloadProxyOptions(profileChoices)...)
			case "libredmm":
				displayName = "LibreDMM"
				options = []ScraperOption{
					{
						Key:         "request_delay",
						Label:       "Request delay",
						Description: "Delay between requests to avoid rate limiting",
						Type:        "number",
						Min:         ptrInt(0),
						Max:         ptrInt(5000),
						Unit:        "ms",
					},
					{
						Key:         "base_url",
						Label:       "Base URL",
						Description: "LibreDMM base URL",
						Type:        "string",
					},
				}
				options = append(options, scraperFakeUserAgentOptions()...)
				options = append(options, scraperProxyOptions(profileChoices)...)
				options = append(options, scraperDownloadProxyOptions(profileChoices)...)
			case "mgstage":
				displayName = "MGStage"
				// MGStage scraper options
				options = []ScraperOption{
					{
						Key:         "request_delay",
						Label:       "Request delay",
						Description: "Delay between requests to avoid rate limiting (0 = no delay)",
						Type:        "number",
						Min:         ptrInt(0),
						Max:         ptrInt(5000),
						Unit:        "ms",
					},
				}
				options = append(options, scraperFakeUserAgentOptions()...)
				options = append(options, scraperProxyOptions(profileChoices)...)
				options = append(options, scraperDownloadProxyOptions(profileChoices)...)
			case "javlibrary":
				displayName = "JavLibrary"
				options = []ScraperOption{
					{
						Key:         "language",
						Label:       "Language",
						Description: "Language for metadata (affects title, genres, and actress names)",
						Type:        "select",
						Choices: []ScraperChoice{
							{Value: "en", Label: "English"},
							{Value: "ja", Label: "Japanese"},
							{Value: "cn", Label: "Chinese (Simplified)"},
							{Value: "tw", Label: "Chinese (Traditional)"},
						},
					},
					{
						Key:         "request_delay",
						Label:       "Request delay",
						Description: "Delay between requests to avoid rate limiting",
						Type:        "number",
						Min:         ptrInt(0),
						Max:         ptrInt(5000),
						Unit:        "ms",
					},
					{
						Key:         "base_url",
						Label:       "Base URL",
						Description: "JavLibrary base URL (leave default unless you need a mirror/domain override)",
						Type:        "string",
					},
					{
						Key:         "use_flaresolverr",
						Label:       "Use FlareSolverr",
						Description: "Route requests through FlareSolverr to bypass Cloudflare protection (requires FlareSolverr to be configured in Proxy settings)",
						Type:        "boolean",
					},
				}
				options = append(options, scraperFakeUserAgentOptions()...)
				options = append(options, scraperProxyOptions(profileChoices)...)
				options = append(options, scraperDownloadProxyOptions(profileChoices)...)
			case "javdb":
				displayName = "JavDB"
				options = []ScraperOption{
					{
						Key:         "request_delay",
						Label:       "Request delay",
						Description: "Delay between requests to avoid rate limiting",
						Type:        "number",
						Min:         ptrInt(0),
						Max:         ptrInt(5000),
						Unit:        "ms",
					},
					{
						Key:         "base_url",
						Label:       "Base URL",
						Description: "JavDB base URL (leave default unless you need a mirror/domain override)",
						Type:        "string",
					},
					{
						Key:         "use_flaresolverr",
						Label:       "Use FlareSolverr",
						Description: "Route requests through FlareSolverr to bypass Cloudflare protection (often needed for JavDB)",
						Type:        "boolean",
					},
				}
				options = append(options, scraperFakeUserAgentOptions()...)
				options = append(options, scraperProxyOptions(profileChoices)...)
				options = append(options, scraperDownloadProxyOptions(profileChoices)...)
			case "javbus":
				displayName = "JavBus"
				options = []ScraperOption{
					{
						Key:         "language",
						Label:       "Language",
						Description: "Language for metadata output",
						Type:        "select",
						Choices: []ScraperChoice{
							{Value: "ja", Label: "Japanese"},
							{Value: "en", Label: "English"},
							{Value: "zh", Label: "Chinese"},
						},
					},
					{
						Key:         "request_delay",
						Label:       "Request delay",
						Description: "Delay between requests to avoid rate limiting",
						Type:        "number",
						Min:         ptrInt(0),
						Max:         ptrInt(5000),
						Unit:        "ms",
					},
					{
						Key:         "base_url",
						Label:       "Base URL",
						Description: "JavBus base URL (leave default unless you need a mirror/domain override)",
						Type:        "string",
					},
				}
				options = append(options, scraperFakeUserAgentOptions()...)
				options = append(options, scraperProxyOptions(profileChoices)...)
				options = append(options, scraperDownloadProxyOptions(profileChoices)...)
			case "jav321":
				displayName = "Jav321"
				options = []ScraperOption{
					{
						Key:         "request_delay",
						Label:       "Request delay",
						Description: "Delay between requests to avoid rate limiting",
						Type:        "number",
						Min:         ptrInt(0),
						Max:         ptrInt(5000),
						Unit:        "ms",
					},
					{
						Key:         "base_url",
						Label:       "Base URL",
						Description: "Jav321 base URL",
						Type:        "string",
					},
				}
				options = append(options, scraperFakeUserAgentOptions()...)
				options = append(options, scraperProxyOptions(profileChoices)...)
				options = append(options, scraperDownloadProxyOptions(profileChoices)...)
			case "tokyohot":
				displayName = "Tokyo-Hot"
				options = []ScraperOption{
					{
						Key:         "language",
						Label:       "Language",
						Description: "Language for metadata output",
						Type:        "select",
						Choices: []ScraperChoice{
							{Value: "ja", Label: "Japanese"},
							{Value: "en", Label: "English"},
							{Value: "zh", Label: "Chinese"},
						},
					},
					{
						Key:         "request_delay",
						Label:       "Request delay",
						Description: "Delay between requests to avoid rate limiting",
						Type:        "number",
						Min:         ptrInt(0),
						Max:         ptrInt(5000),
						Unit:        "ms",
					},
					{
						Key:         "base_url",
						Label:       "Base URL",
						Description: "Tokyo-Hot base URL",
						Type:        "string",
					},
				}
				options = append(options, scraperFakeUserAgentOptions()...)
				options = append(options, scraperProxyOptions(profileChoices)...)
				options = append(options, scraperDownloadProxyOptions(profileChoices)...)
			case "aventertainment":
				displayName = "AV Entertainment"
				options = []ScraperOption{
					{
						Key:         "language",
						Label:       "Language",
						Description: "Language for metadata output",
						Type:        "select",
						Choices: []ScraperChoice{
							{Value: "en", Label: "English"},
							{Value: "ja", Label: "Japanese"},
						},
					},
					{
						Key:         "request_delay",
						Label:       "Request delay",
						Description: "Delay between requests to avoid rate limiting",
						Type:        "number",
						Min:         ptrInt(0),
						Max:         ptrInt(5000),
						Unit:        "ms",
					},
					{
						Key:         "base_url",
						Label:       "Base URL",
						Description: "AV Entertainment base URL",
						Type:        "string",
					},
					{
						Key:         "scrape_bonus_screens",
						Label:       "Scrape bonus screenshots",
						Description: "Append bonus image files (e.g., 特典ファイル) to screenshots",
						Type:        "boolean",
					},
				}
				options = append(options, scraperFakeUserAgentOptions()...)
				options = append(options, scraperProxyOptions(profileChoices)...)
				options = append(options, scraperDownloadProxyOptions(profileChoices)...)
			case "dlgetchu":
				displayName = "DLGetchu"
				options = []ScraperOption{
					{
						Key:         "request_delay",
						Label:       "Request delay",
						Description: "Delay between requests to avoid rate limiting",
						Type:        "number",
						Min:         ptrInt(0),
						Max:         ptrInt(5000),
						Unit:        "ms",
					},
					{
						Key:         "base_url",
						Label:       "Base URL",
						Description: "DLGetchu base URL",
						Type:        "string",
					},
				}
				options = append(options, scraperFakeUserAgentOptions()...)
				options = append(options, scraperProxyOptions(profileChoices)...)
				options = append(options, scraperDownloadProxyOptions(profileChoices)...)
			case "caribbeancom":
				displayName = "Caribbeancom"
				options = []ScraperOption{
					{
						Key:         "language",
						Label:       "Language",
						Description: "Language for metadata output",
						Type:        "select",
						Choices: []ScraperChoice{
							{Value: "ja", Label: "Japanese"},
							{Value: "en", Label: "English"},
						},
					},
					{
						Key:         "request_delay",
						Label:       "Request delay",
						Description: "Delay between requests to avoid rate limiting",
						Type:        "number",
						Min:         ptrInt(0),
						Max:         ptrInt(5000),
						Unit:        "ms",
					},
					{
						Key:         "base_url",
						Label:       "Base URL",
						Description: "Caribbeancom base URL",
						Type:        "string",
					},
				}
				options = append(options, scraperFakeUserAgentOptions()...)
				options = append(options, scraperProxyOptions(profileChoices)...)
				options = append(options, scraperDownloadProxyOptions(profileChoices)...)
			case "fc2":
				displayName = "FC2"
				options = []ScraperOption{
					{
						Key:         "request_delay",
						Label:       "Request delay",
						Description: "Delay between requests to avoid rate limiting",
						Type:        "number",
						Min:         ptrInt(0),
						Max:         ptrInt(5000),
						Unit:        "ms",
					},
					{
						Key:         "base_url",
						Label:       "Base URL",
						Description: "FC2 base URL",
						Type:        "string",
					},
				}
				options = append(options, scraperFakeUserAgentOptions()...)
				options = append(options, scraperProxyOptions(profileChoices)...)
				options = append(options, scraperDownloadProxyOptions(profileChoices)...)
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

func scraperProxyOptions(profileChoices []ScraperChoice) []ScraperOption {
	return []ScraperOption{
		{
			Key:         "proxy.enabled",
			Label:       "Enable proxy for this scraper",
			Description: "Use proxy for this scraper (inherits global proxy profile when no scraper profile is selected)",
			Type:        "boolean",
		},
		{
			Key:         "proxy.profile",
			Label:       "Proxy profile",
			Description: "Optional scraper-specific proxy profile (leave empty to inherit global default profile)",
			Type:        "select",
			Choices:     profileChoices,
		},
	}
}

func scraperFakeUserAgentOptions() []ScraperOption {
	return []ScraperOption{
		{
			Key:         "use_fake_user_agent",
			Label:       "Use fake User-Agent",
			Description: "Use a browser-like User-Agent string for this scraper",
			Type:        "boolean",
		},
		{
			Key:         "fake_user_agent",
			Label:       "Fake User-Agent",
			Description: "Optional custom fake User-Agent (leave empty to use default browser User-Agent)",
			Type:        "string",
		},
	}
}

func scraperDownloadProxyOptions(profileChoices []ScraperChoice) []ScraperOption {
	return []ScraperOption{
		{
			Key:         "download_proxy.enabled",
			Label:       "Download proxy enabled",
			Description: "Enable scraper-specific download proxy override",
			Type:        "boolean",
		},
		{
			Key:         "download_proxy.profile",
			Label:       "Download proxy profile",
			Description: "Optional scraper-specific download proxy profile (leave empty to inherit scraper/global proxy profile)",
			Type:        "select",
			Choices:     profileChoices,
		},
	}
}

func proxyProfileChoices(cfg *config.Config) []ScraperChoice {
	choices := []ScraperChoice{
		{Value: "", Label: "Inherit Default"},
	}
	if cfg == nil || len(cfg.Scrapers.Proxy.Profiles) == 0 {
		return choices
	}

	names := make([]string, 0, len(cfg.Scrapers.Proxy.Profiles))
	for name := range cfg.Scrapers.Proxy.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		choices = append(choices, ScraperChoice{
			Value: name,
			Label: name,
		})
	}

	return choices
}

// ptrInt returns a pointer to an int value
func ptrInt(v int) *int {
	return &v
}
