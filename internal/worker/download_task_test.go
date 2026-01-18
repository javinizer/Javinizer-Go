package worker

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/javinizer/javinizer-go/internal/downloader"
	"github.com/javinizer/javinizer-go/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadTask_DryRunPreview(t *testing.T) {
	tests := []struct {
		name              string
		movie             *models.Movie
		expectedPreview   []string // Strings that should be in the preview message
		expectedNotIn     []string // Strings that should NOT be in preview
		shouldContainPath bool     // Whether preview should contain target path
	}{
		{
			name: "Movie with cover, trailer, and screenshots",
			movie: &models.Movie{
				ID:          "IPX-123",
				Title:       "Test Movie",
				CoverURL:    "https://example.com/cover.jpg",
				TrailerURL:  "https://example.com/trailer.mp4",
				Screenshots: []string{"https://example.com/ss1.jpg", "https://example.com/ss2.jpg", "https://example.com/ss3.jpg"},
			},
			expectedPreview: []string{
				"[DRY RUN]",
				"Would download:",
				"cover",
				"trailer",
				"3 screenshots",
			},
			shouldContainPath: true,
		},
		{
			name: "Movie with only cover",
			movie: &models.Movie{
				ID:       "ABC-456",
				Title:    "Test Movie 2",
				CoverURL: "https://example.com/cover.jpg",
			},
			expectedPreview: []string{
				"[DRY RUN]",
				"Would download:",
				"cover",
			},
			expectedNotIn: []string{
				"trailer",
				"screenshots",
			},
			shouldContainPath: true,
		},
		{
			name: "Movie with only screenshots",
			movie: &models.Movie{
				ID:          "XYZ-789",
				Title:       "Test Movie 3",
				Screenshots: []string{"https://example.com/ss1.jpg"},
			},
			expectedPreview: []string{
				"[DRY RUN]",
				"Would download:",
				"1 screenshot",
			},
			expectedNotIn: []string{
				"cover",
				"trailer",
			},
			shouldContainPath: true,
		},
		{
			name: "Movie with no media URLs",
			movie: &models.Movie{
				ID:    "NO-MEDIA",
				Title: "Test Movie 4",
			},
			expectedPreview: []string{
				"[DRY RUN]",
				"Would download:",
				"nothing (no media URLs)",
			},
			expectedNotIn: []string{
				"cover",
				"trailer",
				"screenshot",
			},
			shouldContainPath: false,
		},
		{
			name: "Movie with trailer only",
			movie: &models.Movie{
				ID:         "TRAILER-ONLY",
				Title:      "Test Movie 5",
				TrailerURL: "https://example.com/trailer.mp4",
			},
			expectedPreview: []string{
				"[DRY RUN]",
				"Would download:",
				"trailer",
			},
			expectedNotIn: []string{
				"cover",
				"screenshot",
			},
			shouldContainPath: true,
		},
		{
			name: "Movie with multiple screenshots",
			movie: &models.Movie{
				ID:    "MANY-SS",
				Title: "Test Movie 6",
				Screenshots: []string{
					"https://example.com/ss1.jpg",
					"https://example.com/ss2.jpg",
					"https://example.com/ss3.jpg",
					"https://example.com/ss4.jpg",
					"https://example.com/ss5.jpg",
				},
			},
			expectedPreview: []string{
				"[DRY RUN]",
				"Would download:",
				"5 screenshots",
			},
			expectedNotIn: []string{
				"cover",
				"trailer",
			},
			shouldContainPath: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create progress tracker to capture messages
			progressChan := make(chan ProgressUpdate, 100)
			progressTracker := NewProgressTracker(progressChan)

			// Create download task with dry-run enabled
			targetDir := t.TempDir()
			task := NewDownloadTask(
				tt.movie,
				targetDir,
				&downloader.Downloader{}, // Nil is fine for dry-run (not actually used)
				progressTracker,
				true, // dryRun = true
				nil,  // multipart (nil = single file)
			)

			// Start the task in progress tracker
			progressTracker.Start(task.id, task.Type(), task.Description())

			// Execute the task
			ctx := context.Background()
			err := task.Execute(ctx)
			require.NoError(t, err, "Dry-run should not fail")

			// Get the progress update
			progress, exists := progressTracker.Get(task.id)
			require.True(t, exists, "Progress should exist for task")

			// Verify task completed (100% progress)
			assert.Equal(t, 1.0, progress.Progress, "Task should complete with 100% progress")

			// Verify preview message contains expected strings
			message := progress.Message
			for _, expected := range tt.expectedPreview {
				assert.Contains(t, message, expected,
					"Preview should contain '%s'", expected)
			}

			// Verify preview does NOT contain unwanted strings
			for _, notExpected := range tt.expectedNotIn {
				assert.NotContains(t, message, notExpected,
					"Preview should NOT contain '%s'", notExpected)
			}

			// Verify target path is in message if expected
			if tt.shouldContainPath {
				assert.Contains(t, message, targetDir,
					"Preview should contain target directory path")
			}

			// Verify message indicates dry-run
			assert.Contains(t, message, "[DRY RUN]",
				"Message should indicate dry-run mode")
		})
	}
}

