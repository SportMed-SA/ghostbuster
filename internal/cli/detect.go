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
		opts := scanner.Options{
			TranslationsPath: translationsPath,
			SourcePath:       sourcePath,
			Prefix:           prefix,
			Extensions:       extensions,
			ExcludeDirs:      excludeDirs,
		}

		result, err := scanner.FindUnusedKeys(opts)
		if err != nil {
			return err
		}

		switch strings.ToLower(format) {
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
