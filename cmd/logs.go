package cmd

import (
	"fmt"

	"github.com/ihsan-ramadhan/tuckify/internal/service"
	"github.com/spf13/cobra"
)

var logsFollow bool
var logsLines int

var logsCmd = &cobra.Command{
	Use:   "logs <name>",
	Short: "Show logs for a schedule's system service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		srv, err := service.NewService()
		if err != nil {
			return err
		}

		exists, err := srv.Exists(name)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("schedule %q is offline — run 'tuckify start %s' first", name, name)
		}

		return srv.Logs(name, logsFollow, logsLines)
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "follow log output")
	logsCmd.Flags().IntVarP(&logsLines, "lines", "n", 50, "number of lines to show")
	rootCmd.AddCommand(logsCmd)
}
