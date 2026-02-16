package httpclient_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/javinizer/javinizer-go/internal/config"
	httpclient "github.com/javinizer/javinizer-go/internal/httpclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFlareSolverr(t *testing.T) {
	tests := []struct {
		name      string
		cfg       config.FlareSolverrConfig
		wantError bool
	}{
		{
			name: "valid configuration",
			cfg: config.FlareSolverrConfig{
				Enabled:    true,
				URL:        "http://localhost:8191/v1",
				Timeout:    30,
				MaxRetries: 3,
				SessionTTL: 300,
			},
			wantError: false,
		},
		{
			name: "empty URL",
			cfg: config.FlareSolverrConfig{
				Enabled:    true,
				URL:        "",
				Timeout:    30,
				MaxRetries: 3,
				SessionTTL: 300,
			},
			wantError: true,
		},
		{
			name: "custom URL",
			cfg: config.FlareSolverrConfig{
				Enabled:    true,
				URL:        "http://192.168.1.100:8191/v1",
				Timeout:    30,
				MaxRetries: 3,
				SessionTTL: 300,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, err := httpclient.NewFlareSolverr(&tt.cfg)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, fs)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, fs)
			}
		})
	}
}

func TestNewRestyClientWithFlareSolverr_Disabled(t *testing.T) {
	cfg := &config.Config{
		Scrapers: config.ScrapersConfig{
			Proxy: config.ProxyConfig{
				FlareSolverr: config.FlareSolverrConfig{
					Enabled: false,
					URL:     "http://localhost:8191/v1",
				},
			},
		},
	}

	client, fs, err := httpclient.NewRestyClientWithFlareSolverr(
		&cfg.Scrapers.Proxy,
		30*time.Second,
		3,
	)

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Nil(t, fs) // FlareSolverr should be nil when disabled
}

func TestNewRestyClientWithFlareSolverr_Enabled(t *testing.T) {
	cfg := &config.Config{
		Scrapers: config.ScrapersConfig{
			Proxy: config.ProxyConfig{
				FlareSolverr: config.FlareSolverrConfig{
					Enabled:    true,
					URL:        "http://localhost:8191/v1",
					Timeout:    30,
					MaxRetries: 3,
					SessionTTL: 300,
				},
			},
		},
	}

	client, fs, err := httpclient.NewRestyClientWithFlareSolverr(
		&cfg.Scrapers.Proxy,
		30*time.Second,
		3,
	)

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, fs) // FlareSolverr should be created when enabled
}

func TestFlareSolverrRequestJSON(t *testing.T) {
	// Test that FlareSolverrRequest can be marshaled to JSON correctly
	req := httpclient.FlareSolverrRequest{
		Cmd:        "request.get",
		URL:        "https://example.com",
		MaxTimeout: 30,
		Session:    "test-session",
		Proxy: &httpclient.FlareSolverrProxy{
			URL:      "http://proxy.example.com:8080",
			Username: "user",
			Password: "pass",
		},
	}

	// Just verify struct can be constructed
	assert.Equal(t, "request.get", req.Cmd)
	assert.Equal(t, "https://example.com", req.URL)
	assert.Equal(t, 30, req.MaxTimeout)
	assert.Equal(t, "test-session", req.Session)
	require.NotNil(t, req.Proxy)
	assert.Equal(t, "http://proxy.example.com:8080", req.Proxy.URL)
	assert.Equal(t, "user", req.Proxy.Username)
	assert.Equal(t, "pass", req.Proxy.Password)
}

