package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validFlareSolverrConfig() FlareSolverrConfig {
	return FlareSolverrConfig{
		Enabled:    true,
		URL:        "http://localhost:8191",
		Timeout:    30,
		MaxRetries: 2,
		SessionTTL: 300,
	}
}

func TestConfigValidate_AllScraperFlareSolverrOverrides(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Scrapers.Proxy.FlareSolverr = validFlareSolverrConfig()

	override := &ProxyConfig{FlareSolverr: validFlareSolverrConfig()}
	cfg.Scrapers.R18Dev.Proxy = override
	cfg.Scrapers.DMM.Proxy = override
	cfg.Scrapers.LibreDMM.Proxy = override
	cfg.Scrapers.MGStage.Proxy = override
	cfg.Scrapers.JavLibrary.Proxy = override
	cfg.Scrapers.JavDB.Proxy = override
	cfg.Scrapers.JavBus.Proxy = override
	cfg.Scrapers.Jav321.Proxy = override
	cfg.Scrapers.TokyoHot.Proxy = override
	cfg.Scrapers.AVEntertainment.Proxy = override
	cfg.Scrapers.DLGetchu.Proxy = override
	cfg.Scrapers.Caribbeancom.Proxy = override
	cfg.Scrapers.FC2.Proxy = override

	require.NoError(t, cfg.Validate())
}

func TestValidateHTTPBaseURL_Branches(t *testing.T) {
	require.NoError(t, validateHTTPBaseURL("openai.base_url", "https://api.openai.com/v1"))

	err := validateHTTPBaseURL("openai.base_url", "://bad")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be a valid http(s) URL")

	err = validateHTTPBaseURL("openai.base_url", "ftp://example.com")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be a valid http(s) URL")
}

func TestValidateProxyProfileConfig_Branches(t *testing.T) {
	require.NoError(t, validateProxyProfileConfig(nil))

	cfg := DefaultConfig()
	cfg.Scrapers.Proxy.Enabled = true
	cfg.Scrapers.Proxy.DefaultProfile = ""
	err := validateProxyProfileConfig(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "default_profile is required")

	cfg = DefaultConfig()
	cfg.Scrapers.Proxy.DefaultProfile = "missing"
	err = validateProxyProfileConfig(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "references unknown profile")

	cfg = DefaultConfig()
	cfg.Scrapers.Proxy.Profiles = map[string]ProxyProfile{
		"default": {URL: "http://proxy.local"},
	}
	cfg.Scrapers.Proxy.DefaultProfile = "default"
	cfg.Scrapers.R18Dev.Proxy = &ProxyConfig{Enabled: true}
	err = validateProxyProfileConfig(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "scrapers.r18dev.proxy.profile is required")

	cfg.Scrapers.R18Dev.Proxy = &ProxyConfig{Profile: "unknown"}
	err = validateProxyProfileConfig(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `references unknown profile "unknown"`)
}

func TestValidateNoLegacyProxyDirectFields(t *testing.T) {
	require.NoError(t, validateNoLegacyProxyDirectFields("scrapers.proxy", nil))

	err := validateNoLegacyProxyDirectFields("scrapers.proxy", &ProxyConfig{UseMainProxy: true})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "use_main_proxy is no longer supported")

	err = validateNoLegacyProxyDirectFields("scrapers.proxy", &ProxyConfig{URL: "http://proxy.local"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "direct proxy fields")
}

func TestApplyNamedProxyProfile_Branches(t *testing.T) {
	profiles := map[string]ProxyProfile{
		"default": {
			URL:      "http://proxy.local:8080",
			Username: "user",
			Password: "pass",
			FlareSolverr: FlareSolverrConfig{
				Enabled:    true,
				URL:        "http://localhost:8191",
				Timeout:    30,
				MaxRetries: 1,
				SessionTTL: 300,
			},
		},
	}

	target := &ProxyConfig{}
	applyNamedProxyProfile(nil, profiles, "default") // no-op
	applyNamedProxyProfile(target, profiles, "")     // no-op
	applyNamedProxyProfile(target, nil, "default")   // no-op
	applyNamedProxyProfile(target, profiles, "none") // no-op
	assert.Equal(t, "", target.URL)

	applyNamedProxyProfile(target, profiles, "default")
	assert.Equal(t, "http://proxy.local:8080", target.URL)
	assert.Equal(t, "user", target.Username)
	assert.Equal(t, "pass", target.Password)
	assert.True(t, target.FlareSolverr.Enabled)
}

func TestValidateTranslationConfig_Branches(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Metadata.Translation.Enabled = true
	cfg.Metadata.Translation.Provider = "openai"
	cfg.Metadata.Translation.OpenAI.BaseURL = "not-a-url"
	err := cfg.validateTranslationConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "openai.base_url")

	cfg = DefaultConfig()
	cfg.Metadata.Translation.Enabled = true
	cfg.Metadata.Translation.Provider = "deepl"
	cfg.Metadata.Translation.DeepL.Mode = "invalid"
	err = cfg.validateTranslationConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deepl.mode")

	cfg = DefaultConfig()
	cfg.Metadata.Translation.Enabled = true
	cfg.Metadata.Translation.Provider = "deepl"
	cfg.Metadata.Translation.DeepL.Mode = "free"
	cfg.Metadata.Translation.DeepL.BaseURL = "ftp://deepl.example"
	err = cfg.validateTranslationConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deepl.base_url")

	cfg = DefaultConfig()
	cfg.Metadata.Translation.Enabled = true
	cfg.Metadata.Translation.Provider = "google"
	cfg.Metadata.Translation.Google.Mode = "invalid"
	err = cfg.validateTranslationConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "google.mode")

	cfg = DefaultConfig()
	cfg.Metadata.Translation.Enabled = true
	cfg.Metadata.Translation.Provider = "google"
	cfg.Metadata.Translation.Google.Mode = "free"
	cfg.Metadata.Translation.Google.BaseURL = "ftp://google.example"
	err = cfg.validateTranslationConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "google.base_url")

	cfg = DefaultConfig()
	cfg.Metadata.Translation.Enabled = true
	cfg.Metadata.Translation.Provider = "unknown"
	err = cfg.validateTranslationConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "provider must be one of")
}
