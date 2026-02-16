package javdb

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewScraper(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Scrapers.JavDB.Enabled = true

	scraper := New(cfg)
	require.NotNil(t, scraper)
	assert.Equal(t, "javdb", scraper.Name())
	assert.True(t, scraper.IsEnabled())
}

func TestSearch_Disabled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Scrapers.JavDB.Enabled = false

	scraper := New(cfg)
	_, err := scraper.Search("IPX-123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "disabled")
}

func TestSearch_Success(t *testing.T) {
	searchHTML := `
<html>
	<body>
		<div class="movie-list">
			<div class="item">
				<a href="/v/abc123">
					<div class="video-title"><strong>IPX-123</strong> Test Movie Title</div>
					<div class="uid">IPX-123</div>
				</a>
			</div>
		</div>
	</body>
</html>
`
	detailHTML := `
<html>
	<head><title>IPX-123 Test Movie Title - JavDB</title></head>
	<body>
		<h2 class="title is-4"><strong>IPX-123</strong> Test Movie Title</h2>
		<div class="column-video-cover"><img class="video-cover" src="//img.example.com/cover.jpg" /></div>
		<div class="movie-panel-info">
			<div class="panel-block"><strong>番號:</strong><span class="value">IPX-123</span></div>
			<div class="panel-block"><strong>日期:</strong><span class="value">2024-01-02</span></div>
			<div class="panel-block"><strong>時長:</strong><span class="value">120分鐘</span></div>
			<div class="panel-block"><strong>導演:</strong><span class="value"><a>Director Name</a></span></div>
			<div class="panel-block"><strong>片商:</strong><span class="value"><a>Maker Name</a></span></div>
			<div class="panel-block"><strong>發行:</strong><span class="value"><a>Label Name</a></span></div>
			<div class="panel-block"><strong>系列:</strong><span class="value"><a>Series Name</a></span></div>
			<div class="panel-block"><strong>演員:</strong><span class="value"><a>Actress One</a><a>Actress Two</a></span></div>
			<div class="panel-block"><strong>類別:</strong><span class="value"><a>Drama</a><a>Schoolgirl</a></span></div>
			<div class="panel-block"><strong>評分:</strong><span class="value">4.1分 (123人評價)</span></div>
		</div>
		<span itemprop="description">This is a test description.</span>
		<div class="tile-images preview-images">
			<a href="//img.example.com/shot1.jpg"><img src="//img.example.com/thumb1.jpg" /></a>
			<a href="https://img.example.com/shot2.jpg"><img src="https://img.example.com/thumb2.jpg" /></a>
		</div>
		<video id="preview-video"><source src="//video.example.com/trailer.mp4" /></video>
	</body>
</html>
`

	client := resty.New()
	client.SetTransport(&staticRoundTripper{
		responses: map[string]string{
			"https://javdb.test/search?q=IPX-123&f=all": searchHTML,
			"https://javdb.test/v/abc123":               detailHTML,
		},
	})

	scraper := &Scraper{
		client:       client,
		cfg:          &config.JavDBConfig{Enabled: true},
		enabled:      true,
		baseURL:      "https://javdb.test",
		requestDelay: 0,
	}
	scraper.lastRequestTime.Store(time.Time{})

	result, err := scraper.Search("IPX-123")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "javdb", result.Source)
	assert.Equal(t, "IPX-123", result.ID)
	assert.Equal(t, "IPX-123", result.ContentID)
	assert.Equal(t, "Test Movie Title", result.Title)
	assert.Equal(t, "This is a test description.", result.Description)
	assert.Equal(t, 120, result.Runtime)
	assert.Equal(t, "Director Name", result.Director)
	assert.Equal(t, "Maker Name", result.Maker)
	assert.Equal(t, "Label Name", result.Label)
	assert.Equal(t, "Series Name", result.Series)
	assert.Equal(t, "https://img.example.com/cover.jpg", result.CoverURL)
	assert.Equal(t, "https://video.example.com/trailer.mp4", result.TrailerURL)
	assert.Len(t, result.ScreenshotURL, 2)
	assert.Len(t, result.Genres, 2)
	assert.Len(t, result.Actresses, 2)
	assert.Less(t, result.Actresses[0].DMMID, 0)
	assert.Less(t, result.Actresses[1].DMMID, 0)
	require.NotNil(t, result.ReleaseDate)
	assert.Equal(t, "2024-01-02", result.ReleaseDate.Format("2006-01-02"))
	require.NotNil(t, result.Rating)
	assert.InDelta(t, 8.2, result.Rating.Score, 0.001)
	assert.Equal(t, 123, result.Rating.Votes)
}