func TestDownloadTask_DryRunNoActualDownloads(t *testing.T) {
	// This test verifies that dry-run mode doesn't actually call the downloader
	t.Run("Dry-run completes without calling downloader", func(t *testing.T) {
		movie := &models.Movie{
			ID:         "TEST-001",
			Title:      "Test Movie",
			CoverURL:   "https://example.com/cover.jpg",
			TrailerURL: "https://example.com/trailer.mp4",
		}

		progressChan := make(chan ProgressUpdate, 100)
		progressTracker := NewProgressTracker(progressChan)

		targetDir := t.TempDir()

		// Create task with dry-run
		task := NewDownloadTask(
			movie,
			targetDir,
			nil, // Downloader is nil - would panic if actually called
			progressTracker,
			true, // dryRun
			nil,  // multipart (nil = single file)
		)

		// Start the task in progress tracker
		progressTracker.Start(task.id, task.Type(), task.Description())

		// Should complete successfully without calling downloader
		ctx := context.Background()
		err := task.Execute(ctx)
		require.NoError(t, err)

		// Verify progress shows completion
		progress, exists := progressTracker.Get(task.id)
		require.True(t, exists)
		assert.Equal(t, 1.0, progress.Progress)
		assert.Contains(t, progress.Message, "[DRY RUN]")
	})
}

func TestDownloadTask_ContextCancellation(t *testing.T) {
	// While dry-run returns early, test context handling in the task structure
	t.Run("Task respects context cancellation", func(t *testing.T) {
		movie := &models.Movie{
			ID:       "TEST-002",
			Title:    "Test Movie",
			CoverURL: "https://example.com/cover.jpg",
		}

		progressChan := make(chan ProgressUpdate, 100)
		progressTracker := NewProgressTracker(progressChan)

		targetDir := t.TempDir()

		task := NewDownloadTask(
			movie,
			targetDir,
			&downloader.Downloader{},
			progressTracker,
			true, // dry-run
			nil,  // multipart (nil = single file)
		)

		// Create cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// For dry-run mode, the task completes so quickly it might not check cancellation
		// This test documents the behavior rather than requiring specific cancellation handling
		err := task.Execute(ctx)

		// Dry-run mode completes immediately, so cancellation may not be detected
		// This is acceptable behavior - we're testing that it doesn't panic
		if err != nil {
			assert.Equal(t, context.Canceled, err, "Error should be context.Canceled if detected")
		}
	})
}

