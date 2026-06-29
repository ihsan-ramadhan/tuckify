package cmd

import (
	"fmt"

	"github.com/ihsan-ramadhan/tuckify/internal/service"
	"github.com/ihsan-ramadhan/tuckify/internal/store"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Remove a saved schedule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		found, err := store.Delete(name)
		if err != nil {
			return fmt.Errorf("delete from store: %w", err)
		}
		if !found {
			return fmt.Errorf("schedule %q not found", name)
		}

		srv, err := service.NewService()
		if err != nil {
			return err
		}
		exists, _ := srv.Exists(name)
		if exists {
			if err := srv.Uninstall(name); err != nil {
				fmt.Printf("warning: could not remove system service for %q: %v\n", name, err)
			}
		}

		fmt.Printf("deleted schedule %q\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