func TestSearch_Success_EnglishLabels(t *testing.T) {
	searchHTML := `
<html>
	<body>
		<div class="movie-list">
			<div class="item">
				<a href="/v/live123">
					<div class="video-title"><strong>SSNI-344</strong> Sample Title</div>
				</a>
			</div>
		</div>
	</body>
</html>
`
	detailHTML := `
<html>
	<body>
		<h2 class="title is-4"><strong>SSNI-344</strong> Sample Title</h2>
		<div class="column-video-cover"><img class="video-cover" src="https://img.example.com/cover.jpg" /></div>
		<nav class="panel movie-panel-info">
			<div class="panel-block"><strong>ID:</strong><span class="value"><a href="/video_codes/SSNI">SSNI</a>-344</span></div>
			<div class="panel-block"><strong>Released Date:</strong><span class="value">2018-11-19</span></div>
			<div class="panel-block"><strong>Duration:</strong><span class="value">150 minute(s)</span></div>
			<div class="panel-block"><strong>Maker:</strong><span class="value"><a>Maker Name</a></span></div>
			<div class="panel-block"><strong>Publisher:</strong><span class="value"><a>Publisher Name</a></span></div>
			<div class="panel-block"><strong>Tags:</strong><span class="value"><a>Big Tits</a><a>Rape</a></span></div>
			<div class="panel-block"><strong>Actor(s):</strong><span class="value"><a>Actress One</a></span></div>
		</nav>
	</body>
</html>
`

	client := resty.New()
	client.SetTransport(&staticRoundTripper{
		responses: map[string]string{
			"https://javdb.test/search?q=SSNI-344&f=all": searchHTML,
			"https://javdb.test/v/live123":               detailHTML,
		},
	})

	scraper := &Scraper{
		client:       client,
		cfg:          &config.JavDBConfig{Enabled: true},
		enabled:      true,
		baseURL:      "https://javdb.test",
		requestDelay: 0,
	}
	scraper.lastRequestTime.Store(time.Time{})

	result, err := scraper.Search("SSNI-344")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "SSNI-344", result.ID)
	assert.Equal(t, 150, result.Runtime)
	assert.Equal(t, "Publisher Name", result.Label)
	assert.Equal(t, "Maker Name", result.Maker)
	assert.Equal(t, []string{"Big Tits", "Rape"}, result.Genres)
	assert.Len(t, result.Actresses, 1)
	assert.Less(t, result.Actresses[0].DMMID, 0)
}

func TestSearch_ScreenshotSkipsLoginLink(t *testing.T) {
	searchHTML := `
<html><body><div class="movie-list"><div class="item"><a href="/v/img123"><div class="video-title"><strong>ABC-123</strong> Movie</div></a></div></div></body></html>
`
	detailHTML := `
<html>
	<body>
		<h2 class="title is-4"><strong>ABC-123</strong> Movie</h2>
		<div class="movie-panel-info">
			<div class="panel-block"><strong>ID:</strong><span class="value">ABC-123</span></div>
		</div>
		<div class="tile-images preview-images">
			<a class="preview-video-container" href="/login"><img src="https://img.example.com/trailer-thumb.jpg" /></a>
			<a class="tile-item" href="https://img.example.com/shot1.jpg"><img src="https://img.example.com/thumb1.jpg" /></a>
		</div>
	</body>
</html>
`

	client := resty.New()
	client.SetTransport(&staticRoundTripper{
		responses: map[string]string{
			"https://javdb.test/search?q=ABC-123&f=all": searchHTML,
			"https://javdb.test/v/img123":               detailHTML,
		},
	})

	scraper := &Scraper{
		client:       client,
		cfg:          &config.JavDBConfig{Enabled: true},
		enabled:      true,
		baseURL:      "https://javdb.test",
		requestDelay: 0,
	}
	scraper.lastRequestTime.Store(time.Time{})

	result, err := scraper.Search("ABC-123")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.ScreenshotURL, 1)
	assert.Equal(t, "https://img.example.com/shot1.jpg", result.ScreenshotURL[0])
}

type staticRoundTripper struct {
	responses map[string]string
}

func (s *staticRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if body, ok := s.responses[req.URL.String()]; ok {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}, nil
	}

	return &http.Response{
		StatusCode: http.StatusNotFound,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("not found")),
		Request:    req,
	}, nil
}
