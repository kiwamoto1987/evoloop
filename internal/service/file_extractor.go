package service

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// MaxRelevantFileSize is the maximum size of a single file to include in the prompt context.
const MaxRelevantFileSize = 10000

// MaxTotalFileSize is the maximum total size of all relevant files combined.
const MaxTotalFileSize = 50000

// ExtractFilePaths parses error output to find referenced file paths.
func ExtractFilePaths(output string, projectRoot string) []string {
	patterns := []*regexp.Regexp{
		// Go test: "    file_test.go:10: message"
		regexp.MustCompile(`\s+(\S+\.go):(\d+)`),
		// Go build/vet: "./path/file.go:10:5: message"
		regexp.MustCompile(`\.?/?(\S+\.go):(\d+)`),
		// Generic: "path/to/file.ext:line"
		regexp.MustCompile(`(\S+\.\w+):(\d+)`),
	}

	seen := make(map[string]bool)
	var paths []string

	for _, line := range strings.Split(output, "\n") {
		for _, pattern := range patterns {
			matches := pattern.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) < 2 {
					continue
				}
				filePath := match[1]

				// Resolve to absolute path
				absPath := filePath
				if !filepath.IsAbs(filePath) {
					absPath = filepath.Join(projectRoot, filePath)
				}

				// Verify the file exists
				if _, err := os.Stat(absPath); err != nil {
					continue
				}

				if !seen[absPath] {
					seen[absPath] = true
					paths = append(paths, absPath)
				}
			}
		}
	}

	return paths
}

// ReadRelevantFiles reads files up to size limits and returns a map of relative path to content.
func ReadRelevantFiles(paths []string, projectRoot string) map[string]string {
	contents := make(map[string]string)
	totalSize := 0

	for _, absPath := range paths {
		if totalSize >= MaxTotalFileSize {
			break
		}

		info, err := os.Stat(absPath)
		if err != nil || info.Size() > MaxRelevantFileSize {
			continue
		}

		data, err := os.ReadFile(absPath)
		if err != nil {
			continue
		}

		if totalSize+len(data) > MaxTotalFileSize {
			break
		}

		relPath, err := filepath.Rel(projectRoot, absPath)
		if err != nil {
			relPath = absPath
		}

		contents[relPath] = string(data)
		totalSize += len(data)
	}

	return contents
}