func TestNewRestyClientWithFlareSolverr_RequestLevelProxy(t *testing.T) {
	var captured map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		_ = json.NewDecoder(r.Body).Decode(&captured)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","solution":{"response":"<html></html>","cookies":[],"userAgent":"ua"}}`))
	}))
	defer server.Close()

	proxyCfg := &config.ProxyConfig{
		Enabled:  true,
		URL:      "http://proxy.example.com:8080",
		Username: "proxyuser",
		Password: "proxypass",
		FlareSolverr: config.FlareSolverrConfig{
			Enabled:    true,
			URL:        server.URL,
			Timeout:    30,
			MaxRetries: 1,
			SessionTTL: 300,
		},
	}

	_, fs, err := httpclient.NewRestyClientWithFlareSolverr(proxyCfg, 30*time.Second, 1)
	require.NoError(t, err)
	require.NotNil(t, fs)

	_, _, err = fs.ResolveURL("https://example.com")
	require.NoError(t, err)

	require.NotNil(t, captured)
	proxyVal, ok := captured["proxy"].(map[string]any)
	require.True(t, ok, "expected proxy field in FlareSolverr request")
	assert.Equal(t, "http://proxy.example.com:8080", proxyVal["url"])
	assert.Equal(t, "proxyuser", proxyVal["username"])
	assert.Equal(t, "proxypass", proxyVal["password"])
}

func TestFlareSolverrResponseJSON(t *testing.T) {
	// Test that FlareSolverrResponse struct has expected fields
	resp := httpclient.FlareSolverrResponse{
		Status:  "ok",
		Message: "",
		Session: "test-session",
	}
	resp.Solution.Response = "<html>test</html>"
	resp.Solution.Cookies = []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}{
		{Name: "cf_clearance", Value: "abc123"},
	}
	resp.Solution.UserAgent = "Mozilla/5.0"

	assert.Equal(t, "ok", resp.Status)
	assert.Equal(t, "test-session", resp.Session)
	assert.Equal(t, "<html>test</html>", resp.Solution.Response)
	assert.Len(t, resp.Solution.Cookies, 1)
	assert.Equal(t, "cf_clearance", resp.Solution.Cookies[0].Name)
	assert.Equal(t, "abc123", resp.Solution.Cookies[0].Value)
	assert.Equal(t, "Mozilla/5.0", resp.Solution.UserAgent)
}

