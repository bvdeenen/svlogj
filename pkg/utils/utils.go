package utils

import (
	"fmt"
	"iter"
	"os"
	"slices"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

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

func NoFilesEmptyCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func SocklogDir() string {
	v := os.Getenv("SVLOGJ_SOCKLOGDIR")
	if len(v) != 0 {
		return v
	}
	return "/var/log/socklog"
}
