package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"ghostbuster/internal/scanner"
)

func TestRunDetectFailsWhenReferencedKeyIsMissing(t *testing.T) {
	tempDir := t.TempDir()
	translationsDir := filepath.Join(tempDir, "translations")
	sourceDir := filepath.Join(tempDir, "src")

	if err := os.MkdirAll(translationsDir, 0o755); err != nil {
		t.Fatalf("mkdir translations dir: %v", err)
	}
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("mkdir source dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(translationsDir, "en.json"), []byte(`{"_globalTranslations":{"existing":"Existing"}}`), 0o644); err != nil {
		t.Fatalf("write translation file: %v", err)
	}
	source := `const existing = "_globalTranslations.existing";
const missing = "_globalTranslations.missing";`
	if err := os.WriteFile(filepath.Join(sourceDir, "app.ts"), []byte(source), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	err := runDetect(detectRunConfig{
		Options: scanner.Options{
			TranslationsPath: translationsDir,
			SourcePath:       sourceDir,
			Prefix:           "_globalTranslations.",
			Extensions:       []string{".ts"},
		},
		Format: "json",
	})

	var exitErr cliExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected CLI exit error, got %v", err)
	}
	if exitErr.ExitCode() != issuesFoundExitCode {
		t.Fatalf("exit code mismatch: got %d, want %d", exitErr.ExitCode(), issuesFoundExitCode)
	}
}

func TestHasMissingTranslationsIncludesPerFileGaps(t *testing.T) {
	if !hasMissingTranslations(nil, []scanner.MissingTranslation{{
		File: "de.json",
		Keys: []string{"_globalTranslations.missing"},
	}}) {
		t.Fatal("expected a per-file translation gap to be treated as a missing translation")
	}
}
