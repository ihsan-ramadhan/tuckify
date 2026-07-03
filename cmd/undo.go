package cmd

import (
	"fmt"
	"os"

	"github.com/ihsan-ramadhan/tuckify/internal/history"
	"github.com/spf13/cobra"
)

var undoCmd = &cobra.Command{
	Use:   "undo",
	Short: "Undo the last tuckify run",
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := history.Undo()
		if err != nil {
			return err
		}
		if n == 0 {
			fmt.Fprintln(os.Stderr, "nothing to undo")
			return nil
		}
		fmt.Printf("reverted %d file(s)\n", n)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(undoCmd)
}
