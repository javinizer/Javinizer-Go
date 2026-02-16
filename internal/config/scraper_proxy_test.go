package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveScraperProxy(t *testing.T) {
	global := ProxyConfig{
		Enabled:  true,
		URL:      "http://global-proxy.example.com:8080",
		Username: "global-user",
		Password: "global-pass",
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
		assert.Equal(t, global, *resolved)
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
}
