package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

func getServiceBackend() string {
	switch runtime.GOOS {
	case "linux":
		if _, err := exec.LookPath("systemctl"); err == nil {
			return "systemd"
		}
		return "crontab"
	case "darwin":
		return "launchd"
	case "windows":
		return "schtasks"
	default:
		return "unknown"
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print tuckify version and system details",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("tuckify version %s (%s/%s)\n", rootCmd.Version, runtime.GOOS, getServiceBackend())
	},
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("tuckify version {{.Version}} (%s/%s)\n", runtime.GOOS, getServiceBackend()))
	rootCmd.AddCommand(versionCmd)
}
