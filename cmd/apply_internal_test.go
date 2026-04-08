package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyPatchToProject_Success(t *testing.T) {
	dir := t.TempDir()

	// Create original file
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create patch file
	patchContent := `--- a/hello.txt
+++ b/hello.txt
@@ -1 +1 @@
-hello
+hello world
`
	patchPath := filepath.Join(dir, "test.patch")
	if err := os.WriteFile(patchPath, []byte(patchContent), 0644); err != nil {
		t.Fatal(err)
	}

	err := applyPatchToProject(dir, patchPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was modified
	content, err := os.ReadFile(filepath.Join(dir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "hello world\n" {
		t.Errorf("expected 'hello world\\n', got %q", string(content))
	}
}

func TestApplyPatchToProject_InvalidPatch(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello\n"), 0644); err != nil {
		t.Fatal(err)
	}

	patchPath := filepath.Join(dir, "bad.patch")
	if err := os.WriteFile(patchPath, []byte("not a valid patch"), 0644); err != nil {
		t.Fatal(err)
	}

	err := applyPatchToProject(dir, patchPath)
	if err == nil {
		t.Fatal("expected error for invalid patch")
	}
}

func TestApplyPatchToProject_MissingPatchFile(t *testing.T) {
	dir := t.TempDir()

	err := applyPatchToProject(dir, "/nonexistent/patch.patch")
	if err == nil {
		t.Fatal("expected error for missing patch file")
	}
}
