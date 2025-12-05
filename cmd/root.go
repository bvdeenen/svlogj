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
		svlog.Svlog(types.ParseConfig{
			Facility: utils.GetString(flags, facility_flag),
			Level:    utils.GetString(flags, level_flag),
			Entity:   utils.GetString(flags, entity_flag),
			Service:  utils.GetString(flags, service_flag),
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
	rootCmd.Flags().BoolP(generate_config_flag, "c", false, "Generate Config File from the socklog configuration and current logs content")
	rootCmd.Flags().StringP(generate_completion_flag, "p", "", "Generate Completion for bash, zsh or fish")
	rootCmd.Flags().StringP(facility_flag, "f", "", "select facility")
	rootCmd.Flags().StringP(level_flag, "l", "", "select level")
	rootCmd.Flags().StringP(entity_flag, "e", "", "select entity")
	rootCmd.Flags().StringP(service_flag, "s", "", "select service")

	err := rootCmd.RegisterFlagCompletionFunc(facility_flag,
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
