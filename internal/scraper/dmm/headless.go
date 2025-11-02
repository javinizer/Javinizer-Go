package dmm

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/httpclient"
	"github.com/javinizer/javinizer-go/internal/logging"
)

// basicAuth creates a Basic Authentication header value
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// FetchWithHeadless fetches a URL using headless Chrome with age verification cookies
func FetchWithHeadless(url string, timeout int, proxyConfig *config.ProxyConfig) (string, error) {
	if timeout <= 0 {
		timeout = 30 // Default timeout
	}

	logging.Debugf("DMM Headless: Starting headless browser for %s (timeout: %ds)", url, timeout)

	// Create allocator options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("disable-extensions", true),
	)

	// Only override Chrome binary path if explicitly set via environment variable
	// This allows chromedp to use its built-in discovery on dev machines while
	// respecting the container's CHROME_BIN setting in production
	chromeBin := os.Getenv("CHROME_BIN")
	if chromeBin == "" {
		chromeBin = os.Getenv("CHROME_PATH")
	}
	if chromeBin != "" {
		logging.Debugf("DMM Headless: Using explicit Chrome binary: %s", chromeBin)
		opts = append(opts, chromedp.ExecPath(chromeBin))
	}

	// Add proxy if configured
	if proxyConfig != nil && proxyConfig.Enabled && proxyConfig.URL != "" {
		sanitizedURL := httpclient.SanitizeProxyURL(proxyConfig.URL)
		logging.Debugf("DMM Headless: Using proxy %s", sanitizedURL)
		opts = append(opts, chromedp.ProxyServer(proxyConfig.URL))
	}

	// Create context with custom allocator
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Set timeout
	ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	var bodyHTML string

	// Prepare actions list
	actions := []chromedp.Action{
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Set cookies for age verification
			logging.Debug("DMM Headless: Setting age verification cookies")
			expr := network.SetCookie("age_check_done", "1").
				WithDomain(".dmm.co.jp")
			if err := expr.Do(ctx); err != nil {
				return fmt.Errorf("failed to set age_check_done cookie: %w", err)
			}
			expr = network.SetCookie("cklg", "ja").
				WithDomain(".dmm.co.jp")
			if err := expr.Do(ctx); err != nil {
				return fmt.Errorf("failed to set cklg cookie: %w", err)
			}

			// Set proxy authentication if proxy is enabled and credentials are provided
			if proxyConfig != nil && proxyConfig.Enabled && proxyConfig.URL != "" &&
				proxyConfig.Username != "" && proxyConfig.Password != "" {
				logging.Debug("DMM Headless: Setting proxy authentication credentials")
				// Use CDP Network domain to set proxy auth credentials
				// Note: This only works for HTTP/HTTPS proxies with Basic auth
				// SOCKS5 proxies authenticate during handshake (not supported via headers)
				err := network.SetExtraHTTPHeaders(network.Headers{
					"Proxy-Authorization": "Basic " + basicAuth(proxyConfig.Username, proxyConfig.Password),
				}).Do(ctx)
				if err != nil {
					return fmt.Errorf("failed to set proxy authentication: %w", err)
				}
			}
			return nil
		}),
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.Sleep(3 * time.Second), // Wait for JavaScript rendering
		chromedp.InnerHTML("body", &bodyHTML),
	}

	// Run chromedp actions
	err := chromedp.Run(ctx, actions...)
	if err != nil {
		return "", fmt.Errorf("headless browser failed: %w", err)
	}

	logging.Debugf("DMM Headless: ✓ Successfully fetched %d characters", len(bodyHTML))
	return bodyHTML, nil
}
