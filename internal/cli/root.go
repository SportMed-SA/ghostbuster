package cli

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
)

type exitCoder interface {
	ExitCode() int
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ghostbuster",
	Short: "A simple and straightforward tool to find unused translation keys in your project.",
	Long:  `ghostbuster is a simple command-line tool designed to help developers finding unused translation keys in their projects. It scans nested JSON translation files and frontend source code to identify translation keys that are defined but not used anywhere in the codebase. This helps to keep your translation files clean and maintainable by removing obsolete keys.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(os.Args) == 1 {
			if !isInteractiveTerminal() {
				return errors.New("interactive mode requires a TTY; use a subcommand in non-interactive contexts")
			}
			return runInteractiveRoot(cmd)
		}

		return cmd.Help()
	},
}

func isInteractiveTerminal() bool {
	stdinInfo, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	stdoutInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return (stdinInfo.Mode()&os.ModeCharDevice) != 0 && (stdoutInfo.Mode()&os.ModeCharDevice) != 0
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		if ec, ok := err.(exitCoder); ok {
			os.Exit(ec.ExitCode())
		}
		os.Exit(1)
	}
}

func init() {
	configureHelpOutput(rootCmd)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ghostbuster.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
