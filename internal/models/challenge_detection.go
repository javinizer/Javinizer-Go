package models

import "strings"

// IsCloudflareChallengePage detects Cloudflare anti-bot/interstitial challenge pages
// from HTML content. These pages are often returned with HTTP 200.
func IsCloudflareChallengePage(body string) bool {
	lower := strings.ToLower(strings.TrimSpace(body))
	if lower == "" {
		return false
	}

	highConfidence := []string{
		"/cdn-cgi/challenge-platform",
		"cf-browser-verification",
		"cf_chl_",
		"cf-challenge",
		"checking your browser before accessing",
		"enable javascript and cookies to continue",
		"ddos protection by cloudflare",
	}
	for _, marker := range highConfidence {
		if strings.Contains(lower, marker) {
			return true
		}
	}

	// Fallback scoring for softer variants.
	score := 0
	for _, marker := range []string{
		"cloudflare",
		"attention required",
		"just a moment",
		"ray id",
		"cf-ray",
		"/cdn-cgi/",
		"captcha",
		"turnstile",
	} {
		if strings.Contains(lower, marker) {
			score++
		}
	}

	return score >= 3
}
