package scanner

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFindUnusedKeys(t *testing.T) {
	tempDir := t.TempDir()
	translationsDir := filepath.Join(tempDir, "translations")
	sourceDir := filepath.Join(tempDir, "src")

	if err := os.MkdirAll(translationsDir, 0o755); err != nil {
		t.Fatalf("mkdir translations dir: %v", err)
	}
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("mkdir source dir: %v", err)
	}

	translations := `{
  "_globalTranslations": {
    "home": {
      "title": "Home",
      "subtitle": "Subtitle"
    },
    "footer": "Footer"
  },
  "other": {
    "ignored": "value"
  }
}`
	source := `const used = "_globalTranslations.home.title";
const unknown = '_globalTranslations.unknown.key';`

	if err := os.WriteFile(filepath.Join(translationsDir, "en.json"), []byte(translations), 0o644); err != nil {
		t.Fatalf("write translation file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "app.ts"), []byte(source), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	result, err := FindUnusedKeys(Options{
		TranslationsPath: translationsDir,
		SourcePath:       sourceDir,
		Prefix:           "_globalTranslations.",
		Extensions:       []string{".ts"},
		ExcludeDirs:      []string{"node_modules"},
	})
	if err != nil {
		t.Fatalf("find unused keys: %v", err)
	}

	if result.TranslationKeyCount != 3 {
		t.Fatalf("translation key count mismatch: got %d, want %d", result.TranslationKeyCount, 3)
	}
	if result.UsedKeyCount != 2 {
		t.Fatalf("used key count mismatch: got %d, want %d", result.UsedKeyCount, 2)
	}

	wantUnused := []string{"_globalTranslations.footer", "_globalTranslations.home.subtitle"}
	if !reflect.DeepEqual(result.UnusedKeys, wantUnused) {
		t.Fatalf("unused keys mismatch: got %v, want %v", result.UnusedKeys, wantUnused)
	}

	wantMissing := []string{"_globalTranslations.unknown.key"}
	if !reflect.DeepEqual(result.MissingKeys, wantMissing) {
		t.Fatalf("missing keys mismatch: got %v, want %v", result.MissingKeys, wantMissing)
	}
}

func TestFindUnusedKeysReportsMissingTranslationsByFile(t *testing.T) {
	tempDir := t.TempDir()
	translationsDir := filepath.Join(tempDir, "translations")
	sourceDir := filepath.Join(tempDir, "src")

	if err := os.MkdirAll(translationsDir, 0o755); err != nil {
		t.Fatalf("mkdir translations dir: %v", err)
	}
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("mkdir source dir: %v", err)
	}

	enFile := filepath.Join(translationsDir, "en.json")
	deFile := filepath.Join(translationsDir, "de.json")
	if err := os.WriteFile(enFile, []byte(`{"_globalTranslations":{"title":"Title","cta":"Continue"}}`), 0o644); err != nil {
		t.Fatalf("write English translations: %v", err)
	}
	if err := os.WriteFile(deFile, []byte(`{"_globalTranslations":{"title":"Titel"}}`), 0o644); err != nil {
		t.Fatalf("write German translations: %v", err)
	}
	source := `const title = "_globalTranslations.title";
const cta = "_globalTranslations.cta";
const missing = "_globalTranslations.missingEverywhere";`
	if err := os.WriteFile(filepath.Join(sourceDir, "app.ts"), []byte(source), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	result, err := FindUnusedKeys(Options{
		TranslationsPath: translationsDir,
		SourcePath:       sourceDir,
		Prefix:           "_globalTranslations.",
		Extensions:       []string{".ts"},
	})
	if err != nil {
		t.Fatalf("find unused keys: %v", err)
	}

	wantMissing := []string{"_globalTranslations.missingEverywhere"}
	if !reflect.DeepEqual(result.MissingKeys, wantMissing) {
		t.Fatalf("missing keys mismatch: got %v, want %v", result.MissingKeys, wantMissing)
	}
	wantMissingByFile := []MissingTranslation{{
		File: deFile,
		Keys: []string{"_globalTranslations.cta"},
	}}
	if !reflect.DeepEqual(result.MissingByFile, wantMissingByFile) {
		t.Fatalf("missing translations by file mismatch: got %v, want %v", result.MissingByFile, wantMissingByFile)
	}
}

