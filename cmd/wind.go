package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var windCmd = &cobra.Command{
	Use:   "wind [id|title]",
	Short: "Restore files from a snapshot",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		if err := doWind(key); err != nil {
			return err
		}
		fmt.Println("done")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(windCmd)
}
