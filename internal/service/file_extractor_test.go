package service_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kiwamoto1987/evoloop/internal/service"
)

func TestExtractFilePaths_GoTestOutput(t *testing.T) {
	dir := t.TempDir()

	// Create files that match the error output
	if err := os.WriteFile(filepath.Join(dir, "demo_test.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	output := `--- FAIL: TestDemo_IntentionalFailure (0.00s)
    demo_test.go:10: got 2, want 3
FAIL
FAIL	github.com/example/project/internal/domain	0.002s`

	paths := service.ExtractFilePaths(output, dir)

	if len(paths) == 0 {
		t.Fatal("expected at least 1 file path extracted")
	}

	found := false
	for _, p := range paths {
		if filepath.Base(p) == "demo_test.go" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected demo_test.go in paths, got %v", paths)
	}
}

func TestExtractFilePaths_GoBuildOutput(t *testing.T) {
	dir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(dir, "internal"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "internal", "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	output := `internal/main.go:15:2: undefined: someFunc`

	paths := service.ExtractFilePaths(output, dir)

	if len(paths) == 0 {
		t.Fatal("expected at least 1 file path extracted")
	}
}

func TestExtractFilePaths_NonExistentFile(t *testing.T) {
	dir := t.TempDir()

	output := `    nonexistent.go:10: some error`

	paths := service.ExtractFilePaths(output, dir)

	if len(paths) != 0 {
		t.Errorf("expected 0 paths for non-existent file, got %v", paths)
	}
}

func TestExtractFilePaths_NoDuplicates(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "file.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	output := `    file.go:10: error one
    file.go:20: error two`

	paths := service.ExtractFilePaths(output, dir)

	if len(paths) != 1 {
		t.Errorf("expected 1 unique path, got %d: %v", len(paths), paths)
	}
}

func TestReadRelevantFiles_ReadsContent(t *testing.T) {
	dir := t.TempDir()

	content := "package main\n\nfunc main() {}\n"
	filePath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result := service.ReadRelevantFiles([]string{filePath}, dir)

	if len(result) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result))
	}
	if result["main.go"] != content {
		t.Errorf("unexpected content: %q", result["main.go"])
	}
}

func TestReadRelevantFiles_SkipsLargeFiles(t *testing.T) {
	dir := t.TempDir()

	// Create a file larger than MaxRelevantFileSize
	largeContent := make([]byte, service.MaxRelevantFileSize+1)
	filePath := filepath.Join(dir, "large.go")
	if err := os.WriteFile(filePath, largeContent, 0644); err != nil {
		t.Fatal(err)
	}

	result := service.ReadRelevantFiles([]string{filePath}, dir)

	if len(result) != 0 {
		t.Errorf("expected 0 files (skipped large file), got %d", len(result))
	}
}

func TestReadRelevantFiles_RespectsMaxTotal(t *testing.T) {
	dir := t.TempDir()

	// Create multiple files that together exceed MaxTotalFileSize
	var paths []string
	for i := 0; i < 10; i++ {
		content := make([]byte, service.MaxTotalFileSize/5)
		name := filepath.Join(dir, "file"+string(rune('0'+i))+".go")
		if err := os.WriteFile(name, content, 0644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, name)
	}

	result := service.ReadRelevantFiles(paths, dir)

	// Should have read at most 5 files (50000 / 10000 = 5)
	if len(result) > 5 {
		t.Errorf("expected at most 5 files within total limit, got %d", len(result))
	}
}
