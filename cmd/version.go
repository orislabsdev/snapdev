package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// versionCmd prints version and build metadata.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print snapdev version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("snapdev %s (%s/%s)\n", Version, runtime.GOOS, runtime.GOARCH)
	},
}