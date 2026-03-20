package scanner

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const maxJSONDepth = 20

// Options controls translation and source scanning behavior.
type Options struct {
	TranslationsPath string
	SourcePath       string
	Prefix           string
	Extensions       []string
	ExcludeDirs      []string
}

// Result is the command output model for both text and JSON formats.
type Result struct {
	TranslationKeyCount int      `json:"translationKeyCount"`
	UsedKeyCount        int      `json:"usedKeyCount"`
	UnusedKeys          []string `json:"unusedKeys"`
	UnknownUsedKeys     []string `json:"unknownUsedKeys"`
}

// FindUnusedKeys parses translation keys and scans source files for used keys.
func FindUnusedKeys(opts Options) (Result, error) {
	if opts.Prefix == "" {
		return Result{}, fmt.Errorf("prefix must not be empty")
	}

	translationFiles, err := collectTranslationFiles(opts.TranslationsPath)
	if err != nil {
		return Result{}, err
	}

	translationKeys := make(map[string]struct{})
	for _, path := range translationFiles {
		keys, err := parseTranslationJSON(path, opts.Prefix)
		if err != nil {
			return Result{}, err
		}
		for key := range keys {
			translationKeys[key] = struct{}{}
		}
	}

	usedKeys, err := scanUsedKeys(opts.SourcePath, opts.Prefix, opts.Extensions, opts.ExcludeDirs)
	if err != nil {
		return Result{}, err
	}

	unusedKeys := make([]string, 0)
	for key := range translationKeys {
		if _, used := usedKeys[key]; !used {
			unusedKeys = append(unusedKeys, key)
		}
	}

	unknownUsed := make([]string, 0)
	for key := range usedKeys {
		if _, known := translationKeys[key]; !known {
			unknownUsed = append(unknownUsed, key)
		}
	}

	sort.Strings(unusedKeys)
	sort.Strings(unknownUsed)

	return Result{
		TranslationKeyCount: len(translationKeys),
		UsedKeyCount:        len(usedKeys),
		UnusedKeys:          unusedKeys,
		UnknownUsedKeys:     unknownUsed,
	}, nil
}

// Collect all JSON translation files from the given path, which can be a single file or a directory.
func collectTranslationFiles(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat translations path %q: %w", path, err)
	}

	if !info.IsDir() {
		if strings.EqualFold(filepath.Ext(path), ".json") {
			return []string{path}, nil
		}

		return nil, fmt.Errorf("translations file must be .json: %s", path)
	}

	files := make([]string, 0)
	err = filepath.WalkDir(path, func(filePath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			return nil
		}

		if strings.EqualFold(filepath.Ext(filePath), ".json") {
			files = append(files, filePath)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walk translations path %q: %w", path, err)
	}

	sort.Strings(files)
	if len(files) == 0 {
		return nil, fmt.Errorf("no JSON translation files found in %s", path)
	}

	return files, nil
}

// Parses a JSON translation file and returns the translation keys.
func parseTranslationJSON(path, prefix string) (map[string]struct{}, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read translation file %q: %w", path, err)
	}

	var raw any

	// Takes the JSON content and stores it in a generic interface. If the content is not valid JSON, an error is returned.
	if err := json.Unmarshal(content, &raw); err != nil {
		return nil, fmt.Errorf("parse JSON in %q: %w", path, err)
	}

	// Checks if the root of the JSON is a map with string keys and any values. If so, the result is stored in the variable
	// root. If not, ok will be false.
	root, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("translation file %q must contain a JSON object at root", path)
	}

	keys := make(map[string]struct{})
	if err := flattenMap(root, "", prefix, keys, 0); err != nil {
		return nil, err
	}
	return keys, nil
}

// Flattens the nested JSON structure into a flat map of translation keys with a maximum depth limit.
// The parentKey is used to build the full key path, and the prefix is used to filter keys that should be included in the result.
func flattenMap(node map[string]any, parentKey, prefix string, out map[string]struct{}, depth int) error {
	if depth > maxJSONDepth {
		return fmt.Errorf("JSON nesting exceeds maximum depth of %d", maxJSONDepth)
	}

	for key, value := range node {
		nextKey := key
		if parentKey != "" {
			nextKey = parentKey + "." + key
		}

		switch typed := value.(type) {
		case map[string]any:
			if err := flattenMap(typed, nextKey, prefix, out, depth+1); err != nil {
				return err
			}
		case string:
			if strings.HasPrefix(nextKey, prefix) {
				out[nextKey] = struct{}{}
			}
		}
	}
	return nil
}
