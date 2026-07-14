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
	Short: "Find unused and missing translation keys in your project.",
	Long:  `ghostbuster scans nested JSON translation files and frontend source code to identify keys that are unused, referenced but undefined, or missing from individual translation files. This helps keep translations clean, complete, and maintainable.`,
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
