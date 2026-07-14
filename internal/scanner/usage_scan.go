package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Scans the source directory for used translation keys based on the provided prefix and file extensions.
func scanUsedKeys(sourcePath, prefix string, exts, excludedDirs []string) (map[string]struct{}, error) {
	if sourcePath == "" {
		return nil, fmt.Errorf("source path must not be empty")
	}

	extensionsSet := make(map[string]struct{}, len(exts))
	for _, ext := range exts {
		normalized := strings.ToLower(strings.TrimSpace(ext))
		if normalized == "" {
			continue
		}

		if !strings.HasPrefix(normalized, ".") {
			normalized = "." + normalized
		}

		extensionsSet[normalized] = struct{}{}
	}

	excludeSet := make(map[string]struct{}, len(excludedDirs))
	for _, dir := range excludedDirs {
		normalized := strings.TrimSpace(dir)
		if normalized == "" {
			continue
		}

		excludeSet[normalized] = struct{}{}
	}

	keyPatterns := quotedTranslationKeyPatterns(prefix)
	used := make(map[string]struct{})
	err := filepath.WalkDir(sourcePath, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			if _, excluded := excludeSet[d.Name()]; excluded {
				return fs.SkipDir
			}

			return nil
		}

		if len(extensionsSet) > 0 {
			if _, ok := extensionsSet[strings.ToLower(filepath.Ext(path))]; !ok {
				return nil
			}
		}

		fileKeys, err := extractUsedKeysFromFile(path, keyPatterns)
		if err != nil {
			return err
		}

		for key := range fileKeys {
			used[key] = struct{}{}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("scan source path %q: %w", sourcePath, err)
	}

	return used, nil
}

func quotedTranslationKeyPatterns(prefix string) []*regexp.Regexp {
	escapedPrefix := regexp.QuoteMeta(prefix)
	return []*regexp.Regexp{
		regexp.MustCompile(`"(` + escapedPrefix + `[^"\\]*(?:\\.[^"\\]*)*)"`),
		regexp.MustCompile(`'(` + escapedPrefix + `[^'\\]*(?:\\.[^'\\]*)*)'`),
		regexp.MustCompile("`(" + escapedPrefix + "[^`\\\\]*(?:\\\\.[^`\\\\]*)*)`"),
	}
}

// Extracts quoted translation keys from a source file.
func extractUsedKeysFromFile(path string, keyPatterns []*regexp.Regexp) (map[string]struct{}, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read source file %q: %w", path, err)
	}

	used := make(map[string]struct{})
	for _, pattern := range keyPatterns {
		matches := pattern.FindAllStringSubmatch(string(content), -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			used[match[1]] = struct{}{}
		}
	}

	return used, nil
}
