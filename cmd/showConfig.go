// Package cmd
package cmd

import (
	"fmt"
	"svlogj/pkg/config"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// createConfigCmd represents the createConfig command
var showConfigCommand = &cobra.Command{
	Use:   "show-config",
	Short: "shows the config on stdout",
	Long:  ` `,
	Run: func(cmd *cobra.Command, args []string) {
		conf := config.LoadConfig()
		bytes, err := yaml.Marshal(conf)
		cobra.CheckErr(err)
		fmt.Println(string(bytes))
	},
}

func init() {
	rootCmd.AddCommand(showConfigCommand)
}
