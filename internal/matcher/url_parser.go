package matcher

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ParsedInput represents the result of parsing user input
type ParsedInput struct {
	ID          string // Extracted movie ID
	ScraperHint string // Suggested scraper ("dmm", "r18dev", or "")
	IsURL       bool   // true if input was a URL
}

// ParseInput determines if input is a URL or ID and extracts the movie ID
func ParseInput(input string) (*ParsedInput, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("input cannot be empty")
	}

	// DMM URL detection
	if strings.Contains(input, "dmm.co.jp") {
		contentID := extractDMMContentID(input)
		if contentID == "" {
			return nil, fmt.Errorf("failed to extract content ID from DMM URL")
		}
		return &ParsedInput{
			ID:          contentID,
			ScraperHint: "dmm",
			IsURL:       true,
		}, nil
	}

	// R18.dev URL detection
	if strings.Contains(input, "r18.dev") || strings.Contains(input, "r18.com") {
		id := extractR18DevID(input)
		if id == "" {
			return nil, fmt.Errorf("failed to extract ID from R18.dev URL")
		}
		return &ParsedInput{
			ID:          id,
			ScraperHint: "r18dev",
			IsURL:       true,
		}, nil
	}

	// LibreDMM URL detection
	if strings.Contains(input, "libredmm.com") {
		id := extractLibreDMMID(input)
		if id == "" {
			return nil, fmt.Errorf("failed to extract ID from LibreDMM URL")
		}
		return &ParsedInput{
			ID:          id,
			ScraperHint: "libredmm",
			IsURL:       true,
		}, nil
	}

	// Not a URL - treat as JAV ID
	return &ParsedInput{
		ID:          input,
		ScraperHint: "",
		IsURL:       false,
	}, nil
}

// extractDMMContentID extracts content ID from DMM URL
// Supports both www.dmm.co.jp (cid=) and video.dmm.co.jp (id=) formats
func extractDMMContentID(url string) string {
	// Try cid= format first (www.dmm.co.jp)
	cidRegex := regexp.MustCompile(`cid=([^/?&]+)`)
	matches := cidRegex.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}

	// Try id= format (video.dmm.co.jp)
	idRegex := regexp.MustCompile(`[?&]id=([^/?&]+)`)
	matches = idRegex.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// extractR18DevID extracts movie ID from R18.dev URL
// Example: https://r18.dev/videos/vod/movies/detail/-/id=ipx00535/
func extractR18DevID(url string) string {
	// Pattern: /id=([^/?&]+)/
	idRegex := regexp.MustCompile(`/id=([^/?&]+)`)
	matches := idRegex.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// extractLibreDMMID extracts movie ID from LibreDMM URLs.
// Supports:
//   - https://www.libredmm.com/movies/IPX-535
//   - https://www.libredmm.com/movies/IPX-535.json
//   - https://www.libredmm.com/search?q=IPX535&format=json
func extractLibreDMMID(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	if q := strings.TrimSpace(u.Query().Get("q")); q != "" {
		return q
	}

	re := regexp.MustCompile(`/movies/([^/?&]+)`)
	matches := re.FindStringSubmatch(u.Path)
	if len(matches) > 1 {
		return strings.TrimSuffix(matches[1], ".json")
	}

	return ""
}
