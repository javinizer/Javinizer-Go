package batch

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/models"
	"github.com/javinizer/javinizer-go/internal/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBatchJob_IncludesCompletedAndEndedAt(t *testing.T) {
	cfg := &config.Config{}
	deps := createTestDeps(t, cfg, "")

	job := deps.JobQueue.CreateJob([]string{"/path/to/IPX-700.mp4"})
	started := time.Now().UTC().Add(-2 * time.Minute)
	ended := time.Now().UTC()
	job.UpdateFileResult("/path/to/IPX-700.mp4", &worker.FileResult{
		FilePath:  "/path/to/IPX-700.mp4",
		MovieID:   "IPX-700",
		Status:    worker.JobStatusCompleted,
		Data:      &models.Movie{ID: "IPX-700", Title: "Done"},
		StartedAt: started,
		EndedAt:   &ended,
	})
	job.MarkCompleted()

	router := gin.New()
	router.GET("/batch/:id", getBatchJob(deps))

	req := httptest.NewRequest("GET", "/batch/"+job.ID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)

	var response BatchJobResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	require.NotNil(t, response.CompletedAt)
	require.Contains(t, response.Results, "/path/to/IPX-700.mp4")

	result := response.Results["/path/to/IPX-700.mp4"]
	assert.NotNil(t, result.EndedAt)
	assert.Equal(t, "completed", result.Status)
	assert.Equal(t, "IPX-700", result.MovieID)
}

func TestBatchScrape_InvalidPresetReturnsBadRequest(t *testing.T) {
	initTestWebSocket(t)

	cfg := &config.Config{
		Matching: config.MatchingConfig{
			RegexEnabled: false,
		},
		API: config.APIConfig{
			Security: config.SecurityConfig{
				AllowedDirectories: []string{"/path"},
			},
		},
	}
	deps := createTestDeps(t, cfg, "")

	router := gin.New()
	router.POST("/batch/scrape", batchScrape(deps))

	body := `{"files":["/path/to/IPX-535.mp4"],"preset":"not-a-preset"}`
	req := httptest.NewRequest("POST", "/batch/scrape", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
}
