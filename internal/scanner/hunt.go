package scanner

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

// HuntOptions controls how unused keys are removed from translation files.
type HuntOptions struct {
	Options
	CreateBackup bool
}

// HuntResult describes what was changed by a hunt run.
type HuntResult struct {
	DetectedUnusedCount int      `json:"detectedUnusedCount"`
	RemovedCount        int      `json:"removedCount"`
	RemovedKeys         []string `json:"removedKeys"`
	FilesModified       []string `json:"filesModified"`
	BackupsCreated      []string `json:"backupsCreated"`
}

// HuntUnusedKeys finds and removes all currently unused translation keys.
func HuntUnusedKeys(opts HuntOptions) (HuntResult, error) {
	detectResult, err := FindUnusedKeys(opts.Options)
	if err != nil {
		return HuntResult{}, err
	}

	unusedSet := make(map[string]struct{}, len(detectResult.UnusedKeys))
	for _, key := range detectResult.UnusedKeys {
		unusedSet[key] = struct{}{}
	}

	result := HuntResult{
		DetectedUnusedCount: len(detectResult.UnusedKeys),
	}
	if len(unusedSet) == 0 {
		return result, nil
	}

	translationFiles, err := collectTranslationFiles(opts.TranslationsPath)
	if err != nil {
		return HuntResult{}, err
	}

	removedKeys := make(map[string]struct{})
	for _, path := range translationFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			return HuntResult{}, fmt.Errorf("read translation file %q: %w", path, err)
		}

		var raw any
		if err := json.Unmarshal(content, &raw); err != nil {
			return HuntResult{}, fmt.Errorf("parse JSON in %q: %w", path, err)
		}

		root, ok := raw.(map[string]any)
		if !ok {
			return HuntResult{}, fmt.Errorf("translation file %q must contain a JSON object at root", path)
		}

		removedInFile := removeUnusedFromMap(root, "", unusedSet, removedKeys)
		if removedInFile == 0 {
			continue
		}

		if opts.CreateBackup {
			backupPath, err := createBackup(path, content)
			if err != nil {
				return HuntResult{}, err
			}
			result.BackupsCreated = append(result.BackupsCreated, backupPath)
		}

		updated, err := json.MarshalIndent(root, "", "  ")
		if err != nil {
			return HuntResult{}, fmt.Errorf("marshal updated JSON in %q: %w", path, err)
		}

		info, err := os.Stat(path)
		if err != nil {
			return HuntResult{}, fmt.Errorf("stat translation file %q: %w", path, err)
		}

		if err := os.WriteFile(path, append(updated, '\n'), info.Mode().Perm()); err != nil {
			return HuntResult{}, fmt.Errorf("write updated translation file %q: %w", path, err)
		}

		result.RemovedCount += removedInFile
		result.FilesModified = append(result.FilesModified, path)
	}

	for key := range removedKeys {
		result.RemovedKeys = append(result.RemovedKeys, key)
	}

	sort.Strings(result.RemovedKeys)
	sort.Strings(result.FilesModified)
	sort.Strings(result.BackupsCreated)
	return result, nil
}

func removeUnusedFromMap(node map[string]any, parentKey string, unusedSet map[string]struct{}, removedKeys map[string]struct{}) int {
	removed := 0
	for key, value := range node {
		nextKey := key
		if parentKey != "" {
			nextKey = parentKey + "." + key
		}

		switch typed := value.(type) {
		case map[string]any:
			removed += removeUnusedFromMap(typed, nextKey, unusedSet, removedKeys)
		case string:
			if _, shouldRemove := unusedSet[nextKey]; shouldRemove {
				delete(node, key)
				removed++
				removedKeys[nextKey] = struct{}{}
			}
		}
	}
	return removed
}

func createBackup(path string, content []byte) (string, error) {
	backupPath := fmt.Sprintf("%s.bak.%d", path, time.Now().UnixNano())

	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("stat translation file %q: %w", path, err)
	}

	if err := os.WriteFile(backupPath, content, info.Mode().Perm()); err != nil {
		return "", fmt.Errorf("create backup for %q: %w", path, err)
	}

	return backupPath, nil
}
