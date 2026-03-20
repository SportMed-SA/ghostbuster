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
	restoreTranslationsPath string
	restoreFormat           string
)

var restoreCmd = &cobra.Command{
	Use:          "restore",
	Short:        "Restore translation files from backup files created by hunt",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRestore(restoreRunConfig{
			RestoreOptions: scanner.RestoreOptions{TranslationsPath: restoreTranslationsPath},
			Format:         restoreFormat,
		})
	},
}

type restoreRunConfig struct {
	RestoreOptions scanner.RestoreOptions
	Format         string
}

func runRestore(cfg restoreRunConfig) error {
	result, err := scanner.RestoreFromBackup(cfg.RestoreOptions)
	if err != nil {
		return err
	}

	switch strings.ToLower(cfg.Format) {
	case "text":
		printRestoreTextResult(result)
	case "json":
		payload, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal JSON output: %w", err)
		}
		fmt.Println(string(payload))
	default:
		return errors.New("unsupported format: use text or json")
	}

	return nil
}

func printRestoreTextResult(result scanner.RestoreResult) {
	sort.Strings(result.FilesRestored)
	sort.Strings(result.BackupsUsed)
	sort.Strings(result.MissingBackups)

	fmt.Printf("Translation files: %d\n", result.TranslationFileCount)
	fmt.Printf("Files restored: %d\n", result.RestoredCount)
	fmt.Printf("Missing backups: %d\n", len(result.MissingBackups))

	if len(result.FilesRestored) > 0 {
		fmt.Println()
		fmt.Println("Restored translation files:")
		for _, file := range result.FilesRestored {
			fmt.Printf("- %s\n", file)
		}
	}

	if len(result.BackupsUsed) > 0 {
		fmt.Println()
		fmt.Println("Backups used:")
		for _, file := range result.BackupsUsed {
			fmt.Printf("- %s\n", file)
		}
	}

	if len(result.MissingBackups) > 0 {
		fmt.Println()
		fmt.Println("Translation files without backup:")
		for _, file := range result.MissingBackups {
			fmt.Printf("- %s\n", file)
		}
	}
}

func init() {
	rootCmd.AddCommand(restoreCmd)

	restoreCmd.Flags().StringVarP(&restoreTranslationsPath, "translations", "t", "", "Path to translation JSON file or directory")
	restoreCmd.Flags().StringVar(&restoreFormat, "format", "text", "Output format: text or json")

	_ = restoreCmd.MarkFlagRequired("translations")
}
