package scanner

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRestoreFromBackupRestoresExactBackup(t *testing.T) {
	tempDir := t.TempDir()
	translationsDir := filepath.Join(tempDir, "translations")
	if err := os.MkdirAll(translationsDir, 0o755); err != nil {
		t.Fatalf("mkdir translations dir: %v", err)
	}

	translationFile := filepath.Join(translationsDir, "en.json")
	original := []byte("{\n  \"key\": \"current\"\n}\n")
	backup := []byte("{\n  \"key\": \"backup\"\n}\n")

	if err := os.WriteFile(translationFile, original, 0o644); err != nil {
		t.Fatalf("write translation file: %v", err)
	}
	if err := os.WriteFile(translationFile+".bak", backup, 0o644); err != nil {
		t.Fatalf("write backup file: %v", err)
	}

	result, err := RestoreFromBackup(RestoreOptions{TranslationsPath: translationsDir})
	if err != nil {
		t.Fatalf("restore from backup: %v", err)
	}

	if result.TranslationFileCount != 1 {
		t.Fatalf("translation file count mismatch: got %d, want 1", result.TranslationFileCount)
	}
	if result.RestoredCount != 1 {
		t.Fatalf("restored count mismatch: got %d, want 1", result.RestoredCount)
	}
	if len(result.MissingBackups) != 0 {
		t.Fatalf("expected no missing backups, got %v", result.MissingBackups)
	}

	updated, err := os.ReadFile(translationFile)
	if err != nil {
		t.Fatalf("read restored translation file: %v", err)
	}
	if string(updated) != string(backup) {
		t.Fatalf("translation file content mismatch: got %q, want %q", string(updated), string(backup))
	}
}

func TestRestoreFromBackupUsesLatestTimestampedBackup(t *testing.T) {
	tempDir := t.TempDir()
	translationsDir := filepath.Join(tempDir, "translations")
	if err := os.MkdirAll(translationsDir, 0o755); err != nil {
		t.Fatalf("mkdir translations dir: %v", err)
	}

	translationFile := filepath.Join(translationsDir, "en.json")
	if err := os.WriteFile(translationFile, []byte("{\n  \"key\": \"current\"\n}\n"), 0o644); err != nil {
		t.Fatalf("write translation file: %v", err)
	}

	olderBackup := translationFile + ".bak.1"
	newerBackup := translationFile + ".bak.2"

	if err := os.WriteFile(olderBackup, []byte("{\n  \"key\": \"older\"\n}\n"), 0o644); err != nil {
		t.Fatalf("write older backup: %v", err)
	}
	if err := os.WriteFile(newerBackup, []byte("{\n  \"key\": \"newer\"\n}\n"), 0o644); err != nil {
		t.Fatalf("write newer backup: %v", err)
	}

	olderTime := time.Now().Add(-2 * time.Minute)
	newerTime := time.Now().Add(-1 * time.Minute)
	if err := os.Chtimes(olderBackup, olderTime, olderTime); err != nil {
		t.Fatalf("set older backup time: %v", err)
	}
	if err := os.Chtimes(newerBackup, newerTime, newerTime); err != nil {
		t.Fatalf("set newer backup time: %v", err)
	}

	result, err := RestoreFromBackup(RestoreOptions{TranslationsPath: translationsDir})
	if err != nil {
		t.Fatalf("restore from backup: %v", err)
	}

	if result.RestoredCount != 1 {
		t.Fatalf("restored count mismatch: got %d, want 1", result.RestoredCount)
	}
	if len(result.BackupsUsed) != 1 || result.BackupsUsed[0] != newerBackup {
		t.Fatalf("expected newest timestamped backup to be used, got %v", result.BackupsUsed)
	}

	updated, err := os.ReadFile(translationFile)
	if err != nil {
		t.Fatalf("read restored translation file: %v", err)
	}
	if string(updated) != "{\n  \"key\": \"newer\"\n}\n" {
		t.Fatalf("translation file content mismatch: got %q", string(updated))
	}
}

func TestRestoreFromBackupMissingBackup(t *testing.T) {
	tempDir := t.TempDir()
	translationsDir := filepath.Join(tempDir, "translations")
	if err := os.MkdirAll(translationsDir, 0o755); err != nil {
		t.Fatalf("mkdir translations dir: %v", err)
	}

	translationFile := filepath.Join(translationsDir, "en.json")
	content := "{\n  \"key\": \"current\"\n}\n"
	if err := os.WriteFile(translationFile, []byte(content), 0o644); err != nil {
		t.Fatalf("write translation file: %v", err)
	}

	result, err := RestoreFromBackup(RestoreOptions{TranslationsPath: translationsDir})
	if err != nil {
		t.Fatalf("restore from backup: %v", err)
	}

	if result.RestoredCount != 0 {
		t.Fatalf("expected no restored files, got %d", result.RestoredCount)
	}
	if len(result.MissingBackups) != 1 || result.MissingBackups[0] != translationFile {
		t.Fatalf("missing backup mismatch: got %v", result.MissingBackups)
	}

	updated, err := os.ReadFile(translationFile)
	if err != nil {
		t.Fatalf("read translation file: %v", err)
	}
	if string(updated) != content {
		t.Fatalf("translation file should remain unchanged when backup is missing")
	}
}
