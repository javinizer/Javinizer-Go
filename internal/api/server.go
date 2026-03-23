package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/javinizer/javinizer-go/internal/aggregator"
	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/database"
	"github.com/javinizer/javinizer-go/internal/logging"
	"github.com/javinizer/javinizer-go/internal/matcher"
	"github.com/javinizer/javinizer-go/internal/models"
	ws "github.com/javinizer/javinizer-go/internal/websocket"
	"github.com/javinizer/javinizer-go/internal/worker"
)

var (
	wsHub         *ws.Hub
	wsHubCancel   context.CancelFunc // Track cancel function to clean up old hubs
	wsUpgrader    websocket.Upgrader
	wsHubShutdown chan struct{} // Signal when hub goroutine exits
)

// ServerDependencies holds all dependencies needed to create the API server
// Access to Config, Registry, Aggregator, and Matcher must be synchronized
// to prevent data races during config reload.
type ServerDependencies struct {
	mu          sync.RWMutex                  // Protects Registry, Aggregator, Matcher during reload
	config      atomic.Pointer[config.Config] // Thread-safe config access
	ConfigFile  string
	Registry    *models.ScraperRegistry
	DB          *database.DB
	Aggregator  *aggregator.Aggregator
	MovieRepo   *database.MovieRepository
	ActressRepo *database.ActressRepository
	HistoryRepo *database.HistoryRepository
	Matcher     *matcher.Matcher
	JobQueue    *worker.JobQueue
	Auth        *AuthManager
	wsCancel    context.CancelFunc // Cancel function for WebSocket hub context
}

// GetConfig returns the current configuration (thread-safe)
func (d *ServerDependencies) GetConfig() *config.Config {
	cfg := d.config.Load()
	if cfg == nil {
		logging.Errorf("CRITICAL: GetConfig() called before SetConfig() - this is a programming error")
		panic("GetConfig() called with nil config - ensure SetConfig() is called during ServerDependencies initialization")
	}
	return cfg
}

// SetConfig atomically sets the configuration (thread-safe)
func (d *ServerDependencies) SetConfig(cfg *config.Config) {
	if cfg == nil {
		logging.Errorf("CRITICAL: SetConfig() called with nil config - this is a programming error")
		panic("SetConfig() called with nil config - config must not be nil")
	}
	d.config.Store(cfg)
}

// Shutdown gracefully shuts down server resources, including the WebSocket hub
func (d *ServerDependencies) Shutdown() {
	if d.wsCancel != nil {
		d.wsCancel()
	}
}

// GetRegistry returns the current scraper registry (thread-safe)
func (d *ServerDependencies) GetRegistry() *models.ScraperRegistry {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Registry
}

// GetAggregator returns the current aggregator (thread-safe)
func (d *ServerDependencies) GetAggregator() *aggregator.Aggregator {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Aggregator
}

// GetMatcher returns the current matcher (thread-safe)
func (d *ServerDependencies) GetMatcher() *matcher.Matcher {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Matcher
}

// resolveSwaggerPath returns the path to swagger.json, checking multiple locations
// Returns Docker path first (/app/docs/swagger/swagger.json), then falls back to local dev path
func resolveSwaggerPath() string {
	// Try Docker path first (production deployment)
	dockerPath := "/app/docs/swagger/swagger.json"
	if _, err := os.Stat(dockerPath); err == nil {
		return dockerPath
	}

	// Fall back to local development path
	return "./docs/swagger/swagger.json"
}

// isSameOrigin checks if the origin matches the request host (same-origin)
// Compares scheme, hostname, and port (with default port normalization)
func isSameOrigin(origin string, r *http.Request) bool {
	if origin == "" {
		// No Origin header (e.g., some non-browser clients) - treat as same-origin
		return true
	}

	u, err := url.Parse(origin)
	if err != nil {
		return false
	}

	// Get request scheme
	reqScheme := "http"
	if r.TLS != nil {
		reqScheme = "https"
	}

	// Normalize origin port (explicit default ports)
	originPort := u.Port()
	if originPort == "" {
		switch u.Scheme {
		case "http":
			originPort = "80"
		case "https":
			originPort = "443"
		}
	}

	// Parse request host using net.SplitHostPort for proper IPv6 support
	// Handle both "host:port" and "[ipv6]:port" formats
	reqHost := r.Host
	reqPort := ""
	if host, port, err := net.SplitHostPort(r.Host); err == nil {
		reqHost = host
		reqPort = port
	}
	if reqPort == "" {
		if reqScheme == "http" {
			reqPort = "80"
		} else {
			reqPort = "443"
		}
	}

	// Compare scheme, hostname, and port components directly
	return strings.EqualFold(u.Scheme, reqScheme) &&
		strings.EqualFold(u.Hostname(), reqHost) &&
		originPort == reqPort
}

