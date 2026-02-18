package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveScraperProxy(t *testing.T) {
	global := ProxyConfig{
		Enabled:        true,
		URL:            "http://global-proxy.example.com:8080",
		Username:       "global-user",
		Password:       "global-pass",
		DefaultProfile: "main",
		Profiles: map[string]ProxyProfile{
			"main": {
				URL:      "http://main-proxy.example.com:8080",
				Username: "main-user",
				Password: "main-pass",
			},
			"backup": {
				URL:      "http://backup-proxy.example.com:8080",
				Username: "backup-user",
				Password: "backup-pass",
			},
		},
		FlareSolverr: FlareSolverrConfig{
			Enabled:    true,
			URL:        "http://global-fs:8191/v1",
			Timeout:    30,
			MaxRetries: 3,
			SessionTTL: 300,
		},
	}

	t.Run("uses global when override is nil", func(t *testing.T) {
		resolved := ResolveScraperProxy(global, nil)
		assert.Equal(t, "http://main-proxy.example.com:8080", resolved.URL)
		assert.Equal(t, "main-user", resolved.Username)
		assert.Equal(t, "main-pass", resolved.Password)
	})

	t.Run("uses scraper override when provided", func(t *testing.T) {
		override := &ProxyConfig{
			Enabled: true,
			URL:     "http://scraper-proxy.example.com:8080",
			FlareSolverr: FlareSolverrConfig{
				Enabled:    true,
				URL:        "http://scraper-fs:8191/v1",
				Timeout:    45,
				MaxRetries: 2,
				SessionTTL: 600,
			},
		}

		resolved := ResolveScraperProxy(global, override)
		assert.Equal(t, *override, *resolved)
	})

	t.Run("reuses global proxy when use_main_proxy is set", func(t *testing.T) {
		override := &ProxyConfig{
			Enabled:      true,
			UseMainProxy: true,
		}

		resolved := ResolveScraperProxy(global, override)
		assert.Equal(t, "http://main-proxy.example.com:8080", resolved.URL)
		assert.Equal(t, "main-user", resolved.Username)
		assert.Equal(t, "main-pass", resolved.Password)
	})

	t.Run("inherits global flaresolverr when scraper override omits it", func(t *testing.T) {
		override := &ProxyConfig{
			Enabled:  true,
			URL:      "http://scraper-proxy.example.com:8080",
			Username: "scraper-user",
			Password: "scraper-pass",
			// FlareSolverr block intentionally omitted (zero-value)
		}

		resolved := ResolveScraperProxy(global, override)
		assert.Equal(t, override.Enabled, resolved.Enabled)
		assert.Equal(t, override.URL, resolved.URL)
		assert.Equal(t, override.Username, resolved.Username)
		assert.Equal(t, override.Password, resolved.Password)
		assert.Equal(t, global.FlareSolverr, resolved.FlareSolverr)
	})

	t.Run("uses scraper profile when provided", func(t *testing.T) {
		override := &ProxyConfig{
			Enabled: true,
			Profile: "backup",
		}

		resolved := ResolveScraperProxy(global, override)
		assert.Equal(t, "http://backup-proxy.example.com:8080", resolved.URL)
		assert.Equal(t, "backup-user", resolved.Username)
		assert.Equal(t, "backup-pass", resolved.Password)
	})

	t.Run("inherits global proxy URL and credentials when enabled override omits URL", func(t *testing.T) {
		override := &ProxyConfig{
			Enabled: true,
			// URL/credentials intentionally omitted
		}

		resolved := ResolveScraperProxy(global, override)
		assert.Equal(t, true, resolved.Enabled)
		assert.Equal(t, "http://main-proxy.example.com:8080", resolved.URL)
		assert.Equal(t, "main-user", resolved.Username)
		assert.Equal(t, "main-pass", resolved.Password)
		assert.Equal(t, global.FlareSolverr, resolved.FlareSolverr)
	})

	t.Run("reuses global proxy but allows flaresolverr override", func(t *testing.T) {
		override := &ProxyConfig{
			Enabled:      true,
			UseMainProxy: true,
			FlareSolverr: FlareSolverrConfig{
				Enabled:    true,
				URL:        "http://scraper-fs:8191/v1",
				Timeout:    40,
				MaxRetries: 2,
				SessionTTL: 500,
			},
		}

		resolved := ResolveScraperProxy(global, override)
		assert.Equal(t, global.Enabled, resolved.Enabled)
		assert.Equal(t, "http://main-proxy.example.com:8080", resolved.URL)
		assert.Equal(t, "main-user", resolved.Username)
		assert.Equal(t, "main-pass", resolved.Password)
		assert.Equal(t, override.FlareSolverr, resolved.FlareSolverr)
	})
}
