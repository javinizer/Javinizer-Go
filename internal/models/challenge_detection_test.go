package models

import "testing"

func TestIsCloudflareChallengePage(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{
			name: "cloudflare challenge platform marker",
			in:   `<html><head></head><body><script src="/cdn-cgi/challenge-platform/h/b/orchestrate/chl_page/v1"></script></body></html>`,
			want: true,
		},
		{
			name: "checking browser marker",
			in:   `<html><title>Just a moment...</title><body>Checking your browser before accessing example.com</body></html>`,
			want: true,
		},
		{
			name: "normal page",
			in:   `<html><title>Movie Detail</title><body><h1>ABW-102</h1></body></html>`,
			want: false,
		},
		{
			name: "empty",
			in:   ``,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCloudflareChallengePage(tt.in)
			if got != tt.want {
				t.Fatalf("IsCloudflareChallengePage() = %v, want %v", got, tt.want)
			}
		})
	}
}
