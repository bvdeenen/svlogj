package utils

import (
	"slices"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func GetBool(flags *pflag.FlagSet, name string) bool {
	g, _ := flags.GetBool(name)
	return g
}
func GetString(flags *pflag.FlagSet, name string) string {
	g, _ := flags.GetString(name)
	return g
}

func RemoveEmptyStrings(l []string) []string {
	return slices.Collect(func(yield func(string) bool) {
		for _, v := range l {
			if len(v) != 0 {
				if !yield(v) {
					return
				}
			}
		}
	})
}

func NoFilesEmptyCompletion(cmd *cobra.Command, args []string, complete string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}
