package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mew",
	Short: "Mew: simple local-first snapshot tool",
	Long:  "Mew is a tiny local-first immutable snapshot tool (snap & wind).",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(os.Stderr, "use 'mew snap <title>' or 'mew wind <id|title>'")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