func TestFindUnusedKeysDoesNotTreatPrefixAsUsage(t *testing.T) {
	tempDir := t.TempDir()
	translationsDir := filepath.Join(tempDir, "translations")
	sourceDir := filepath.Join(tempDir, "src")

	if err := os.MkdirAll(translationsDir, 0o755); err != nil {
		t.Fatalf("mkdir translations dir: %v", err)
	}
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("mkdir source dir: %v", err)
	}

	translations := `{
  "_globalTranslations": {
    "erhaltenSieIhr": "Erhalten Sie Ihr",
    "erhaltenSieIhrMassgeschneidertes": "Erhalten Sie Ihr massgeschneidertes"
  }
}`
	source := `const label = "_globalTranslations.erhaltenSieIhrMassgeschneidertes";`

	if err := os.WriteFile(filepath.Join(translationsDir, "de.json"), []byte(translations), 0o644); err != nil {
		t.Fatalf("write translation file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "app.ts"), []byte(source), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	result, err := FindUnusedKeys(Options{
		TranslationsPath: translationsDir,
		SourcePath:       sourceDir,
		Prefix:           "_globalTranslations.",
		Extensions:       []string{".ts"},
		ExcludeDirs:      []string{"node_modules"},
	})
	if err != nil {
		t.Fatalf("find unused keys: %v", err)
	}

	wantUnused := []string{"_globalTranslations.erhaltenSieIhr"}
	if !reflect.DeepEqual(result.UnusedKeys, wantUnused) {
		t.Fatalf("unused keys mismatch: got %v, want %v", result.UnusedKeys, wantUnused)
	}
}

func TestScanUsedKeysHonorsExcludedDirs(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "src")
	excludedDir := filepath.Join(sourceDir, "node_modules")
	includedDir := filepath.Join(sourceDir, "components")

	if err := os.MkdirAll(excludedDir, 0o755); err != nil {
		t.Fatalf("mkdir excluded dir: %v", err)
	}
	if err := os.MkdirAll(includedDir, 0o755); err != nil {
		t.Fatalf("mkdir included dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(excludedDir, "vendor.ts"), []byte(`const k = "_globalTranslations.in_vendor";`), 0o644); err != nil {
		t.Fatalf("write excluded file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(includedDir, "view.ts"), []byte("const k = `_globalTranslations.in_app`;"), 0o644); err != nil {
		t.Fatalf("write included file: %v", err)
	}

	used, err := scanUsedKeys(sourceDir, "_globalTranslations.", []string{".ts"}, []string{"node_modules"})
	if err != nil {
		t.Fatalf("scan used keys: %v", err)
	}

	if _, ok := used["_globalTranslations.in_vendor"]; ok {
		t.Fatalf("excluded directory key should not be scanned")
	}
	if _, ok := used["_globalTranslations.in_app"]; !ok {
		t.Fatalf("included directory key should be scanned")
	}
}

func TestScanUsedKeysIgnoresUnrelatedApostrophes(t *testing.T) {
	tempDir := t.TempDir()
	sourceFile := filepath.Join(tempDir, "component.ts")
	source := `/**
 * The form includes the user's changes.
 */
const options = {
  headerKey: '_globalTranslations.deleteAccountHeader',
  messageKey: '_globalTranslations.deleteAccountMessage',
  successMessageKey: '_globalTranslations.deleteAccountSuccessMessage',
};`
	if err := os.WriteFile(sourceFile, []byte(source), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	used, err := scanUsedKeys(tempDir, "_globalTranslations.", []string{".ts"}, nil)
	if err != nil {
		t.Fatalf("scan used keys: %v", err)
	}

	want := []string{
		"_globalTranslations.deleteAccountHeader",
		"_globalTranslations.deleteAccountMessage",
		"_globalTranslations.deleteAccountSuccessMessage",
	}
	for _, key := range want {
		if _, ok := used[key]; !ok {
			t.Errorf("expected %q to be detected, got %v", key, used)
		}
	}
}
