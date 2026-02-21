package compat

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ScanProject scans all user source files for compatibility issues.
func ScanProject(dir string) ([]Issue, error) {
	var issues []Issue

	// Find source files
	patterns := []string{"*.js", "*.css", "*.html", "*.htm"}
	var files []string

	for _, pattern := range patterns {
		matches, _ := filepath.Glob(filepath.Join(dir, "src", pattern))
		files = append(files, matches...)
		// Also check subdirectories
		filepath.Walk(filepath.Join(dir, "src"), func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			for _, p := range patterns {
				if matched, _ := filepath.Match(p, filepath.Base(path)); matched {
					// Avoid duplicates
					for _, f := range files {
						if f == path {
							return nil
						}
					}
					files = append(files, path)
					break
				}
			}
			_ = ext
			return nil
		})
	}

	for _, file := range files {
		fileIssues, err := scanFile(file, dir)
		if err != nil {
			continue
		}
		issues = append(issues, fileIssues...)
	}

	return issues, nil
}

func scanFile(path, projectDir string) ([]Issue, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	if ext == "htm" {
		ext = "html"
	}

	relPath, _ := filepath.Rel(projectDir, path)

	var issues []Issue
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		for _, rule := range Rules {
			if !matchesFileType(rule.FileTypes, ext) {
				continue
			}

			for _, pattern := range rule.Patterns {
				re, err := regexp.Compile(pattern)
				if err != nil {
					continue
				}
				if re.MatchString(line) {
					issues = append(issues, Issue{
						File:     relPath,
						Line:     lineNum,
						Rule:     rule,
						Severity: rule.Severity,
						Title:    rule.Title,
						Fix:      rule.Fix,
						AutoFix:  rule.AutoFix,
					})
					break // Only report once per rule per line
				}
			}
		}
	}

	return issues, nil
}

func matchesFileType(types []string, ext string) bool {
	for _, t := range types {
		if t == ext {
			return true
		}
	}
	return false
}
