package aggregator

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAggregate_AppliesConfiguredTranslation(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"content": `["タイトル翻訳","説明翻訳"]`,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(response))
	}))
	defer ts.Close()

	cfg := config.DefaultConfig()
	cfg.Scrapers.Priority = []string{"r18dev"}
	cfg.Metadata.Priority.Title = []string{"r18dev"}
	cfg.Metadata.Priority.Description = []string{"r18dev"}
	cfg.Metadata.Translation.Enabled = true
	cfg.Metadata.Translation.Provider = "openai"
	cfg.Metadata.Translation.SourceLanguage = "en"
	cfg.Metadata.Translation.TargetLanguage = "ja"
	cfg.Metadata.Translation.ApplyToPrimary = true
	cfg.Metadata.Translation.OverwriteExistingTarget = true
	cfg.Metadata.Translation.OpenAI.BaseURL = ts.URL
	cfg.Metadata.Translation.OpenAI.APIKey = "k"
	cfg.Metadata.Translation.OpenAI.Model = "m"
	cfg.Metadata.Translation.Fields = config.TranslationFieldsConfig{
		Title:       true,
		Description: true,
	}

	agg := New(cfg)

	results := []*models.ScraperResult{
		{
			Source:      "r18dev",
			Language:    "en",
			ID:          "IPX-001",
			ContentID:   "ipx001",
			Title:       "Original Title",
			Description: "Original Description",
		},
	}

	movie, err := agg.Aggregate(results)
	require.NoError(t, err)
	require.NotNil(t, movie)

	assert.Equal(t, "タイトル翻訳", movie.Title)
	assert.Equal(t, "説明翻訳", movie.Description)

	require.Len(t, movie.Translations, 2)
	langMap := map[string]models.MovieTranslation{}
	for _, tr := range movie.Translations {
		langMap[tr.Language] = tr
	}
	assert.Equal(t, "Original Title", langMap["en"].Title)
	assert.Equal(t, "タイトル翻訳", langMap["ja"].Title)
	assert.Equal(t, "translation:openai", langMap["ja"].SourceName)
}

func TestAggregate_TranslationFailureDoesNotFailAggregate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer ts.Close()

	cfg := config.DefaultConfig()
	cfg.Scrapers.Priority = []string{"r18dev"}
	cfg.Metadata.Priority.Title = []string{"r18dev"}
	cfg.Metadata.Translation.Enabled = true
	cfg.Metadata.Translation.Provider = "openai"
	cfg.Metadata.Translation.SourceLanguage = "en"
	cfg.Metadata.Translation.TargetLanguage = "ja"
	cfg.Metadata.Translation.OpenAI.BaseURL = ts.URL
	cfg.Metadata.Translation.OpenAI.APIKey = "k"
	cfg.Metadata.Translation.OpenAI.Model = "m"
	cfg.Metadata.Translation.Fields = config.TranslationFieldsConfig{Title: true}

	agg := New(cfg)

	results := []*models.ScraperResult{{
		Source:    "r18dev",
		Language:  "en",
		ID:        "IPX-002",
		ContentID: "ipx002",
		Title:     "Original Title",
	}}

	movie, err := agg.Aggregate(results)
	require.NoError(t, err)
	require.NotNil(t, movie)
	assert.Equal(t, "Original Title", movie.Title)
	assert.Len(t, movie.Translations, 1)
}

func TestMergeTranslationFields(t *testing.T) {
	t.Run("overwrites all non-empty incoming fields", func(t *testing.T) {
		current := models.MovieTranslation{
			Language:      "en",
			Title:         "Old Title",
			OriginalTitle: "Old Original",
			Description:   "Old Description",
			Director:      "Old Director",
			Maker:         "Old Maker",
			Label:         "Old Label",
			Series:        "Old Series",
			SourceName:    "old-source",
		}
		incoming := models.MovieTranslation{
			Language:      "ja",
			Title:         "New Title",
			OriginalTitle: "New Original",
			Description:   "New Description",
			Director:      "New Director",
			Maker:         "New Maker",
			Label:         "New Label",
			Series:        "New Series",
			SourceName:    "new-source",
		}

		merged := mergeTranslationFields(current, incoming)
		assert.Equal(t, "ja", merged.Language)
		assert.Equal(t, "New Title", merged.Title)
		assert.Equal(t, "New Original", merged.OriginalTitle)
		assert.Equal(t, "New Description", merged.Description)
		assert.Equal(t, "New Director", merged.Director)
		assert.Equal(t, "New Maker", merged.Maker)
		assert.Equal(t, "New Label", merged.Label)
		assert.Equal(t, "New Series", merged.Series)
		assert.Equal(t, "new-source", merged.SourceName)
	})

	t.Run("keeps existing values when incoming fields are empty", func(t *testing.T) {
		current := models.MovieTranslation{
			Language:      "en",
			Title:         "Old Title",
			OriginalTitle: "Old Original",
			Description:   "Old Description",
			Director:      "Old Director",
			Maker:         "Old Maker",
			Label:         "Old Label",
			Series:        "Old Series",
			SourceName:    "old-source",
		}
		incoming := models.MovieTranslation{
			Language: "fr",
		}

		merged := mergeTranslationFields(current, incoming)
		assert.Equal(t, "fr", merged.Language)
		assert.Equal(t, "Old Title", merged.Title)
		assert.Equal(t, "Old Original", merged.OriginalTitle)
		assert.Equal(t, "Old Description", merged.Description)
		assert.Equal(t, "Old Director", merged.Director)
		assert.Equal(t, "Old Maker", merged.Maker)
		assert.Equal(t, "Old Label", merged.Label)
		assert.Equal(t, "Old Series", merged.Series)
		assert.Equal(t, "old-source", merged.SourceName)
	})
}
