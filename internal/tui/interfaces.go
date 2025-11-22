package tui

import (
	"github.com/javinizer/javinizer-go/internal/worker"
)

// PoolInterface defines the contract for worker pool operations.
// Implemented by: *worker.Pool
type PoolInterface interface {
	Submit(task worker.Task) error
	Wait() error
	Stop()
}

// ProgressTrackerInterface defines the contract for progress tracking operations.
// Implemented by: *worker.ProgressTracker
type ProgressTrackerInterface interface {
	Update(id string, progress float64, message string, bytesProcessed int64)
	Complete(id string, message string)
	Fail(id string, err error)
}

// DownloaderInterface defines the contract for media download operations used by ProcessingCoordinator.
// Only includes methods directly called by processor.go (SetDownloadExtrafanart).
// Implemented by: *downloader.Downloader
type DownloaderInterface interface {
	SetDownloadExtrafanart(enabled bool)
}
