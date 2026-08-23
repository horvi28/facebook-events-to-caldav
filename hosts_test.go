package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadHostConfig(t *testing.T) {
	yaml := `
Music:
  color: "#1DB954"
  hosts:
    - https://fb.com/a
    - https://fb.com/b
Sports:
  hosts:
    - https://fb.com/c
    - https://fb.com/a   # duplicate, already under Music
`
	path := filepath.Join(t.TempDir(), "hosts.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	config, err := loadHostConfig(path)
	if err != nil {
		t.Fatalf("loadHostConfig: %v", err)
	}

	if len(config.Hosts) != 3 {
		t.Errorf("Hosts = %v, want 3 unique", config.Hosts)
	}
	if category := config.CategoryOf["https://fb.com/a"]; category != "Music" {
		t.Errorf("First category should win in case of duplicates, got %q", category)
	}
	if category := config.CategoryOf["https://fb.com/c"]; category != "Sports" {
		t.Errorf("Profile 'c' should have category of 'Sports', got %q", category)
	}
	if color := config.ColorOf["Music"]; color != "#1DB954" {
		t.Errorf("Music should be #1DB954, got %q", color)
	}
	if color := config.ColorOf["Sports"]; color != "" {
		t.Errorf("Sports color should be empty, got %q", color)
	}
}
