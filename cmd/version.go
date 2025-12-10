// Package cmd 
package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// createConfigCmd represents the createConfig command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "returns the version",
	Long: ` `,
	Run: func(cmd *cobra.Command, args []string) {
		b, _ := debug.ReadBuildInfo()
		fmt.Println(b.Main.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
