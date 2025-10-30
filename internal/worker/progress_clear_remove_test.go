package worker

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProgressTracker_Clear(t *testing.T) {
	t.Run("Clear removes all tasks", func(t *testing.T) {
		ch := make(chan ProgressUpdate, 10)
		tracker := NewProgressTracker(ch)

		// Add multiple tasks
		tracker.Start("task-1", TaskTypeScrape, "Starting 1")
		tracker.Start("task-2", TaskTypeDownload, "Starting 2")
		tracker.Start("task-3", TaskTypeOrganize, "Starting 3")

		// Verify tasks exist
		all := tracker.GetAll()
		assert.Len(t, all, 3, "Expected 3 tasks before clear")

		// Clear all tasks
		tracker.Clear()

		// Verify all tasks are gone
		all = tracker.GetAll()
		assert.Empty(t, all, "Expected no tasks after clear")

		// Verify stats reflect empty state
		stats := tracker.Stats()
		assert.Equal(t, 0, stats.Total)
		assert.Equal(t, 0, stats.Running)
		assert.Equal(t, 0, stats.Success)
		assert.Equal(t, 0, stats.Failed)
	})

	t.Run("Clear on empty tracker is safe", func(t *testing.T) {
		ch := make(chan ProgressUpdate, 10)
		tracker := NewProgressTracker(ch)

		// Clear when already empty
		tracker.Clear()

		all := tracker.GetAll()
		assert.Empty(t, all)
	})

	t.Run("Can add tasks after clear", func(t *testing.T) {
		ch := make(chan ProgressUpdate, 10)
		tracker := NewProgressTracker(ch)

		// Add, clear, add again
		tracker.Start("task-1", TaskTypeScrape, "Starting 1")
		tracker.Clear()
		tracker.Start("task-2", TaskTypeDownload, "Starting 2")

		all := tracker.GetAll()
		assert.Len(t, all, 1, "Expected 1 task after clear and re-add")

		// Check task ID
		if len(all) > 0 {
			assert.Equal(t, "task-2", all[0].ID)
		}
	})

	t.Run("Clear is thread-safe", func(t *testing.T) {
		ch := make(chan ProgressUpdate, 100)
		tracker := NewProgressTracker(ch)

		// Add many tasks
		for i := 0; i < 50; i++ {
			tracker.Start(string(rune('A'+i%26)), TaskTypeScrape, "Starting")
		}

		// Clear concurrently from multiple goroutines
		done := make(chan bool, 10)
		for g := 0; g < 10; g++ {
			go func() {
				tracker.Clear()
				done <- true
			}()
		}

		// Wait for all clears
		for g := 0; g < 10; g++ {
			<-done
		}

		// Should be empty
		all := tracker.GetAll()
		assert.Empty(t, all)
	})
}

