// Package cmd 
package cmd

import (
	"os"
	"svlogj/pkg/config"
	"svlogj/pkg/svlog"
	"svlogj/pkg/types"
	"svlogj/pkg/utils"

	"github.com/spf13/cobra"
)


const facilityFlag = "facility"
const levelFlag = "level"
const entityFlag = "entity"
const serviceFlag = "service"
const timeConfigFlag = "time-config"
const followFlag = "follow"
const monochromeFlag = "monochrome"
const ansiColorFlag = "ansi-color"

// GREP style context
const afterFlag = "after"
const beforeFlag = "before"
const contextFlag = "context"

var conf types.Config

var rootCmd = &cobra.Command{
	Use:               "svlogj",
	Short:             "Frontend for svlogtail in Void Linux",
	ValidArgsFunction: utils.NoFilesEmptyCompletion,
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		context := utils.GetInt(flags, contextFlag, 0, 20)
		svlog.Svlog(types.ParseConfig{
			Facility:  utils.GetString(flags, facilityFlag),
			Level:     utils.GetString(flags, levelFlag),
			Entity:    utils.GetString(flags, entityFlag),
			Service:   utils.GetString(flags, serviceFlag),
			AnsiColor: utils.GetString(flags, ansiColorFlag),
			Grep: types.Grep{
				After:   max(context, utils.GetInt(flags, afterFlag, 0, 20)),
				Before:  max(context, utils.GetInt(flags, beforeFlag, 0, 20)),
				Context: context,
			},
			TimeConfig: utils.GetString(flags, timeConfigFlag),
			Follow:     utils.GetBool(flags, followFlag),
			Monochrome: utils.GetBool(flags, monochromeFlag),
		})
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}


func init() {
	rootCmd.Flags().StringP(facilityFlag, "f", "", "select facility")
	rootCmd.Flags().StringP(levelFlag, "l", "", "select level")
	rootCmd.Flags().StringP(entityFlag, "e", "", "select entity")
	rootCmd.Flags().StringP(serviceFlag, "s", "", "select service")
	// GREP flags
	rootCmd.Flags().IntP(afterFlag, "A", 0, "grep after")
	rootCmd.Flags().IntP(beforeFlag, "B", 0, "grep before")
	rootCmd.Flags().IntP(contextFlag, "C", 0, "grep context")
	rootCmd.Flags().String(timeConfigFlag, envTimeConfig(), "timeconfig")
	rootCmd.Flags().Bool(followFlag, envFollow(), "follow")
	rootCmd.Flags().Bool(monochromeFlag, envMonochrome(), "monochrome output")
	rootCmd.Flags().String(ansiColorFlag, envAnsiColor(), "ansi color for match")

	err := rootCmd.RegisterFlagCompletionFunc(timeConfigFlag,
		func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return []string{"uptime_s", "local"}, cobra.ShellCompDirectiveNoFileComp
		})

	err = rootCmd.RegisterFlagCompletionFunc(facilityFlag,
		func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			if len(conf.Facilities) == 0 {
				conf = config.LoadConfig()
			}
			return conf.Facilities, cobra.ShellCompDirectiveNoFileComp
		})
	err = rootCmd.RegisterFlagCompletionFunc(levelFlag,
		func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			if len(conf.Levels) == 0 {
				conf = config.LoadConfig()
			}
			return conf.Levels, cobra.ShellCompDirectiveNoFileComp
		})
	err = rootCmd.RegisterFlagCompletionFunc(entityFlag,
		func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			if len(conf.Entities) == 0 {
				conf = config.LoadConfig()
			}
			return conf.Entities, cobra.ShellCompDirectiveNoFileComp
		})
	err = rootCmd.RegisterFlagCompletionFunc(serviceFlag,
		func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			if len(conf.Services) == 0 {
				conf = config.LoadConfig()
			}
			return conf.Services, cobra.ShellCompDirectiveNoFileComp
		})
	cobra.CheckErr(err)
}

func envFollow() bool {
	return len(os.Getenv("SVLOGJ_FOLLOW"))!=0
}

func envMonochrome() bool {
	return len(os.Getenv("SVLOGJ_MONOCHROME"))!=0
}

func envAnsiColor() string {
	v := os.Getenv("SVLOGJ_ANSICOLOR")
	if len(v)!=0 {
		return v
	}
	return "1;33"
}

func envTimeConfig() string {
	return os.Getenv("SVLOGJ_TIMECONFIG")
}
