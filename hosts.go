package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// category is one entry in the hosts YAML file: a color and the profiles under it
// This is the structure that is given by the user in the hosts.yaml file
type category struct {
	Color string   `yaml:"color"`
	Hosts []string `yaml:"hosts"`
}

// HostConfig is the flattened view of the hosts.yaml file used by the rest of the app for easier handling
type HostConfig struct {
	Hosts      []string          // All host URLs to scrape, deduplicated (the first category will win)
	CategoryOf map[string]string // Host URL -> category name
	ColorOf    map[string]string // Category name -> hex color (it is empty if unset)
}

func loadHostConfig(hostsFilePath string) (*HostConfig, error) {
	data, err := os.ReadFile(hostsFilePath)
	if err != nil {
		return nil, fmt.Errorf("read hosts file %s: %w", hostsFilePath, err)
	}

	// Parse the hosts.yaml as is
	var rawConfig map[string]category
	if err := yaml.Unmarshal(data, &rawConfig); err != nil {
		return nil, fmt.Errorf("parse hosts file %s: %w", hostsFilePath, err)
	}

	// Init the hostConfig that will be used in the rest of the app
	hostConfig := &HostConfig{
		Hosts:      []string{},
		CategoryOf: map[string]string{},
		ColorOf:    map[string]string{},
	}

	for categoryName, categoryObject := range rawConfig {
		hostConfig.ColorOf[categoryName] = categoryObject.Color

		for _, host := range categoryObject.Hosts {
			// Check if the host is already added
			if _, seen := hostConfig.CategoryOf[host]; seen {
				continue // already added under another category, skip it
			}
			// Flatten the hosts URLs and add the category to the map
			hostConfig.Hosts = append(hostConfig.Hosts, host)
			hostConfig.CategoryOf[host] = categoryName
		}
	}

	return hostConfig, nil
}
