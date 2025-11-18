package api

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/javinizer/javinizer-go/internal/aggregator"
	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/database"
	"github.com/javinizer/javinizer-go/internal/matcher"
	"github.com/javinizer/javinizer-go/internal/models"
	"github.com/javinizer/javinizer-go/internal/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	// Test that NewServer properly initializes all routes
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Logging: config.LoggingConfig{
			Level: "info",
		},
		Matching: config.MatchingConfig{
			RegexEnabled: false,
		},
	}

	registry := models.NewScraperRegistry()
	mat, err := matcher.NewMatcher(&cfg.Matching)
	require.NoError(t, err)

	deps := &ServerDependencies{
		ConfigFile:  "/tmp/config.yaml",
		Registry:    registry,
		DB:          nil, // Not needed for route testing
		Aggregator:  aggregator.New(cfg),
		MovieRepo:   newMockMovieRepo(),
		ActressRepo: newMockActressRepo(),
		Matcher:     mat,
		JobQueue:    worker.NewJobQueue(),
	}
	// Initialize atomic config pointer
	deps.SetConfig(cfg)

	router := NewServer(deps)
	defer cleanupServerHub(t, deps)
	require.NotNil(t, router)

	// Test that routes are registered
	routes := router.Routes()
	routePaths := make(map[string]bool)
	for _, route := range routes {
		routePaths[route.Path] = true
	}

	// Verify critical endpoints exist
	expectedRoutes := []string{
		"/health",
		"/ws/progress",
		"/api/v1/scrape",
		"/api/v1/movies/:id",
		"/api/v1/movies",
		"/api/v1/actresses/search",
		"/api/v1/config",
		"/api/v1/scrapers",
		"/api/v1/scan",
		"/api/v1/browse",
		"/api/v1/cwd",
		"/api/v1/batch/scrape",
		"/api/v1/batch/:id",
		"/api/v1/batch/:id/cancel",
		"/api/v1/batch/:id/movies/:movieId",
		"/api/v1/batch/:id/movies/:movieId/preview",
		"/api/v1/batch/:id/organize",
	}

	for _, route := range expectedRoutes {
		assert.True(t, routePaths[route], "Route %s should be registered", route)
	}
}

