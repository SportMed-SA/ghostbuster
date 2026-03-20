package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHuntUnusedKeysCreatesBackupAndRemovesKeys(t *testing.T) {
	tempDir := t.TempDir()
	translationsDir := filepath.Join(tempDir, "translations")
	sourceDir := filepath.Join(tempDir, "src")

	if err := os.MkdirAll(translationsDir, 0o755); err != nil {
		t.Fatalf("mkdir translations dir: %v", err)
	}
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("mkdir source dir: %v", err)
	}

	translationFile := filepath.Join(translationsDir, "en.json")
	translations := `{
  "_globalTranslations": {
    "used": "A",
    "unused": "B"
  }
}`
	source := `const title = "_globalTranslations.used";`

	if err := os.WriteFile(translationFile, []byte(translations), 0o644); err != nil {
		t.Fatalf("write translation file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "app.ts"), []byte(source), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	result, err := HuntUnusedKeys(HuntOptions{
		Options: Options{
			TranslationsPath: translationsDir,
			SourcePath:       sourceDir,
			Prefix:           "_globalTranslations.",
			Extensions:       []string{".ts"},
		},
		CreateBackup: true,
	})
	if err != nil {
		t.Fatalf("hunt unused keys: %v", err)
	}

	if result.RemovedCount != 1 {
		t.Fatalf("removed count mismatch: got %d, want 1", result.RemovedCount)
	}
	if len(result.BackupsCreated) != 1 {
		t.Fatalf("backup count mismatch: got %d, want 1", len(result.BackupsCreated))
	}

	backupPath := result.BackupsCreated[0]
	if !strings.HasPrefix(backupPath, translationFile+".bak.") {
		t.Fatalf("expected timestamped backup path, got %q", backupPath)
	}
	if _, err := os.Stat(backupPath); err != nil {
		t.Fatalf("expected backup file: %v", err)
	}

	detectResult, err := FindUnusedKeys(Options{
		TranslationsPath: translationsDir,
		SourcePath:       sourceDir,
		Prefix:           "_globalTranslations.",
		Extensions:       []string{".ts"},
	})
	if err != nil {
		t.Fatalf("detect after hunt: %v", err)
	}
	if len(detectResult.UnusedKeys) != 0 {
		t.Fatalf("expected no unused keys after hunt, got %v", detectResult.UnusedKeys)
	}
}

func TestHuntUnusedKeysNoBackup(t *testing.T) {
	tempDir := t.TempDir()
	translationsDir := filepath.Join(tempDir, "translations")
	sourceDir := filepath.Join(tempDir, "src")

	if err := os.MkdirAll(translationsDir, 0o755); err != nil {
		t.Fatalf("mkdir translations dir: %v", err)
	}
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("mkdir source dir: %v", err)
	}

	translationFile := filepath.Join(translationsDir, "en.json")
	translations := `{
  "_globalTranslations": {
    "unused": "B"
  }
}`

	if err := os.WriteFile(translationFile, []byte(translations), 0o644); err != nil {
		t.Fatalf("write translation file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "app.ts"), []byte("const x = 'nothing';"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	result, err := HuntUnusedKeys(HuntOptions{
		Options: Options{
			TranslationsPath: translationsDir,
			SourcePath:       sourceDir,
			Prefix:           "_globalTranslations.",
			Extensions:       []string{".ts"},
		},
		CreateBackup: false,
	})
	if err != nil {
		t.Fatalf("hunt unused keys: %v", err)
	}

	if result.RemovedCount != 1 {
		t.Fatalf("removed count mismatch: got %d, want 1", result.RemovedCount)
	}
	if len(result.BackupsCreated) != 0 {
		t.Fatalf("expected no backups, got %d", len(result.BackupsCreated))
	}
	if matches, err := filepath.Glob(translationFile + ".bak*"); err != nil {
		t.Fatalf("glob backup files: %v", err)
	} else if len(matches) > 0 {
		t.Fatalf("unexpected backup file created: %v", matches)
	}
}