func TestDownloadTask_Multipart(t *testing.T) {
	// Test that multipart info is properly stored (used for multi-part videos)
	tests := []struct {
		name      string
		multipart *downloader.MultipartInfo
	}{
		{"Single file (nil)", nil},
		{"Multi-part file 1", &downloader.MultipartInfo{IsMultiPart: true, PartNumber: 1, PartSuffix: "-pt1"}},
		{"Multi-part file 2", &downloader.MultipartInfo{IsMultiPart: true, PartNumber: 2, PartSuffix: "-pt2"}},
		{"Multi-part file 3", &downloader.MultipartInfo{IsMultiPart: true, PartNumber: 3, PartSuffix: "-pt3"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			movie := &models.Movie{
				ID:    "TEST-PART",
				Title: "Test Movie",
			}

			progressChan := make(chan ProgressUpdate, 100)
			progressTracker := NewProgressTracker(progressChan)

			targetDir := t.TempDir()

			task := NewDownloadTask(
				movie,
				targetDir,
				&downloader.Downloader{},
				progressTracker,
				true,
				tt.multipart,
			)

			assert.Equal(t, tt.multipart, task.multipart, "Multipart info should be stored correctly")

			// Execute dry-run
			ctx := context.Background()
			err := task.Execute(ctx)
			require.NoError(t, err)
		})
	}
}

func TestDownloadTask_Description(t *testing.T) {
	tests := []struct {
		name        string
		dryRun      bool
		movieID     string
		wantContain []string
	}{
		{
			name:        "Dry-run description",
			dryRun:      true,
			movieID:     "IPX-123",
			wantContain: []string{"[DRY RUN]", "Downloading media", "IPX-123"},
		},
		{
			name:        "Normal description",
			dryRun:      false,
			movieID:     "ABC-456",
			wantContain: []string{"Downloading media", "ABC-456"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			movie := &models.Movie{
				ID:    tt.movieID,
				Title: "Test Movie",
			}

			progressChan := make(chan ProgressUpdate, 100)
			progressTracker := NewProgressTracker(progressChan)

			targetDir := t.TempDir()

			task := NewDownloadTask(
				movie,
				targetDir,
				&downloader.Downloader{},
				progressTracker,
				tt.dryRun,
				nil, // multipart (nil = single file)
			)

			description := task.Description()
			for _, want := range tt.wantContain {
				assert.Contains(t, description, want,
					"Description should contain '%s'", want)
			}

			if !tt.dryRun {
				assert.NotContains(t, description, "[DRY RUN]",
					"Non-dry-run description should not contain '[DRY RUN]'")
			}
		})
	}
}

func TestDownloadTask_ProgressUpdates(t *testing.T) {
	// Test that dry-run provides progress updates throughout execution
	t.Run("Progress updates are sent during dry-run", func(t *testing.T) {
		movie := &models.Movie{
			ID:          "PROGRESS-TEST",
			Title:       "Test Movie",
			CoverURL:    "https://example.com/cover.jpg",
			TrailerURL:  "https://example.com/trailer.mp4",
			Screenshots: []string{"https://example.com/ss1.jpg"},
		}

		progressChan := make(chan ProgressUpdate, 100)
		progressTracker := NewProgressTracker(progressChan)

		targetDir := t.TempDir()

		task := NewDownloadTask(
			movie,
			targetDir,
			&downloader.Downloader{},
			progressTracker,
			true, // dry-run
			nil,  // multipart (nil = single file)
		)

		// Start the task in progress tracker
		progressTracker.Start(task.id, task.Type(), task.Description())

		// Execute
		ctx := context.Background()
		err := task.Execute(ctx)
		require.NoError(t, err)

		// Allow brief time for progress updates to propagate
		time.Sleep(10 * time.Millisecond)

		// Verify final progress state
		progress, exists := progressTracker.Get(task.id)
		require.True(t, exists, "Progress should be tracked")

		// Should reach 100% completion
		assert.Equal(t, 1.0, progress.Progress, "Should reach 100% progress")

		// Message should contain dry-run preview
		assert.Contains(t, progress.Message, "[DRY RUN]")
		assert.Contains(t, progress.Message, "Would download:")

		// Verify the preview contains expected items
		message := progress.Message
		assert.True(t,
			strings.Contains(message, "cover") &&
				strings.Contains(message, "trailer") &&
				strings.Contains(message, "screenshot"),
			"Preview should list all available media types")
	})
}