// isOriginAllowed checks if an origin is allowed based on configuration
// Note: This does NOT handle same-origin checking - use isSameOrigin for that
// Note: Wildcard "*" is explicitly NOT supported - it must be an exact origin match
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	// Check each allowed origin
	for _, allowed := range allowedOrigins {
		// Explicitly reject wildcard - this is a security measure to prevent
		// cross-site WebSocket hijacking and CSRF attacks
		if allowed == "*" {
			logging.Warn("Ignoring wildcard '*' in AllowedOrigins - only exact origin matches are supported for security")
			continue
		}
		if allowed == origin {
			return true
		}
	}

	return false
}

// acceptsHTML checks if the request Accept header includes text/html with q>0
// Used to distinguish browser requests from API clients
// Properly parses Accept header to respect quality values (q-values) per RFC 9110
func acceptsHTML(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	if accept == "" {
		return false
	}

	// Parse Accept header: split by comma and check each media type
	parts := strings.Split(accept, ",")
	for _, part := range parts {
		part = strings.TrimSpace(strings.ToLower(part))

		// Extract media type and parameters
		segments := strings.Split(part, ";")
		if len(segments) == 0 {
			continue
		}

		mediaType := strings.TrimSpace(segments[0])

		// Check if this is text/html
		if mediaType == "text/html" {
			// Parse parameters to find q-value
			qValue := 1.0 // Default quality is 1.0 if not specified

			for i := 1; i < len(segments); i++ {
				param := strings.TrimSpace(segments[i])
				if strings.HasPrefix(param, "q=") {
					// Extract q-value
					qStr := strings.TrimPrefix(param, "q=")
					qStr = strings.TrimSpace(qStr)
					// Simple parsing: check if it starts with "0" or "0."
					if qStr == "0" || qStr == "0.0" || qStr == "0.00" || qStr == "0.000" {
						qValue = 0.0
						break
					}
					// Any other value means q > 0
					break
				}
			}

			// Only accept if q > 0
			if qValue > 0 {
				return true
			}
			// If q=0, continue checking other media types
		}
	}

	return false
}

// NewServer creates and configures the Gin router with all API endpoints
func NewServer(deps *ServerDependencies) *gin.Engine {
	// Clean up existing WebSocket hub if it exists (important for tests that call NewServer multiple times)
	if wsHubCancel != nil {
		wsHubCancel() // Cancel the old hub's context
		if wsHubShutdown != nil {
			// Wait for the old hub goroutine to fully exit (max 500ms)
			select {
			case <-wsHubShutdown:
				// Old hub shut down successfully
			case <-time.After(500 * time.Millisecond):
				// Timeout - old hub didn't shut down gracefully, but proceed anyway
				logging.Warnf("Old WebSocket hub did not shut down within timeout")
			}
		}
	}

	// Initialize job queue and WebSocket hub
	wsHub = ws.NewHub()
	wsHubShutdown = make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	wsHubCancel = cancel
	deps.wsCancel = cancel

	go func() {
		wsHub.Run(ctx)
		close(wsHubShutdown) // Signal that hub goroutine has exited
	}()

	// Configure WebSocket upgrader with dynamic origin checking from config
	// Read allowedOrigins from deps.GetConfig() each time to respect config reloads
	wsUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")

			// Read current allowed origins from config (respects config reloads)
			allowedOrigins := deps.GetConfig().API.Security.AllowedOrigins

			// Empty config → allow same-origin only (secure default)
			if len(allowedOrigins) == 0 {
				return isSameOrigin(origin, r)
			}

			// Check for exact origin match
			// Note: Wildcard "*" is NOT supported for security reasons - it would disable CSRF protection
			if isOriginAllowed(origin, allowedOrigins) {
				return true
			}

			// Reject
			logging.Debugf("WebSocket: Rejected origin %s (not in allowed list)", origin)
			return false
		},
	}

	if deps.GetConfig().Logging.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	webAssets := loadWebUIAssets()

	registerCORSMiddleware(router, deps)
	registerDocumentationRoutes(router)
	registerCoreRoutes(router, deps)
	registerAPIV1Routes(router, deps)
	registerStaticWebRoutes(router, webAssets)
	registerNoRouteHandler(router, webAssets)
	logRegisteredRoutes(router)

	return router
}

// serveScalarDocs serves the Scalar API documentation UI
func serveScalarDocs(c *gin.Context) {
	html := `<!doctype html>
<html>
  <head>
    <title>Javinizer API Documentation</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <script
      id="api-reference"
      data-url="/docs/openapi.json"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// LogServerInfo logs information about the running server
func LogServerInfo(cfg *config.Config) {
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logging.Infof("Starting API server on %s", addr)
	logging.Infof("📚 API Documentation (Scalar): http://%s/docs", addr)
	logging.Infof("📖 Swagger UI: http://%s/swagger/index.html", addr)
	logging.Infof("🏥 Health check: http://%s/health", addr)
}