func TestNewServer_CORSHeaders(t *testing.T) {
	cfg := &config.Config{
		API: config.APIConfig{
			Security: config.SecurityConfig{
				AllowedOrigins: []string{"*"}, // Allow all origins for test
			},
		},
		Logging: config.LoggingConfig{
			Level: "info",
		},
		Matching: config.MatchingConfig{
			RegexEnabled: false,
		},
	}

	registry := models.NewScraperRegistry()
	mat, err := matcher.NewMatcher(&cfg.Matching)
	require.NoError(t, err)

	deps := &ServerDependencies{
		ConfigFile:  "/tmp/config.yaml",
		Registry:    registry,
		Aggregator:  aggregator.New(cfg),
		MovieRepo:   newMockMovieRepo(),
		ActressRepo: newMockActressRepo(),
		Matcher:     mat,
		JobQueue:    worker.NewJobQueue(),
	}
	// Initialize atomic config pointer
	deps.SetConfig(cfg)

	router := NewServer(deps)
	defer cleanupServerHub(t, deps)

	// Test OPTIONS request (CORS preflight)
	req := httptest.NewRequest("OPTIONS", "/api/v1/movies", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 204 and proper CORS headers
	assert.Equal(t, 204, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
}

func TestNewServer_StaticFiles(t *testing.T) {
	cfg := &config.Config{
		Logging: config.LoggingConfig{
			Level: "info",
		},
		Matching: config.MatchingConfig{
			RegexEnabled: false,
		},
	}

	registry := models.NewScraperRegistry()
	mat, err := matcher.NewMatcher(&cfg.Matching)
	require.NoError(t, err)

	deps := &ServerDependencies{
		ConfigFile:  "/tmp/config.yaml",
		Registry:    registry,
		Aggregator:  aggregator.New(cfg),
		MovieRepo:   newMockMovieRepo(),
		ActressRepo: newMockActressRepo(),
		Matcher:     mat,
		JobQueue:    worker.NewJobQueue(),
	}
	// Initialize atomic config pointer
	deps.SetConfig(cfg)

	router := NewServer(deps)
	defer cleanupServerHub(t, deps)

	// Test that docs endpoint is registered
	req := httptest.NewRequest("GET", "/docs", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return HTML (even if OpenAPI file doesn't exist in test)
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "Javinizer API Documentation")
}

func TestServeScalarDocs(t *testing.T) {
	cfg := &config.Config{
		Logging: config.LoggingConfig{
			Level: "info",
		},
		Matching: config.MatchingConfig{
			RegexEnabled: false,
		},
	}

	registry := models.NewScraperRegistry()
	mat, err := matcher.NewMatcher(&cfg.Matching)
	require.NoError(t, err)

	deps := &ServerDependencies{
		ConfigFile:  "/tmp/config.yaml",
		Registry:    registry,
		Aggregator:  aggregator.New(cfg),
		MovieRepo:   newMockMovieRepo(),
		ActressRepo: newMockActressRepo(),
		Matcher:     mat,
		JobQueue:    worker.NewJobQueue(),
	}
	// Initialize atomic config pointer
	deps.SetConfig(cfg)

	router := NewServer(deps)
	defer cleanupServerHub(t, deps)

	req := httptest.NewRequest("GET", "/docs", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))

	body := w.Body.String()
	assert.Contains(t, body, "<!doctype html>")
	assert.Contains(t, body, "Javinizer API Documentation")
	assert.Contains(t, body, "@scalar/api-reference")
	assert.Contains(t, body, "/docs/openapi.json")
}

func TestLogServerInfo(t *testing.T) {
	// Test that LogServerInfo doesn't panic
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
	}

	// This should not panic
	assert.NotPanics(t, func() {
		LogServerInfo(cfg)
	})
}

func TestNewServer_GinMode(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		wantMode string
	}{
		{
			name:     "debug mode",
			logLevel: "debug",
			wantMode: "debug",
		},
		{
			name:     "release mode",
			logLevel: "info",
			wantMode: "release",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Logging: config.LoggingConfig{
					Level: tt.logLevel,
				},
				Matching: config.MatchingConfig{
					RegexEnabled: false,
				},
			}

			registry := models.NewScraperRegistry()
			mat, err := matcher.NewMatcher(&cfg.Matching)
			require.NoError(t, err)

			deps := &ServerDependencies{
				ConfigFile:  "/tmp/config.yaml",
				Registry:    registry,
				Aggregator:  aggregator.New(cfg),
				MovieRepo:   newMockMovieRepo(),
				ActressRepo: newMockActressRepo(),
				Matcher:     mat,
				JobQueue:    worker.NewJobQueue(),
			}
			// Initialize atomic config pointer
			deps.SetConfig(cfg)

			router := NewServer(deps)
			defer cleanupServerHub(t, deps)
			require.NotNil(t, router)

			// Router should be created without panic
		})
	}
}

