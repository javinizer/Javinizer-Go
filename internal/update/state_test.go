package update

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateStore_LoadSave(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "update_cache.json")

	store := NewStateStore(statePath, DefaultCheckInterval)

	// Test loading empty state
	state, err := store.LoadState()
	assert.NoError(t, err)
	assert.Nil(t, state)

	// Test saving state
	state = &UpdateState{
		Version:    "v1.6.0",
		CheckedAt:  NowISO8601(),
		Available:  true,
		Prerelease: false,
		Source:     "fresh",
	}

	err = store.SaveState(state)
	require.NoError(t, err)

	// Test loading saved state
	loaded, err := store.LoadState()
	assert.NoError(t, err)
	assert.NotNil(t, loaded)
	assert.Equal(t, "v1.6.0", loaded.Version)
	assert.True(t, loaded.Available)
	assert.False(t, loaded.Prerelease)
	assert.Equal(t, "fresh", loaded.Source)
}

func TestStateStore_ShouldCheck(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "update_cache.json")

	// Test with no state
	store := NewStateStore(statePath, 24*time.Hour)
	assert.True(t, store.ShouldCheck(), "should check when no state exists")

	// Test with fresh state (within interval)
	state := &UpdateState{
		CheckedAt: NowISO8601(),
	}
	store.SetState(state)
	assert.False(t, store.ShouldCheck(), "should not check within interval")

	// Test with old state (past interval)
	state = &UpdateState{
		CheckedAt: time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC3339),
	}
	store.SetState(state)
	assert.True(t, store.ShouldCheck(), "should check when past interval")

	// Test with invalid timestamp (should fail open and re-check)
	state = &UpdateState{
		CheckedAt: "not-a-timestamp",
	}
	store.SetState(state)
	assert.True(t, store.ShouldCheck(), "should check when timestamp is invalid")
}

func TestStateStore_Clear(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "update_cache.json")

	store := NewStateStore(statePath, DefaultCheckInterval)

	// Save state
	state := &UpdateState{
		Version:   "v1.6.0",
		CheckedAt: NowISO8601(),
	}
	require.NoError(t, store.SaveState(state))

	// Clear
	err := store.ClearState()
	assert.NoError(t, err)

	// Verify state is cleared
	assert.Nil(t, store.GetState())
}

func TestLoadStateFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "update_cache.json")

	// Test with non-existent file
	state, err := LoadStateFromFile(statePath)
	assert.NoError(t, err)
	assert.Nil(t, state)

	// Test with valid file
	state = &UpdateState{
		Version:   "v1.6.0",
		CheckedAt: NowISO8601(),
	}
	err = SaveStateToFile(statePath, state)
	require.NoError(t, err)

	loaded, err := LoadStateFromFile(statePath)
	assert.NoError(t, err)
	assert.Equal(t, "v1.6.0", loaded.Version)
}

func TestLoadStateFromFile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "update_cache.json")

	require.NoError(t, os.WriteFile(statePath, []byte(`{not-json`), 0644))

	state, err := LoadStateFromFile(statePath)
	assert.Error(t, err)
	assert.Nil(t, state)
}

func TestNowISO8601(t *testing.T) {
	now := NowISO8601()
	// Should be parseable as RFC3339
	parsed, err := time.Parse(time.RFC3339, now)
	assert.NoError(t, err)
	assert.WithinDuration(t, time.Now(), parsed, time.Second)
}

func TestStateStore_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "update_cache.json")

	store := NewStateStore(statePath, DefaultCheckInterval)

	// Pre-populate with a state to test concurrent access
	state := &UpdateState{
		Version:   "v1.6.0",
		CheckedAt: NowISO8601(),
	}
	store.SetState(state)

	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func() {
			_, _ = store.LoadState()
			s := store.GetState()
			if s != nil {
				_ = s.Version
				_ = s.CheckedAt
			}
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	// Verify no race conditions - state should still be accessible
	s := store.GetState()
	assert.NotNil(t, s)
	assert.Equal(t, "v1.6.0", s.Version)
}

func TestUpdateState_JSON(t *testing.T) {
	state := UpdateState{
		Version:    "v1.6.0",
		CheckedAt:  "2026-03-21T10:00:00Z",
		Available:  true,
		Prerelease: false,
		Source:     "fresh",
		Error:      "rate limited",
	}

	data, err := json.Marshal(state)
	require.NoError(t, err)

	var loaded UpdateState
	err = json.Unmarshal(data, &loaded)
	assert.NoError(t, err)
	assert.Equal(t, state, loaded)
}

func TestNewDefaultStateStore_UsesRuntimeDataDir(t *testing.T) {
	tempDataDir := t.TempDir()
	t.Setenv("JAVINIZER_DATA_DIR", tempDataDir)

	store := NewDefaultStateStore()
	require.NotNil(t, store)
	assert.Equal(t, filepath.Join(tempDataDir, "update_cache.json"), store.path)
	assert.Equal(t, DefaultCheckInterval, store.interval)
}

func TestStateStore_LoadState_PathIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "update_cache.json")
	require.NoError(t, os.Mkdir(statePath, 0o755))

	store := NewStateStore(statePath, DefaultCheckInterval)
	state, err := store.LoadState()
	require.Error(t, err)
	assert.Nil(t, state)
}

func TestStateStore_LoadState_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "update_cache.json")
	require.NoError(t, os.WriteFile(statePath, []byte(`{bad json`), 0o644))

	store := NewStateStore(statePath, DefaultCheckInterval)
	state, err := store.LoadState()
	require.Error(t, err)
	assert.Nil(t, state)
}

func TestStateStore_SaveState_MkdirAllFailure(t *testing.T) {
	tmpDir := t.TempDir()
	parentAsFile := filepath.Join(tmpDir, "not-a-directory")
	require.NoError(t, os.WriteFile(parentAsFile, []byte("x"), 0o644))

	store := NewStateStore(filepath.Join(parentAsFile, "update_cache.json"), DefaultCheckInterval)
	err := store.SaveState(&UpdateState{Version: "v1.0.0"})
	require.Error(t, err)
}

func TestStateStore_SaveState_RenameFailureCleansTempFile(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "as-directory")
	require.NoError(t, os.Mkdir(targetDir, 0o755))

	store := NewStateStore(targetDir, DefaultCheckInterval)
	err := store.SaveState(&UpdateState{Version: "v1.0.0"})
	require.Error(t, err)

	_, statErr := os.Stat(targetDir + ".tmp")
	assert.True(t, os.IsNotExist(statErr))
}

func TestSaveStateToFile_MkdirAllFailure(t *testing.T) {
	tmpDir := t.TempDir()
	parentAsFile := filepath.Join(tmpDir, "not-a-directory")
	require.NoError(t, os.WriteFile(parentAsFile, []byte("x"), 0o644))

	err := SaveStateToFile(filepath.Join(parentAsFile, "update_cache.json"), &UpdateState{Version: "v1.0.0"})
	require.Error(t, err)
}

func TestSaveStateToFile_RenameFailureCleansTempFile(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "as-directory")
	require.NoError(t, os.Mkdir(targetDir, 0o755))

	err := SaveStateToFile(targetDir, &UpdateState{Version: "v1.0.0"})
	require.Error(t, err)

	_, statErr := os.Stat(targetDir + ".tmp")
	assert.True(t, os.IsNotExist(statErr))
}
