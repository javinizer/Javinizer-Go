package api

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	ws "github.com/javinizer/javinizer-go/internal/websocket"
)

var (
	wsTestOnce sync.Once
	wsTestMu   sync.Mutex
)

// cleanupServerHub cleans up the global hub created by NewServer
func cleanupServerHub(t *testing.T, deps *ServerDependencies) {
	t.Helper()
	if deps.wsCancel != nil {
		deps.wsCancel()
		// Wait for hub to shut down gracefully (max 500ms)
		time.Sleep(100 * time.Millisecond)
	}
}

// initTestWebSocket initializes the package-level wsHub and wsUpgrader for testing.
// This prevents nil pointer panics in processBatchJob and similar functions.
// Note: wsHub is initialized once and reused across tests to avoid race conditions
// with background goroutines. wsUpgrader is always reinitialized to ensure test
// isolation when tests run in different orders (some tests call NewServer which sets
// stricter origin checking, so we need to reset to test-friendly settings).
func initTestWebSocket(t *testing.T) {
	t.Helper()

	wsTestMu.Lock()
	defer wsTestMu.Unlock()

	// Always reinitialize wsUpgrader for testing (allow all origins)
	// This ensures test isolation even if NewServer() was called by another test.
	// The mutex prevents race conditions during reinitialization.
	wsUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins in tests
		},
	}

	// Only initialize wsHub if not already initialized
	if wsHub == nil {
		wsHub = ws.NewHub()
		ctx, cancel := context.WithCancel(context.Background())

		// Create channel to signal when Run completes
		done := make(chan struct{})
		go func() {
			wsHub.Run(ctx)
			close(done)
		}()

		// Clean up on test completion - ensure hub stops gracefully
		t.Cleanup(func() {
			cancel()
			// Wait for hub goroutine to fully exit (max 500ms)
			select {
			case <-done:
				// Hub shut down successfully
			case <-time.After(500 * time.Millisecond):
				// Timeout waiting for shutdown (shouldn't happen in tests)
			}
		})
	}
}
