package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"ghostbuster/internal/scanner"

	"github.com/spf13/cobra"
)

const (
	issuesFoundExitCode = 2
)

type cliExitError struct {
	msg  string
	code int
}

func (e cliExitError) Error() string {
	return e.msg
}

func (e cliExitError) ExitCode() int {
	return e.code
}

var (
	translationsPath string
	sourcePath       string
	format           string
	prefix           string
	extensions       []string
	excludeDirs      []string
)

var detectCmd = &cobra.Command{
	Use:          "detect",
	Short:        "Find translation key ghosts in your codebase and translation files",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDetect(detectRunConfig{
			Options: scanner.Options{
				TranslationsPath: translationsPath,
				SourcePath:       sourcePath,
				Prefix:           prefix,
				Extensions:       extensions,
				ExcludeDirs:      excludeDirs,
			},
			Format: format,
		})
	},
}

type detectRunConfig struct {
	Options scanner.Options
	Format  string
}

func runDetect(cfg detectRunConfig) error {
	result, err := scanner.FindUnusedKeys(cfg.Options)
	if err != nil {
		return err
	}

	switch strings.ToLower(cfg.Format) {
	case "text":
		printTextResult(result)
	case "json":
		payload, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal JSON output: %w", err)
		}
		fmt.Println(string(payload))
	default:
		return errors.New("unsupported format: use text or json")
	}

	if len(result.UnusedKeys) > 0 || hasMissingTranslations(result.MissingKeys, result.MissingByFile) {
		return cliExitError{msg: "translation key issues found", code: issuesFoundExitCode}
	}

	return nil
}

func printTextResult(result scanner.Result) {
	sort.Strings(result.UnusedKeys)

	fmt.Printf("Translation keys: %d\n", result.TranslationKeyCount)
	fmt.Printf("Used keys: %d\n", result.UsedKeyCount)
	fmt.Printf("Unused keys: %d\n", len(result.UnusedKeys))
	printMissingTranslationCounts(result.MissingKeys, result.MissingByFile)
	if len(result.UnusedKeys) > 0 {
		fmt.Println()
		fmt.Println("Unused translation keys:")
		for _, key := range result.UnusedKeys {
			fmt.Printf("- %s\n", key)
		}
	}

	printMissingTranslationDetails(result.MissingKeys, result.MissingByFile)
}

func hasMissingTranslations(missingKeys []string, missingByFile []scanner.MissingTranslation) bool {
	return len(missingKeys) > 0 || len(missingByFile) > 0
}

func printMissingTranslationCounts(missingKeys []string, missingByFile []scanner.MissingTranslation) {
	missingInFiles := 0
	for _, missing := range missingByFile {
		missingInFiles += len(missing.Keys)
	}

	fmt.Printf("Missing referenced keys: %d\n", len(missingKeys))
	fmt.Printf("Missing translations in files: %d\n", missingInFiles)
}

func printMissingTranslationDetails(missingKeys []string, missingByFile []scanner.MissingTranslation) {
	sort.Strings(missingKeys)
	for i := range missingByFile {
		sort.Strings(missingByFile[i].Keys)
	}

	if len(missingKeys) > 0 {
		fmt.Println()
		fmt.Println("Referenced keys not found in any translation file:")
		for _, key := range missingKeys {
			fmt.Printf("- %s\n", key)
		}
	}

	if len(missingByFile) > 0 {
		fmt.Println()
		fmt.Println("Referenced keys missing from translation files:")
		for _, missing := range missingByFile {
			fmt.Printf("- %s:\n", missing.File)
			for _, key := range missing.Keys {
				fmt.Printf("  - %s\n", key)
			}
		}
	}
}

// Sets up the command flags and add the command to the root command.
func init() {
	rootCmd.AddCommand(detectCmd)

	detectCmd.Flags().StringVarP(&translationsPath, "translations", "t", "", "Path to translation JSON file or directory")
	detectCmd.Flags().StringVarP(&sourcePath, "source", "s", "", "Path to frontend source directory")
	detectCmd.Flags().StringVar(&format, "format", "text", "Output format: text or json")
	detectCmd.Flags().StringVar(&prefix, "prefix", "_globalTranslations.", "Translation key prefix to analyze")
	detectCmd.Flags().StringSliceVar(&extensions, "ext", []string{".js", ".jsx", ".ts", ".tsx", ".vue", ".svelte", ".html"}, "Source file extensions to scan")
	detectCmd.Flags().StringSliceVar(&excludeDirs, "exclude-dir", []string{"node_modules", "dist", "build", "coverage", ".next"}, "Directory names to skip while scanning source files")

	// Those functions return an error if the flag doesn't exist, but we just defined them, so we can ignore the error.
	_ = detectCmd.MarkFlagRequired("translations")
	_ = detectCmd.MarkFlagRequired("source")
}
