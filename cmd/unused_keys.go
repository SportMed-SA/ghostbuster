package cmd

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
	unusedFoundExitCode = 2
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
	unusedTranslationsPath string
	unusedSourcePath       string
	unusedFormat           string
	unusedPrefix           string
	unusedExts             []string
	unusedExcludeDirs      []string
)

var unusedKeysCmd = &cobra.Command{
	Use:   "unused-keys",
	Short: "Find unused translation keys",
	Long:  "Scan nested JSON translations and frontend files to report unused translation keys.",
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := scanner.Options{
			TranslationsPath: unusedTranslationsPath,
			SourcePath:       unusedSourcePath,
			Prefix:           unusedPrefix,
			Extensions:       unusedExts,
			ExcludeDirs:      unusedExcludeDirs,
		}

		result, err := scanner.FindUnusedKeys(opts)
		if err != nil {
			return err
		}

		switch strings.ToLower(unusedFormat) {
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

		if len(result.UnusedKeys) > 0 {
			return cliExitError{msg: "unused translation keys found", code: unusedFoundExitCode}
		}
		return nil
	},
}

func printTextResult(result scanner.Result) {
	sort.Strings(result.UnusedKeys)
	sort.Strings(result.UnknownUsedKeys)

	fmt.Printf("Translation keys: %d\n", result.TranslationKeyCount)
	fmt.Printf("Used keys: %d\n", result.UsedKeyCount)
	fmt.Printf("Unused keys: %d\n", len(result.UnusedKeys))
	if len(result.UnusedKeys) > 0 {
		fmt.Println()
		fmt.Println("Unused translation keys:")
		for _, key := range result.UnusedKeys {
			fmt.Printf("- %s\n", key)
		}
	}

	if len(result.UnknownUsedKeys) > 0 {
		fmt.Println()
		fmt.Println("Used keys not found in translations:")
		for _, key := range result.UnknownUsedKeys {
			fmt.Printf("- %s\n", key)
		}
	}
}

func init() {
	rootCmd.AddCommand(unusedKeysCmd)

	unusedKeysCmd.Flags().StringVarP(&unusedTranslationsPath, "translations", "t", ".", "Path to translation JSON file or directory")
	unusedKeysCmd.Flags().StringVarP(&unusedSourcePath, "source", "s", ".", "Path to frontend source directory")
	unusedKeysCmd.Flags().StringVar(&unusedFormat, "format", "text", "Output format: text or json")
	unusedKeysCmd.Flags().StringVar(&unusedPrefix, "prefix", "_globalTranslations.", "Translation key prefix to analyze")
	unusedKeysCmd.Flags().StringSliceVar(&unusedExts, "ext", []string{".js", ".jsx", ".ts", ".tsx", ".vue", ".svelte", ".html"}, "Source file extensions to scan")
	unusedKeysCmd.Flags().StringSliceVar(&unusedExcludeDirs, "exclude-dir", []string{"node_modules", "dist", "build", "coverage", ".next"}, "Directory names to skip while scanning source files")

	_ = unusedKeysCmd.MarkFlagRequired("translations")
	_ = unusedKeysCmd.MarkFlagRequired("source")
}
