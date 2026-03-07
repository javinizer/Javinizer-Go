package fc2

import (
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/models"
	"github.com/stretchr/testify/assert"
)

func testConfig() *config.Config {
	cfg := config.DefaultConfig()
	cfg.Scrapers.FC2.Enabled = true
	cfg.Scrapers.FC2.RequestDelay = 0
	cfg.Scrapers.Proxy.Enabled = false
	return cfg
}

func TestScraperInterfaceCompliance(t *testing.T) {
	s := New(testConfig())
	var _ models.Scraper = s
	var _ models.ScraperQueryResolver = s
}

func TestNameAndEnabled(t *testing.T) {
	cfg := testConfig()
	s := New(cfg)

	assert.Equal(t, "fc2", s.Name())
	assert.True(t, s.IsEnabled())

	cfg.Scrapers.FC2.Enabled = false
	s = New(cfg)
	assert.False(t, s.IsEnabled())
}

func TestResolveSearchQuery(t *testing.T) {
	s := New(testConfig())

	tests := []struct {
		name  string
		input string
		want  string
		ok    bool
	}{
		{name: "canonical", input: "FC2-PPV-4847718", want: "FC2-PPV-4847718", ok: true},
		{name: "compact", input: "FC2PPV4847718", want: "FC2-PPV-4847718", ok: true},
		{name: "ppv short", input: "PPV-4847718", want: "FC2-PPV-4847718", ok: true},
		{name: "article url", input: "https://adult.contents.fc2.com/article/4847718/", want: "FC2-PPV-4847718", ok: true},
		{name: "plain article id", input: "4847718", want: "FC2-PPV-4847718", ok: true},
		{name: "invalid", input: "ABP-123", want: "", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := s.ResolveSearchQuery(tt.input)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetURL(t *testing.T) {
	s := New(testConfig())

	u, err := s.GetURL("PPV-4847718")
	assert.NoError(t, err)
	assert.Equal(t, "https://adult.contents.fc2.com/article/4847718/", u)

	u, err = s.GetURL("https://adult.contents.fc2.com/article/4847718/?lang=en")
	assert.NoError(t, err)
	assert.Equal(t, "https://adult.contents.fc2.com/article/4847718/", u)

	_, err = s.GetURL("ABP-123")
	assert.Error(t, err)
}

func TestParseDetailPage(t *testing.T) {
	html := `
<!doctype html>
<html>
<head>
<meta property="og:title" content="FC2-PPV-4847718 Sample Title">
<meta property="og:description" content="FC2-PPV-4847718 Sample description text">
<meta property="og:image" content="//storage.example.com/cover.png">
<meta property="og:video" content="https://adult.contents.fc2.com/embed/4847718/">
<script type="application/ld+json">{"@type":"Product","aggregateRating":{"ratingValue":4.9,"reviewCount":204}}</script>
</head>
<body>
  <div class="items_article_MainitemThumb">
    <span><p class="items_article_info">30:39</p></span>
  </div>
  <div class="items_article_headerInfo">
    <ul><li>by <a href="https://adult.contents.fc2.com/users/demo/">Demo Seller</a></li></ul>
  </div>
  <section class="items_article_TagArea">
    <a class="tag tagTag">素人</a>
    <a class="tag tagTag">中出し</a>
    <a class="tag tagTag">素人</a>
  </section>
  <div class="items_article_softDevice"><p>販売日 : 2026/02/13</p></div>
  <div class="items_article_softDevice"><p>商品ID : FC2 PPV 4847718</p></div>
  <ul class="items_article_SampleImagesArea">
    <li><a href="//contents-thumbnail2.fc2.com/w1280/sample1.png"></a></li>
    <li><a href="/sample2.png"></a></li>
  </ul>
</body>
</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	assert.NoError(t, err)

	result := parseDetailPage(doc, html, "https://adult.contents.fc2.com/article/4847718/", "4847718")
	if assert.NotNil(t, result) {
		assert.Equal(t, "fc2", result.Source)
		assert.Equal(t, "FC2-PPV-4847718", result.ID)
		assert.Equal(t, "FC2-PPV-4847718", result.ContentID)
		assert.Equal(t, "Sample Title", result.Title)
		assert.Equal(t, "Sample Title", result.OriginalTitle)
		assert.Equal(t, "Sample description text", result.Description)
		assert.Equal(t, 31, result.Runtime)
		assert.Equal(t, "Demo Seller", result.Maker)
		assert.Equal(t, "https://storage.example.com/cover.png", result.CoverURL)
		assert.Equal(t, "https://storage.example.com/cover.png", result.PosterURL)
		assert.Equal(t, "https://adult.contents.fc2.com/embed/4847718/", result.TrailerURL)
		assert.Equal(t, []string{"素人", "中出し"}, result.Genres)
		assert.Equal(t, []string{
			"https://contents-thumbnail2.fc2.com/w1280/sample1.png",
			"https://adult.contents.fc2.com/sample2.png",
		}, result.ScreenshotURL)

		if assert.NotNil(t, result.ReleaseDate) {
			expected := time.Date(2026, 2, 13, 0, 0, 0, 0, time.UTC)
			assert.True(t, result.ReleaseDate.Equal(expected))
		}
		if assert.NotNil(t, result.Rating) {
			assert.InDelta(t, 4.9, result.Rating.Score, 0.0001)
			assert.Equal(t, 204, result.Rating.Votes)
		}
	}
}

func TestParseRuntime(t *testing.T) {
	assert.Equal(t, 31, parseRuntime("30:39"))
	assert.Equal(t, 65, parseRuntime("1:04:31"))
	assert.Equal(t, 120, parseRuntime("120min"))
	assert.Equal(t, 0, parseRuntime(""))
}

func TestIsNotFoundPage(t *testing.T) {
	assert.True(t, isFC2NotFoundPage("申し訳ありません、お探しの商品が見つかりませんでした"))
	assert.False(t, isFC2NotFoundPage("<html><title>normal item page</title></html>"))
}
