package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFileAtomic_Success(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	// Create source file with test content
	srcPath := filepath.Join(tmpDir, "source.txt")
	testContent := []byte("test content for atomic copy")
	if err := os.WriteFile(srcPath, testContent, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy to destination
	dstPath := filepath.Join(tmpDir, "destination.txt")
	if err := CopyFileAtomic(srcPath, dstPath); err != nil {
		t.Fatalf("CopyFileAtomic failed: %v", err)
	}

	// Verify destination exists
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Error("Destination file does not exist")
	}

	// Verify content matches
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(dstContent) != string(testContent) {
		t.Errorf("Content mismatch: got %q, want %q", string(dstContent), string(testContent))
	}

	// Verify temp file was cleaned up
	tmpPath := dstPath + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("Temp file was not cleaned up")
	}
}

func TestCopyFileAtomic_SourceNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	srcPath := filepath.Join(tmpDir, "nonexistent.txt")
	dstPath := filepath.Join(tmpDir, "destination.txt")

	err := CopyFileAtomic(srcPath, dstPath)
	if err == nil {
		t.Fatal("Expected error for nonexistent source, got nil")
	}

	// Verify destination was not created
	if _, err := os.Stat(dstPath); !os.IsNotExist(err) {
		t.Error("Destination file should not exist when source is missing")
	}
}

func TestCopyFileAtomic_CreatesNestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file
	srcPath := filepath.Join(tmpDir, "source.txt")
	testContent := []byte("test content")
	if err := os.WriteFile(srcPath, testContent, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy to nested directory that doesn't exist yet
	dstPath := filepath.Join(tmpDir, "level1", "level2", "level3", "destination.txt")

	// Should succeed by auto-creating directories
	err := CopyFileAtomic(srcPath, dstPath)
	if err != nil {
		t.Fatalf("CopyFileAtomic should auto-create nested directories, got error: %v", err)
	}

	// Verify destination file exists
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Error("Destination file does not exist after copy")
	}

	// Verify content matches
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(dstContent) != string(testContent) {
		t.Errorf("Content mismatch: got %q, want %q", string(dstContent), string(testContent))
	}

	// Verify intermediate directories were created
	dir := filepath.Join(tmpDir, "level1", "level2", "level3")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("Intermediate directories were not created")
	}
}

func TestCopyFileAtomic_LargeFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large file test in short mode")
	}

	tmpDir := t.TempDir()

	// Create a 10MB file
	srcPath := filepath.Join(tmpDir, "large_source.bin")
	size := 10 * 1024 * 1024 // 10MB
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}

	if err := os.WriteFile(srcPath, data, 0644); err != nil {
		t.Fatalf("Failed to create large source file: %v", err)
	}

	// Copy
	dstPath := filepath.Join(tmpDir, "large_dest.bin")
	if err := CopyFileAtomic(srcPath, dstPath); err != nil {
		t.Fatalf("CopyFileAtomic failed for large file: %v", err)
	}

	// Verify size matches
	srcInfo, _ := os.Stat(srcPath)
	dstInfo, _ := os.Stat(dstPath)

	if srcInfo.Size() != dstInfo.Size() {
		t.Errorf("File size mismatch: got %d, want %d", dstInfo.Size(), srcInfo.Size())
	}
}

func TestCopyFileAtomic_OverwriteExisting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file
	srcPath := filepath.Join(tmpDir, "source.txt")
	newContent := []byte("new content")
	if err := os.WriteFile(srcPath, newContent, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Create existing destination file with old content
	dstPath := filepath.Join(tmpDir, "destination.txt")
	oldContent := []byte("old content")
	if err := os.WriteFile(dstPath, oldContent, 0644); err != nil {
		t.Fatalf("Failed to create destination file: %v", err)
	}

	// Copy should overwrite
	if err := CopyFileAtomic(srcPath, dstPath); err != nil {
		t.Fatalf("CopyFileAtomic failed: %v", err)
	}

	// Verify new content
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(dstContent) != string(newContent) {
		t.Errorf("Content was not overwritten: got %q, want %q", string(dstContent), string(newContent))
	}
}
