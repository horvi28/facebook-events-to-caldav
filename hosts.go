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

	// Parse into a yaml.Node instead of a map, because Go randomizes map
	// iteration order. The categories have to be walked in the order they appear
	// in the file, otherwise a host listed under two categories would get a
	// random one of them, and with it a random color, on every run.
	// However the expected behaviour would be to choose the first category where the host is listed.
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("parse hosts file %s: %w", hostsFilePath, err)
	}

	// Init the hostConfig that will be used in the rest of the app
	hostConfig := &HostConfig{
		Hosts:      []string{},
		CategoryOf: map[string]string{},
		ColorOf:    map[string]string{},
	}

	// An empty file parses into a document without any content
	if len(document.Content) == 0 {
		return hostConfig, nil
	}

	categories := document.Content[0]
	if categories.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("parse hosts file %s: expected categories as a mapping at the top level", hostsFilePath)
	}

	// A mapping node keeps its children as alternating key and value nodes,
	// so every second node is a category name and the one after it its content
	for index := 0; index+1 < len(categories.Content); index += 2 {
		categoryName := categories.Content[index].Value

		var categoryObject category
		if err := categories.Content[index+1].Decode(&categoryObject); err != nil {
			return nil, fmt.Errorf("parse category %s in hosts file %s: %w", categoryName, hostsFilePath, err)
		}

		hostConfig.ColorOf[categoryName] = categoryObject.Color

		for _, host := range categoryObject.Hosts {
			// Check if the host is already added
			if _, seen := hostConfig.CategoryOf[host]; seen {
				continue // already added under an earlier category, skip it
			}
			// Flatten the hosts URLs and add the category to the map
			hostConfig.Hosts = append(hostConfig.Hosts, host)
			hostConfig.CategoryOf[host] = categoryName
		}
	}

	return hostConfig, nil
}
