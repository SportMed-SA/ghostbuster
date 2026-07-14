package scanner

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
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

func TestHuntUnusedKeysReportsMissingKeysWithoutModifyingTranslations(t *testing.T) {
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
	original := []byte(`{"_globalTranslations":{"existing":"Existing"}}`)
	if err := os.WriteFile(translationFile, original, 0o644); err != nil {
		t.Fatalf("write translation file: %v", err)
	}
	source := `const existing = "_globalTranslations.existing";
const missing = "_globalTranslations.missing";`
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

	wantMissing := []string{"_globalTranslations.missing"}
	if !reflect.DeepEqual(result.MissingKeys, wantMissing) {
		t.Fatalf("missing keys mismatch: got %v, want %v", result.MissingKeys, wantMissing)
	}
	if result.RemovedCount != 0 || len(result.FilesModified) != 0 || len(result.BackupsCreated) != 0 {
		t.Fatalf("missing keys must not modify translations: %+v", result)
	}
	updated, err := os.ReadFile(translationFile)
	if err != nil {
		t.Fatalf("read translation file: %v", err)
	}
	if !reflect.DeepEqual(updated, original) {
		t.Fatalf("translation file changed: got %q, want %q", updated, original)
	}
}

func TestHuntUnusedKeysRemovesOnlyUnusedKeys(t *testing.T) {
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
    "used": "Keep me",
    "unused": "Remove me"
  }
}`
	source := `const label = "_globalTranslations.used";`

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

	if result.DetectedUnusedCount != 1 {
		t.Fatalf("detected unused count mismatch: got %d, want 1", result.DetectedUnusedCount)
	}
	if result.RemovedCount != 1 {
		t.Fatalf("removed count mismatch: got %d, want 1", result.RemovedCount)
	}

	wantRemovedKeys := []string{"_globalTranslations.unused"}
	if !reflect.DeepEqual(result.RemovedKeys, wantRemovedKeys) {
		t.Fatalf("removed keys mismatch: got %v, want %v", result.RemovedKeys, wantRemovedKeys)
	}

	wantFilesModified := []string{translationFile}
	if !reflect.DeepEqual(result.FilesModified, wantFilesModified) {
		t.Fatalf("files modified mismatch: got %v, want %v", result.FilesModified, wantFilesModified)
	}

	updated, err := os.ReadFile(translationFile)
	if err != nil {
		t.Fatalf("read updated translation file: %v", err)
	}
	updatedContent := string(updated)
	if !strings.Contains(updatedContent, `"used": "Keep me"`) {
		t.Fatalf("used translation should remain after hunt, got %q", updatedContent)
	}
	if strings.Contains(updatedContent, "unused") || strings.Contains(updatedContent, "Remove me") {
		t.Fatalf("unused translation should be removed after hunt, got %q", updatedContent)
	}
}

func TestHuntUnusedKeysPreservesAmpersandInExistingValues(t *testing.T) {
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
    "used": "Save & continue",
    "unused": "Remove me"
  }
}`
	source := `const label = "_globalTranslations.used";`

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
		CreateBackup: false,
	})
	if err != nil {
		t.Fatalf("hunt unused keys: %v", err)
	}
	if result.RemovedCount != 1 {
		t.Fatalf("removed count mismatch: got %d, want 1", result.RemovedCount)
	}

	updated, err := os.ReadFile(translationFile)
	if err != nil {
		t.Fatalf("read updated translation file: %v", err)
	}
	updatedContent := string(updated)
	if !strings.Contains(updatedContent, "Save & continue") {
		t.Fatalf("expected literal ampersand to be preserved, got %q", updatedContent)
	}
	if strings.Contains(updatedContent, `\u0026`) {
		t.Fatalf("expected ampersand not to be escaped as \\u0026, got %q", updatedContent)
	}
}

func TestHuntUnusedKeysBackupPreservesOriginalBytes(t *testing.T) {
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
	original := []byte(`{
  "_globalTranslations": {
    "used": "Save & continue",
    "unused": "Remove me"
  }
}`)
	source := `const label = "_globalTranslations.used";`

	if err := os.WriteFile(translationFile, original, 0o644); err != nil {
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
	if len(result.BackupsCreated) != 1 {
		t.Fatalf("backup count mismatch: got %d, want 1", len(result.BackupsCreated))
	}

	backupContent, err := os.ReadFile(result.BackupsCreated[0])
	if err != nil {
		t.Fatalf("read backup file: %v", err)
	}
	if !bytes.Equal(backupContent, original) {
		t.Fatalf("backup content mismatch: got %q, want %q", string(backupContent), string(original))
	}
}
