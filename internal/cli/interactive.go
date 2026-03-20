package cli

import (
	"errors"
	"fmt"
	"strings"

	"ghostbuster/internal/scanner"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

const (
	actionDetect  = "detect"
	actionHunt    = "hunt"
	actionRestore = "restore"
	actionHelp    = "help"
)

var (
	defaultPrefix      = "_globalTranslations."
	defaultExtensions  = []string{".js", ".jsx", ".ts", ".tsx", ".vue", ".svelte", ".html"}
	defaultExcludeDirs = []string{"node_modules", "dist", "build", "coverage", ".next"}
)

func runInteractiveRoot(cmd *cobra.Command) error {
	var action string

	if err := askAction(&action); err != nil {
		return err
	}

	switch action {
	case actionDetect:
		return runInteractiveDetect()
	case actionHunt:
		return runInteractiveHunt()
	case actionRestore:
		return runInteractiveRestore()
	case actionHelp:
		return runInteractiveHelp(cmd)
	default:
		return errors.New("unknown interactive action")
	}
}

func runInteractiveDetect() error {
	var translationsPath string
	var sourcePath string
	outputFormat := "text"
	prefix := defaultPrefix
	extCSV := strings.Join(defaultExtensions, ",")
	excludeCSV := strings.Join(defaultExcludeDirs, ",")

	if err := askRequiredPath("Translations path (--translations)", &translationsPath); err != nil {
		return err
	}
	if err := askRequiredPath("Source path (--source)", &sourcePath); err != nil {
		return err
	}
	if err := askFormat(&outputFormat); err != nil {
		return err
	}
	if err := askInput("Prefix (--prefix)", &prefix); err != nil {
		return err
	}
	if err := askInput("Extensions comma-separated (--ext)", &extCSV); err != nil {
		return err
	}
	if err := askInput("Excluded directories comma-separated (--exclude-dir)", &excludeCSV); err != nil {
		return err
	}

	extensions := parseCSVList(extCSV)
	excludeDirs := parseCSVList(excludeCSV)

	preview := fmt.Sprintf(
		"ghostbuster detect --translations %q --source %q --format %s --prefix %q --ext %q --exclude-dir %q",
		translationsPath,
		sourcePath,
		outputFormat,
		prefix,
		strings.Join(extensions, ","),
		strings.Join(excludeDirs, ","),
	)

	confirmed, err := confirmExecution(preview)
	if err != nil {
		return err
	}
	if !confirmed {
		return nil
	}

	return runDetect(detectRunConfig{
		Options: scanner.Options{
			TranslationsPath: translationsPath,
			SourcePath:       sourcePath,
			Prefix:           prefix,
			Extensions:       extensions,
			ExcludeDirs:      excludeDirs,
		},
		Format: outputFormat,
	})
}

func runInteractiveHunt() error {
	var translationsPath string
	var sourcePath string
	outputFormat := "text"
	prefix := defaultPrefix
	extCSV := strings.Join(defaultExtensions, ",")
	excludeCSV := strings.Join(defaultExcludeDirs, ",")
	createBackup := true

	if err := askRequiredPath("Translations path (--translations)", &translationsPath); err != nil {
		return err
	}
	if err := askRequiredPath("Source path (--source)", &sourcePath); err != nil {
		return err
	}
	if err := askFormat(&outputFormat); err != nil {
		return err
	}
	if err := askInput("Prefix (--prefix)", &prefix); err != nil {
		return err
	}
	if err := askInput("Extensions comma-separated (--ext)", &extCSV); err != nil {
		return err
	}
	if err := askInput("Excluded directories comma-separated (--exclude-dir)", &excludeCSV); err != nil {
		return err
	}
	if err := askConfirm("Create timestamped backups before modifying files?", &createBackup); err != nil {
		return err
	}

	extensions := parseCSVList(extCSV)
	excludeDirs := parseCSVList(excludeCSV)

	preview := fmt.Sprintf(
		"ghostbuster hunt --translations %q --source %q --format %s --prefix %q --ext %q --exclude-dir %q %s",
		translationsPath,
		sourcePath,
		outputFormat,
		prefix,
		strings.Join(extensions, ","),
		strings.Join(excludeDirs, ","),
		backupFlagPreview(createBackup),
	)

	confirmed, err := confirmExecution(preview)
	if err != nil {
		return err
	}
	if !confirmed {
		return nil
	}

	return runHunt(huntRunConfig{
		HuntOptions: scanner.HuntOptions{
			Options: scanner.Options{
				TranslationsPath: translationsPath,
				SourcePath:       sourcePath,
				Prefix:           prefix,
				Extensions:       extensions,
				ExcludeDirs:      excludeDirs,
			},
			CreateBackup: createBackup,
		},
		Format: outputFormat,
	})
}

func runInteractiveRestore() error {
	var translationsPath string
	outputFormat := "text"

	if err := askRequiredPath("Translations path (--translations)", &translationsPath); err != nil {
		return err
	}
	if err := askFormat(&outputFormat); err != nil {
		return err
	}

	preview := fmt.Sprintf(
		"ghostbuster restore --translations %q --format %s",
		translationsPath,
		outputFormat,
	)

	confirmed, err := confirmExecution(preview)
	if err != nil {
		return err
	}
	if !confirmed {
		return nil
	}

	return runRestore(restoreRunConfig{
		RestoreOptions: scanner.RestoreOptions{TranslationsPath: translationsPath},
		Format:         outputFormat,
	})
}

func runInteractiveHelp(root *cobra.Command) error {
	var topic string

	if err := askHelpTopic(&topic); err != nil {
		return err
	}

	preview := "ghostbuster --help"
	switch topic {
	case "detect", "hunt", "restore":
		preview = "ghostbuster " + topic + " --help"
	}

	confirmed, err := confirmExecution(preview)
	if err != nil {
		return err
	}
	if !confirmed {
		return nil
	}

	switch topic {
	case "ghostbuster":
		return root.Help()
	case "detect":
		return detectCmd.Help()
	case "hunt":
		return huntCmd.Help()
	case "restore":
		return restoreCmd.Help()
	default:
		return errors.New("unknown help topic")
	}
}

func askAction(value *string) error {
	return runSingleFieldForm(
		huh.NewSelect[string]().
			Title("Welcome to Ghostbuster").
			Description("Choose what you want to run").
			Options(
				huh.NewOption("Detect unused keys", actionDetect),
				huh.NewOption("Hunt and remove unused keys", actionHunt),
				huh.NewOption("Restore from backups", actionRestore),
				huh.NewOption("Show command help", actionHelp),
			).
			Value(value),
	)
}

func askHelpTopic(value *string) error {
	return runSingleFieldForm(
		huh.NewSelect[string]().
			Title("Help topic").
			Description("Choose which command help to display").
			Options(
				huh.NewOption("ghostbuster", "ghostbuster"),
				huh.NewOption("detect", "detect"),
				huh.NewOption("hunt", "hunt"),
				huh.NewOption("restore", "restore"),
			).
			Value(value),
	)
}

func askRequiredPath(title string, value *string) error {
	return runSingleFieldForm(
		huh.NewInput().
			Title(title).
			Value(value).
			Validate(func(v string) error {
				if strings.TrimSpace(v) == "" {
					return errors.New("this field is required")
				}
				return nil
			}),
	)
}

func askInput(title string, value *string) error {
	return runSingleFieldForm(
		huh.NewInput().
			Title(title).
			Value(value),
	)
}

func askFormat(value *string) error {
	return runSingleFieldForm(
		huh.NewSelect[string]().
			Title("Output format (--format)").
			Options(
				huh.NewOption("text", "text"),
				huh.NewOption("json", "json"),
			).
			Value(value),
	)
}

func askConfirm(title string, value *bool) error {
	return runSingleFieldForm(
		huh.NewConfirm().
			Title(title).
			Value(value),
	)
}

func runSingleFieldForm(field huh.Field) error {
	form := huh.NewForm(huh.NewGroup(field))
	return form.Run()
}

func confirmExecution(preview string) (bool, error) {
	confirmed := false

	form := huh.NewForm(huh.NewGroup(huh.NewNote().Title("Command preview").Description(preview)))
	if err := form.Run(); err != nil {
		return false, err
	}

	if err := runSingleFieldForm(
		huh.NewConfirm().
			Title("Execute this command now?").
			Affirmative("Run").
			Negative("Cancel").
			Value(&confirmed),
	); err != nil {
		return false, err
	}

	return confirmed, nil
}

func parseCSVList(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func backupFlagPreview(createBackup bool) string {
	if createBackup {
		return ""
	}
	return "--no-backup"
}
