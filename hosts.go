package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func getHostList(hostsFilePath string) (hostList []string, err error) {
	file, err := os.Open(hostsFilePath)
	if err != nil {
		return nil, fmt.Errorf("open hosts file %s: %w", hostsFilePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close hosts file %s: %w", hostsFilePath, closeErr)
		}
	}()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments (lines starting with #)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read hosts file %s: %w", hostsFilePath, err)
	}

	return lines, nil
}
