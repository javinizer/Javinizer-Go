package worker

import (
	"errors"
	"strings"
	"testing"

	"github.com/javinizer/javinizer-go/internal/models"
)

func TestFormatScraperFailure502(t *testing.T) {
	msg := formatScraperFailure("libredmm", errors.New("LibreDMM returned status code 502"))

	if !strings.Contains(msg, "temporarily unavailable") {
		t.Fatalf("expected unavailable message, got: %s", msg)
	}
	if !strings.Contains(msg, "host may be down") {
		t.Fatalf("expected host down hint, got: %s", msg)
	}
	if !strings.Contains(msg, "502") {
		t.Fatalf("expected status code in message, got: %s", msg)
	}
}

func TestFormatScraperFailure502_PassthroughReadableMessage(t *testing.T) {
	msg := formatScraperFailure(
		"libredmm",
		models.NewScraperStatusError(
			"LibreDMM",
			502,
			"LibreDMM is temporarily unavailable (HTTP 502 Bad Gateway; host may be down)",
		),
	)

	if !strings.Contains(msg, "libredmm: LibreDMM is temporarily unavailable") {
		t.Fatalf("expected passthrough message, got: %s", msg)
	}
	if strings.Contains(msg, "details:") {
		t.Fatalf("expected no nested details wrapper, got: %s", msg)
	}
}

func TestBuildScraperNoResultsError_NotFoundOnly(t *testing.T) {
	msg := buildScraperNoResultsError([]scraperFailure{
		{Scraper: "r18dev", Err: errors.New("movie ABW-102 not found on R18.dev")},
		{Scraper: "dmm", Err: errors.New("movie ABW-102 not found on DMM")},
	})

	if !strings.HasPrefix(msg, "Movie not found on configured scrapers:") {
		t.Fatalf("expected not-found prefix, got: %s", msg)
	}
}

func TestBuildScraperNoResultsError_AvailabilityIssues(t *testing.T) {
	msg := buildScraperNoResultsError([]scraperFailure{
		{Scraper: "libredmm", Err: errors.New("LibreDMM returned status code 502")},
		{Scraper: "r18dev", Err: errors.New("R18.dev returned status code 503")},
	})

	if !strings.HasPrefix(msg, "Movie lookup failed due to source availability issues:") {
		t.Fatalf("expected availability prefix, got: %s", msg)
	}
	if !strings.Contains(msg, "source temporarily unavailable") {
		t.Fatalf("expected unavailable detail, got: %s", msg)
	}
}

func TestBuildScraperNoResultsError_RateLimited(t *testing.T) {
	msg := buildScraperNoResultsError([]scraperFailure{
		{Scraper: "javdb", Err: errors.New("JavDB returned status code 429")},
	})

	if !strings.Contains(msg, "rate-limited") {
		t.Fatalf("expected rate-limit detail, got: %s", msg)
	}
	if !strings.Contains(msg, "429") {
		t.Fatalf("expected status code in rate-limit detail, got: %s", msg)
	}
}
