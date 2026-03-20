package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func configureHelpOutput(root *cobra.Command) {
	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		printCommandHelp(cmd)
	})

	root.SetUsageFunc(func(cmd *cobra.Command) error {
		printCommandHelp(cmd)
		return nil
	})
}

func printCommandHelp(cmd *cobra.Command) {
	out := cmd.OutOrStdout()

	if cmd.Long != "" {
		fmt.Fprintln(out, cmd.Long)
	} else if cmd.Short != "" {
		fmt.Fprintln(out, cmd.Short)
	}

	fmt.Fprintln(out)
	fmt.Fprintf(out, "Usage:\n  %s\n", cmd.UseLine())

	if cmd.HasAvailableSubCommands() {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Available Commands:")
		for _, child := range cmd.Commands() {
			if child.IsAvailableCommand() && !child.IsAdditionalHelpTopicCommand() {
				fmt.Fprintf(out, "  %-18s %s\n", child.Name(), child.Short)
			}
		}
	}

	requiredLocal, optionalLocal := splitFlags(cmd.LocalFlags())
	requiredInherited, optionalInherited := splitFlags(cmd.InheritedFlags())

	requiredFlags := append(requiredLocal, requiredInherited...)
	optionalFlags := append(optionalLocal, optionalInherited...)

	sort.Strings(requiredFlags)
	sort.Strings(optionalFlags)

	if len(requiredFlags) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Required Flags:")
		for _, line := range requiredFlags {
			fmt.Fprintln(out, line)
		}
	}

	if len(optionalFlags) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Optional Flags:")
		for _, line := range optionalFlags {
			fmt.Fprintln(out, line)
		}
	}

	fmt.Fprintln(out)
	fmt.Fprintf(out, "Run \"%s [command] --help\" for more information about a command.\n", cmd.CommandPath())
}

func splitFlags(flags *pflag.FlagSet) (required []string, optional []string) {
	if flags == nil {
		return nil, nil
	}

	flags.VisitAll(func(flag *pflag.Flag) {
		line := formatFlagLine(flag)
		if isRequiredFlag(flag) {
			required = append(required, line)
			return
		}
		optional = append(optional, line)
	})

	return required, optional
}

func formatFlagLine(flag *pflag.Flag) string {
	parts := make([]string, 0, 2)

	if flag.Shorthand != "" {
		parts = append(parts, fmt.Sprintf("-%s", flag.Shorthand))
	}

	parts = append(parts, fmt.Sprintf("--%s", flag.Name))

	typeName := flag.Value.Type()
	if typeName != "bool" {
		parts[len(parts)-1] = parts[len(parts)-1] + " " + typeName
	}

	line := fmt.Sprintf("  %-26s %s", strings.Join(parts, ", "), flag.Usage)
	if flag.DefValue != "" {
		line += fmt.Sprintf(" (default %q)", flag.DefValue)
	}

	return line
}

func isRequiredFlag(flag *pflag.Flag) bool {
	if flag.Annotations == nil {
		return false
	}

	values, ok := flag.Annotations[cobra.BashCompOneRequiredFlag]
	if !ok {
		return false
	}

	for _, value := range values {
		if strings.EqualFold(value, "true") {
			return true
		}
	}

	return false
}
