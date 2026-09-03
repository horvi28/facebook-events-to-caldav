package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setCalDAVEnvironment sets every CalDAV variable, as if they came from the
// environment instead of a file.
func setCalDAVEnvironment(t *testing.T) {
	t.Helper()

	t.Setenv("CALDAV_URL", "http://localhost:5232")
	t.Setenv("CALDAV_USERNAME", "user-from-environment")
	t.Setenv("CALDAV_PASSWORD", "password")
	t.Setenv("CALDAV_CALENDAR_PATH", "/user/calendar/")
}

// TestLoadConfigWithoutEnvFile covers a container, where the variables are given
// to the process directly and no .env file exists at all.
func TestLoadConfigWithoutEnvFile(t *testing.T) {
	t.Chdir(t.TempDir())
	setCalDAVEnvironment(t)

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig without a .env file: %v", err)
	}

	if config.Username != "user-from-environment" {
		t.Errorf("username = %q, want the one from the environment", config.Username)
	}
}

func TestLoadConfigFromEnvFile(t *testing.T) {
	directory := t.TempDir()
	envFile := "CALDAV_URL=http://localhost:5232\n" +
		"CALDAV_USERNAME=user-from-file\n" +
		"CALDAV_PASSWORD=password\n" +
		"CALDAV_CALENDAR_PATH=/user/calendar/\n"
	if err := os.WriteFile(filepath.Join(directory, ".env"), []byte(envFile), fileMode); err != nil {
		t.Fatalf("write .env file: %v", err)
	}
	t.Chdir(directory)

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}

	if config.Username != "user-from-file" {
		t.Errorf("username = %q, want the one from the .env file", config.Username)
	}
}

// TestLoadConfigWithoutTheVariables makes sure that a missing .env file is only
// tolerated as long as the variables arrive some other way.
func TestLoadConfigWithoutTheVariables(t *testing.T) {
	t.Chdir(t.TempDir())
	// Emptied instead of unset, so that the test does not depend on what the
	// shell running it happens to have set
	t.Setenv("CALDAV_URL", "")
	t.Setenv("CALDAV_USERNAME", "")
	t.Setenv("CALDAV_PASSWORD", "")
	t.Setenv("CALDAV_CALENDAR_PATH", "")

	_, err := loadConfig()
	if err == nil {
		t.Fatal("expected an error when neither a .env file nor the variables are there")
	}
	if !strings.Contains(err.Error(), "CALDAV_URL") {
		t.Errorf("error %q does not name a missing variable", err)
	}
}

// TestLoadConfigWithABrokenEnvFile makes sure a .env file that is there but
// unreadable is still reported, instead of being treated like a missing one.
func TestLoadConfigWithABrokenEnvFile(t *testing.T) {
	directory := t.TempDir()
	// A directory named .env can be opened but never parsed
	if err := os.Mkdir(filepath.Join(directory, ".env"), 0o750); err != nil {
		t.Fatalf("create .env directory: %v", err)
	}
	t.Chdir(directory)
	setCalDAVEnvironment(t)

	if _, err := loadConfig(); err == nil {
		t.Fatal("expected an unreadable .env file to be an error")
	}
}
