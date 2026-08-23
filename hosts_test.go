package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// loadHostConfigFromYAML writes the given YAML into a temporary hosts file and
// loads it.
func loadHostConfigFromYAML(t *testing.T, content string) *HostConfig {
	t.Helper()

	path := filepath.Join(t.TempDir(), "hosts.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write hosts file: %v", err)
	}

	config, err := loadHostConfig(path)
	if err != nil {
		t.Fatalf("loadHostConfig: %v", err)
	}

	return config
}

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
		t.Fatalf("write hosts file: %v", err)
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

const categoryCountForOrderTests = 20

// TestLoadHostConfigKeepsCategoryOrder pins down the ordering in the YAML.
func TestLoadHostConfigKeepsCategoryOrder(t *testing.T) {
	// Construct the YAML file manually with a specified host order
	var hostsYAML strings.Builder
	expectedHosts := make([]string, 0, categoryCountForOrderTests)
	for index := range categoryCountForOrderTests {
		host := fmt.Sprintf("https://fb.com/%02d", index)
		fmt.Fprintf(&hostsYAML, "Category%02d:\n  hosts:\n    - %s\n", index, host)
		expectedHosts = append(expectedHosts, host)
	}

	config := loadHostConfigFromYAML(t, hostsYAML.String())

	if !slices.Equal(config.Hosts, expectedHosts) {
		t.Errorf("Hosts should follow the order of the file, got %v, want %v", config.Hosts, expectedHosts)
	}
}

// TestLoadHostConfigFirstCategoryWinsForDuplicates covers the rule that a
// profile listed several times keeps the category it appears under first.
// The duplicate is put under every category, so only walking them in file
// order can give the expected answer.
func TestLoadHostConfigFirstCategoryWinsForDuplicates(t *testing.T) {
	const duplicatedHost = "https://fb.com/duplicate"

	// Add the duplicate to every host, then a unique host to every category
	var hostsYAML strings.Builder
	for index := range categoryCountForOrderTests {
		uniqueHost := fmt.Sprintf("https://fb.com/only-%02d", index)
		fmt.Fprintf(&hostsYAML, "Category%02d:\n  hosts:\n    - %s\n    - %s\n", index, duplicatedHost, uniqueHost)
	}

	config := loadHostConfigFromYAML(t, hostsYAML.String())

	if category := config.CategoryOf[duplicatedHost]; category != "Category00" {
		t.Errorf("Duplicated host should keep its first category 'Category00', got %q", category)
	}

	// A duplicate must exist only once in the host list
	occurrences := 0
	for _, host := range config.Hosts {
		if host == duplicatedHost {
			occurrences++
		}
	}
	if occurrences != 1 {
		t.Errorf("Duplicated host should be listed once, got %d times", occurrences)
	}

	// Every category should have a color entry (although all of them is an empty string)
	if len(config.ColorOf) != categoryCountForOrderTests {
		t.Errorf("Every category should have a color entry, got %d", len(config.ColorOf))
	}
}

func TestLoadHostConfigWithEmptyFile(t *testing.T) {
	config := loadHostConfigFromYAML(t, "")

	if len(config.Hosts) != 0 {
		t.Errorf("An empty file should give no hosts, got %v", config.Hosts)
	}
	if len(config.CategoryOf) != 0 {
		t.Errorf("An empty file should give no categories, got %v", config.CategoryOf)
	}
	if len(config.ColorOf) != 0 {
		t.Errorf("An empty file should give no colors, got %v", config.ColorOf)
	}
}

func TestLoadHostConfigWithoutCategoryMapping(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hosts.yaml")
	// A list of hosts without any category around them is not a valid file
	if err := os.WriteFile(path, []byte("- https://fb.com/a\n"), 0o644); err != nil {
		t.Fatalf("write hosts file: %v", err)
	}

	if _, err := loadHostConfig(path); err == nil {
		t.Error("A file without categories should fail, got no error")
	}
}
