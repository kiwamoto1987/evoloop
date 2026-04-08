package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kiwamoto1987/evoloop/cmd"
)

func TestInitCommand_CreatesConfig(t *testing.T) {
	dir := t.TempDir()

	// Change to temp dir for the test
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(original) }()

	if err := cmd.RunInit(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify config file exists
	configPath := filepath.Join(dir, ".evoloop", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config.yaml was not created")
	}

	// Verify runtime directories exist
	for _, subdir := range []string{"logs", "patches", "prompts", "reports"} {
		runtimeDir := filepath.Join(dir, ".evoloop", "runtime", subdir)
		if _, err := os.Stat(runtimeDir); os.IsNotExist(err) {
			t.Errorf("runtime directory %s was not created", subdir)
		}
	}
}

func TestInitCommand_AlreadyExists(t *testing.T) {
	dir := t.TempDir()

	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(original) }()

	// First init should succeed
	if err := cmd.RunInit(); err != nil {
		t.Fatalf("first init failed: %v", err)
	}

	// Second init should fail
	if err := cmd.RunInit(); err == nil {
		t.Fatal("expected error for duplicate init")
	}
}
