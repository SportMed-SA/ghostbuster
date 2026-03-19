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

	wantUnknown := []string{"_globalTranslations.unknown.key"}
	if !reflect.DeepEqual(result.UnknownUsedKeys, wantUnknown) {
		t.Fatalf("unknown used keys mismatch: got %v, want %v", result.UnknownUsedKeys, wantUnknown)
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
