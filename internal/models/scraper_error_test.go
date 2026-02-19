package models

import (
	"errors"
	"testing"
)

func TestNewScraperStatusErrorClassifiesKinds(t *testing.T) {
	tests := []struct {
		status int
		want   ScraperErrorKind
	}{
		{status: 404, want: ScraperErrorKindNotFound},
		{status: 429, want: ScraperErrorKindRateLimited},
		{status: 403, want: ScraperErrorKindBlocked},
		{status: 502, want: ScraperErrorKindUnavailable},
		{status: 418, want: ScraperErrorKindUnknown},
	}

	for _, tt := range tests {
		err := NewScraperStatusError("Test", tt.status, "")
		if err.Kind != tt.want {
			t.Fatalf("status %d -> kind %q, want %q", tt.status, err.Kind, tt.want)
		}
	}
}

func TestAsScraperErrorUnwrap(t *testing.T) {
	inner := NewScraperStatusError("Test", 502, "upstream bad gateway")
	outer := errors.New("wrapped: " + inner.Error())

	if got, ok := AsScraperError(outer); ok || got != nil {
		t.Fatalf("plain wrapped string should not match ScraperError")
	}

	wrapped := errors.Join(errors.New("context"), inner)
	got, ok := AsScraperError(wrapped)
	if !ok || got == nil {
		t.Fatalf("expected AsScraperError to find joined scraper error")
	}
	if got.StatusCode != 502 {
		t.Fatalf("unexpected status code: %d", got.StatusCode)
	}
}