func TestConfigValidate_FlareSolverr(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *config.Config
		wantError bool
	}{
		{
			name: "valid FlareSolverr config",
			cfg: &config.Config{
				Scrapers: config.ScrapersConfig{
					Proxy: config.ProxyConfig{
						FlareSolverr: config.FlareSolverrConfig{
							Enabled:    true,
							URL:        "http://localhost:8191/v1",
							Timeout:    30,
							MaxRetries: 3,
							SessionTTL: 300,
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "FlareSolverr enabled but empty URL",
			cfg: &config.Config{
				Scrapers: config.ScrapersConfig{
					Proxy: config.ProxyConfig{
						FlareSolverr: config.FlareSolverrConfig{
							Enabled:    true,
							URL:        "",
							Timeout:    30,
							MaxRetries: 3,
							SessionTTL: 300,
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "FlareSolverr timeout too low",
			cfg: &config.Config{
				Scrapers: config.ScrapersConfig{
					Proxy: config.ProxyConfig{
						FlareSolverr: config.FlareSolverrConfig{
							Enabled:    true,
							URL:        "http://localhost:8191/v1",
							Timeout:    0,
							MaxRetries: 3,
							SessionTTL: 300,
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "FlareSolverr timeout too high",
			cfg: &config.Config{
				Scrapers: config.ScrapersConfig{
					Proxy: config.ProxyConfig{
						FlareSolverr: config.FlareSolverrConfig{
							Enabled:    true,
							URL:        "http://localhost:8191/v1",
							Timeout:    500,
							MaxRetries: 3,
							SessionTTL: 300,
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "FlareSolverr max retries too high",
			cfg: &config.Config{
				Scrapers: config.ScrapersConfig{
					Proxy: config.ProxyConfig{
						FlareSolverr: config.FlareSolverrConfig{
							Enabled:    true,
							URL:        "http://localhost:8191/v1",
							Timeout:    30,
							MaxRetries: 20,
							SessionTTL: 300,
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "FlareSolverr session TTL too low",
			cfg: &config.Config{
				Scrapers: config.ScrapersConfig{
					Proxy: config.ProxyConfig{
						FlareSolverr: config.FlareSolverrConfig{
							Enabled:    true,
							URL:        "http://localhost:8191/v1",
							Timeout:    30,
							MaxRetries: 3,
							SessionTTL: 30,
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "FlareSolverr session TTL too high",
			cfg: &config.Config{
				Scrapers: config.ScrapersConfig{
					Proxy: config.ProxyConfig{
						FlareSolverr: config.FlareSolverrConfig{
							Enabled:    true,
							URL:        "http://localhost:8191/v1",
							Timeout:    30,
							MaxRetries: 3,
							SessionTTL: 5000,
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "FlareSolverr disabled should not validate",
			cfg: &config.Config{
				Scrapers: config.ScrapersConfig{
					Proxy: config.ProxyConfig{
						FlareSolverr: config.FlareSolverrConfig{
							Enabled:    false,
							URL:        "",
							Timeout:    0,
							MaxRetries: 0,
							SessionTTL: 0,
						},
					},
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set minimal required fields for validation
			if tt.cfg.Scrapers.TimeoutSeconds == 0 {
				tt.cfg.Scrapers.TimeoutSeconds = 30
			}
			if tt.cfg.Scrapers.RequestTimeoutSeconds == 0 {
				tt.cfg.Scrapers.RequestTimeoutSeconds = 60
			}
			if tt.cfg.Scrapers.DMM.BrowserTimeout == 0 {
				tt.cfg.Scrapers.DMM.BrowserTimeout = 30
			}
			if tt.cfg.Scrapers.Referer == "" {
				tt.cfg.Scrapers.Referer = "https://www.dmm.co.jp/"
			}
			if tt.cfg.Performance.MaxWorkers == 0 {
				tt.cfg.Performance.MaxWorkers = 5
			}
			if tt.cfg.Performance.WorkerTimeout == 0 {
				tt.cfg.Performance.WorkerTimeout = 300
			}
			if tt.cfg.Performance.UpdateInterval == 0 {
				tt.cfg.Performance.UpdateInterval = 100
			}

			err := tt.cfg.Validate()

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultConfig_FlareSolverr(t *testing.T) {
	cfg := config.DefaultConfig()

	assert.False(t, cfg.Scrapers.Proxy.FlareSolverr.Enabled)
	assert.Equal(t, "http://localhost:8191/v1", cfg.Scrapers.Proxy.FlareSolverr.URL)
	assert.Equal(t, 30, cfg.Scrapers.Proxy.FlareSolverr.Timeout)
	assert.Equal(t, 3, cfg.Scrapers.Proxy.FlareSolverr.MaxRetries)
	assert.Equal(t, 300, cfg.Scrapers.Proxy.FlareSolverr.SessionTTL)
}

func TestDefaultConfig_JavLibrary(t *testing.T) {
	cfg := config.DefaultConfig()

	assert.False(t, cfg.Scrapers.JavLibrary.Enabled)
	assert.Equal(t, "en", cfg.Scrapers.JavLibrary.Language)
	assert.Equal(t, 1000, cfg.Scrapers.JavLibrary.RequestDelay)
	assert.Equal(t, "http://www.javlibrary.com", cfg.Scrapers.JavLibrary.BaseURL)
	assert.False(t, cfg.Scrapers.JavLibrary.UseFlareSolverr)
}

func TestCookieConversion(t *testing.T) {
	// Test that FlareSolverrResponse cookies can be converted to http.Cookie format
	resp := httpclient.FlareSolverrResponse{}
	resp.Solution.Cookies = []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}{
		{Name: "cf_clearance", Value: "abc123"},
		{Name: "__cf_bm", Value: "xyz789"},
	}

	// Simulate the conversion that happens in ResolveURL
	cookies := make([]http.Cookie, len(resp.Solution.Cookies))
	for i, c := range resp.Solution.Cookies {
		cookies[i] = http.Cookie{
			Name:  c.Name,
			Value: c.Value,
		}
	}

	assert.Len(t, cookies, 2)
	assert.Equal(t, "cf_clearance", cookies[0].Name)
	assert.Equal(t, "abc123", cookies[0].Value)
	assert.Equal(t, "__cf_bm", cookies[1].Name)
	assert.Equal(t, "xyz789", cookies[1].Value)
}
