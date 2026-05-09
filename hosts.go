package main

import (
	"bufio"
	"fmt"
	"os"
)

const hostsFilePath = "hosts.txt"

func getHostList() (hostList []string, err error) {
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
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read hosts file %s: %w", hostsFilePath, err)
	}

	return lines, nil
}