func TestNewServer_AllEndpointsAccessible(t *testing.T) {
	// Integration test - verify all endpoints are accessible (no 404)
	cfg := &config.Config{
		Logging: config.LoggingConfig{
			Level: "info",
		},
		Matching: config.MatchingConfig{
			RegexEnabled: false,
		},
		Scrapers: config.ScrapersConfig{
			Priority: []string{"r18dev"},
		},
	}

	registry := models.NewScraperRegistry()
	registry.Register(&mockScraper{name: "r18dev", enabled: true})

	mat, err := matcher.NewMatcher(&cfg.Matching)
	require.NoError(t, err)

	// Use in-memory database for testing
	dbCfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "sqlite",
			DSN:  ":memory:",
		},
		Logging: config.LoggingConfig{
			Level: "error",
		},
	}
	db, err := database.New(dbCfg)
	require.NoError(t, err)

	deps := &ServerDependencies{
		ConfigFile:  "/tmp/config.yaml",
		Registry:    registry,
		DB:          db,
		Aggregator:  aggregator.New(cfg),
		MovieRepo:   database.NewMovieRepository(db),
		ActressRepo: database.NewActressRepository(db),
		Matcher:     mat,
		JobQueue:    worker.NewJobQueue(),
	}
	// Initialize atomic config pointer
	deps.SetConfig(cfg)

	router := NewServer(deps)
	defer cleanupServerHub(t, deps)

	// Test GET endpoints
	getEndpoints := []string{
		"/health",
		"/api/v1/movies",
		"/api/v1/config",
		"/api/v1/scrapers",
		"/api/v1/cwd",
		"/docs",
	}

	for _, endpoint := range getEndpoints {
		t.Run("GET "+endpoint, func(t *testing.T) {
			req := httptest.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Should not return 404
			assert.NotEqual(t, 404, w.Code, "Endpoint %s should exist", endpoint)
		})
	}
}

func TestNewServer_SecurityHeaders(t *testing.T) {
	// Test that server properly handles security-related scenarios
	cfg := &config.Config{
		API: config.APIConfig{
			Security: config.SecurityConfig{
				AllowedOrigins: []string{"*"}, // Allow all origins for test
			},
		},
		Logging: config.LoggingConfig{
			Level: "info",
		},
		Matching: config.MatchingConfig{
			RegexEnabled: false,
		},
	}

	registry := models.NewScraperRegistry()
	mat, err := matcher.NewMatcher(&cfg.Matching)
	require.NoError(t, err)

	deps := &ServerDependencies{
		ConfigFile:  "/tmp/config.yaml",
		Registry:    registry,
		Aggregator:  aggregator.New(cfg),
		MovieRepo:   newMockMovieRepo(),
		ActressRepo: newMockActressRepo(),
		Matcher:     mat,
		JobQueue:    worker.NewJobQueue(),
	}
	// Initialize atomic config pointer
	deps.SetConfig(cfg)

	router := NewServer(deps)
	defer cleanupServerHub(t, deps)

	t.Run("CORS allows all origins", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		req.Header.Set("Origin", "http://evil.com")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// CORS is wide open (for development)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("Large request body handling", func(t *testing.T) {
		// Test that server can handle large request bodies without crashing
		largeBody := strings.Repeat("x", 1024*1024) // 1MB
		req := httptest.NewRequest("POST", "/api/v1/scrape", strings.NewReader(largeBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return 400 (bad request) not 500 (server error)
		assert.True(t, w.Code == 400 || w.Code == 413, "Should handle large body gracefully")
	})
}

func TestNewServer_InvalidRoutes(t *testing.T) {
	cfg := &config.Config{
		Logging: config.LoggingConfig{
			Level: "info",
		},
		Matching: config.MatchingConfig{
			RegexEnabled: false,
		},
	}

	registry := models.NewScraperRegistry()
	mat, err := matcher.NewMatcher(&cfg.Matching)
	require.NoError(t, err)

	deps := &ServerDependencies{
		ConfigFile:  "/tmp/config.yaml",
		Registry:    registry,
		Aggregator:  aggregator.New(cfg),
		MovieRepo:   newMockMovieRepo(),
		ActressRepo: newMockActressRepo(),
		Matcher:     mat,
		JobQueue:    worker.NewJobQueue(),
	}
	// Initialize atomic config pointer
	deps.SetConfig(cfg)

	router := NewServer(deps)
	defer cleanupServerHub(t, deps)

	invalidRoutes := []string{
		"/nonexistent",
		"/api/v1/invalid",
		"/api/v2/movies",
		"/../../../etc/passwd",
	}

	for _, route := range invalidRoutes {
		t.Run("Invalid:"+route, func(t *testing.T) {
			req := httptest.NewRequest("GET", route, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Should return 404
			assert.Equal(t, 404, w.Code)
		})
	}
}
