/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"svlogj/pkg/config"
	"svlogj/pkg/svlog"
	"svlogj/pkg/types"
	"svlogj/pkg/utils"

	"github.com/spf13/cobra"
)

const generate_config_flag = "generate-config"
const generate_completion_flag = "generate-completion"
const facility_flag = "facility"
const level_flag = "level"
const entity_flag = "entity"
const service_flag = "service"
const time_config_flag = "time-config"
const follow_flag = "follow"
const monochrome_flag = "monochrome"
const ansi_color_flag = "ansi-color"

// GREP style context
const after_flag = "after"
const before_flag = "before"
const context_flag = "context"

var conf types.Config

var rootCmd = &cobra.Command{
	Use:               "svlogj",
	Short:             "Frontend for svlogtail",
	ValidArgsFunction: utils.NoFilesEmptyCompletion,
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		if utils.GetBool(flags, generate_config_flag) {
			config.ParseAndStoreConfig()
			return
		}
		if shell := utils.GetString(flags, generate_completion_flag); len(shell) != 0 {
			switch shell {
			case "bash":
				_ = cmd.Root().GenBashCompletion(cmd.OutOrStdout())
				break
			case "zsh":
				_ = cmd.Root().GenZshCompletion(cmd.OutOrStdout())
				break
			case "fish":
				_ = cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
				break
			default:
				break
			}
			return
		}
		context := utils.GetInt(flags, context_flag, 0, 20)
		svlog.Svlog(types.ParseConfig{
			Facility:  utils.GetString(flags, facility_flag),
			Level:     utils.GetString(flags, level_flag),
			Entity:    utils.GetString(flags, entity_flag),
			Service:   utils.GetString(flags, service_flag),
			AnsiColor: utils.GetString(flags, ansi_color_flag),
			Grep: types.Grep{
				max(context, utils.GetInt(flags, after_flag, 0, 20)),
				max(context, utils.GetInt(flags, before_flag, 0, 20)),
				context,
			},
			TimeConfig: utils.GetString(flags, time_config_flag),
			Follow:     utils.GetBool(flags, follow_flag),
			Monochrome: utils.GetBool(flags, monochrome_flag),
		})
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	utils.Check(rootCmd.Execute())
}

func init() {
	// Here you will define your flags and configuration settings.
	rootCmd.Flags().Bool(generate_config_flag, false, "Generate Config File from the socklog configuration and current logs content")
	rootCmd.Flags().String(generate_completion_flag, "", "Generate Completion for bash, zsh or fish")
	rootCmd.Flags().StringP(facility_flag, "f", "", "select facility")
	rootCmd.Flags().StringP(level_flag, "l", "", "select level")
	rootCmd.Flags().StringP(entity_flag, "e", "", "select entity")
	rootCmd.Flags().StringP(service_flag, "s", "", "select service")
	// GREP flags
	rootCmd.Flags().IntP(after_flag, "A", 0, "grep after")
	rootCmd.Flags().IntP(before_flag, "B", 0, "grep before")
	rootCmd.Flags().IntP(context_flag, "C", 0, "grep context")
	rootCmd.Flags().String(time_config_flag, "", "timeconfig")
	rootCmd.Flags().Bool(follow_flag, false, "follow")
	rootCmd.Flags().Bool(monochrome_flag, false, "monochrome output")
	rootCmd.Flags().String(ansi_color_flag, "1;33", "ansi color for match")

	err := rootCmd.RegisterFlagCompletionFunc(time_config_flag,
		func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return []string{"uptime_s", "local"}, cobra.ShellCompDirectiveNoFileComp
		})

	err = rootCmd.RegisterFlagCompletionFunc(facility_flag,
		func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			if len(conf.Facilities) == 0 {
				conf = config.LoadConfig()
			}
			return conf.Facilities, cobra.ShellCompDirectiveNoFileComp
		})
	err = rootCmd.RegisterFlagCompletionFunc(level_flag,
		func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			if len(conf.Levels) == 0 {
				conf = config.LoadConfig()
			}
			return conf.Levels, cobra.ShellCompDirectiveNoFileComp
		})
	err = rootCmd.RegisterFlagCompletionFunc(entity_flag,
		func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			if len(conf.Entities) == 0 {
				conf = config.LoadConfig()
			}
			return conf.Entities, cobra.ShellCompDirectiveNoFileComp
		})
	err = rootCmd.RegisterFlagCompletionFunc(service_flag,
		func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			if len(conf.Services) == 0 {
				conf = config.LoadConfig()
			}
			return conf.Services, cobra.ShellCompDirectiveNoFileComp
		})
	utils.Check(err)
}
