package erlang

import (
	"bufio"
	"os"
	"strings"
)

type ToolVersionsParser struct{}

func NewToolVersionsParser() ToolVersionsParser {
	return ToolVersionsParser{}
}

func (p ToolVersionsParser) ParseVersion(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[0] == "erlang" {
			version := parts[1]
			return version, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", nil
}
