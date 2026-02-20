package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var snapCmd = &cobra.Command{
	Use:   "snap [title]",
	Short: "Create an immutable snapshot",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		id, err := doSnap(title)
		if err != nil {
			return err
		}
		fmt.Println(id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(snapCmd)
}
