package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeProxyTestConfig(t *testing.T, yamlContent string) string {
	t.Helper()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	return cfgPath
}

func TestProxyConfig_ProfileBasedLoading(t *testing.T) {
	yamlContent := `
config_version: 2
scrapers:
  proxy:
    enabled: true
    default_profile: "main"
    profiles:
      main:
        url: "http://proxy.example.com:8080"
        username: "user"
        password: "pass"
        flaresolverr:
          enabled: true
          url: "http://flaresolverr-main:8191/v1"
          timeout: 30
          max_retries: 3
          session_ttl: 300
      download:
        url: "socks5://localhost:1080"
        username: "dl-user"
        password: "dl-pass"
output:
  download_proxy:
    enabled: true
    profile: "download"
`

	cfgPath := writeProxyTestConfig(t, yamlContent)
	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected profile-based proxy config to validate, got: %v", err)
	}

	if !cfg.Scrapers.Proxy.Enabled {
		t.Fatal("expected scraper proxy to be enabled")
	}
	if cfg.Scrapers.Proxy.DefaultProfile != "main" {
		t.Fatalf("expected default_profile to be main, got %q", cfg.Scrapers.Proxy.DefaultProfile)
	}
	if cfg.Output.DownloadProxy.Profile != "download" {
		t.Fatalf("expected output.download_proxy.profile to be download, got %q", cfg.Output.DownloadProxy.Profile)
	}

	globalResolved := ResolveGlobalProxy(cfg.Scrapers.Proxy)
	if globalResolved.URL != "http://proxy.example.com:8080" {
		t.Fatalf("expected resolved global proxy URL, got %q", globalResolved.URL)
	}
	if globalResolved.Username != "user" || globalResolved.Password != "pass" {
		t.Fatalf("expected resolved global proxy credentials, got %q/%q", globalResolved.Username, globalResolved.Password)
	}

	downloadResolved := ResolveScraperProxy(cfg.Scrapers.Proxy, &cfg.Output.DownloadProxy)
	if downloadResolved.URL != "socks5://localhost:1080" {
		t.Fatalf("expected resolved download proxy URL, got %q", downloadResolved.URL)
	}
	if downloadResolved.Username != "dl-user" || downloadResolved.Password != "dl-pass" {
		t.Fatalf("expected resolved download proxy credentials, got %q/%q", downloadResolved.Username, downloadResolved.Password)
	}
}

func TestProxyConfig_Disabled(t *testing.T) {
	yamlContent := `
config_version: 2
scrapers:
  proxy:
    enabled: false
output:
  download_proxy:
    enabled: false
`

	cfgPath := writeProxyTestConfig(t, yamlContent)
	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected disabled proxy config to validate, got: %v", err)
	}

	if cfg.Scrapers.Proxy.Enabled {
		t.Error("expected scraper proxy to be disabled")
	}
	if cfg.Output.DownloadProxy.Enabled {
		t.Error("expected download proxy to be disabled")
	}
}

func TestProxyConfig_DefaultValues(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Scrapers.Proxy.Enabled {
		t.Error("expected scraper proxy to be disabled by default")
	}
	if cfg.Scrapers.Proxy.DefaultProfile != "" {
		t.Errorf("expected empty default profile by default, got %q", cfg.Scrapers.Proxy.DefaultProfile)
	}
	if cfg.Output.DownloadProxy.Enabled {
		t.Error("expected download proxy to be disabled by default")
	}
	if cfg.Output.DownloadProxy.Profile != "" {
		t.Errorf("expected empty download proxy profile by default, got %q", cfg.Output.DownloadProxy.Profile)
	}
}

func TestProxyConfig_EnabledRequiresDefaultProfile(t *testing.T) {
	yamlContent := `
config_version: 2
scrapers:
  proxy:
    enabled: true
    profiles:
      main:
        url: "http://proxy.example.com:8080"
`

	cfgPath := writeProxyTestConfig(t, yamlContent)
	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	err = cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error when scrapers.proxy.enabled=true and default_profile is missing")
	}
	if !strings.Contains(err.Error(), "scrapers.proxy.default_profile is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProxyConfig_LegacyDirectFieldsRejected(t *testing.T) {
	yamlContent := `
config_version: 2
scrapers:
  proxy:
    enabled: true
    default_profile: "main"
    profiles:
      main:
        url: "http://proxy.example.com:8080"
    url: "http://legacy-direct-proxy.example.com:8080"
`

	cfgPath := writeProxyTestConfig(t, yamlContent)
	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	err = cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error for legacy direct proxy fields")
	}
	if !strings.Contains(err.Error(), "direct proxy fields (url/username/password) are no longer supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProxyConfig_LegacyUseMainProxyRejected(t *testing.T) {
	yamlContent := `
config_version: 2
scrapers:
  proxy:
    enabled: true
    default_profile: "main"
    profiles:
      main:
        url: "http://proxy.example.com:8080"
  dmm:
    proxy:
      enabled: true
      use_main_proxy: true
`

	cfgPath := writeProxyTestConfig(t, yamlContent)
	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	err = cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error for use_main_proxy")
	}
	if !strings.Contains(err.Error(), "use_main_proxy is no longer supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}
