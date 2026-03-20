package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// RestoreOptions controls how backup files are restored.
type RestoreOptions struct {
	TranslationsPath string
}

// RestoreResult describes what was restored from backups.
type RestoreResult struct {
	FilesRestored        []string `json:"filesRestored"`
	BackupsUsed          []string `json:"backupsUsed"`
	MissingBackups       []string `json:"missingBackups"`
	RestoredCount        int      `json:"restoredCount"`
	TranslationFileCount int      `json:"translationFileCount"`
}

// RestoreFromBackup restores translation files from their latest available backup.
func RestoreFromBackup(opts RestoreOptions) (RestoreResult, error) {
	translationFiles, err := collectTranslationFiles(opts.TranslationsPath)
	if err != nil {
		return RestoreResult{}, err
	}

	result := RestoreResult{TranslationFileCount: len(translationFiles)}

	for _, translationFile := range translationFiles {
		backupPath, ok, err := findBestBackupForTranslation(translationFile)
		if err != nil {
			return RestoreResult{}, err
		}
		if !ok {
			result.MissingBackups = append(result.MissingBackups, translationFile)
			continue
		}

		backupContent, err := os.ReadFile(backupPath)
		if err != nil {
			return RestoreResult{}, fmt.Errorf("read backup file %q: %w", backupPath, err)
		}

		targetInfo, err := os.Stat(translationFile)
		if err != nil {
			return RestoreResult{}, fmt.Errorf("stat translation file %q: %w", translationFile, err)
		}

		if err := os.WriteFile(translationFile, backupContent, targetInfo.Mode().Perm()); err != nil {
			return RestoreResult{}, fmt.Errorf("restore translation file %q from backup %q: %w", translationFile, backupPath, err)
		}

		result.FilesRestored = append(result.FilesRestored, translationFile)
		result.BackupsUsed = append(result.BackupsUsed, backupPath)
		result.RestoredCount++
	}

	sort.Strings(result.FilesRestored)
	sort.Strings(result.BackupsUsed)
	sort.Strings(result.MissingBackups)

	return result, nil
}

func findBestBackupForTranslation(translationFile string) (string, bool, error) {
	exactBackup := translationFile + ".bak"
	if _, err := os.Stat(exactBackup); err == nil {
		return exactBackup, true, nil
	} else if !os.IsNotExist(err) {
		return "", false, fmt.Errorf("stat backup file %q: %w", exactBackup, err)
	}

	backupCandidates, err := filepath.Glob(translationFile + ".bak.*")
	if err != nil {
		return "", false, fmt.Errorf("glob backup files for %q: %w", translationFile, err)
	}
	if len(backupCandidates) == 0 {
		return "", false, nil
	}

	latestBackup := ""
	var latestModTime int64
	for _, candidate := range backupCandidates {
		info, err := os.Stat(candidate)
		if err != nil {
			return "", false, fmt.Errorf("stat backup file %q: %w", candidate, err)
		}

		modUnixNano := info.ModTime().UnixNano()
		if latestBackup == "" || modUnixNano > latestModTime {
			latestBackup = candidate
			latestModTime = modUnixNano
		}
	}

	return latestBackup, true, nil
}
