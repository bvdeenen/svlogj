package utils

import (
	"fmt"
	"iter"
	"os"
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
func GetInt(flags *pflag.FlagSet, name string, lower int, upper int) int {
	g, _ := flags.GetInt(name)
	if g < lower || g > upper {
		err := fmt.Errorf("Flag value of '%s' %d not within [%d, %d]\n", name, g, lower, upper)
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
	return g
}

func RemoveEmptyStrings(l iter.Seq[string]) []string {
	return slices.Collect(func(yield func(string) bool) {
		for v := range l {
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
