package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initForce bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize repository metadata (.mew/)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureDirs(); err != nil {
			return err
		}
		idxPath := indexPath()
		if _, err := os.Stat(idxPath); err == nil && !initForce {
			return fmt.Errorf("%s already exists (use --force to reinitialize)", idxPath)
		}
		// write empty index
		idx := Index{Snaps: []Snapshot{}}
		b, err := json.MarshalIndent(idx, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(idxPath, b, 0o644); err != nil {
			return err
		}
		fmt.Println("initialized .mew/")
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "overwrite existing metadata")
	rootCmd.AddCommand(initCmd)
}
