package javdb

import (
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/httpclient"
	"github.com/javinizer/javinizer-go/internal/logging"
)

// NewHTTPClient creates an HTTP client and FlareSolverr for the JavDB scraper.
// HTTP-01, HTTP-03: Per-scraper HTTP client and FlareSolverr ownership.
// Returns client, flaresolverr, and error.
func NewHTTPClient(cfg *config.ScraperConfig, globalProxy *config.ProxyConfig) (*resty.Client, *httpclient.FlareSolverr, error) {
	// Resolve proxy per-scraper (HTTP-02)
	proxyCfg := config.ResolveScraperProxy(*globalProxy, cfg.Proxy)

	// Use timeout from ScraperConfig, default to 30s
	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// Use RetryCount from ScraperConfig, default to 3
	retryCount := cfg.RetryCount
	if retryCount == 0 {
		retryCount = 3
	}

	var fs *httpclient.FlareSolverr
	var err error

	// Check for FlareSolverr on ScraperConfig directly (HTTP-03)
	if cfg.FlareSolverr.Enabled {
		// Build a ProxyConfig with FlareSolverr for the factory function
		proxyWithFlareSolverr := &config.ProxyConfig{
			Enabled:      proxyCfg.Enabled,
			URL:          proxyCfg.URL,
			Username:     proxyCfg.Username,
			Password:     proxyCfg.Password,
			FlareSolverr: cfg.FlareSolverr,
		}
		_, fs, err = httpclient.NewRestyClientWithFlareSolverr(
			proxyWithFlareSolverr,
			timeout,
			retryCount,
		)
		if err != nil {
			logging.Errorf("JavDB: Failed to create FlareSolverr: %v", err)
			// Fall through to create client without FlareSolverr
		}
	}

	// Create client (with or without FlareSolverr)
	client, err := httpclient.NewRestyClient(proxyCfg, timeout, retryCount)
	if err != nil {
		return nil, nil, err
	}

	// Apply UserAgent from ScraperConfig
	userAgent := config.ResolveScraperUserAgent(
		cfg.UserAgent,
		cfg.UseFakeUserAgent,
		cfg.UserAgent,
	)
	client.SetHeader("User-Agent", userAgent)
	client.SetHeader("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	client.SetHeader("Accept-Language", "en-US,en;q=0.9,ja;q=0.8,zh;q=0.7")
	client.SetHeader("Connection", "keep-alive")
	client.SetHeader("Upgrade-Insecure-Requests", "1")

	return client, fs, nil
}
