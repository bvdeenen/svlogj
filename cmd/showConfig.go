// Package cmd
package cmd

import (
	"fmt"
	"os"
	"strings"
	"svlogj/pkg/config"
	"svlogj/pkg/types"
	"svlogj/pkg/utils"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)
import (
	"github.com/jedib0t/go-pretty/v6/table"
)

const tableFlag = "table"

// createConfigCmd represents the createConfig command
var showConfigCommand = &cobra.Command{
	Use:   "show-config",
	Short: "shows the config on stdout",
	Long:  ` `,
	Run: func(cmd *cobra.Command, args []string) {
		conf := config.LoadConfig()
		if utils.GetBool(cmd.Flags(), tableFlag) {
			showConfigAsTable(conf)

		} else {

			bytes, err := yaml.Marshal(conf)
			cobra.CheckErr(err)
			fmt.Println(string(bytes))
		}
	},
}

func showConfigAsTable(c types.Config) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Facilities", "Levels", "Entities", "Services"})
	for i := range 1000 {
		f := getOrEmpty(c.Facilities, i)
		l := getOrEmpty(c.Levels, i)
		e := getOrEmpty(c.Entities, i)
		s := getOrEmpty(c.Services, i)
		if len(f) == 0 && len(l) == 0 && len(e) == 0 && len(s) == 0 {
			break
		}
		t.AppendRow([]interface{}{f, l, e, s})
	}
	t.Render()

	fmt.Println("The services are the subdirectories in ", utils.SocklogDir())
	t = table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Service", "config"})

	for _, s := range c.ConfigFiles {
		lines := strings.Join(s.Lines, "\n")
		t.AppendRow([]interface{}{s.Service, lines})
		t.AppendSeparator()
	}
	t.Render()

}

func getOrEmpty(stringList []string, i int) string {
	if i >= len(stringList) {
		return ""
	} else {
		return stringList[i]
	}

}

func init() {
	rootCmd.AddCommand(showConfigCommand)
	showConfigCommand.Flags().BoolP(tableFlag, "t", true, "table format")
}
