package api

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/javinizer/javinizer-go/internal/logging"
	webui "github.com/javinizer/javinizer-go/web"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type webUIAssets struct {
	distFS      fs.FS
	staticFS    http.FileSystem
	indexHTML   []byte
	uiAvailable bool
}

func loadWebUIAssets() webUIAssets {
	assets := webUIAssets{}

	distFS, distErr := webui.DistFS()
	if distErr != nil {
		logging.Warnf("Web UI assets unavailable: %v", distErr)
		return assets
	}

	assets.distFS = distFS
	assets.staticFS = http.FS(distFS)

	if _, err := fs.Stat(distFS, "index.html"); err != nil {
		logging.Warnf("Web UI index.html not found in embedded assets: %v", err)
		return assets
	}

	indexBytes, readErr := fs.ReadFile(distFS, "index.html")
	if readErr != nil {
		logging.Warnf("Failed to read embedded Web UI index.html: %v", readErr)
		return assets
	}

	assets.indexHTML = indexBytes
	assets.uiAvailable = true
	return assets
}

// registerCORSMiddleware configures CORS with dynamic origin validation.
func registerCORSMiddleware(router *gin.Engine, deps *ServerDependencies) {
	// Read allowedOrigins from deps.GetConfig() each time to respect config reloads.
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Read current allowed origins from config (respects config reloads)
		allowedOrigins := deps.GetConfig().API.Security.AllowedOrigins

		// Handle CORS based on configuration.
		if len(allowedOrigins) == 0 {
			// Empty config -> allow same-origin only.
			if isSameOrigin(origin, c.Request) {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
				c.Writer.Header().Add("Vary", "Origin")
			}
		} else {
			// Check for exact origin match only.
			// Note: Wildcard "*" is NOT supported for security reasons.
			if isOriginAllowed(origin, allowedOrigins) {
				// Specific origin allowed.
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
				c.Writer.Header().Add("Vary", "Origin")
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		// Allow specific headers - whitelist approach for security.
		allowedHeaders := []string{
			"Content-Type",
			"Authorization",
			"Accept",
			"Origin",
			"X-Requested-With",
		}
		c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
}

func registerDocumentationRoutes(router *gin.Engine) {
	// Serve OpenAPI spec directly for Scalar.
	router.StaticFile("/docs/openapi.json", resolveSwaggerPath())

	// Scalar API documentation (modern, beautiful UI).
	router.GET("/docs", serveScalarDocs)

	// Also provide traditional Swagger UI as fallback.
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func registerCoreRoutes(router *gin.Engine, deps *ServerDependencies) {
	router.GET("/health", healthCheck(deps))
	router.GET("/ws/progress", requireAuthenticated(deps), handleWebSocket(wsHub))
}

func registerAPIV1Routes(router *gin.Engine, deps *ServerDependencies) {
	// API v1 routes (define BEFORE static files to ensure API takes precedence).
	v1 := router.Group("/api/v1")

	registerAuthRoutes(v1, deps)

	protected := v1.Group("")
	protected.Use(requireAuthenticated(deps))

	registerProtectedRoutes(protected, deps)
}

func registerAuthRoutes(v1 *gin.RouterGroup, deps *ServerDependencies) {
	// Authentication endpoints (must remain public for first-run setup/login).
	v1.GET("/auth/status", getAuthStatus(deps))
	v1.POST("/auth/setup", setupAuth(deps))
	v1.POST("/auth/login", loginAuth(deps))
	v1.POST("/auth/logout", logoutAuth(deps))
}

func registerProtectedRoutes(protected *gin.RouterGroup, deps *ServerDependencies) {
	registerMovieRoutes(protected, deps)
	registerActressRoutes(protected, deps)
	registerSystemRoutes(protected, deps)
	registerVersionRoutes(protected, deps)
	registerFileRoutes(protected, deps)
	registerBatchRoutes(protected, deps)
	registerHistoryRoutes(protected, deps)
}

func registerMovieRoutes(protected *gin.RouterGroup, deps *ServerDependencies) {
	protected.POST("/scrape", scrapeMovie(deps))
	protected.GET("/movies/:id", getMovie(deps))
	protected.GET("/movies", listMovies(deps))
	protected.POST("/movies/:id/rescrape", rescrapeMovie(deps))
	protected.POST("/movies/:id/compare-nfo", compareNFO(deps))
}

func registerActressRoutes(protected *gin.RouterGroup, deps *ServerDependencies) {
	protected.GET("/actresses", listActresses(deps.ActressRepo))
	protected.GET("/actresses/:id", getActress(deps.ActressRepo))
	protected.POST("/actresses", createActress(deps.ActressRepo))
	protected.PUT("/actresses/:id", updateActress(deps.ActressRepo))
	protected.DELETE("/actresses/:id", deleteActress(deps.ActressRepo))
	protected.GET("/actresses/search", searchActresses(deps.ActressRepo))
	protected.POST("/actresses/merge/preview", previewActressMerge(deps.ActressRepo))
	protected.POST("/actresses/merge", mergeActresses(deps.ActressRepo))
}

func registerSystemRoutes(protected *gin.RouterGroup, deps *ServerDependencies) {
	protected.GET("/config", getConfig(deps))
	protected.PUT("/config", updateConfig(deps))
	protected.GET("/scrapers", getAvailableScrapers(deps))
	protected.POST("/proxy/test", testProxy(deps))
	protected.POST("/translation/models", getTranslationModels(deps))
}

func registerVersionRoutes(protected *gin.RouterGroup, deps *ServerDependencies) {
	protected.GET("/version", versionStatus(deps))
	protected.POST("/version/check", versionCheck(deps))
}

func registerFileRoutes(protected *gin.RouterGroup, deps *ServerDependencies) {
	protected.GET("/cwd", getCurrentWorkingDirectory(deps))
	protected.POST("/scan", scanDirectory(deps))
	protected.POST("/browse", browseDirectory(deps))
	protected.POST("/browse/autocomplete", autocompletePath(deps))
}

func registerBatchRoutes(protected *gin.RouterGroup, deps *ServerDependencies) {
	protected.POST("/batch/scrape", batchScrape(deps))
	protected.GET("/batch/:id", getBatchJob(deps))
	protected.POST("/batch/:id/cancel", cancelBatchJob(deps))
	protected.PATCH("/batch/:id/movies/:movieId", updateBatchMovie(deps))
	protected.POST("/batch/:id/movies/:movieId/poster-crop", updateBatchMoviePosterCrop(deps))
	protected.POST("/batch/:id/movies/:movieId/exclude", excludeBatchMovie(deps))
	protected.POST("/batch/:id/movies/:movieId/preview", previewOrganize(deps))
	protected.POST("/batch/:id/movies/:movieId/rescrape", rescrapeBatchMovie(deps))
	protected.POST("/batch/:id/organize", organizeJob(deps))
	protected.POST("/batch/:id/update", updateBatchJob(deps))
	// Temp resource endpoints (for review page preview).
	protected.GET("/temp/posters/:jobId/:filename", serveTempPoster())
	protected.GET("/temp/image", serveTempImage(deps))
	// Persistent resource endpoints (for cropped posters stored in database).
	protected.GET("/posters/:filename", serveCroppedPoster())
}

func registerHistoryRoutes(protected *gin.RouterGroup, deps *ServerDependencies) {
	protected.GET("/history", getHistory(deps.HistoryRepo))
	protected.GET("/history/stats", getHistoryStats(deps.HistoryRepo))
	protected.DELETE("/history/:id", deleteHistory(deps.HistoryRepo))
	protected.DELETE("/history", deleteHistoryBulk(deps.HistoryRepo))
}

func registerStaticWebRoutes(router *gin.Engine, assets webUIAssets) {
	// Serve frontend static files from embedded web bundle.
	// Define AFTER API routes so API takes precedence.
	if !assets.uiAvailable {
		return
	}

	if appFS, err := fs.Sub(assets.distFS, "_app"); err == nil {
		router.StaticFS("/_app", http.FS(appFS))
	} else {
		logging.Warnf("Web UI _app assets unavailable: %v", err)
	}

	if _, err := fs.Stat(assets.distFS, "favicon.ico"); err == nil {
		router.GET("/favicon.ico", func(c *gin.Context) {
			c.FileFromFS("favicon.ico", assets.staticFS)
		})
	}

	if _, err := fs.Stat(assets.distFS, "robots.txt"); err == nil {
		router.GET("/robots.txt", func(c *gin.Context) {
			c.FileFromFS("robots.txt", assets.staticFS)
		})
	}
}

func registerNoRouteHandler(router *gin.Engine, assets webUIAssets) {
	// Fallback: serve index.html for browser SPA routing only.
	// API requests to non-existent endpoints should return proper 404 JSON.
	router.NoRoute(func(c *gin.Context) {
		logging.Debugf("NoRoute hit: %s %s (Accept: %s)",
			c.Request.Method,
			c.Request.URL.Path,
			c.Request.Header.Get("Accept"))

		method := c.Request.Method
		if assets.uiAvailable && acceptsHTML(c) {
			// HEAD requests should not return a body per HTTP semantics.
			if method == http.MethodHead {
				c.Status(http.StatusNoContent)
				return
			}
			// Serve SPA for GET requests.
			if method == http.MethodGet {
				c.Data(http.StatusOK, "text/html; charset=utf-8", assets.indexHTML)
				return
			}
		}

		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Not Found",
			"message": "The requested resource does not exist",
			"path":    c.Request.URL.Path,
			"method":  c.Request.Method,
		})
	})
}

func logRegisteredRoutes(router *gin.Engine) {
	logging.Debugf("Registered routes:")
	for _, route := range router.Routes() {
		logging.Debugf("  %s %s", route.Method, route.Path)
	}
}
