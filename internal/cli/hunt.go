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

var (
	huntTranslationsPath string
	huntSourcePath       string
	huntFormat           string
	huntPrefix           string
	huntExtensions       []string
	huntExcludeDirs      []string
	huntNoBackup         bool
)

var huntCmd = &cobra.Command{
	Use:          "hunt",
	Short:        "Find and remove unused translation keys",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHunt(huntRunConfig{
			HuntOptions: scanner.HuntOptions{
				Options: scanner.Options{
					TranslationsPath: huntTranslationsPath,
					SourcePath:       huntSourcePath,
					Prefix:           huntPrefix,
					Extensions:       huntExtensions,
					ExcludeDirs:      huntExcludeDirs,
				},
				CreateBackup: !huntNoBackup,
			},
			Format: huntFormat,
		})
	},
}

type huntRunConfig struct {
	HuntOptions scanner.HuntOptions
	Format      string
}

func runHunt(cfg huntRunConfig) error {
	result, err := scanner.HuntUnusedKeys(cfg.HuntOptions)
	if err != nil {
		return err
	}

	switch strings.ToLower(cfg.Format) {
	case "text":
		printHuntTextResult(result, cfg.HuntOptions.CreateBackup)
	case "json":
		payload, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal JSON output: %w", err)
		}
		fmt.Println(string(payload))
	default:
		return errors.New("unsupported format: use text or json")
	}

	if result.RemovedCount > 0 || hasMissingTranslations(result.MissingKeys, result.MissingByFile) {
		return cliExitError{msg: "translation key issues found", code: issuesFoundExitCode}
	}

	return nil
}

func printHuntTextResult(result scanner.HuntResult, createBackup bool) {
	sort.Strings(result.RemovedKeys)
	sort.Strings(result.FilesModified)
	sort.Strings(result.BackupsCreated)

	fmt.Printf("Detected unused keys: %d\n", result.DetectedUnusedCount)
	fmt.Printf("Removed keys: %d\n", result.RemovedCount)
	fmt.Printf("Files modified: %d\n", len(result.FilesModified))
	if createBackup {
		fmt.Printf("Backups created: %d\n", len(result.BackupsCreated))
	}
	printMissingTranslationCounts(result.MissingKeys, result.MissingByFile)

	if len(result.FilesModified) > 0 {
		fmt.Println()
		fmt.Println("Updated translation files:")
		for _, file := range result.FilesModified {
			fmt.Printf("- %s\n", file)
		}
	}

	if len(result.BackupsCreated) > 0 {
		fmt.Println()
		fmt.Println("Backup files:")
		for _, file := range result.BackupsCreated {
			fmt.Printf("- %s\n", file)
		}
	}

	if len(result.RemovedKeys) > 0 {
		fmt.Println()
		fmt.Println("Removed translation keys:")
		for _, key := range result.RemovedKeys {
			fmt.Printf("- %s\n", key)
		}
	}

	printMissingTranslationDetails(result.MissingKeys, result.MissingByFile)
}

func init() {
	rootCmd.AddCommand(huntCmd)

	huntCmd.Flags().StringVarP(&huntTranslationsPath, "translations", "t", "", "Path to translation JSON file or directory")
	huntCmd.Flags().StringVarP(&huntSourcePath, "source", "s", "", "Path to frontend source directory")
	huntCmd.Flags().StringVar(&huntFormat, "format", "text", "Output format: text or json")
	huntCmd.Flags().StringVar(&huntPrefix, "prefix", "_globalTranslations.", "Translation key prefix to analyze")
	huntCmd.Flags().StringSliceVar(&huntExtensions, "ext", []string{".js", ".jsx", ".ts", ".tsx", ".vue", ".svelte", ".html"}, "Source file extensions to scan")
	huntCmd.Flags().StringSliceVar(&huntExcludeDirs, "exclude-dir", []string{"node_modules", "dist", "build", "coverage", ".next"}, "Directory names to skip while scanning source files")
	huntCmd.Flags().BoolVar(&huntNoBackup, "no-backup", false, "Do not create .bak backup copies before deleting keys")

	_ = huntCmd.MarkFlagRequired("translations")
	_ = huntCmd.MarkFlagRequired("source")
}
