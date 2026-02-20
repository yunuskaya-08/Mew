package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List snapshots",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		idx, err := loadIndex()
		if err != nil {
			return err
		}
		if len(idx.Snaps) == 0 {
			fmt.Println("no snapshots")
			return nil
		}
		for _, s := range idx.Snaps {
			fmt.Printf("%s\t%s\t%s\n", s.ID, s.Time.Format("2006-01-02T15:04:05"), s.Title)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