func TestProgressTracker_Remove(t *testing.T) {
	t.Run("Remove specific task", func(t *testing.T) {
		ch := make(chan ProgressUpdate, 10)
		tracker := NewProgressTracker(ch)

		tracker.Start("task-1", TaskTypeScrape, "Starting 1")
		tracker.Start("task-2", TaskTypeDownload, "Starting 2")
		tracker.Start("task-3", TaskTypeOrganize, "Starting 3")

		// Verify all exist
		require.Len(t, tracker.GetAll(), 3)

		// Remove task-2
		tracker.Remove("task-2")

		all := tracker.GetAll()
		assert.Len(t, all, 2, "Expected 2 tasks after removing one")

		// Check tasks by ID
		foundIDs := make(map[string]bool)
		for _, task := range all {
			foundIDs[task.ID] = true
		}
		assert.True(t, foundIDs["task-1"], "task-1 should exist")
		assert.False(t, foundIDs["task-2"], "task-2 should not exist")
		assert.True(t, foundIDs["task-3"], "task-3 should exist")

		// Verify task-2 cannot be retrieved
		progress, ok := tracker.Get("task-2")
		assert.False(t, ok, "Removed task should not be retrievable")
		assert.Nil(t, progress)
	})

	t.Run("Remove non-existent task is safe", func(t *testing.T) {
		ch := make(chan ProgressUpdate, 10)
		tracker := NewProgressTracker(ch)

		tracker.Start("task-1", TaskTypeScrape, "Starting")

		// Remove non-existent task should not panic
		tracker.Remove("non-existent")

		// Original task should still exist
		all := tracker.GetAll()
		assert.Len(t, all, 1)

		// Check task ID
		if len(all) > 0 {
			assert.Equal(t, "task-1", all[0].ID)
		}
	})

	t.Run("Remove updates stats correctly", func(t *testing.T) {
		ch := make(chan ProgressUpdate, 10)
		tracker := NewProgressTracker(ch)

		tracker.Start("task-1", TaskTypeScrape, "Starting 1")
		tracker.Start("task-2", TaskTypeScrape, "Starting 2")
		tracker.Start("task-3", TaskTypeScrape, "Starting 3")

		// Drain start messages
		for i := 0; i < 3; i++ {
			<-ch
		}

		// Complete one, fail one, leave one running
		tracker.Complete("task-1", "Done")
		tracker.Fail("task-2", assert.AnError)

		statsBefore := tracker.Stats()
		assert.Equal(t, 3, statsBefore.Total)
		assert.Equal(t, 1, statsBefore.Running)
		assert.Equal(t, 1, statsBefore.Success)
		assert.Equal(t, 1, statsBefore.Failed)

		// Remove the completed task
		tracker.Remove("task-1")

		statsAfter := tracker.Stats()
		assert.Equal(t, 2, statsAfter.Total, "Total should decrease by 1")
		assert.Equal(t, 1, statsAfter.Running, "Running count unchanged")
		assert.Equal(t, 0, statsAfter.Success, "Success count should decrease by 1")
		assert.Equal(t, 1, statsAfter.Failed, "Failed count unchanged")
	})

	t.Run("Remove last task leaves empty tracker", func(t *testing.T) {
		ch := make(chan ProgressUpdate, 10)
		tracker := NewProgressTracker(ch)

		tracker.Start("task-1", TaskTypeScrape, "Starting")
		tracker.Remove("task-1")

		all := tracker.GetAll()
		assert.Empty(t, all)

		stats := tracker.Stats()
		assert.Equal(t, 0, stats.Total)
	})

	t.Run("Remove by type filter works after removal", func(t *testing.T) {
		ch := make(chan ProgressUpdate, 10)
		tracker := NewProgressTracker(ch)

		tracker.Start("scrape-1", TaskTypeScrape, "Starting")
		tracker.Start("scrape-2", TaskTypeScrape, "Starting")
		tracker.Start("download-1", TaskTypeDownload, "Starting")

		// Remove one scrape task
		tracker.Remove("scrape-1")

		scrapes := tracker.GetByType(TaskTypeScrape)
		downloads := tracker.GetByType(TaskTypeDownload)

		assert.Len(t, scrapes, 1, "Expected 1 scrape task after removal")
		assert.Len(t, downloads, 1, "Expected 1 download task unchanged")
	})

	t.Run("Remove is thread-safe", func(t *testing.T) {
		ch := make(chan ProgressUpdate, 100)
		tracker := NewProgressTracker(ch)

		// Add many tasks with unique IDs
		numTasks := 50
		for i := 0; i < numTasks; i++ {
			taskID := fmt.Sprintf("task-%d", i)
			tracker.Start(taskID, TaskTypeScrape, "Starting")
		}

		// Verify initial state
		initialCount := len(tracker.GetAll())
		assert.Equal(t, numTasks, initialCount, "Should have all tasks initially")

		// Remove tasks concurrently from multiple goroutines
		done := make(chan bool, 10)
		for g := 0; g < 10; g++ {
			go func(goroutineID int) {
				for i := 0; i < 5; i++ {
					taskID := fmt.Sprintf("task-%d", goroutineID*5+i)
					tracker.Remove(taskID)
				}
				done <- true
			}(g)
		}

		// Wait for all removals
		for g := 0; g < 10; g++ {
			<-done
		}

		// Verify removals occurred (should have removed 50 tasks total from 10 goroutines * 5 each)
		finalCount := len(tracker.GetAll())
		assert.Equal(t, 0, finalCount, "Should have removed all 50 unique tasks")
		assert.Less(t, finalCount, initialCount, "Final count should be less than initial count")
	})

	t.Run("Remove task with progress history", func(t *testing.T) {
		ch := make(chan ProgressUpdate, 10)
		tracker := NewProgressTracker(ch)

		taskID := "task-with-progress"
		tracker.Start(taskID, TaskTypeScrape, "Starting")
		<-ch // Drain start message

		// Update progress multiple times
		tracker.Update(taskID, 0.25, "Quarter done", 256)
		tracker.Update(taskID, 0.50, "Half done", 512)
		tracker.Update(taskID, 0.75, "Almost done", 768)

		// Verify progress is tracked
		progress, ok := tracker.Get(taskID)
		require.True(t, ok)
		assert.Equal(t, 0.75, progress.Progress)
		assert.Equal(t, int64(768), progress.BytesDone)

		// Remove the task
		tracker.Remove(taskID)

		// Verify it's gone
		progress, ok = tracker.Get(taskID)
		assert.False(t, ok)
		assert.Nil(t, progress)
	})
}

// TestProgressTracker_ClearAndRemove_Interaction tests interaction between Clear and Remove
func TestProgressTracker_ClearAndRemove_Interaction(t *testing.T) {
	ch := make(chan ProgressUpdate, 10)
	tracker := NewProgressTracker(ch)

	tracker.Start("task-1", TaskTypeScrape, "Starting 1")
	tracker.Start("task-2", TaskTypeDownload, "Starting 2")

	// Remove one task
	tracker.Remove("task-1")
	assert.Len(t, tracker.GetAll(), 1)

	// Clear should remove the remaining task
	tracker.Clear()
	assert.Empty(t, tracker.GetAll())

	// Removing from empty tracker is safe
	tracker.Remove("task-2")
	assert.Empty(t, tracker.GetAll())
}

// TestProgressTracker_RemoveWhileUpdating tests removing a task while it's being updated
func TestProgressTracker_RemoveWhileUpdating(t *testing.T) {
	ch := make(chan ProgressUpdate, 100)
	tracker := NewProgressTracker(ch)

	taskID := "concurrent-task"
	tracker.Start(taskID, TaskTypeScrape, "Starting")

	done := make(chan bool, 2)

	// Goroutine 1: Keep updating the task
	go func() {
		for i := 0; i < 100; i++ {
			tracker.Update(taskID, float64(i)/100.0, "Updating", int64(i))
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Goroutine 2: Try to remove the task after a delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		tracker.Remove(taskID)
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Task should be removed by now
	_, ok := tracker.Get(taskID)
	assert.False(t, ok, "Task should have been removed")
}
